import { COMMENT_ELEMENT_HTML_LIMIT, COMMENT_ELEMENT_HTML_TRUNCATION, CommentAnchorBridge, CommentProvider, captureLocators, captureVocabularyPreviews, commentAnchor, describeElement, resolveComment, strictAnchorResolver } from "@flanksource/clicky-ui/comments";
import { createContext } from "react";
import { jsxDEV } from "react/jsx-dev-runtime";
// \0rolldown/runtime.js
var __defProp = Object.defineProperty;
var __esmMin = (fn, res, err) => () => {
	if (err) throw err[0];
	try {
		return fn && (res = fn(fn = 0)), res;
	} catch (e) {
		throw err = [e], e;
	}
};
var __exportAll = (all, no_symbols) => {
	let target = {};
	for (var name in all) __defProp(target, name, {
		get: all[name],
		enumerable: true
	});
	if (!no_symbols) __defProp(target, Symbol.toStringTag, { value: "Module" });
	return target;
};
// ../ui/dist/lib/random-uuid.js
var HEX = Array.from({ length: 256 }, (_, byte) => byte.toString(16).padStart(2, "0"));
function randomUUID() {
	const bytes = crypto.getRandomValues(/* @__PURE__ */ new Uint8Array(16));
	bytes[6] = bytes[6] & 15 | 64;
	bytes[8] = bytes[8] & 63 | 128;
	const hex = Array.from(bytes, (byte) => HEX[byte]);
	return `${hex.slice(0, 4).join("")}-${hex.slice(4, 6).join("")}-${hex.slice(6, 8).join("")}-${hex.slice(8, 10).join("")}-${hex.slice(10).join("")}`;
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/bippy/0.6.1/1130005d9ad5ea110510cf11f128db1358e7b7bbb7b2d18e02d0aee399b85c38/node_modules/bippy/dist/rdt-hook.js
var e, t, n, r, i$1, a$1, o$1, s$1, c$1, l$1, u$1, d$1, f$1, v$3, y$3, b$3, x$3, S$3, C$3;
var init_rdt_hook = __esmMin((() => {
	e = (e) => Object.assign(e, { [Symbol.dispose]: e });
	t = `0.6.1`;
	n = `bippy-${t}`;
	r = Object.defineProperty;
	i$1 = Object.prototype.hasOwnProperty;
	a$1 = () => {
		/**
		* @license bippy
		*
		* Copyright (c) Aiden Bai
		*
		* This source code is licensed under the MIT license found in the
		* LICENSE file in the root directory of this source tree.
		*/
	};
	o$1 = (e) => {
		try {
			Function.prototype.toString.call(e).indexOf(`^_^`) > -1 && setTimeout(() => {
				throw Error(`React is running in production mode, but dead code elimination has not been applied. Read how to correctly configure React for production: https://reactjs.org/link/perf-use-production-build`);
			});
		} catch {}
	};
	s$1 = (e = globalThis.__REACT_DEVTOOLS_GLOBAL_HOOK__) => !!(e && `getFiberRoots` in e);
	c$1 = !1;
	u$1 = (e = globalThis.__REACT_DEVTOOLS_GLOBAL_HOOK__) => c$1 ? !0 : (e && typeof e.inject == `function` && (l$1 = e.inject.toString()), !!l$1?.includes(`(injected)`));
	d$1 = /* @__PURE__ */ new Set();
	f$1 = /* @__PURE__ */ new Set();
	v$3 = (e) => {
		e && d$1.add(e);
		let t = /* @__PURE__ */ new Map(), i = 0, s = {
			_instrumentationIsActive: !1,
			_instrumentationSource: n,
			checkDCE: o$1,
			hasUnsupportedRendererAttached: !1,
			inject(e) {
				let n = ++i;
				return t.set(n, e), f$1.add(e), s._instrumentationIsActive || (s._instrumentationIsActive = !0, d$1.forEach((e) => e())), n;
			},
			on: a$1,
			onCommitFiberRoot: a$1,
			onCommitFiberUnmount: a$1,
			onPostCommitFiberRoot: a$1,
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
							f$1.add(e), t.renderers.set(n, e);
						}), y$3(e));
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
			y$3(e);
		}
		return s;
	};
	y$3 = (e) => {
		e && d$1.add(e);
		try {
			let t = globalThis.__REACT_DEVTOOLS_GLOBAL_HOOK__;
			if (!t) return;
			if (!t._instrumentationSource) {
				t.checkDCE = o$1, t.supportsFiber = !0, t.supportsFlight = !0, t.hasUnsupportedRendererAttached = !1, t._instrumentationSource = n, t._instrumentationIsActive = !1;
				let e = s$1(t);
				if (e || (t.on = a$1), t.renderers.size) {
					t._instrumentationIsActive = !0, d$1.forEach((e) => e());
					return;
				}
				let r = t.inject, i = u$1(t);
				i && !e && (c$1 = !0, t.inject({ scheduleRefresh() {} }) && (t._instrumentationIsActive = !0)), t.inject = (e) => {
					let n = r(e);
					return f$1.add(e), i && t.renderers.set(n, e), t._instrumentationIsActive = !0, d$1.forEach((e) => e()), n;
				};
			}
			(t.renderers.size || t._instrumentationIsActive || u$1()) && e?.();
		} catch {}
	};
	b$3 = () => i$1.call(globalThis, `__REACT_DEVTOOLS_GLOBAL_HOOK__`);
	x$3 = (e) => b$3() ? (y$3(e), globalThis.__REACT_DEVTOOLS_GLOBAL_HOOK__) : v$3(e);
	S$3 = () => !!(typeof window < `u` && (window.document?.createElement || window.navigator?.product === `ReactNative`));
	C$3 = () => {
		try {
			S$3() && x$3();
		} catch {}
	};
}));
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/bippy/0.6.1/1130005d9ad5ea110510cf11f128db1358e7b7bbb7b2d18e02d0aee399b85c38/node_modules/bippy/dist/install-hook-only.js
var init_install_hook_only = __esmMin((() => {
	init_rdt_hook();
	/**
	* @license bippy
	*
	* Copyright (c) Aiden Bai
	*
	* This source code is licensed under the MIT license found in the
	* LICENSE file in the root directory of this source tree.
	*/
	C$3();
}));
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/bippy/0.6.1/1130005d9ad5ea110510cf11f128db1358e7b7bbb7b2d18e02d0aee399b85c38/node_modules/bippy/dist/core.js
/**
* @license bippy
*
* Copyright (c) Aiden Bai
*
* This source code is licensed under the MIT license found in the
* LICENSE file in the root directory of this source tree.
*/
function w$3(e, t, n = !1) {
	if (!e) return null;
	let r = t(e);
	if (r instanceof Promise) return (async () => {
		if (await r === !0) return e;
		let i = n ? e.return : e.child;
		for (; i;) {
			let e = await E$3(i, t, n);
			if (e) return e;
			i = n ? null : i.sibling;
		}
		return null;
	})();
	if (r === !0) return e;
	let i = n ? e.return : e.child;
	for (; i;) {
		let e = T$2(i, t, n);
		if (e) return e;
		i = n ? null : i.sibling;
	}
	return null;
}
var _$3, Oe$1, v$2, Ne$1, T$2, E$3, D$3, Ie$1, Re$1, O$2, ze$1, B$4, K$3, q$3, J$3, Y$3, X$3, Z$4, Q$4, Je$2, $$4, Ye$2, Xe$3;
var init_core = __esmMin((() => {
	init_rdt_hook();
	_$3 = (e) => {
		switch (e.tag) {
			case 5:
			case 26:
			case 27: return !0;
			default: return typeof e.type == `string`;
		}
	};
	Oe$1 = (e) => {
		switch (e.tag) {
			case 1:
			case 11:
			case 0:
			case 14:
			case 15: return !0;
			default: return !1;
		}
	};
	v$2 = (e) => !e || typeof e != `object` ? !1 : `pendingProps` in e && !(`containerInfo` in e);
	Ne$1 = (e) => {
		let t = [], n = [];
		for (_$3(e) ? t.push(e) : e.child && n.push(e.child); n.length;) {
			let e = n.pop();
			if (!e) break;
			_$3(e) ? t.push(e) : e.child && n.push(e.child), e.sibling && n.push(e.sibling);
		}
		return t;
	};
	T$2 = (e, t, n = !1) => {
		if (!e) return null;
		if (t(e) === !0) return e;
		let r = n ? e.return : e.child;
		for (; r;) {
			let e = T$2(r, t, n);
			if (e) return e;
			r = n ? null : r.sibling;
		}
		return null;
	};
	E$3 = async (e, t, n = !1) => {
		if (!e) return null;
		if (await t(e) === !0) return e;
		let r = n ? e.return : e.child;
		for (; r;) {
			let e = await E$3(r, t, n);
			if (e) return e;
			r = n ? null : r.sibling;
		}
		return null;
	};
	D$3 = (e) => {
		let t = e;
		return typeof t == `function` ? t : typeof t == `object` && t ? D$3(t.type || t.render) : null;
	};
	Ie$1 = (e) => {
		let t = e;
		if (typeof t == `string`) return t;
		if (typeof t != `function` && !(typeof t == `object` && t)) return null;
		let n = t.displayName || t.name || null;
		if (n) return n;
		let r = D$3(t);
		return r && (r.displayName || r.name) || null;
	};
	Re$1 = () => {
		let e = globalThis.__REACT_DEVTOOLS_GLOBAL_HOOK__;
		return !!e?._instrumentationIsActive || s$1(e) || u$1(e);
	};
	O$2 = /* @__PURE__ */ new Set();
	ze$1 = (e) => {
		let t = e.alternate;
		if (!t) return e;
		if (t.actualStartTime && e.actualStartTime) return t.actualStartTime > e.actualStartTime ? t : e;
		for (let t of O$2) {
			let n = w$3(t.current, (t) => {
				if (t === e) return !0;
			});
			if (n) return n;
		}
		return e;
	};
	B$4 = (t) => {
		if (!b$3()) return null;
		let n = t;
		for (; n.return;) n = n.return;
		let r = Z$4.get(n.stateNode);
		return r === void 0 ? null : x$3().renderers?.get(r) ?? null;
	};
	K$3 = /* @__PURE__ */ new Set();
	q$3 = /* @__PURE__ */ new Set();
	J$3 = /* @__PURE__ */ new Set();
	Y$3 = /* @__PURE__ */ new Set();
	X$3 = /* @__PURE__ */ new WeakMap();
	Z$4 = /* @__PURE__ */ new WeakMap();
	Q$4 = (e) => {
		let t = X$3.get(e) ?? {};
		if (X$3.set(e, t), !t.onCommitFiberRoot || e.onCommitFiberRoot !== t.onCommitFiberRoot) {
			let n = e.onCommitFiberRoot, r = (t, i, a) => {
				if (n?.(t, i, a), X$3.get(e)?.onCommitFiberRoot === r) {
					O$2.add(i), Z$4.set(i, t);
					for (let e of K$3) e(t, i, a);
				}
			};
			t.onCommitFiberRoot = r, e.onCommitFiberRoot = r;
		}
		if (!t.onCommitFiberUnmount || e.onCommitFiberUnmount !== t.onCommitFiberUnmount) {
			let n = e.onCommitFiberUnmount, r = (t, i) => {
				if (n?.(t, i), X$3.get(e)?.onCommitFiberUnmount === r) for (let e of q$3) e(t, i);
			};
			t.onCommitFiberUnmount = r, e.onCommitFiberUnmount = r;
		}
		if (!t.onPostCommitFiberRoot || e.onPostCommitFiberRoot !== t.onPostCommitFiberRoot) {
			let n = e.onPostCommitFiberRoot, r = (t, i) => {
				if (n?.(t, i), X$3.get(e)?.onPostCommitFiberRoot === r) for (let e of J$3) e(t, i);
			};
			t.onPostCommitFiberRoot = r, e.onPostCommitFiberRoot = r;
		}
		if (!t.onScheduleFiberRoot || e.onScheduleFiberRoot !== t.onScheduleFiberRoot) {
			let n = e.onScheduleFiberRoot, r = (t, i, a) => {
				if (n?.(t, i, a), X$3.get(e)?.onScheduleFiberRoot === r) for (let e of Y$3) e(t, i, a);
			};
			t.onScheduleFiberRoot = r, e.onScheduleFiberRoot = r;
		}
	};
	Je$2 = (e$1) => {
		let t = x$3(e$1.onActive);
		t._instrumentationSource = e$1.name ?? "bippy-0.6.1", Q$4(t);
		let { onActive: n$1, onCommitFiberRoot: r, onCommitFiberUnmount: a, onPostCommitFiberRoot: c, onScheduleFiberRoot: l } = e$1;
		return r && K$3.add(r), a && q$3.add(a), c && J$3.add(c), l && Y$3.add(l), e(() => {
			n$1 && d$1.delete(n$1), r && K$3.delete(r), a && q$3.delete(a), c && J$3.delete(c), l && Y$3.delete(l);
		});
	};
	$$4 = /* @__PURE__ */ new Set();
	Ye$2 = (e) => e.startsWith(`__reactContainer$`) || e.startsWith(`__reactInternalInstance$`) || e.startsWith(`__reactFiber`);
	Xe$3 = (e) => {
		let t = globalThis.__REACT_DEVTOOLS_GLOBAL_HOOK__;
		if (t?.renderers) for (let n of t.renderers.values()) try {
			let t = n.findFiberByHostInstance?.(e);
			if (t) return t;
		} catch {}
		if (typeof e == `object` && e) {
			if (`_reactRootContainer` in e) return e._reactRootContainer?._internalRoot?.current?.child;
			let t = e.__internalInstanceHandle ?? e._internalInstanceHandle;
			if (v$2(t)) return t;
			let n = e;
			for (let e of $$4) {
				let t = n[e];
				if (t) return t;
			}
			for (let e of Object.keys(n)) if (Ye$2(e)) return $$4.add(e), n[e] || null;
			for (let t of O$2) {
				if (B$4(t.current)?.findFiberByHostInstance) continue;
				let n = w$3(t.current, (t) => t.stateNode === e);
				if (n) return n;
			}
		}
		return null;
	};
}));
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/bippy/0.6.1/1130005d9ad5ea110510cf11f128db1358e7b7bbb7b2d18e02d0aee399b85c38/node_modules/bippy/dist/index.js
var init_dist = __esmMin((() => {
	init_rdt_hook();
	init_install_hook_only();
	init_core();
}));
/**
* @license bippy
*
* Copyright (c) Aiden Bai
*
* This source code is licensed under the MIT license found in the
* LICENSE file in the root directory of this source tree.
*/
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/bippy/0.6.1/1130005d9ad5ea110510cf11f128db1358e7b7bbb7b2d18e02d0aee399b85c38/node_modules/bippy/dist/get-source.js
/**
* @license bippy
*
* Copyright (c) Aiden Bai
*
* This source code is licensed under the MIT license found in the
* LICENSE file in the root directory of this source tree.
*/
function C$2(e, t) {
	let n = 0, r = 0, i = 0;
	do
		i = S$2[e.next()], n |= (i & 31) << r, r += 5;
	while (i & 32);
	let a = n & 1;
	return n >>>= 1, a && (n = -2147483648 | -n), t + n;
}
function w$2(e, t) {
	return e.pos >= t ? !1 : e.peek() !== oe$2;
}
function T$1(e) {
	let { length: t } = e, n = new ce$2(e), r = [], i = 0, a = 0, o = 0, s = 0, c = 0;
	do {
		let e = n.indexOf(`;`), t = [], l = !0, u = 0;
		for (i = 0; n.pos < e;) {
			let r;
			i = C$2(n, i), i < u && (l = !1), u = i, w$2(n, e) ? (a = C$2(n, a), o = C$2(n, o), s = C$2(n, s), w$2(n, e) ? (c = C$2(n, c), r = [
				i,
				a,
				o,
				s,
				c
			]) : r = [
				i,
				a,
				o,
				s
			]) : r = [i], t.push(r), n.pos++;
		}
		l || E$2(t), r.push(t), n.pos = e + 1;
	} while (n.pos <= t);
	return r;
}
function E$2(e) {
	e.sort(D$2);
}
function D$2(e, t) {
	return e[0] - t[0];
}
var i, a, o, s, c, l, u, d, f, ee$2, p, te$1, m, h, g$2, _$2, v$1, y$2, ne$2, re$2, ie$2, ae$2, b$2, oe$2, x$2, se$2, S$2, ce$2, O$1, k$1, le$2, A$2, j$2, M$3, N$3, ue$2, P$3, de$2, fe, F$3, pe$1, I$3, L$3, R$3, z$3, B$3, V$3, me$1, H$3, U$3, he$1, W$3, G$2, K$2, ge$1, _e$1, q$2, ve$1, ye$1, be$1, xe$1, J$2, Se$1, Ce$1, we$1, Y$2, Te$1, X$2, Ee$1, Z$3, De$1, Oe, Q$3, ke, Ae$1, $$3, je$1;
var init_get_source = __esmMin((() => {
	init_rdt_hook();
	init_core();
	i = /^[a-zA-Z][a-zA-Z\d+\-.]*:/;
	a = [
		`rsc://`,
		`file:///`,
		`webpack-internal://`,
		`webpack://`,
		`node:`,
		`turbopack://`,
		`metro://`,
		`/app-pages-browser/`,
		`/(app-pages-browser)/`
	];
	o = [`rsc://`, `about://React/`];
	s = [
		`<anonymous>`,
		`eval`,
		``
	];
	c = /\.(jsx|tsx|ts|js)$/;
	l = /(\.min|bundle|chunk|vendor|vendors|runtime|polyfill|polyfills)\.(js|mjs|cjs)$|(chunk|bundle|vendor|vendors|runtime|polyfill|polyfills|framework|app|main|index)[-_.][A-Za-z0-9_-]{4,}\.(js|mjs|cjs)$|[\da-f]{8,}\.(js|mjs|cjs)$|[-_.][\da-f]{20,}\.(js|mjs|cjs)$|\/dist\/|\/build\/|\/.next\/|\/out\/|\/node_modules\/|\.webpack\.|\.vite\.|\.turbopack\./i;
	u = /^\?[\w~.-]+(?:=[^&#]*)?(?:&[\w~.-]+(?:=[^&#]*)?)*$/;
	d = /\(at [^)]+\)$/;
	f = [`react_stack_bottom_frame`, `react-stack-bottom-frame`];
	ee$2 = /(^|@)\S+:\d+/;
	p = /^\s*at .*(\S+:\d+|\(native\))/m;
	te$1 = /^(eval@)?(\[native code\])?$/;
	m = (e, t) => {
		if (t?.includeInElement !== !1) {
			let n = e.split(`
`), r = [];
			for (let e of n) if (/^\s*at\s+/.test(e)) {
				let t = _$2(e, void 0)[0];
				t && r.push(t);
			} else if (/^\s*in\s+/.test(e)) {
				let t = e.replace(/^\s*in\s+/, ``).replace(/\s*\(at .*\)$/, ``);
				r.push({
					functionName: t,
					source: e
				});
			} else if (e.match(ee$2)) {
				let t = v$1(e, void 0)[0];
				t && r.push(t);
			}
			return g$2(r, t);
		}
		return e.match(p) ? _$2(e, t) : v$1(e, t);
	};
	h = (e) => {
		if (!e.includes(`:`)) return [
			e,
			void 0,
			void 0
		];
		let t = e.startsWith(`(`) && /:\d+\)$/.test(e) ? e.slice(1, -1) : e, n = /(.+?)(?::(\d+))?(?::(\d+))?$/.exec(t);
		return n ? [
			n[1],
			n[2] || void 0,
			n[3] || void 0
		] : [
			t,
			void 0,
			void 0
		];
	};
	g$2 = (e, t) => t && t.slice != null ? Array.isArray(t.slice) ? e.slice(t.slice[0], t.slice[1]) : e.slice(0, t.slice) : e;
	_$2 = (e, t) => g$2(e.split(`
`).filter((e) => !!e.match(p)), t).map((e) => {
		let t = e;
		t.includes(`(eval `) && (t = t.replace(/eval code/g, `eval`).replace(/(\(eval at [^()]*)|(,.*$)/g, ``));
		let n = t.replace(/^\s+/, ``).replace(/\(eval code/g, `(`).replace(/^.*?\s+/, ``), r = n.match(/ (\(.+\)$)/);
		n = r ? n.replace(r[0], ``) : n;
		let i = h(r ? r[1] : n);
		return {
			functionName: r && n || void 0,
			fileName: [`eval`, `<anonymous>`].includes(i[0]) ? void 0 : i[0],
			lineNumber: i[1] ? +i[1] : void 0,
			columnNumber: i[2] ? +i[2] : void 0,
			source: t
		};
	});
	v$1 = (e, t) => g$2(e.split(`
`).filter((e) => !e.match(te$1)), t).map((e) => {
		let t = e;
		if (t.includes(` > eval`) && (t = t.replace(/ line (\d+)(?: > eval line \d+)* > eval:\d+:\d+/g, `:$1`)), !t.includes(`@`) && !t.includes(`:`)) return { functionName: t };
		{
			let e = /(([^\n\r"\u2028\u2029]*".[^\n\r"\u2028\u2029]*"[^\n\r@\u2028\u2029]*(?:@[^\n\r"\u2028\u2029]*"[^\n\r@\u2028\u2029]*)*(?:[\n\r\u2028\u2029][^@]*)?)?[^@]*)@/, n = t.match(e), r = n && n[1] ? n[1] : void 0, i = h(t.replace(e, ``));
			return {
				functionName: r,
				fileName: i[0],
				lineNumber: i[1] ? +i[1] : void 0,
				columnNumber: i[2] ? +i[2] : void 0,
				source: t
			};
		}
	});
	y$2 = /* @__PURE__ */ new WeakMap();
	ne$2 = (e) => f.some((t) => e.includes(t));
	re$2 = (e) => {
		let t = e.getFunctionName?.() ?? ``;
		if (t) return t;
		let n = e.getTypeName?.() ?? ``, r = e.getMethodName?.() ?? ``;
		return n && r ? `${n}.${r}` : r;
	};
	ie$2 = (e) => {
		let t = [];
		for (let n = 1; n < e.length; n++) {
			let r = e[n], i = re$2(r);
			if (ne$2(i)) return {
				frames: t,
				isTrusted: !0
			};
			if (r.isNative?.()) {
				t.push({ functionName: i || void 0 });
				continue;
			}
			let a = r.getScriptNameOrSourceURL?.() ?? ``;
			!a && r.isEval?.() && (a = r.getEvalOrigin?.() ?? ``), t.push({
				functionName: i && i !== `<anonymous>` ? i : void 0,
				fileName: a && a !== `<anonymous>` ? a : void 0,
				lineNumber: r.getLineNumber?.() ?? void 0,
				columnNumber: r.getColumnNumber?.() ?? void 0,
				enclosingLineNumber: r.getEnclosingLineNumber?.() ?? void 0,
				enclosingColumnNumber: r.getEnclosingColumnNumber?.() ?? void 0,
				source: `    at ${r.toString()}`
			});
		}
		return {
			frames: t,
			isTrusted: !1
		};
	};
	ae$2 = (e) => {
		let t = -1;
		for (let n of f) if (t = e.indexOf(n), t !== -1) break;
		return {
			frames: m(t === -1 ? e : e.slice(0, e.lastIndexOf(`
`, t))).slice(1),
			isTrusted: t !== -1
		};
	};
	b$2 = (e) => {
		let t = y$2.get(e);
		if (t) return t;
		let n = null, r = (e, t) => {
			n = ie$2(t);
			let r = `${e.name || `Error`}: ${e.message || ``}`;
			for (let e of t) r += `\n    at ${e.toString()}`;
			return r;
		}, i = Error.prepareStackTrace;
		Error.prepareStackTrace = r;
		let a;
		try {
			a = String(e.stack);
		} finally {
			Error.prepareStackTrace = i;
		}
		let o = n ?? ae$2(a);
		return y$2.set(e, o), o;
	};
	oe$2 = 44;
	x$2 = `ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/`;
	se$2 = /* @__PURE__ */ new Uint8Array(64);
	S$2 = /* @__PURE__ */ new Uint8Array(128);
	for (let e = 0; e < x$2.length; e++) {
		let t = x$2.charCodeAt(e);
		se$2[e] = t, S$2[t] = e;
	}
	ce$2 = class {
		constructor(e) {
			this.pos = 0, this.buffer = e;
		}
		next() {
			return this.buffer.charCodeAt(this.pos++);
		}
		peek() {
			return this.buffer.charCodeAt(this.pos);
		}
		indexOf(e) {
			let { buffer: t, pos: n } = this, r = t.indexOf(e, n);
			return r === -1 ? t.length : r;
		}
	};
	O$1 = /^[a-zA-Z][a-zA-Z\d+\-.]*:/;
	k$1 = /^data:application\/json[^,]+base64,/;
	le$2 = /(?:\/\/[@#][ \t]+sourceMappingURL=([^\s'"]+?)[ \t]*$)|(?:\/\*[@#][ \t]+sourceMappingURL=([^*]+?)[ \t]*(?:\*\/)[ \t]*$)/;
	A$2 = /* @__PURE__ */ new Map();
	j$2 = /* @__PURE__ */ new Map();
	M$3 = (e, t, n, r, i) => {
		if (n < 0 || n >= e.length) return null;
		let a = e[n];
		if (!a || a.length === 0) return null;
		let o = null, s = 0, c = a.length - 1;
		for (; s <= c;) {
			let e = s + c >> 1;
			a[e][0] <= r ? (o = a[e], s = e + 1) : c = e - 1;
		}
		if (!o || o.length < 4) return null;
		let [, l, u, d] = o;
		if (l === void 0 || u === void 0 || d === void 0) return null;
		let f = t[l];
		return f ? {
			columnNumber: d,
			fileName: f,
			lineNumber: u + 1,
			isIgnoreListed: i?.has(l) ?? !1
		} : null;
	};
	N$3 = (e, t, n) => {
		if (e.sections) {
			let r = t - 1, i = null;
			for (let t of e.sections) if (r > t.offset.line || r === t.offset.line && n >= t.offset.column) i = t;
			else break;
			if (!i) return null;
			let a = r - i.offset.line, o = r === i.offset.line ? n - i.offset.column : n;
			return M$3(i.map.mappings, i.map.sources, a, o, i.map.ignoredSourceIndices);
		}
		return M$3(e.mappings, e.sources, t - 1, n, e.ignoredSourceIndices);
	};
	ue$2 = (e, t) => {
		let n, r = t.length;
		for (; r > 0 && !n;) {
			let e = t.lastIndexOf(`
`, r - 1) + 1, i = t.slice(e, r).match(le$2);
			i && (n = i[1] || i[2]), r = e - 1;
		}
		if (!n) return null;
		let i = O$1.test(n);
		if (!(k$1.test(n) || i || n.startsWith(`/`))) {
			let t = e.split(`/`);
			t[t.length - 1] = n, n = t.join(`/`);
		}
		return n;
	};
	P$3 = (e) => {
		let t = e.ignoreList ?? e.x_google_ignoreList;
		return Array.isArray(t) && t.length > 0 ? new Set(t) : void 0;
	};
	de$2 = (e) => ({
		file: e.file,
		ignoredSourceIndices: P$3(e),
		mappings: T$1(e.mappings),
		names: e.names,
		sourceRoot: e.sourceRoot,
		sources: e.sources,
		sourcesContent: e.sourcesContent,
		version: 3
	});
	fe = (e) => {
		let t = e.sections.map(({ map: e, offset: t }) => ({
			map: {
				...e,
				ignoredSourceIndices: P$3(e),
				mappings: T$1(e.mappings)
			},
			offset: t
		})), n = /* @__PURE__ */ new Set();
		for (let e of t) for (let t of e.map.sources) n.add(t);
		return {
			file: e.file,
			mappings: [],
			names: [],
			sections: t,
			sourceRoot: void 0,
			sources: Array.from(n),
			sourcesContent: void 0,
			version: 3
		};
	};
	F$3 = (e) => {
		if (!e) return !1;
		let t = e.trim();
		if (!t) return !1;
		let n = t.match(O$1);
		if (!n) return !0;
		let r = n[0].toLowerCase();
		return r === `http:` || r === `https:`;
	};
	pe$1 = async (e, t = fetch) => {
		if (!F$3(e)) return null;
		let n = await t(e);
		if (!n.ok) return null;
		let r = await n.text();
		if (!r) return null;
		let i = ue$2(e, r);
		if (!i || !F$3(i) && !k$1.test(i)) return null;
		let a = await t(i);
		if (!a.ok) return null;
		try {
			let e = await a.json();
			return `sections` in e ? fe(e) : de$2(e);
		} catch {
			return null;
		}
	};
	I$3 = async (e, t = !0, n) => {
		if (t && A$2.has(e)) return A$2.get(e) ?? null;
		let r = t ? j$2.get(e) : void 0;
		if (r) return (await r).sourceMap;
		let i = pe$1(e, n).then((e) => ({
			sourceMap: e,
			isTransientFailure: !1
		}), () => ({
			sourceMap: null,
			isTransientFailure: !0
		}));
		t && j$2.set(e, i);
		let { sourceMap: a, isTransientFailure: o } = await i;
		return t && (j$2.delete(e), o || A$2.set(e, a)), a;
	};
	L$3 = async (e, t = !0, n) => await Promise.all(e.map(async (e) => {
		if (!e.fileName) return e;
		let r = await I$3(e.fileName, t, n);
		if (!r || typeof e.lineNumber != `number` || typeof e.columnNumber != `number`) return e;
		let i = N$3(r, e.lineNumber, e.columnNumber);
		return i ? {
			...e,
			source: i.fileName && e.source ? e.source.replace(e.fileName, i.fileName) : e.source,
			fileName: i.fileName,
			lineNumber: i.lineNumber,
			columnNumber: i.columnNumber,
			isIgnoreListed: i.isIgnoreListed,
			isSymbolicated: !0
		} : e;
	}));
	R$3 = (e) => e._debugStack instanceof Error && typeof e._debugStack?.stack == `string`;
	z$3 = (e) => typeof e.tag == `number`;
	B$3 = (e) => e._debugOwner;
	V$3 = (e) => {
		let t = null;
		if (w$3(e, (n) => {
			if (n === e) return !1;
			let r = n._debugOwner;
			return (r === e || e.alternate !== null && r === e.alternate) && n._debugStack instanceof Error ? (t = n._debugStack, !0) : !1;
		}), !t) return null;
		let { frames: n, isTrusted: i } = b$2(t);
		if (!i) return null;
		for (let e = n.length - 1; e >= 0; e--) {
			let t = n[e];
			if (t.fileName) return {
				...t,
				lineNumber: t.enclosingLineNumber || t.lineNumber,
				columnNumber: t.enclosingColumnNumber || t.columnNumber
			};
		}
		return null;
	};
	me$1 = () => {
		let n = x$3();
		for (let e of [...Array.from(f$1), ...Array.from(n.renderers.values())]) {
			let t = e.currentDispatcherRef;
			if (t && typeof t == `object`) return `H` in t ? t.H : t.current;
		}
		return null;
	};
	H$3 = (e) => {
		for (let n of f$1) {
			let t = n.currentDispatcherRef;
			t && typeof t == `object` && (`H` in t ? t.H = e : t.current = e);
		}
	};
	U$3 = (e) => `\n    in ${e}`;
	he$1 = (e, t) => {
		let n = U$3(e);
		return t && (n += ` (at ${t})`), n;
	};
	W$3 = !1;
	G$2 = /* @__PURE__ */ new WeakMap();
	K$2 = (e, t) => {
		if (!e || W$3) return ``;
		let r = G$2.get(e);
		if (r !== void 0) return r;
		let i = Error.prepareStackTrace;
		Error.prepareStackTrace = void 0, W$3 = !0;
		let a = me$1();
		H$3(null);
		let o = console.error, s = console.warn;
		console.error = () => {}, console.warn = () => {};
		try {
			let r = { DetermineComponentFrameRoot() {
				let n;
				try {
					if (t) {
						let t = function() {
							throw Error();
						};
						if (Object.defineProperty(t.prototype, `props`, { set: function() {
							throw Error();
						} }), typeof Reflect == `object` && Reflect.construct) {
							try {
								Reflect.construct(t, []);
							} catch (e) {
								n = e;
							}
							Reflect.construct(e, [], t);
						} else {
							try {
								t.call();
							} catch (e) {
								n = e;
							}
							e.call(t.prototype);
						}
					} else {
						try {
							throw Error();
						} catch (e) {
							n = e;
						}
						let t = e();
						t && typeof t.catch == `function` && t.catch(() => {});
					}
				} catch (e) {
					if (e instanceof Error && n instanceof Error && typeof e.stack == `string`) return [e.stack, n.stack];
				}
				return [null, null];
			} };
			r.DetermineComponentFrameRoot.displayName = `DetermineComponentFrameRoot`, Object.getOwnPropertyDescriptor(r.DetermineComponentFrameRoot, `name`)?.configurable && Object.defineProperty(r.DetermineComponentFrameRoot, `name`, { value: `DetermineComponentFrameRoot` });
			let [i, a] = r.DetermineComponentFrameRoot();
			if (i && a) {
				let t = i.split(`
`), r = a.split(`
`), o = 0, s = 0;
				for (; o < t.length && !t[o].includes(`DetermineComponentFrameRoot`);) o++;
				for (; s < r.length && !r[s].includes(`DetermineComponentFrameRoot`);) s++;
				if (o === t.length || s === r.length) for (o = t.length - 1, s = r.length - 1; o >= 1 && s >= 0 && t[o] !== r[s];) s--;
				for (; o >= 1 && s >= 0; o--, s--) if (t[o] !== r[s]) {
					if (o !== 1 || s !== 1) do
						if (o--, s--, s < 0 || t[o] !== r[s]) {
							let r = `\n${t[o].replace(` at new `, ` at `)}`, i = Ie$1(e);
							return i && r.includes(`<anonymous>`) && (r = r.replace(`<anonymous>`, i)), G$2.set(e, r), r;
						}
					while (o >= 1 && s >= 0);
					break;
				}
			}
		} finally {
			W$3 = !1, Error.prepareStackTrace = i, H$3(a), console.error = o, console.warn = s;
		}
		let c = e ? Ie$1(e) : ``, l = c ? U$3(c) : ``;
		return G$2.set(e, l), l;
	};
	ge$1 = (e, t) => {
		let n = e.tag, r = ``;
		switch (n) {
			case 28:
				r = U$3(`Activity`);
				break;
			case 1:
				r = K$2(e.type, !0);
				break;
			case 11:
				r = K$2(e.type.render, !1);
				break;
			case 0:
			case 15:
				r = K$2(e.type, !1);
				break;
			case 5:
			case 26:
			case 27:
				r = U$3(e.type);
				break;
			case 16:
				r = U$3(`Lazy`);
				break;
			case 13:
				r = e.child !== t && t !== null ? U$3(`Suspense Fallback`) : U$3(`Suspense`);
				break;
			case 19:
				r = U$3(`SuspenseList`);
				break;
			case 30:
				r = U$3(`ViewTransition`);
				break;
			default: return ``;
		}
		return r;
	};
	_e$1 = (e) => {
		try {
			let t = ``, n = e, r = null;
			do {
				t += ge$1(n, r);
				let e = n._debugInfo;
				if (e && Array.isArray(e)) for (let n = e.length - 1; n >= 0; n--) {
					let r = e[n];
					typeof r.name == `string` && (t += he$1(r.name, r.env));
				}
				r = n, n = n.return;
			} while (n);
			return t;
		} catch (e) {
			return e instanceof Error ? `\nError generating stack: ${e.message}\n${e.stack}` : ``;
		}
	};
	q$2 = (e) => {
		let t = Error.prepareStackTrace;
		Error.prepareStackTrace = void 0;
		let n = e;
		if (!n) return ``;
		Error.prepareStackTrace = t, n.startsWith(`Error: react-stack-top-frame
`) && (n = n.slice(29));
		let r = n.indexOf(`
`);
		r !== -1 && (n = n.slice(r + 1));
		let i = Math.max(n.indexOf(`react_stack_bottom_frame`), n.indexOf(`react-stack-bottom-frame`));
		if (i !== -1 && (i = n.lastIndexOf(`
`, i)), i !== -1) n = n.slice(0, i);
		else return ``;
		return n;
	};
	ve$1 = (e) => !!(e.functionName && e.fileName && J$2(e.fileName));
	ye$1 = (e, t) => e.fileName === t.fileName && e.lineNumber === t.lineNumber && e.columnNumber === t.columnNumber;
	be$1 = (e) => {
		let t = /* @__PURE__ */ new Map();
		for (let n of e) for (let e of n.stackFrames) {
			if (!ve$1(e)) continue;
			let n = e.functionName, r = t.get(n) ?? [];
			r.some((t) => ye$1(t, e)) || (r.push(e), t.set(n, r));
		}
		return t;
	};
	xe$1 = (e, t, n) => {
		if (!e.functionName) return {
			...e,
			isServer: !0
		};
		let r = t.get(e.functionName);
		if (!r || r.length === 0) return {
			...e,
			isServer: !0
		};
		let i = n.get(e.functionName) ?? 0, a = r[i % r.length];
		return n.set(e.functionName, i + 1), {
			...e,
			isServer: !0,
			fileName: a.fileName,
			lineNumber: a.lineNumber,
			columnNumber: a.columnNumber,
			source: e.source?.replace(`(at Server)`, `(${a.fileName}:${a.lineNumber}:${a.columnNumber})`)
		};
	};
	J$2 = (e) => o.some((t) => e.startsWith(t));
	Se$1 = (e) => !e.isServer && e.fileName && J$2(e.fileName) ? {
		...e,
		isServer: !0
	} : e;
	Ce$1 = (e) => {
		let t = [], n = e;
		for (; n;) if (z$3(n)) {
			let e = n;
			if (n = B$3(e), n && R$3(e)) {
				let { frames: n, isTrusted: r } = b$2(e._debugStack);
				if (r) for (let e of n) t.push(Se$1(e));
			}
		} else {
			let e = n;
			if (n = e.owner, n && e.debugStack instanceof Error) for (let n of b$2(e.debugStack).frames) t.push({
				...n,
				isServer: !0
			});
		}
		return t;
	};
	we$1 = (e) => {
		let t = [];
		return w$3(e, (e) => {
			if (!R$3(e)) return;
			let r = typeof e.type == `string` ? e.type : Ie$1(e.type) || `<anonymous>`;
			t.push({
				componentName: r,
				stackFrames: m(q$2(e._debugStack?.stack))
			});
		}, !0), t;
	};
	Y$2 = async (e, t = !0, n) => {
		let r = we$1(e), i = m(_e$1(e)), a = be$1(r), o = /* @__PURE__ */ new Map();
		return L$3(i.map((e) => (e.source?.includes(`(at Server)`) ?? !1) || e.source != null && d.test(e.source) ? xe$1(e, a, o) : e).filter((e, t, n) => {
			if (t === 0) return !0;
			let r = n[t - 1];
			return e.functionName !== r.functionName;
		}), t, n);
	};
	Te$1 = (e) => !!e.fileName && !e.isIgnoreListed;
	X$2 = async (e, t = !0, r) => {
		let i = Ce$1(e);
		if (i.length > 0) {
			let a = V$3(e) ?? {};
			a.functionName = Ie$1(e.type) ?? a.functionName;
			let o = await L$3([a, ...i], t, r);
			if (o.some((e, t) => t > 0 && Te$1(e))) return o;
		}
		return Y$2(e, t, r);
	};
	Ee$1 = (e) => {
		let t = e._debugSource;
		return t ? typeof t == `object` && !!t && `fileName` in t && typeof t.fileName == `string` && `lineNumber` in t && typeof t.lineNumber == `number` : !1;
	};
	Z$3 = (e) => e.fileName ? {
		fileName: e.fileName,
		lineNumber: e.lineNumber,
		columnNumber: e.columnNumber,
		functionName: e.functionName
	} : null;
	De$1 = (e) => {
		if (!R$3(e)) return null;
		let { frames: t, isTrusted: n } = b$2(e._debugStack);
		if (!n) return null;
		for (let e of t) if (e.fileName) return e;
		return null;
	};
	Oe = async (e, t = !0, n) => {
		if (Ee$1(e)) return e._debugSource || null;
		let r = De$1(e) ?? V$3(e);
		if (r) {
			let [e] = await L$3([r], t, n), i = Z$3(e);
			if (i) return i;
		}
		let i = await Y$2(e, t, n);
		for (let e of i) if (e.fileName) return Z$3(e);
		return null;
	};
	Q$3 = (e) => e.split(`/`).filter(Boolean).length;
	ke = (e) => e.split(`/`).filter(Boolean)[0] ?? null;
	Ae$1 = (e) => {
		let t = e.indexOf(`/`, 1);
		if (t === -1 || Q$3(e.slice(0, t)) !== 1) return e;
		let n = e.slice(t);
		if (!c.test(n) || Q$3(n) < 2) return e;
		let r = ke(n);
		return !r || r.startsWith(`@`) || r.length > 4 ? e : n;
	};
	$$3 = (e) => {
		if (!e || s.some((t) => t === e)) return ``;
		let t = e, n = t.startsWith(`http://`) || t.startsWith(`https://`);
		if (n) try {
			t = new URL(t).pathname;
		} catch {}
		if (n && (t = Ae$1(t)), t.startsWith(`about://React/`)) {
			let e = t.slice(14), n = e.indexOf(`/`), r = e.indexOf(`:`);
			t = n !== -1 && (r === -1 || n < r) ? e.slice(n + 1) : e;
		}
		let r = !0;
		for (; r;) {
			r = !1;
			for (let e of a) if (t.startsWith(e)) {
				t = t.slice(e.length), e === `file:///` && (t = `/${t.replace(/^\/+/, ``)}`), r = !0;
				break;
			}
		}
		if (i.test(t)) {
			let e = t.match(i);
			e && (t = t.slice(e[0].length));
		}
		if (t.startsWith(`//`)) {
			let e = t.indexOf(`/`, 2);
			t = e === -1 ? `` : t.slice(e);
		}
		let o = t.indexOf(`?`);
		if (o !== -1) {
			let e = t.slice(o);
			u.test(e) && (t = t.slice(0, o));
		}
		return t;
	};
	je$1 = (e) => {
		let t = $$3(e);
		return !(!t || !c.test(t) || l.test(t));
	};
})), g$1, ee$1, _$1, y$1, b$1, x$1, S$1, C$1, w$1, E$1, D$1, A$1, ne$1, re$1, ie$1, ae$1, oe$1, se$1, ce$1, le$1, ue$1, j$1, M$2, N$2, P$2, F$2, I$2, L$2, R$2, z$2, B$2, V$2, H$2, U$2, W$2;
var init_source = __esmMin((() => {
	init_get_source();
	g$1 = Symbol.for(`react.context`);
	ee$1 = Symbol.for(`react.memo_cache_sentinel`);
	_$1 = [];
	y$1 = null;
	b$1 = null;
	x$1 = null;
	S$1 = 0;
	C$1 = null;
	w$1 = Error("Suspense Exception: This is not a real error! It's an implementation detail of `use` to interrupt the current render.");
	E$1 = () => {
		/**
		* @license bippy
		*
		* Copyright (c) Aiden Bai
		*
		* This source code is licensed under the MIT license found in the
		* LICENSE file in the root directory of this source tree.
		*/
		let e = b$1;
		return e !== null && (b$1 = e.next), e;
	};
	D$1 = (e) => {
		if (y$1 === null) return e._currentValue;
		if (x$1 === null) throw Error(`Context reads do not line up with context dependencies.`);
		if (Object.prototype.hasOwnProperty.call(x$1, `memoizedValue`)) {
			let e = x$1.memoizedValue;
			return x$1 = x$1.next, e;
		}
		return e._currentValue;
	};
	A$1 = (e, t, n, r = null) => {
		_$1.push({
			displayName: r,
			primitive: e,
			stackError: Error(),
			value: t,
			dispatcherHookName: n
		});
	};
	ne$1 = (e) => {
		if (typeof e == `object` && e) {
			let t = e;
			if (typeof t.then == `function`) {
				let e = C$1 !== null && S$1 < C$1.length ? C$1[S$1++] : t;
				switch (e.status) {
					case `fulfilled`: return A$1(`Promise`, e.value, `Use`), e.value;
					case `rejected`: throw e.reason;
				}
				throw A$1(`Unresolved`, e, `Use`), w$1;
			}
			if (t.$$typeof === g$1 && `_currentValue` in t) {
				let e = t, n = D$1(e);
				return A$1(`Context (use)`, n, `Use`, e.displayName || `Context`), n;
			}
		}
		throw Error(`An unsupported type was passed to use(): ` + String(e));
	};
	re$1 = (e) => {
		let t = D$1(e);
		return A$1(`Context`, t, `Context`, e.displayName || null), t;
	};
	ie$1 = (e) => {
		let t = E$1(), n = t === null ? typeof e == `function` ? e() : e : t.memoizedState;
		return A$1(`State`, n, `State`), [n, () => {}];
	};
	ae$1 = (e, t, n) => {
		let r = E$1(), i = r === null ? n === void 0 ? t : n(t) : r.memoizedState;
		return A$1(`Reducer`, i, `Reducer`), [i, () => {}];
	};
	oe$1 = (e) => {
		let t = E$1(), n = t === null ? { current: e } : t.memoizedState;
		return A$1(`Ref`, n.current, `Ref`), n;
	};
	se$1 = () => {
		let e = E$1();
		return A$1(`CacheRefresh`, e === null ? () => {} : e.memoizedState, `CacheRefresh`), () => {};
	};
	ce$1 = (e) => {
		E$1(), A$1(`LayoutEffect`, e, `LayoutEffect`);
	};
	le$1 = (e) => {
		E$1(), A$1(`InsertionEffect`, e, `InsertionEffect`);
	};
	ue$1 = (e) => {
		E$1(), A$1(`Effect`, e, `Effect`);
	};
	j$1 = (e) => {
		E$1();
		let t;
		typeof e == `object` && e && `current` in e && (t = e.current), A$1(`ImperativeHandle`, t, `ImperativeHandle`);
	};
	M$2 = (e, t) => {
		A$1(`DebugValue`, typeof t == `function` ? t(e) : e, `DebugValue`);
	};
	N$2 = (e) => {
		let t = E$1();
		return A$1(`Callback`, t === null ? e : t.memoizedState[0], `Callback`), e;
	};
	P$2 = (e) => {
		let t = E$1(), n = t === null ? e() : t.memoizedState[0];
		return A$1(`Memo`, n, `Memo`), n;
	};
	F$2 = (e, t) => {
		let n = E$1();
		E$1();
		let r = n === null ? t() : n.memoizedState;
		return A$1(`SyncExternalStore`, r, `SyncExternalStore`), r;
	};
	I$2 = () => {
		let e = E$1();
		E$1();
		let t = e === null ? !1 : e.memoizedState;
		return A$1(`Transition`, t, `Transition`), [t, () => {}];
	};
	L$2 = (e) => {
		let t = E$1(), n = t === null ? e : t.memoizedState;
		return A$1(`DeferredValue`, n, `DeferredValue`), n;
	};
	R$2 = () => {
		let e = E$1(), t = e === null ? `` : e.memoizedState;
		return A$1(`Id`, t, `Id`), t;
	};
	z$2 = (e) => {
		let t = y$1;
		if (t == null) return [];
		let n = t.updateQueue?.memoCache;
		if (n == null) return [];
		let r = n.data[n.index];
		return r === void 0 && (r = n.data[n.index] = Array.from({ length: e }, () => ee$1)), n.index++, r;
	};
	B$2 = (e) => {
		let t = E$1(), n = t === null ? e : t.memoizedState;
		return A$1(`Optimistic`, n, `Optimistic`), [n, () => {}];
	};
	V$2 = (e, t) => {
		let n, r = null;
		if (e !== null) {
			let t = e.memoizedState;
			if (typeof t == `object` && t && `then` in t && typeof t.then == `function`) {
				let e = t;
				switch (e.status) {
					case `fulfilled`:
						n = e.value;
						break;
					case `rejected`:
						r = e.reason;
						break;
					default: r = w$1, n = e;
				}
			} else n = t;
		} else n = t;
		return {
			value: n,
			error: r
		};
	};
	H$2 = (e) => (t, n) => {
		let r = E$1();
		E$1(), E$1();
		let i = Error(), { value: a, error: o } = V$2(r, n);
		if (_$1.push({
			displayName: null,
			primitive: e,
			stackError: i,
			value: a,
			dispatcherHookName: e
		}), o !== null) throw o;
		return [
			a,
			() => {},
			!1
		];
	};
	U$2 = H$2(`ActionState`);
	W$2 = {
		readContext: D$1,
		use: ne$1,
		useCallback: N$2,
		useContext: re$1,
		useEffect: ue$1,
		useImperativeHandle: j$1,
		useLayoutEffect: ce$1,
		useInsertionEffect: le$1,
		useMemo: P$2,
		useReducer: ae$1,
		useRef: oe$1,
		useState: ie$1,
		useDebugValue: M$2,
		useDeferredValue: L$2,
		useTransition: I$2,
		useSyncExternalStore: F$2,
		useId: R$2,
		useHostTransitionStatus: () => {
			let e = D$1({ _currentValue: null });
			return A$1(`HostTransitionStatus`, e, `HostTransitionStatus`), e;
		},
		useFormState: H$2(`FormState`),
		useActionState: U$2,
		useOptimistic: B$2,
		useMemoCache: z$2,
		useCacheRefresh: se$1,
		useEffectEvent: (e) => (E$1(), A$1(`EffectEvent`, e, `EffectEvent`), e)
	};
	typeof Proxy > `u` || new Proxy(W$2, { get(e, t) {
		if (Object.prototype.hasOwnProperty.call(e, t)) return e[t];
		let n = Error(`Missing method in Dispatcher: ` + t);
		throw n.name = `ReactDebugToolsUnsupportedHookError`, n;
	} });
}));
// ../ui/dist/react-grab/freeze-updates-Cd2QP_tz.js
/** @license MIT
* Copyright (c) 2025 Aiden Bai
* See LICENSE in the package root.
*/
function wr$2(e) {
	return Object.fromEntries([...e.attributes].filter((e) => e.name.startsWith(`data-grab-`) && e.name.length > 10).map((e) => [e.name.slice(10), e.value]));
}
function Tr$2(e) {
	let t = e.ownerDocument, n = {
		page: [],
		elements: []
	};
	for (let e of [t.documentElement, t.body]) {
		if (!e) continue;
		let t = wr$2(e);
		Object.keys(t).length && n.page.push({
			selector: e.tagName.toLowerCase(),
			attributes: t
		});
	}
	for (let r = e; r && r !== t.documentElement && r !== t.body;) {
		let e = wr$2(r);
		Object.keys(e).length && n.elements.unshift({
			selector: Nn$2(r),
			attributes: e
		});
		let t = r.getRootNode();
		r = r.parentElement ?? (t.nodeType === 11 && `host` in t ? t.host : null);
	}
	return n;
}
function Er$2(e) {
	return [[`Page context`, e.page], [`Element context`, e.elements]].filter(([, e]) => e.length).map(([e, t]) => `## ${e}\n\n${t.map((e) => `${e.selector}\n${Object.entries(e.attributes).map(([e, t]) => `- **${e}:** ${t}`).join(`
`)}`).join(`

`)}`).join(`

`);
}
var g, ee, te, _, ne, re, ie, ae, oe, se, ce, le, ue, de, pe, me, he, ge, _e, ve, ye, be, xe, Se, Ce, we, Te, Ee, De, Ae, je, Me, Ne, Pe, Fe, Ie, Re, ze, Be, Ve, He, Ue, We$1, Ge$1, Ke$1, qe$1, Je$1, Ye$1, Xe$2, Ze$2, Qe$2, et$2, tt$2, v, y, b, rt$2, x, it$2, at$2, ot$2, st$2, ct$2, lt$2, ut$2, dt$2, ft$2, S, C, pt$2, mt$2, ht$2, gt$2, _t$2, w, T, E, yt$2, bt$2, xt$2, St$2, Ct$2, wt$2, Tt$2, Et$2, Dt$2, Ot$2, D, kt$2, At$2, jt$2, Mt$2, Nt$2, Pt$2, Ft$2, It$2, Lt$2, Rt$2, zt$2, Bt$2, Vt$2, O, Ht$2, Ut$2, Wt$2, Gt$2, Kt$2, qt$2, Jt$2, Yt$2, Xt$2, Zt$2, Qt$2, $t$2, en$2, tn$2, nn$2, rn$2, an$2, on$2, sn$2, cn$2, ln$2, un$2, dn$2, fn$2, k, pn$2, mn$2, hn$2, gn$2, _n$2, vn$2, yn$2, bn$2, xn$2, Sn$2, Cn$2, wn$2, Tn$2, En$2, Dn$2, On$2, kn$2, An$2, jn$2, Mn$2, Nn$2, Pn$2, Fn$2, In$2, Ln$2, Rn$2, zn$2, Bn$2, Vn$2, Hn$2, Un$2, Wn$2, Gn$2, Kn$2, qn$2, Jn$2, Yn$2, Xn$2, Zn$2, Qn$2, $n$2, er$2, tr$2, nr$2, rr$2, ir$2, ar$2, or$2, sr$2, cr$2, lr$2, ur$2, dr$2, fr$2, mr$2, gr$2, _r$2, vr$2, yr$2, br$2, xr$2, Sr$2, Cr$2, Dr$2, A, Or$2, j, kr$2, Ar$2, jr$2, Mr$2, Nr$2, Pr$2, Fr$2, Ir$2, Lr$2, Rr$2, zr$2, Br$2, Vr$2, Hr$2, Ur$2, Wr$2, Gr$2, Kr$2, qr$2, Jr$2, Yr$2, Xr$2, Zr$2, Qr$2, $r$2, ei$2, ti$2, ni$2, ri$2, ai$2, oi$2, si$2, M$1, ci$2, li$2, ui$2, di$2, N$1, fi$2, pi$2, mi$2, hi$2, gi$2, _i$2, vi$2, yi$2, bi$2, xi$2, Si$2, Ci$2, wi$2, Ti$2, Ei$2, Di$2, Oi$2, ki$2, Ai$2, ji$2, Mi$2, Ni$2, Pi$2, Fi$2, Ii$2, Li$2, P$1, F$1, Ri$2, zi$2, Bi$2, Vi$2, Hi$2, Ui$2, Wi$2, Gi$2, Ki$2, I$1, qi$2, L$1, Ji$2, R$1, Yi$2, Xi$2, Zi$2, Qi$2, $i$2, ea$2, ta$2, na$2, ra$2, ia$2, aa$2, oa$2, sa$2, ca$2, la$2, ua$2, da$1, fa$1, pa$1, ha$1, ga$1, _a$1, va$1, ya$1, ba$1, xa$1, z$1, B$1, V$1, Sa$1, Ca$1, wa$1, Ta$1, Ea$1, Da$1, Oa$1, ka$1, H$1, Aa$1, U$1, W$1, ja$1, Ma$1, G$1, K$1, Na$1, Pa$1, Fa$1, q$1, Ia$1, La$1, Ra$1, J$1, za$1, Ba$1, Va$1, Y$1, X$1, Ha$1, Ua$1, Wa$1, Ga$1, Ka$1, qa$1, Ja$1, Ya$1, Xa$1, Za$1, Qa$1, $a$1, eo$1, to$1, no$1, ro$1, io$1, ao$1, oo$1, so$1, co$1, lo$1, uo$1, fo$1, po$1, mo$1, ho$1, go$1, _o$1, Z$2, vo$1, yo$1, bo$1, xo$1, So$1, Co$1, wo$1, To$1, Eo$1, Do$1, Oo$1, ko$1, Ao$1, jo$1, Mo$1, No$1, Po$1, Q$2, Fo$1, Io$1, Lo$1, Ro$1, zo$1, Bo$1, Vo$1, $$2, Ho$1, Uo$1, Wo$1, Go$1, Ko$1, qo$1, Jo$1, Yo$1, Xo$1, Zo$1, Qo$1, $o$1, es$1, ts$1, ns$1, rs$1, is$1, as$1, os$1, ss$1, cs$1, ls$1, us$1, ds$1, fs$1, ps$1, ms$1, hs$1;
var init_freeze_updates_Cd2QP_tz = __esmMin((() => {
	init_dist();
	init_source();
	g = null;
	ee = () => {
		if (g !== null) return g;
		try {
			g = window.matchMedia(`(color-gamut: p3)`).matches;
		} catch {
			g = !1;
		}
		return g;
	};
	te = ee();
	_ = (e) => te ? `color(display-p3 0.84 0.19 0.78 / ${e})` : `rgba(210, 57, 192, ${e})`;
	ne = `var(--rg-panel-bg)`;
	re = `var(--rg-shadow)`;
	ie = -1e3;
	ae = .95;
	oe = 1500;
	se = [
		`/components/ui/`,
		`/packages/ui/`,
		`/design-system/`,
		`/design-systems/`,
		`/primitives/`
	];
	ce = 2147483647;
	le = 2147483645;
	ue = .7;
	de = 1e3 / 60;
	pe = .01;
	me = _(.4);
	he = _(.05);
	ge = _(.5);
	_e = _(.08);
	ve = _(.15);
	ye = .2;
	be = [
		`id`,
		`class`,
		`aria-label`,
		`data-testid`,
		`role`,
		`name`,
		`title`
	];
	xe = /* @__PURE__ */ new Set([
		`id`,
		`data-testid`,
		`aria-label`,
		`href`,
		`src`,
		`alt`,
		`type`,
		`name`,
		`placeholder`,
		`role`,
		`for`,
		`action`,
		`method`,
		`title`,
		`disabled`,
		`checked`,
		`readonly`,
		`required`,
		`selected`,
		`open`
	]);
	Se = /* @__PURE__ */ new Set([
		`a`,
		`code`,
		`pre`
	]);
	Ce = /* @__PURE__ */ new Set([
		`script`,
		`style`,
		`template`,
		`noscript`
	]);
	we = [
		`Meta`,
		`Control`,
		`Shift`,
		`Alt`
	];
	Te = /* @__PURE__ */ new Set([
		`ArrowUp`,
		`ArrowDown`,
		`ArrowLeft`,
		`ArrowRight`
	]);
	Ee = `data-react-grab-frozen`;
	De = `data-react-grab-same-origin-frame`;
	Ae = .5;
	je = `data-react-grab-input`;
	Me = .5;
	Ne = `copy`;
	Pe = `comment`;
	Fe = `edit`;
	Ie = `superellipse(1.25)`;
	Re = .01;
	ze = 1e3;
	Be = {
		left: -9999,
		top: -9999
	};
	Ve = {
		left: `left center`,
		right: `right center`,
		top: `center top`,
		bottom: `center bottom`
	};
	He = 1e3;
	Ue = `#00000000`;
	We$1 = `transparent`;
	Ge$1 = 2e3;
	Ke$1 = `cubic-bezier(0.34, 1.56, 0.64, 1)`;
	qe$1 = 12e3;
	Je$1 = 9e3;
	Ye$1 = 1e4;
	Xe$2 = 8e3;
	Ze$2 = 5e3;
	Qe$2 = 1e3;
	et$2 = 1.2;
	tt$2 = -9999;
	v = (e) => typeof e == `object` && !!e && `nodeType` in e && e.nodeType === Node.ELEMENT_NODE;
	y = (e) => v(e) && e.namespaceURI === `http://www.w3.org/1999/xhtml`;
	b = (e) => y(e) && e.tagName === `IFRAME`;
	rt$2 = /* @__PURE__ */ new WeakMap();
	x = (e) => {
		let t = Number.parseFloat(e);
		return Number.isFinite(t) ? t : 0;
	};
	it$2 = (e) => {
		let t = performance.now(), n = rt$2.get(e);
		if (n && t - n.timestamp < 16) return n.metrics;
		let r = e.ownerDocument.defaultView?.getComputedStyle(e), i;
		if (!r) i = {
			contentOffsetX: e.clientLeft,
			contentOffsetY: e.clientTop,
			height: e.offsetHeight,
			width: e.offsetWidth
		};
		else {
			let t = x(r.borderLeftWidth), n = x(r.borderRightWidth), a = x(r.borderTopWidth), o = x(r.borderBottomWidth), s = x(r.paddingLeft), c = x(r.paddingRight), l = x(r.paddingTop), u = x(r.paddingBottom), d = x(r.width), f = x(r.height), p = t + n + s + c, m = a + o + l + u, h = r.boxSizing === `border-box` ? d : d + p, g = r.boxSizing === `border-box` ? f : f + m;
			i = {
				contentOffsetX: t + s,
				contentOffsetY: a + l,
				height: g > 0 ? g : e.offsetHeight,
				width: h > 0 ? h : e.offsetWidth
			};
		}
		return rt$2.set(e, {
			metrics: i,
			timestamp: t
		}), i;
	};
	at$2 = (e, t) => t > 0 ? e / t : 1;
	ot$2 = (e) => {
		if (!e) return null;
		try {
			return e.frameElement;
		} catch {
			return null;
		}
	};
	st$2 = (e, t, n) => {
		let r = t, i = n, a = 1, o = 1, s = e;
		for (; s && s !== window;) {
			let e = ot$2(s);
			if (!e || !b(e)) break;
			let t = e.getBoundingClientRect(), n = it$2(e), c = at$2(t.width, n.width), l = at$2(t.height, n.height);
			r = t.left + (n.contentOffsetX + r) * c, i = t.top + (n.contentOffsetY + i) * l, a *= c, o *= l, s = e.ownerDocument.defaultView;
		}
		return {
			x: r,
			y: i,
			scaleX: a,
			scaleY: o
		};
	};
	ct$2 = (e) => {
		let t = e.ownerDocument.defaultView;
		return t ? t.getComputedStyle(e) : window.getComputedStyle(e);
	};
	lt$2 = (e, t) => e.replace(/([\d.]+)px/g, (e, n) => {
		let r = Number(n) * Math.abs(t);
		return `${Number(r.toFixed(3))}px`;
	});
	ut$2 = (e, t, n) => {
		let [r, i] = e.split(`/`).map((e) => e.trim()), a = lt$2(r, t), o = lt$2(i ?? r, n);
		return i === void 0 && a === o ? a : `${a} / ${o}`;
	};
	dt$2 = /* @__PURE__ */ new WeakMap();
	ft$2 = (e, t) => {
		dt$2.set(e, t);
	};
	S = (e) => dt$2.get(e) ?? null;
	C = (e) => S(e)?.getFiber() ?? Xe$3(e);
	pt$2 = /* @__PURE__ */ new WeakMap();
	mt$2 = /* @__PURE__ */ new WeakMap();
	ht$2 = () => {
		pt$2 = /* @__PURE__ */ new WeakMap(), mt$2 = /* @__PURE__ */ new WeakMap();
	};
	gt$2 = (e, t, n) => {
		let r = mt$2.get(e);
		if (r && n - r.timestamp < 200) return r.borderRadius;
		let i = (t ?? ct$2(e)).borderRadius || `0px`;
		return mt$2.set(e, {
			borderRadius: i,
			timestamp: n
		}), i;
	};
	_t$2 = (e) => {
		let t = performance.now(), n = pt$2.get(e);
		if (n && t - n.timestamp < 16) return n.bounds;
		let r = S(e);
		if (r) {
			let n = r.getBounds();
			return pt$2.set(e, {
				bounds: n,
				timestamp: t
			}), n;
		}
		let i = e.getBoundingClientRect(), a = st$2(e.ownerDocument.defaultView, i.left, i.top), o = {
			borderRadius: ut$2(gt$2(e, null, t), a.scaleX, a.scaleY),
			height: i.height * a.scaleY,
			width: i.width * a.scaleX,
			x: a.x,
			y: a.y
		};
		return pt$2.set(e, {
			bounds: o,
			timestamp: t
		}), o;
	};
	w = (e) => {
		let t = e.ownerDocument?.defaultView;
		return !!(t && e instanceof t.ShadowRoot);
	};
	T = class extends Error {
		constructor(e, t) {
			super(e, t), this.name = `ReactGrabError`;
		}
	};
	E = class extends T {
		constructor(e, t) {
			super(e, { cause: t }), this.name = `RecoverableError`;
		}
	};
	yt$2 = class extends E {
		constructor(e, t, n) {
			super(`Failed to open source file "${e}"`, n), this.name = `OpenFileError`, this.filePath = e, this.lineNumber = t;
		}
	};
	bt$2 = class extends E {
		constructor(e, t, n) {
			super(`Plugin hook "${t}" failed for "${e}"`, n), this.name = `PluginHookError`, this.pluginName = e, this.hookName = t;
		}
	};
	xt$2 = class extends E {
		constructor(e, t) {
			super(`Plugin cleanup failed for "${e}"`, t), this.name = `PluginCleanupError`, this.pluginName = e;
		}
	};
	St$2 = class extends E {
		constructor(e, t) {
			super(`Plugin setup failed for "${e}"`, t), this.name = `PluginSetupError`, this.pluginName = e;
		}
	};
	Ct$2 = class extends E {
		constructor(e, t) {
			super(`Action "${e}" failed`, t), this.name = `ContextMenuActionError`, this.actionId = e;
		}
	};
	wt$2 = class extends E {
		constructor(e, t) {
			super(`Action "${e}" enabled check failed`, t), this.name = `ContextMenuActionEnabledError`, this.actionId = e;
		}
	};
	Tt$2 = class extends T {
		constructor() {
			super(`Can't generate CSS selector for non-element node type.`), this.name = `NonElementNodeError`;
		}
	};
	Et$2 = class extends T {
		constructor(e) {
			super(`Timeout: Can't find a unique selector after ${e}ms`), this.name = `SelectorTimeoutError`, this.timeoutMs = e;
		}
	};
	Dt$2 = class extends T {
		constructor() {
			super(`Selector was not found.`), this.name = `SelectorNotFoundError`;
		}
	};
	Ot$2 = class extends T {
		constructor() {
			super(`Failed to copy`), this.name = `CopyFailedError`;
		}
	};
	D = (e) => S(e)?.getTagName() ?? (e.tagName || ``).toLowerCase();
	kt$2 = /* @__PURE__ */ new WeakMap();
	At$2 = (e = document) => {
		let t = kt$2.get(e);
		if (t !== void 0) return t;
		let n = e.querySelector(`script[nonce], style[nonce]`), r = n?.nonce || n?.getAttribute(`nonce`) || null;
		return r && kt$2.set(e, r), r;
	};
	jt$2 = [
		[`data-rr-block`, ``],
		[`data-rr-ignore`, ``],
		[`data-rr-mask`, ``],
		[`data-sentry-block`, ``],
		[`data-sentry-ignore`, ``],
		[`data-sentry-mask`, ``],
		[`data-dd-privacy`, `hidden`],
		[`data-fs-exclude`, ``],
		[`data-lr-exclude`, ``],
		[`data-hj-suppress`, ``],
		[`data-recording-disable`, ``],
		[`data-clarity-mask`, `true`],
		[`data-heap-redact-text`, ``]
	];
	Mt$2 = (e) => {
		for (let [t, n] of jt$2) e.setAttribute(t, n);
		e.setAttribute(`data-testid`, `react-grab-overlay`), e.classList.add(`ph-no-capture`);
	};
	Nt$2 = (e) => typeof e == `object` && !!e && `nodeType` in e && e.nodeType === 9;
	Pt$2 = (e) => {
		if (e.assignedSlot) return e.assignedSlot;
		if (e.parentElement) return e.parentElement;
		let t = e.getRootNode();
		return w(t) ? t.host : Nt$2(t) ? ot$2(t.defaultView) : null;
	};
	Ft$2 = null;
	It$2 = (e) => {
		Ft$2 = e;
	};
	Lt$2 = () => Ft$2;
	Rt$2 = (e) => {
		if (!Ft$2) return !0;
		let t = e;
		for (; t;) {
			if (t === Ft$2) return !0;
			t = Pt$2(t);
		}
		return !1;
	};
	zt$2 = (e) => e;
	Bt$2 = (e) => e === void 0 || !Number.isFinite(e) ? 3 : Math.max(0, Math.floor(e));
	Vt$2 = /^(?:\.\/)?\/?\([a-z][a-z0-9-]*\)\//;
	O = (e) => {
		let t = $$3(e);
		return t = t.replace(Vt$2, ``), t.startsWith(`./`) && (t = t.slice(2)), t;
	};
	Ht$2 = (e) => {
		try {
			return decodeURIComponent(e);
		} catch {
			return e;
		}
	};
	Ut$2 = /(?:^|[/\\])node_modules[/\\]/;
	Wt$2 = /[/\\]\.vite[/\\]deps[^/\\]*[/\\]/;
	Gt$2 = /\.[mc]?[jt]sx?$/i;
	Kt$2 = /^chunk-[A-Za-z0-9_-]+$/;
	qt$2 = /[/\\]/;
	Jt$2 = /^(.+?)@v?\d/;
	Yt$2 = (e) => e.split(qt$2).filter(Boolean);
	Xt$2 = (e) => {
		let [t, n] = Yt$2(e);
		return !t || t.startsWith(`.`) ? null : t.startsWith(`@`) ? n ? `${t}/${n}` : null : t;
	};
	Zt$2 = (e) => {
		let t = Yt$2(e)[0];
		if (!t) return null;
		let n = t.replace(Gt$2, ``);
		if (Kt$2.test(n)) return null;
		if (!n.startsWith(`@`)) return n;
		let r = n.indexOf(`_`);
		return r === -1 ? null : `${n.slice(0, r)}/${n.slice(r + 1)}`;
	};
	Qt$2 = (e, t, n) => {
		let r = e.split(t);
		return r.length > 1 ? n(r[r.length - 1]) : null;
	};
	$t$2 = (e) => e?.match(Jt$2)?.[1] ?? null;
	en$2 = (e) => {
		let t;
		try {
			t = new URL(e);
		} catch {
			return null;
		}
		if (!t.hostname) return null;
		let n = Yt$2(t.pathname).map(Ht$2);
		for (let [e, t] of n.entries()) {
			if (t.startsWith(`@`)) {
				let r = $t$2(n[e + 1]);
				if (r) return `${t}/${r}`;
				continue;
			}
			let r = $t$2(t);
			if (r) return r;
		}
		return null;
	};
	tn$2 = (e) => Qt$2(e, Wt$2, Zt$2) ?? Qt$2(e, Ut$2, Xt$2);
	nn$2 = (e) => {
		if (!e) return null;
		let t = $$3(e);
		if (!t) return null;
		let n = Ht$2(t);
		return tn$2(n) || en$2(e) || null;
	};
	rn$2 = /^@[A-Za-z0-9][A-Za-z0-9._-]*$/;
	an$2 = /^[A-Za-z0-9][A-Za-z0-9._-]*$/;
	on$2 = /* @__PURE__ */ new Set([
		`app`,
		`web`,
		`website`,
		`frontend`,
		`client`,
		`src`
	]);
	sn$2 = /* @__PURE__ */ new Set([
		`app`,
		`src`,
		`components`,
		`pages`,
		`features`,
		`modules`,
		`hooks`,
		`lib`,
		`utils`,
		`ui`,
		`shared`,
		`common`,
		`core`,
		`styles`,
		`assets`
	]);
	cn$2 = (e) => {
		let t = e;
		for (; t.startsWith(`../`) || t.startsWith(`./`);) t = t.slice(t.startsWith(`../`) ? 3 : 2);
		return t;
	};
	ln$2 = (e) => {
		let t = cn$2(Ht$2($$3(e)));
		if (t.startsWith(`/`)) return null;
		let [n, r, ...i] = Yt$2(t);
		return !n || !r || i.length === 0 || !rn$2.test(n) || sn$2.has(n.slice(1)) || !an$2.test(r) || Gt$2.test(r) || on$2.has(r) ? null : `${n}/${r}`;
	};
	un$2 = (e) => e ? nn$2(e) ?? ln$2(e) : null;
	dn$2 = (e) => {
		if (!e) return {
			origin: `unknown`,
			packageName: null
		};
		let t = un$2(e);
		return t ? {
			origin: `package`,
			packageName: t
		} : je$1(e) ? {
			origin: `app`,
			packageName: null
		} : {
			origin: `unknown`,
			packageName: null
		};
	};
	fn$2 = /* @__PURE__ */ new Set([
		`role`,
		`name`,
		`aria-label`,
		`rel`,
		`href`
	]);
	k = (e) => {
		if (!/^[a-z-]{3,}$/i.test(e)) return !1;
		let t = e.split(/-|[A-Z]/);
		for (let e of t) if (e.length <= 2 || /[^aeiou]{4,}/i.test(e)) return !1;
		return !0;
	};
	pn$2 = (e, t) => {
		let n = fn$2.has(e) || e.startsWith(`data-`) && k(e), r = k(t) && t.length < 100 || t.startsWith(`#`) && k(t.slice(1));
		return n && r;
	};
	mn$2 = (e) => {
		let t = e[0].name;
		for (let n = 1; n < e.length; n++) t = `${e[n].name} > ${t}`;
		return t;
	};
	hn$2 = (e) => {
		let t = 0;
		for (let n of e) t += n.penalty;
		return t;
	};
	gn$2 = (e, t) => hn$2(e) - hn$2(t);
	_n$2 = (e, t) => {
		let n = e.parentNode;
		if (!n) return;
		let r = n.firstChild;
		if (!r) return;
		let i = 0;
		for (; r && (v(r) && (t === void 0 || r.tagName.toLowerCase() === t) && i++, r !== e);) r = r.nextSibling;
		return i;
	};
	vn$2 = (e, t) => e === `html` ? `html` : `${e}:nth-child(${t})`;
	yn$2 = (e, t) => e === `html` ? `html` : `${e}:nth-of-type(${t})`;
	bn$2 = (e, t) => {
		let n = [], r = e.getAttribute(`id`), i = e.tagName.toLowerCase();
		r && k(r) && n.push({
			name: `#${CSS.escape(r)}`,
			penalty: 0
		});
		for (let t of e.classList) k(t) && n.push({
			name: `.${CSS.escape(t)}`,
			penalty: 1
		});
		for (let r of e.attributes) t(r.name, r.value) && n.push({
			name: `[${CSS.escape(r.name)}="${CSS.escape(r.value)}"]`,
			penalty: 2
		});
		n.push({
			name: i,
			penalty: 5
		});
		let a = _n$2(e, i);
		a !== void 0 && n.push({
			name: yn$2(i, a),
			penalty: 10
		});
		let o = _n$2(e);
		return o !== void 0 && n.push({
			name: vn$2(i, o),
			penalty: 50
		}), n;
	};
	xn$2 = (e, t = 1e4, n = []) => {
		if (t <= 0) return [];
		if (e.length === 0) return [n];
		let r = [];
		for (let i of e[0]) {
			let a = t - r.length;
			if (a <= 0) break;
			r.push(...xn$2(e.slice(1), a, [...n, i]));
		}
		return r;
	};
	Sn$2 = (e, t) => {
		let n = t.getRootNode();
		return w(n) ? n : Nt$2(e) ? e : e.ownerDocument;
	};
	Cn$2 = (e, t) => t.querySelectorAll(mn$2(e)).length === 1;
	wn$2 = (e, t) => {
		let n = e, r = [];
		for (; n && n !== t;) {
			let e = n.tagName.toLowerCase(), t = _n$2(n, e);
			if (t === void 0) return;
			r.push({
				name: yn$2(e, t),
				penalty: 10
			}), n = n.parentElement;
		}
		return Cn$2(r, t) ? r : void 0;
	};
	Tn$2 = (e, t, n, r) => {
		if (e.nodeType !== Node.ELEMENT_NODE) throw new Tt$2();
		if (e.tagName.toLowerCase() === `html`) return `html`;
		let i = Sn$2(t, e), a = Date.now(), o = [], s = e, c = 0, l;
		for (; s && s !== i && !l;) if (o.push(bn$2(s, r)), s = s.parentElement, c++, c >= 3) {
			let t = xn$2(o);
			t.sort(gn$2);
			for (let r of t) {
				if (Date.now() - a > n) {
					let t = wn$2(e, i);
					if (!t) throw new Et$2(n);
					return mn$2(t);
				}
				if (Cn$2(r, i)) {
					l = r;
					break;
				}
			}
		}
		if (!l && c < 3) {
			let e = xn$2(o);
			e.sort(gn$2);
			for (let t of e) {
				if (Date.now() - a > n) break;
				if (Cn$2(t, i)) {
					l = t;
					break;
				}
			}
		}
		if (!l) throw new Dt$2();
		return mn$2(l);
	};
	En$2 = (e) => e.ownerDocument.body ?? e.ownerDocument.documentElement;
	Dn$2 = /* @__PURE__ */ new Set([
		`data-testid`,
		`data-test-id`,
		`data-test`,
		`data-cy`,
		`data-qa`,
		`aria-label`,
		`href`,
		`src`,
		`role`,
		`name`,
		`title`,
		`alt`
	]);
	On$2 = (e) => e.length > 0 && e.length <= 120;
	kn$2 = (e, t) => {
		try {
			let n = e.getRootNode(), r = (w(n) ? n : e.ownerDocument).querySelectorAll(t);
			return r.length === 1 && r[0] === e;
		} catch {
			return !1;
		}
	};
	An$2 = (e) => {
		let t = e.getAttribute(`id`);
		if (t) {
			let n = `#${CSS.escape(t)}`;
			if (kn$2(e, n)) return n;
		}
		for (let t of Dn$2) {
			let n = e.getAttribute(t);
			if (!n || !On$2(n)) continue;
			let r = `[${t}=${JSON.stringify(n)}]`;
			if (kn$2(e, r)) return r;
			let i = `${e.tagName.toLowerCase()}${r}`;
			if (kn$2(e, i)) return i;
		}
		return null;
	};
	jn$2 = (e) => {
		let t = [], n = e.getRootNode(), r = w(n) ? n : En$2(e), i = e;
		for (; i;) {
			let e = i.getAttribute(`id`);
			if (e) {
				t.unshift(`#${CSS.escape(e)}`);
				break;
			}
			let n = i.parentNode;
			if (!n) {
				t.unshift(i.tagName.toLowerCase());
				break;
			}
			let a = Array.from(n.children).indexOf(i) + 1;
			if (t.unshift(`${i.tagName.toLowerCase()}:nth-child(${a})`), n === r) {
				v(r) && t.unshift(r.tagName.toLowerCase());
				break;
			}
			i = v(n) ? n : null;
		}
		return t.join(` > `);
	};
	Mn$2 = (e) => {
		let t = An$2(e);
		if (t) return t;
		try {
			let t = Tn$2(e, En$2(e), 200, (e, t) => pn$2(e, t) || Dn$2.has(e) && On$2(t));
			if (t) return t;
		} catch {}
		return jn$2(e);
	};
	Nn$2 = (e) => {
		let t = S(e);
		if (t) return t.getSelector();
		let n = Mn$2(e), r = e.getRootNode();
		if (w(r)) return `${Nn$2(r.host)} >>> ${n}`;
		let i = ot$2(e.ownerDocument.defaultView);
		return i ? `${Nn$2(i)} >>iframe>> ${n}` : n;
	};
	Pn$2 = (e) => {
		let { body: t, documentElement: n } = e.ownerDocument;
		if (e === t || e === n) return !0;
		if (!t) return !1;
		let r = t.getElementsByTagName(`*`).length;
		return r !== 0 && e.getElementsByTagName(`*`).length / r >= .5;
	};
	Fn$2 = (e) => {
		let t = e.closest(`[id],[data-testid],[data-test-id],[data-test],[data-cy],[data-qa],[aria-label],a[href],button,input,select,textarea,[role="button"],[role="link"],[role="checkbox"],[role="radio"],[role="switch"],[role="tab"],[role="menuitem"],[role="option"],[role="textbox"],[role="combobox"],[role="slider"],[role="spinbutton"]`);
		return !t || Pn$2(t) ? e : t;
	};
	In$2 = [
		/\/assets\/[^/?#]+-[a-z0-9_-]{6,}\.(?:c|m)?js(?:[?#]|$)/,
		/\/_next\/static\/.*\.(?:c|m)?js(?:[?#]|$)/,
		/\/static\/chunks\/.*\.(?:c|m)?js(?:[?#]|$)/
	];
	Ln$2 = (e) => {
		if (!e) return !1;
		let t = `/${O(e)}`.toLowerCase();
		return In$2.some((e) => e.test(t));
	};
	Rn$2 = (e) => {
		if (!e) return !1;
		let t = `/${O(e)}/`.toLowerCase();
		return se.some((e) => t.includes(e));
	};
	Bn$2 = (e) => (e && (zn$2 = void 0), zn$2 ??= typeof document < `u` && !!(document.getElementById(`__NEXT_DATA__`) || document.querySelector(`nextjs-portal`)), zn$2);
	Vn$2 = (e) => e.map((e) => `\n  in ${e}`).join(``);
	Un$2 = () => {
		if (Hn$2 !== void 0) return Hn$2;
		let e = document.querySelector(`script[src*="/_next/"]`)?.src, t = e ? new URL(e).pathname : ``, n = t.indexOf(`/_next/`);
		return Hn$2 = n > 0 ? t.slice(0, n) : ``, Hn$2;
	};
	Wn$2 = [`about://React/`, `rsc://React/`];
	Gn$2 = (e) => Wn$2.some((t) => e.startsWith(t));
	Kn$2 = (e) => {
		for (let t of Wn$2) {
			if (!e.startsWith(t)) continue;
			let n = e.indexOf(`/`, t.length);
			if (n === -1) continue;
			let r = n + 1, i = e.lastIndexOf(`?`), a = i > r ? e.slice(r, i) : e.slice(r);
			return Ht$2(a);
		}
		return e;
	};
	qn$2 = (e) => {
		if (typeof e != `object` || !e || !(`status` in e) || e.status !== `fulfilled` || !(`value` in e) || typeof e.value != `object` || e.value === null || !(`originalStackFrame` in e.value)) return null;
		let t = e.value.originalStackFrame;
		return typeof t != `object` || !t || !(`file` in t) || typeof t.file != `string` || !t.file || `ignored` in t && t.ignored ? null : {
			file: t.file,
			line1: `line1` in t && typeof t.line1 == `number` ? t.line1 : null,
			column1: `column1` in t && typeof t.column1 == `number` ? t.column1 : null
		};
	};
	Jn$2 = async (e, t) => {
		let n = [], r = [];
		for (let t = 0; t < e.length; t++) {
			let i = e[t];
			i.isServer && i.fileName && (n.push(t), r.push({
				file: Kn$2(i.fileName),
				methodName: i.functionName ?? `<unknown>`,
				line1: i.lineNumber ?? null,
				column1: i.columnNumber ?? null,
				arguments: []
			}));
		}
		if (r.length === 0) return e;
		let i = new AbortController(), a = setTimeout(() => i.abort(), 5e3), o = () => i.abort();
		t?.aborted && i.abort(), t?.addEventListener(`abort`, o);
		try {
			let t = await fetch(`${Un$2()}/__nextjs_original-stack-frames`, {
				method: `POST`,
				headers: { "Content-Type": `application/json` },
				body: JSON.stringify({
					frames: r,
					isServer: !0,
					isEdgeServer: !1,
					isAppDirectory: !0
				}),
				priority: `high`,
				signal: i.signal
			});
			if (!t.ok) return e;
			let a = await t.json();
			if (!Array.isArray(a)) return e;
			let o = [...e];
			for (let t = 0; t < n.length; t++) {
				let r = qn$2(a[t]);
				if (!r) continue;
				let i = n[t];
				o[i] = {
					...e[i],
					fileName: r.file,
					lineNumber: r.line1 ?? void 0,
					columnNumber: r.column1 ?? void 0,
					isSymbolicated: !0
				};
			}
			return o;
		} catch {
			return e;
		} finally {
			clearTimeout(a), t?.removeEventListener(`abort`, o);
		}
	};
	Yn$2 = (e) => {
		let t = /* @__PURE__ */ new Map();
		return w$3(e, (e) => {
			if (!R$3(e)) return !1;
			let n = q$2(e._debugStack.stack);
			if (!n) return !1;
			for (let e of m(n)) e.functionName && e.fileName && Gn$2(e.fileName) && (t.has(e.functionName) || t.set(e.functionName, {
				...e,
				isServer: !0
			}));
			return !1;
		}, !0), t;
	};
	Xn$2 = (e, t) => {
		if (!t.some((e) => e.isServer && !e.fileName && e.functionName)) return t;
		let n = Yn$2(e);
		return n.size === 0 ? t : t.map((e) => {
			if (!e.isServer || e.fileName || !e.functionName) return e;
			let t = n.get(e.functionName);
			return t ? {
				...e,
				fileName: t.fileName,
				lineNumber: t.lineNumber,
				columnNumber: t.columnNumber
			} : e;
		});
	};
	Zn$2 = 0;
	Qn$2 = [];
	$n$2 = () => Zn$2 < 3 ? (Zn$2 += 1, Promise.resolve()) : new Promise((e) => {
		Qn$2.push(e);
	});
	er$2 = () => {
		let e = Qn$2.shift();
		e ? e() : --Zn$2;
	};
	tr$2 = async (e, t, n = 8e3) => {
		await $n$2();
		let r = new AbortController(), i, a = new Promise((e) => {
			i = setTimeout(() => {
				r.abort(), e(t);
			}, n);
		});
		try {
			let t = e(r.signal);
			return t.catch(() => {}), await Promise.race([t, a]);
		} finally {
			clearTimeout(i), er$2();
		}
	};
	nr$2 = (e, t) => e.length > t ? `${e.slice(0, t)}...` : e;
	rr$2 = (e) => e.startsWith(`data-react-grab-`);
	ir$2 = (e) => e.replace(/\s+/g, ` `).trim();
	ar$2 = (e) => {
		let t = [];
		for (let n of e.childNodes) {
			if (n.nodeType !== Node.TEXT_NODE) continue;
			let e = ir$2(n.textContent ?? ``);
			e && t.push(e);
		}
		return t.join(` `);
	};
	or$2 = (e) => e.getAttribute(`aria-hidden`) === `true` || e.hasAttribute(`hidden`) ? !0 : Ce.has(e.tagName.toLowerCase());
	sr$2 = (e, t, n) => {
		if (e.nodeType === Node.TEXT_NODE) {
			let r = ir$2(e.textContent ?? ``);
			return r ? (t.push(r), n - r.length) : n;
		}
		if (!v(e) || or$2(e)) return n;
		for (let r of e.childNodes) if (n = sr$2(r, t, n), n <= 0) break;
		return n;
	};
	cr$2 = (e, t) => {
		if (or$2(e)) return ``;
		let n = ar$2(e);
		if (!Se.has(t) || n && e.children.length === 0) return n;
		let r = [];
		return sr$2(e, r, 100), r.join(` `);
	};
	lr$2 = (e) => nr$2(e, 15);
	ur$2 = (e) => {
		let t = [];
		for (let n of be) {
			let r = e.getAttribute(n);
			r && t.push(`${n}="${r}"`);
		}
		return t.length > 0 ? ` ${t.join(` `)}` : ``;
	};
	dr$2 = (e) => e === `class` || e === `className` || e === `style`;
	fr$2 = (e) => {
		let t = [], n = [], r = ``;
		for (let { name: i, value: a } of e.attributes) rr$2(i) || (dr$2(i) ? i !== `style` && a && (r = ` class="${lr$2(a)}"`) : xe.has(i) ? t.push(a ? ` ${i}="${a}"` : ` ${i}`) : a && n.push(` ${i}="${lr$2(a)}"`));
		return t.join(``) + n.join(``) + r;
	};
	mr$2 = (e) => {
		let t = S(e);
		if (t) return t.getPreview();
		let n = D(e);
		if (!y(e)) return `<${n}${ur$2(e)} />`;
		let r = fr$2(e), i = cr$2(e, n), a = nr$2(i, 100);
		return a ? `<${n}${r}>${a}</${n}>` : `<${n}${r} />`;
	};
	gr$2 = /* @__PURE__ */ new Set([
		`_`,
		`$`,
		`motion.`,
		`styled.`,
		`chakra.`,
		`ark.`,
		`Primitive.`,
		`Slot.`
	]);
	_r$2 = /* @__PURE__ */ new Set(/* @__PURE__ */ `AppRouter.AppRouterAnnouncer.AppDevOverlay.AppDevOverlayErrorBoundary.ClientPageRoot.ClientSegmentRoot.DevRootHTTPAccessFallbackBoundary.ErrorBoundary.ErrorBoundaryHandler.GracefulDegradeBoundary.HTTPAccessErrorFallback.HTTPAccessFallbackBoundary.HTTPAccessFallbackErrorBoundary.HandleRedirect.Head.HistoryUpdater.HotReload.InnerLayoutRouter.InnerScrollAndFocusHandler.InnerScrollAndFocusHandlerOld.InnerScrollAndMaybeFocusHandler.InnerScrollHandlerNew.LoadableComponent.LoadingBoundary.LoadingBoundaryProvider.NotAllowedRootHTTPFallbackError.OfflineProvider.OuterLayoutRouter.RedirectBoundary.RedirectErrorBoundary.RenderFromTemplateContext.RenderValidationBoundaryAtThisLevel.ReplaySsrOnlyErrors.RootErrorBoundary.RootLevelDevOverlayElement.Router.ScrollAndFocusHandler.ScrollAndMaybeFocusHandler.SegmentBoundaryTrigger.SegmentBoundaryTriggerNode.SegmentStateProvider.SegmentTrieNode.SegmentViewNode.SegmentViewStateNode.ServerRoot.body.html`.split(`.`));
	vr$2 = /* @__PURE__ */ new Set([
		`<anonymous>`,
		`<unknown>`,
		`Anonymous`,
		`Unknown`
	]);
	yr$2 = /* @__PURE__ */ new Set([
		`Suspense`,
		`Fragment`,
		`StrictMode`,
		`Profiler`,
		`SuspenseList`
	]);
	br$2 = /* @__PURE__ */ new Set([
		`MotionDOMComponent`,
		`Slot`,
		`SlotClone`
	]);
	xr$2 = [
		`.Consumer`,
		`.Context`,
		`.Provider`,
		`.Slot`,
		`.SlotClone`,
		`.Slottable`,
		`ProviderProvider`
	];
	Sr$2 = (e) => {
		if (vr$2.has(e) || _r$2.has(e) || yr$2.has(e) || br$2.has(e)) return !0;
		for (let t of xr$2) if (e.endsWith(t)) return !0;
		for (let t of gr$2) if (e.startsWith(t)) return !0;
		return !1;
	};
	Cr$2 = (e) => !(!e || Sr$2(e));
	Dr$2 = (e) => !(e.length <= 1 || Sr$2(e) || e[0] !== e[0].toUpperCase());
	A = (e) => e && Dr$2(e) ? e : null;
	Or$2 = (e) => !Rn$2(e) && !Ln$2(e);
	j = (e) => {
		if (!Re$1()) return e;
		let t = e;
		for (; t?.ownerDocument === e.ownerDocument;) {
			if (C(t)) return t;
			if (t.parentElement) {
				t = t.parentElement;
				continue;
			}
			let e = t.getRootNode();
			t = w(e) ? e.host : null;
		}
		return e;
	};
	kr$2 = (e) => {
		let t = e.return?.child ?? null;
		for (; t;) {
			if (t !== e && t.key !== null) return !0;
			t = t.sibling;
		}
		return !1;
	};
	Ar$2 = (e) => {
		let t = e, n = 0;
		for (; t;) {
			if (t.key !== null && kr$2(t)) return String(t.key);
			if (Oe$1(t) && (n += 1, n === 2)) break;
			t = t.return;
		}
		return null;
	};
	jr$2 = (e) => Re$1() ? Ar$2(C(j(e))) : null;
	Mr$2 = /* @__PURE__ */ new WeakMap();
	Nr$2 = /* @__PURE__ */ new WeakMap();
	Pr$2 = (e) => (t) => fetch(t, {
		signal: e,
		priority: `high`
	});
	Fr$2 = (e) => tr$2(async (t) => {
		try {
			let n = C(e);
			if (!n) return null;
			let r = await X$2(n, !0, Pr$2(t));
			if (Bn$2()) {
				let e = Xn$2(n, r);
				return await Jn$2(e, t);
			}
			return r;
		} catch {
			return null;
		}
	}, null);
	Ir$2 = (e) => {
		if (!Re$1()) return Promise.resolve([]);
		let t = j(e), n = Mr$2.get(t);
		if (n) return n;
		let r = Fr$2(t).then((e) => (e === null && Mr$2.delete(t), e));
		return Mr$2.set(t, r), r;
	};
	Lr$2 = async (e) => {
		if (!Re$1()) return null;
		let t = await Ir$2(e);
		if (!t) return null;
		for (let e of t) {
			let t = A(e.functionName);
			if (t) return t;
		}
		return null;
	};
	Rr$2 = (e) => e.find((e) => !!A(e.functionName)) ?? e[0] ?? null;
	zr$2 = (e) => !e || !Oe$1(e) ? null : A(Ie$1(e.type));
	Br$2 = (e) => tr$2(async (t) => {
		let n = C(j(e));
		if (!n) return null;
		try {
			let e = await Oe(n, !0, Pr$2(t));
			return e?.fileName ? {
				filePath: O(e.fileName),
				lineNumber: e.lineNumber ?? null,
				columnNumber: e.columnNumber ?? null,
				componentName: A(e.functionName) ?? zr$2(n._debugOwner),
				origin: dn$2(e.fileName).origin
			} : null;
		} catch {
			return null;
		}
	}, null);
	Vr$2 = (e) => {
		let t = j(e), n = Nr$2.get(t);
		if (n) return n;
		let r = Br$2(t).then((e) => (e || Nr$2.delete(t), e));
		return Nr$2.set(t, r), r;
	};
	Hr$2 = (e, t) => {
		let n = (e, t) => {
			let n = Rr$2(e);
			return n?.fileName ? {
				filePath: O(n.fileName),
				lineNumber: n.lineNumber ?? null,
				columnNumber: n.columnNumber ?? null,
				componentName: A(n.functionName),
				origin: t
			} : null;
		}, r = t.filter((e) => dn$2(e.fileName).origin === `app`), i = r.filter((e) => Or$2(e.fileName));
		return e?.origin === `app` && Or$2(e.filePath) ? e : n(i, `app`) || (e?.origin === `app` && !Ln$2(e.filePath) ? e : n(r, `app`) || (e?.origin === `app` || e?.origin === `package` ? e : n(t.filter((e) => dn$2(e.fileName).origin === `package`), `package`)));
	};
	Ur$2 = async (e) => {
		let t = await Vr$2(e);
		return t?.origin === `app` && Or$2(t.filePath) ? t : Hr$2(t, await Ir$2(e) ?? []);
	};
	Wr$2 = (e) => Gr$2(j(e), 1)[0] ?? null;
	Gr$2 = (e, n, r = () => !0) => {
		if (!Re$1()) return [];
		let i = C(e);
		if (!i) return [];
		let a = [];
		return w$3(i, (e) => {
			if (a.length >= n) return !0;
			if (Oe$1(e)) {
				let n = Ie$1(e.type);
				n && Cr$2(n) && r(n) && a.push(n);
			}
			return !1;
		}, !0), a;
	};
	Kr$2 = [
		`/src/app/`,
		`/src/pages/`,
		`/app/`,
		`/pages/`
	];
	qr$2 = (e, t) => {
		let n = O(e);
		if (!t || !n.startsWith(`/`)) return n;
		for (let e of Kr$2) {
			let t = n.indexOf(e);
			if (t !== -1) return `/./${n.slice(t + 1)}`;
		}
		return n;
	};
	Jr$2 = (e, t) => {
		let n = qr$2(e.filePath, t), r = t && e.lineNumber ? `${n}:${e.lineNumber}${e.columnNumber ? `:${e.columnNumber}` : ``}` : n;
		return e.componentName ? `\n  in ${e.componentName} (at ${r})` : `\n  in ${r}`;
	};
	Yr$2 = {
		isAppSource: !1,
		consumesBudget: !1
	};
	Xr$2 = (e, t, n, r) => {
		let i = t.packageName, a = t.origin === `app` ? e.fileName : null;
		if (e.isServer && !a && (n || !e.functionName)) {
			let e = i ? `${i} at Server` : `at Server`;
			return {
				text: `\n  in ${n ?? `<anonymous>`} (${e})`,
				...Yr$2
			};
		}
		return !a && n ? {
			text: i ? `\n  in ${n} (${i})` : `\n  in ${n}`,
			...Yr$2
		} : i ? {
			text: `\n  in ${i}`,
			...Yr$2
		} : a ? {
			text: Jr$2({
				componentName: n,
				filePath: a,
				lineNumber: e.lineNumber ?? null,
				columnNumber: e.columnNumber ?? null
			}, r),
			isAppSource: !0,
			consumesBudget: Or$2(a)
		} : null;
	};
	Zr$2 = (e, t = {}, n = null) => {
		let r = Bt$2(t.maxLines), i = Math.max(r, 20), a = Bn$2(), o = [], s = /* @__PURE__ */ new Set(), c = null, l = !1, u = !1, d = !1, f = 0, p = (e) => {
			e && s.add(e);
		};
		if (n) {
			let e = n.origin === `app` && Or$2(n.filePath);
			u = e, e && (f += 1), p(n.componentName), o.push(Jr$2(n, a));
		}
		for (let t of e) {
			if (f >= r || o.length >= i) break;
			let e = dn$2(t.fileName), s = A(t.functionName), m = e.packageName ? `${e.packageName}:${s ?? ``}:${t.isServer ? `server` : `client`}` : null;
			if (m && m === c) continue;
			if (!l && s && s === n?.componentName) {
				l = !0;
				continue;
			}
			let h = Xr$2(t, e, s, a);
			h !== null && h.text !== o[o.length - 1] && (h.isAppSource && h.consumesBudget && (u = !0), h.consumesBudget && (f += 1, d = !0), p(s), o.push(h.text), c = m);
		}
		return {
			text: o.join(``),
			shouldAppendSelectorHint: !u,
			hasBudgetedStackFrame: d,
			renderedComponentNames: s
		};
	};
	Qr$2 = (e, t) => {
		let n = Hr$2(e, t);
		return n?.origin === `app` ? n : null;
	};
	$r$2 = (e, t, n) => {
		let r = Gr$2(j(e), n, (e) => Dr$2(e) && !t.renderedComponentNames.has(e));
		return r.length === 0 ? t : {
			...t,
			text: `${t.text}${Vn$2(r)}`
		};
	};
	ei$2 = async (e, t = {}) => {
		let n = await Vr$2(e), r = await Ir$2(e) ?? [], i = Qr$2(n, r), a = Bt$2(t.maxLines), o = Zr$2(r, t, i);
		if (o.text) return o.hasBudgetedStackFrame ? o : $r$2(e, o, a);
		let s = Gr$2(j(e), a);
		return {
			text: Vn$2(s),
			shouldAppendSelectorHint: !0,
			hasBudgetedStackFrame: !1,
			renderedComponentNames: new Set(s)
		};
	};
	ti$2 = async (e, t = {}) => {
		let n = Er$2(Tr$2(e));
		return [(await ei$2(e, t)).text, n].filter(Boolean).join(`

`);
	};
	ni$2 = (e, t) => {
		let n = jr$2(e), r = n === null ? `` : `\n  key: "${n}"`, i = t.shouldAppendSelectorHint ? `\n  selector: ${Nn$2(Fn$2(e))}` : ``;
		return `${t.text}${r}${i}`;
	};
	ri$2 = async (e, t = {}) => {
		let n = Er$2(Tr$2(e)), r = await ei$2(e, t);
		return [`${mr$2(e)}${ni$2(e, r).replace(/\n\s+/g, ` `)}`, n].filter(Boolean).join(`

`);
	};
	ai$2 = typeof window < `u`;
	oi$2 = (e) => 0;
	si$2 = (e) => {};
	M$1 = ai$2 ? (Object.getOwnPropertyDescriptor(Window.prototype, `requestAnimationFrame`)?.value ?? window.requestAnimationFrame).bind(window) : oi$2;
	ci$2 = ai$2 ? (Object.getOwnPropertyDescriptor(Window.prototype, `cancelAnimationFrame`)?.value ?? window.cancelAnimationFrame).bind(window) : si$2;
	li$2 = () => ai$2 ? new Promise((e) => M$1(() => e())) : Promise.resolve();
	ui$2 = (e) => e.replace(/&/g, `&amp;`).replace(/</g, `&lt;`).replace(/>/g, `&gt;`).replace(/"/g, `&quot;`);
	di$2 = (e, t) => {
		let n = t?.componentName ?? `div`, r = {
			version: `0.1.50`,
			content: e,
			entries: t?.entries ?? [{
				tagName: t?.tagName,
				componentName: n,
				content: e,
				commentText: t?.commentText
			}],
			timestamp: Date.now()
		}, i = (t) => {
			t.preventDefault(), t.clipboardData?.setData(`text/plain`, e), t.clipboardData?.setData(`text/html`, `<meta charset='utf-8'><pre><code>${ui$2(e)}</code></pre>`), t.clipboardData?.setData(`application/x-react-grab`, JSON.stringify(r));
		}, a = document.createElement(`textarea`);
		document.addEventListener(`copy`, i);
		try {
			return a.value = e, a.style.position = `fixed`, a.style.left = `-9999px`, a.ariaHidden = `true`, document.body.appendChild(a), a.select(), typeof document.execCommand == `function` && document.execCommand(`copy`);
		} finally {
			document.removeEventListener(`copy`, i), a.remove();
		}
	};
	N$1 = (e) => {
		try {
			return e.contentDocument;
		} catch {
			return null;
		}
	};
	fi$2 = (e, t, n) => {
		let r = e.getBoundingClientRect(), i = it$2(e), a = at$2(r.width, i.width), o = at$2(r.height, i.height);
		return {
			x: (t - r.left) / a - i.contentOffsetX,
			y: (n - r.top) / o - i.contentOffsetY
		};
	};
	pi$2 = (e, t, n) => {
		let r = e.elementFromPoint(t, n);
		for (; r;) {
			let e = r.shadowRoot?.elementFromPoint(t, n);
			if (e && e !== r) r = e;
			else {
				if (b(r)) {
					let e = N$1(r);
					if (!e) return r;
					let i = fi$2(r, t, n);
					return pi$2(e, i.x, i.y) ?? r;
				}
				return r;
			}
		}
		return null;
	};
	mi$2 = (e, t) => pi$2(document, e, t);
	hi$2 = (e) => {
		let t = e;
		for (; t;) {
			if (t.hasAttribute(`data-react-grab-ignore`)) return !0;
			t = Pt$2(t);
		}
		return !1;
	};
	gi$2 = typeof Element < `u` && typeof Element.prototype.checkVisibility == `function`;
	_i$2 = {
		checkOpacity: !0,
		checkVisibilityCSS: !0,
		opacityProperty: !0,
		visibilityProperty: !0
	};
	vi$2 = {
		checkVisibilityCSS: !0,
		visibilityProperty: !0
	};
	yi$2 = (e, t) => {
		if (gi$2 && !t) return e.checkVisibility(_i$2) ? !0 : e.checkVisibility(vi$2) ? ct$2(e).opacity !== `0` : !1;
		let n = t ?? ct$2(e);
		return n.display !== `none` && n.visibility !== `hidden` && n.opacity !== `0`;
	};
	bi$2 = (e) => {
		let t = D(e);
		return t === `html` || t === `body`;
	};
	xi$2 = [`data-react-grab`, `data-react-grab-demo`];
	Si$2 = (e) => xi$2.some((t) => e.hasAttribute(t));
	Ci$2 = (e) => {
		if (Si$2(e)) return !0;
		let t = e.getRootNode();
		return w(t) && Si$2(t.host);
	};
	wi$2 = (e) => {
		let t = parseInt(e.zIndex, 10);
		return e.pointerEvents === `none` && e.position === `fixed` && !isNaN(t) && t >= 2147483600;
	};
	Ti$2 = (e) => {
		let t = e.backgroundColor;
		return t === `transparent` || t === `rgba(0, 0, 0, 0)`;
	};
	Ei$2 = (e) => {
		let t = e.position;
		if (t !== `fixed` && t !== `absolute`) return !1;
		if (Ti$2(e) || parseFloat(e.opacity) < .1) return !0;
		let n = parseInt(e.zIndex, 10);
		return !isNaN(n) && n > 1e3;
	};
	Di$2 = /* @__PURE__ */ new WeakMap();
	Oi$2 = () => {
		Di$2 = /* @__PURE__ */ new WeakMap();
	};
	ki$2 = (e) => {
		let t = S(e);
		if (t) return t.isConnected();
		if (bi$2(e) || Ci$2(e) || hi$2(e)) return !1;
		let n = performance.now(), r = Di$2.get(e);
		if (r && n - r.timestamp < 50) return r.isVisible;
		if (!yi$2(e)) return Di$2.set(e, {
			isVisible: !1,
			timestamp: n
		}), !1;
		if (e.clientWidth / (e.ownerDocument.defaultView?.innerWidth ?? window.innerWidth) >= .9 && e.clientHeight / (e.ownerDocument.defaultView?.innerHeight ?? window.innerHeight) >= .9) {
			let t = ct$2(e);
			if (wi$2(t) || Ei$2(t)) return !1;
		}
		return Di$2.set(e, {
			isVisible: !0,
			timestamp: n
		}), !0;
	};
	Ai$2 = (e, t, n) => {
		let r = !0;
		for (let i of e.elementsFromPoint(t, n)) {
			if (hi$2(i) && (r = !1), !Rt$2(i)) continue;
			let a = i.shadowRoot;
			if (r && a && a !== e) {
				let e = Ai$2(a, t, n);
				if (e) return e;
			}
			if (r && b(i)) {
				let e = N$1(i);
				if (e) {
					let r = fi$2(i, t, n), a = Ai$2(e, r.x, r.y);
					if (a) return a;
				}
			}
			if (ki$2(i)) return i;
		}
		return null;
	};
	ji$2 = (e, t) => Ai$2(document, e, t);
	Mi$2 = (e, t, n, r) => {
		let i = e.elementsFromPoint(t, n), a = !0;
		for (let o of i) {
			if (hi$2(o) && (a = !1), a && o.shadowRoot && o.shadowRoot !== e && Mi$2(o.shadowRoot, t, n, r), a && b(o)) {
				let e = N$1(o);
				if (e) {
					let i = fi$2(o, t, n);
					Mi$2(e, i.x, i.y, r);
				}
			}
			r.add(o);
		}
	};
	Ni$2 = (e, t) => {
		let n = /* @__PURE__ */ new Set();
		return Mi$2(document, e, t, n), [...n];
	};
	Pi$2 = (e, t, n) => e >= n.left && e <= n.right && t >= n.top && t <= n.bottom;
	Fi$2 = (e, t, n = document) => {
		let r = n.createElement(`style`);
		r.setAttribute(e, ``);
		let i = At$2(n);
		return i && (r.nonce = i), Mt$2(r), r.textContent = t, n.head.appendChild(r), r;
	};
	Ii$2 = `html { pointer-events: none !important; }
iframe[${De}] { pointer-events: auto !important; }`;
	Li$2 = /* @__PURE__ */ new Set();
	P$1 = /* @__PURE__ */ new Map();
	F$1 = !1;
	Ri$2 = (e) => {
		if (P$1.has(e)) return;
		let t = Fi$2(`data-react-grab-frozen-pseudo`, Ii$2, e), n = Fi$2(`data-react-grab-hittest-override`, `html, body { pointer-events: auto !important; }`, e);
		n.disabled = !0, P$1.set(e, {
			pointerEventsStyle: t,
			hitTestOverrideStyle: n
		});
	};
	zi$2 = (e) => {
		let t = P$1.get(e);
		t && (t.pointerEventsStyle.remove(), t.hitTestOverrideStyle.remove(), P$1.delete(e));
	};
	Bi$2 = () => F$1;
	Vi$2 = (e) => (Li$2.add(e), F$1 && Ri$2(e), () => {
		Li$2.delete(e), zi$2(e);
	});
	Hi$2 = () => {
		if (!F$1) {
			F$1 = !0, Li$2.add(document);
			for (let e of Li$2) Ri$2(e);
		}
	};
	Ui$2 = () => {
		if (F$1) {
			F$1 = !1;
			for (let e of [...P$1.keys()]) zi$2(e);
		}
	};
	Wi$2 = () => {
		if (F$1) for (let e of P$1.values()) e.pointerEventsStyle.disabled || (e.pointerEventsStyle.disabled = !0), e.hitTestOverrideStyle.disabled && (e.hitTestOverrideStyle.disabled = !1);
	};
	Gi$2 = () => {
		if (F$1) for (let e of P$1.values()) e.pointerEventsStyle.disabled && (e.pointerEventsStyle.disabled = !1), e.hitTestOverrideStyle.disabled || (e.hitTestOverrideStyle.disabled = !0);
	};
	Ki$2 = /* @__PURE__ */ new WeakMap();
	I$1 = /* @__PURE__ */ new WeakMap();
	qi$2 = [
		`name`,
		`position`,
		`rotation`,
		`scale`,
		`color`,
		`visible`,
		`args`
	];
	L$1 = (e) => typeof e == `object` && !!e;
	Ji$2 = (e) => typeof e == `object` && !!e || typeof e == `function`;
	R$1 = (e, t) => typeof e[t] == `function`;
	Yi$2 = (e) => L$1(e) && typeof e.x == `number` && typeof e.y == `number` && typeof e.z == `number` && R$1(e, `set`) && R$1(e, `clone`) && R$1(e, `applyMatrix4`) && R$1(e, `project`);
	Xi$2 = (e) => L$1(e) && R$1(e, `clone`) && R$1(e, `premultiply`);
	Zi$2 = (e) => L$1(e) && `boundingBox` in e && R$1(e, `computeBoundingBox`);
	Qi$2 = (e) => L$1(e) && typeof e.type == `string` && L$1(e.props) && $i$2(e.object);
	$i$2 = (e) => L$1(e) && e.isObject3D === !0 && typeof e.uuid == `string` && typeof e.name == `string` && typeof e.type == `string` && typeof e.visible == `boolean` && Xi$2(e.matrixWorld) && R$1(e, `updateWorldMatrix`);
	ea$2 = (e) => L$1(e) && e.isCamera === !0;
	ta$2 = (e) => $i$2(e) && e.isScene === !0 && Array.isArray(e.children);
	na$2 = (e) => L$1(e) && typeof e.tagName == `string` && e.tagName.toLowerCase() === `canvas` && R$1(e, `getContext`);
	ra$2 = (e) => L$1(e) && na$2(e.domElement);
	ia$2 = (e) => L$1(e) && R$1(e, `set`);
	aa$2 = (e) => L$1(e) && R$1(e, `setFromCamera`) && R$1(e, `intersectObjects`);
	oa$2 = (e) => L$1(e) && ra$2(e.gl) && ta$2(e.scene) && ea$2(e.camera) && aa$2(e.raycaster) && ia$2(e.pointer);
	sa$2 = (e) => Qi$2(e.__r3f) ? e.__r3f : null;
	ca$2 = (e) => {
		let t = e.current.stateNode;
		if (!L$1(t) || !Ji$2(t.containerInfo)) return null;
		let n = Reflect.get(t.containerInfo, `getState`);
		if (typeof n != `function`) return null;
		let r = n();
		return oa$2(r) ? r : null;
	};
	Je$2({
		name: `react-grab-three-selection`,
		onCommitFiberRoot: (e, t) => {
			let n = ca$2(t);
			n && I$1.set(n.gl.domElement, {
				isReactThreeFiber: !0,
				state: n
			});
		}
	});
	la$2 = (e) => {
		let t = e;
		for (; t;) {
			if (sa$2(t)) return t;
			t = t.parent;
		}
		return null;
	};
	ua$2 = (e) => {
		let t = e;
		for (; t;) {
			let e = sa$2(t);
			if (e && typeof e.eventCount == `number` && e.eventCount > 0) return !0;
			t = t.parent;
		}
		return !1;
	};
	da$1 = (e, t) => {
		let n = e;
		for (; n;) {
			if (n === t) return !0;
			n = n.parent;
		}
		return !1;
	};
	fa$1 = (e, t, i) => {
		let a = typeof i.instanceId == `number` && Number.isInteger(i.instanceId) ? i.instanceId : null, o = Ki$2.get(t);
		o || (o = /* @__PURE__ */ new Map(), Ki$2.set(t, o));
		let s = o.get(a);
		if (s) return s.canvas = e.gl.domElement, s.intersectionPoint = Yi$2(i.point) ? i.point : null, s.object = t, s.rootState = e, s.element;
		let c = sa$2(t), l = c ? Xe$3(c) : null;
		if (c && !l) return null;
		let u = (c?.type || t.type || `object-3d`).toLowerCase(), d = e.gl.domElement.ownerDocument.createElement(u), f = {
			canvas: e.gl.domElement,
			element: d,
			fiber: l,
			instance: c,
			instanceId: a,
			intersectionPoint: Yi$2(i.point) ? i.point : null,
			object: t,
			rootState: e,
			tagName: u
		}, p = d.getBoundingClientRect.bind(d);
		return d.getBoundingClientRect = () => {
			let e = xa$1(f), t = d.ownerDocument.defaultView?.DOMRect;
			return t ? new t(e.x, e.y, e.width, e.height) : p();
		}, ft$2(d, {
			hostElement: f.canvas,
			supportsDomEditing: !1,
			getBounds: () => xa$1(f),
			getFiber: () => f.fiber ? ze$1(f.fiber) : null,
			getPreview: () => ga$1(f),
			getSelector: () => _a$1(f),
			getTagName: () => f.tagName,
			isConnected: () => f.canvas.isConnected && da$1(f.object, f.rootState.scene)
		}), o.set(a, f), d;
	};
	pa$1 = (e, t, n) => {
		if (!na$2(e)) return e;
		let r = I$1.get(e);
		if (!r) return e;
		let i = _t$2(e);
		if (i.width <= 0 || i.height <= 0) return e;
		let a = (t - i.x) / i.width * 2 - 1, o = -((n - i.y) / i.height) * 2 + 1;
		if (a < -1 || a > 1 || o < -1 || o > 1) return e;
		try {
			r.state.pointer.set(a, o), r.state.raycaster.setFromCamera(r.state.pointer, r.state.camera);
			let e = r.state.raycaster.intersectObjects(r.state.scene.children, !0);
			if (r.isReactThreeFiber) for (let t of e) {
				if (!ua$2(t.object)) continue;
				let e = la$2(t.object);
				if (!e || e.visible === !1) continue;
				let n = fa$1(r.state, e, t);
				if (n) return n;
			}
			for (let t of e) {
				let e = r.isReactThreeFiber ? la$2(t.object) : t.object;
				if (!e || e.visible === !1) continue;
				let n = fa$1(r.state, e, t);
				if (n) return n;
			}
		} catch {}
		return e;
	};
	ha$1 = (e) => typeof e == `string` ? JSON.stringify(e) : typeof e == `number` || typeof e == `boolean` ? `{${String(e)}}` : Array.isArray(e) && e.length <= 4 && e.every((e) => typeof e == `number` || typeof e == `string`) ? `{${JSON.stringify(e)}}` : null;
	ga$1 = (e) => {
		let t = e.instance?.props ?? {}, n = [];
		for (let r of qi$2) {
			let i = t[r];
			r === `name` && i === void 0 && e.object.name && (i = e.object.name);
			let a = ha$1(i);
			a && n.push(`${r}=${a}`);
		}
		return e.instanceId !== null && n.push(`instanceId={${e.instanceId}}`), `<${e.tagName}${n.length > 0 ? ` ${n.join(` `)}` : ``} />`;
	};
	_a$1 = (e) => {
		if (e.object.name) return `${e.tagName}[name=${JSON.stringify(e.object.name)}]`;
		let t = e.instanceId === null ? `` : `:${e.instanceId}`;
		return `${e.tagName}[uuid=${JSON.stringify(`${e.object.uuid}${t}`)}]`;
	};
	va$1 = (e) => {
		e.object.updateWorldMatrix(!0, !1);
		let t = e.object.matrixWorld.clone();
		if (e.instanceId === null || !e.object.getMatrixAt) return t;
		let n = e.object.matrixWorld.clone();
		return e.object.getMatrixAt(e.instanceId, n), n.premultiply(t);
	};
	ya$1 = (e) => {
		if (!Zi$2(e.geometry)) return null;
		e.geometry.boundingBox || e.geometry.computeBoundingBox();
		let t = e.geometry.boundingBox;
		return !t || !Yi$2(t.min) || !Yi$2(t.max) ? null : t;
	};
	ba$1 = (e, t) => {
		let n = t.x + t.width / 2, r = t.y + t.height / 2;
		if (e.intersectionPoint) {
			let i = e.intersectionPoint.clone().project(e.rootState.camera);
			Number.isFinite(i.x) && Number.isFinite(i.y) && (n = t.x + (i.x + 1) / 2 * t.width, r = t.y + (1 - i.y) / 2 * t.height);
		}
		return {
			x: n - 8,
			y: r - 8,
			width: 16,
			height: 16,
			borderRadius: `0px`
		};
	};
	xa$1 = (e) => {
		let t = _t$2(e.canvas), n = ya$1(e.object);
		if (!n) return ba$1(e, t);
		let r = va$1(e), i = [n.min.x, n.max.x], a = [n.min.y, n.max.y], o = [n.min.z, n.max.z], s = 1 / 0, c = 1 / 0, l = -1 / 0, u = -1 / 0, d = n.min.clone();
		for (let n of i) for (let i of a) for (let a of o) {
			if (d.set(n, i, a).applyMatrix4(r).project(e.rootState.camera), !Number.isFinite(d.x) || !Number.isFinite(d.y)) continue;
			let o = t.x + (d.x + 1) / 2 * t.width, f = t.y + (1 - d.y) / 2 * t.height;
			s = Math.min(s, o), c = Math.min(c, f), l = Math.max(l, o), u = Math.max(u, f);
		}
		if (!Number.isFinite(s) || !Number.isFinite(c) || !Number.isFinite(l) || !Number.isFinite(u)) return ba$1(e, t);
		let f = Math.max(t.x, s), p = Math.max(t.y, c), m = Math.min(t.x + t.width, l), h = Math.min(t.y + t.height, u);
		return m <= f || h <= p ? ba$1(e, t) : {
			x: f,
			y: p,
			width: m - f,
			height: h - p,
			borderRadius: `0px`
		};
	};
	z$1 = null;
	B$1 = null;
	V$1 = null;
	Sa$1 = () => {
		V$1 !== null && clearTimeout(V$1), V$1 = setTimeout(() => {
			V$1 = null, Gi$2();
		}, 100);
	};
	Ca$1 = () => {
		V$1 !== null && (clearTimeout(V$1), V$1 = null);
	};
	wa$1 = (e, t, n, r) => {
		let i = Math.abs(e - n), a = Math.abs(t - r);
		return i <= 2 && a <= 2;
	};
	Ta$1 = (e, t) => {
		if (!Number.isFinite(e) || !Number.isFinite(t)) return [];
		Ca$1(), Wi$2();
		try {
			let n = Ni$2(e, t), r = Lt$2() ? n.filter(Rt$2) : n, i = [];
			for (let n of r) i.push(pa$1(n, e, t));
			return i;
		} finally {
			Sa$1();
		}
	};
	Ea$1 = (e, t) => {
		if (!Number.isFinite(e) || !Number.isFinite(t)) return null;
		let n = performance.now();
		if (B$1) {
			let r = B$1.element, i = n - B$1.timestamp < 16;
			if (r.isConnected && i && Pi$2(e, t, B$1.bounds) && !N$1(r)) return r;
			B$1 = null, z$1?.element === r && (z$1 = null);
		}
		if (z$1) {
			let r = wa$1(e, t, z$1.clientX, z$1.clientY), i = n - z$1.timestamp < 16;
			if (r || i) return z$1.element;
		}
		Ca$1(), Wi$2();
		try {
			let r = null, i = mi$2(e, t), a = i ? pa$1(i, e, t) : null;
			if (r = i && a && ki$2(a) && Rt$2(i) ? a : ji$2(e, t), r && b(r) && !N$1(r)) {
				let e = _t$2(r);
				B$1 = {
					element: r,
					timestamp: n,
					bounds: {
						left: e.x,
						top: e.y,
						right: e.x + e.width,
						bottom: e.y + e.height
					}
				};
			} else B$1 = null;
			return z$1 = {
				clientX: e,
				clientY: t,
				element: r,
				timestamp: n
			}, r;
		} finally {
			Sa$1();
		}
	};
	Da$1 = () => {
		Ca$1(), Gi$2(), z$1 = null, B$1 = null;
	};
	Oa$1 = async (e, t) => {
		let n = Bn$2(), r = new URLSearchParams({ file: e }), i = n ? `line1` : `line`, a = n ? `column1` : `column`;
		t && r.set(i, String(t)), r.set(a, `1`);
		let o = n ? `${Un$2()}/__nextjs_launch-editor` : `/__open-in-editor`;
		return (await fetch(`${o}?${r}`)).ok;
	};
	ka$1 = async (e, t, n) => {
		try {
			let r = O(e);
			if (await Oa$1(r, t).catch(() => !1)) return;
			let i = t ? `&line=${t}` : ``, a = `https://react-grab.com/open-file?url=${encodeURIComponent(r)}${i}`, o = n ? n(a, r, t) : a;
			window.open(o, `_blank`, `noopener,noreferrer`);
		} catch (n) {
			throw n instanceof yt$2 ? n : new yt$2(e, t, n);
		}
	};
	H$1 = (e) => {
		console.warn(`[react-grab]`, e);
	};
	Aa$1 = (e, t) => e.length === t.length && e.every((e, n) => e === t[n]);
	U$1 = !1;
	W$1 = /* @__PURE__ */ new Map();
	ja$1 = -1;
	Ma$1 = /* @__PURE__ */ new WeakSet();
	G$1 = /* @__PURE__ */ new Map();
	K$1 = /* @__PURE__ */ new Map();
	Na$1 = (e) => Ma$1.has(e) ? !0 : !U$1 || !(`gsapVersions` in window) || !((/* @__PURE__ */ Error()).stack ?? ``).includes(`_tick`) ? !1 : (Ma$1.add(e), !0);
	typeof window < `u` && (window.requestAnimationFrame = (e) => {
		if (!Na$1(e)) return M$1(e);
		if (U$1) {
			let t = ja$1--;
			return W$1.set(t, e), t;
		}
		let t = M$1((n) => {
			if (U$1) {
				let n = ja$1--;
				W$1.set(n, e), G$1.set(t, n);
			} else e(n);
		});
		return t;
	}, window.cancelAnimationFrame = (e) => {
		if (W$1.has(e)) {
			W$1.delete(e);
			return;
		}
		let t = K$1.get(e);
		if (t !== void 0) {
			ci$2(t.nativeId), K$1.delete(e);
			return;
		}
		let n = G$1.get(e);
		n === void 0 ? ci$2(e) : (W$1.delete(n), G$1.delete(e));
	});
	Pa$1 = () => {
		if (!U$1) {
			U$1 = !0, W$1.clear(), G$1.clear();
			for (let [e, { nativeId: t, callback: n }] of K$1) ci$2(t), W$1.set(e, n);
			K$1.clear();
		}
	};
	Fa$1 = () => {
		if (U$1) {
			U$1 = !1;
			for (let [e, t] of W$1.entries()) {
				let n = M$1((n) => {
					K$1.delete(e), t(n);
				});
				K$1.set(e, {
					nativeId: n,
					callback: t
				});
			}
			W$1.clear(), G$1.clear();
		}
	};
	q$1 = (e, t) => {
		if (e.length !== 0) throw e.length === 1 ? e[0] : AggregateError(e, t);
	};
	Ia$1 = `
[${Ee}],
[${Ee}] * {
  animation-play-state: paused !important;
  transition: none !important;
}
`;
	La$1 = (e) => e?.namespaceURI === `http://www.w3.org/2000/svg` && e.tagName === `svg`;
	Ra$1 = null;
	J$1 = [];
	za$1 = [];
	Ba$1 = [];
	Va$1 = /* @__PURE__ */ new Set();
	Y$1 = /* @__PURE__ */ new Map();
	X$1 = /* @__PURE__ */ new Map();
	Ha$1 = [];
	Ua$1 = () => Y$1.size > 0;
	Wa$1 = () => {
		Ra$1 ||= Fi$2(`data-react-grab-frozen-styles`, Ia$1);
	};
	Ga$1 = (e) => {
		let t = /* @__PURE__ */ new Set();
		for (let n of e) {
			let e = La$1(n) ? n : n.closest(`svg`);
			La$1(e) && t.add(e);
			for (let e of n.querySelectorAll(`svg`)) La$1(e) && t.add(e);
		}
		return [...t];
	};
	Ka$1 = (e, t) => {
		let n = Reflect.get(e, t);
		typeof n == `function` && n.call(e);
	};
	qa$1 = (e, t) => {
		for (let n of e) {
			let e = X$1.get(n) ?? 0;
			e === 0 ? (X$1.set(n, 1), t.push(n), Ka$1(n, `pauseAnimations`)) : (X$1.set(n, e + 1), t.push(n));
		}
	};
	Ja$1 = (e, t) => {
		let n = [];
		for (let r of e) {
			let e = X$1.get(r);
			if (e) {
				if (e === 1) try {
					Ka$1(r, `unpauseAnimations`), X$1.delete(r);
				} catch (e) {
					n.push(r), t.push(e);
				}
				else X$1.set(r, e - 1);
			}
		}
		return n;
	};
	Ya$1 = (e) => {
		let t = [];
		for (let n of e) for (let e of n.getAnimations({ subtree: !0 })) e.playState === `running` && t.push(e);
		return t;
	};
	Xa$1 = (e) => {
		for (let t of e) try {
			t.finish();
		} catch {}
	};
	Za$1 = (e) => {
		for (let t of e) try {
			t.finish();
		} catch {
			try {
				t.play();
			} catch {}
		}
	};
	Qa$1 = (e) => {
		if (!(e.effect instanceof KeyframeEffect)) return !1;
		let t = e.effect.target;
		if (!v(t)) return !1;
		let n = t.getRootNode();
		return w(n) && Si$2(n.host);
	};
	$a$1 = (e) => {
		if (e.length === 0 || Aa$1(e, Ba$1)) return;
		eo$1();
		let t = [...e];
		Wa$1(), J$1 = t, za$1 = [];
		try {
			let e = Ga$1(J$1);
			qa$1(e, za$1);
			for (let e of J$1) e.setAttribute(Ee, ``);
			let n = Ya$1(J$1);
			Ha$1 = [];
			for (let e of n) e.pause(), Ha$1.push(e);
			Ba$1 = t;
		} catch (e) {
			try {
				eo$1();
			} catch (t) {
				throw AggregateError([e, t], `Rolling back animation freeze failed`);
			}
			throw e;
		}
	};
	eo$1 = () => {
		if (J$1.length === 0 && za$1.length === 0 && Ha$1.length === 0) return;
		let e = J$1, t = za$1, n = Ha$1, r = [], i = [];
		Ba$1 = [];
		for (let t of e) try {
			t.removeAttribute(Ee);
		} catch (e) {
			i.push(t), r.push(e);
		}
		J$1 = i, za$1 = Ja$1(t, r), Za$1(n), Ha$1 = [], q$1(r, `Unfreezing element animations failed`);
	};
	to$1 = (e) => e.length === 0 ? (eo$1(), () => {}) : ($a$1(e), eo$1);
	no$1 = (e) => {
		let t = [];
		for (let n of e.getAnimations()) Qa$1(n) || n.playState === `running` && t.push(n);
		return {
			targetDocument: e,
			runningAnimations: t
		};
	};
	ro$1 = (e) => {
		let t = Y$1.get(e);
		if (!t) return [];
		let n = [], r = t.styleElement;
		if (r) {
			try {
				if (t.didUseCssFreeze) {
					r.textContent = `
*, *::before, *::after {
  transition: none !important;
}
`;
					let t = [];
					for (let n of e.getAnimations()) Qa$1(n) || t.push(n);
					Xa$1(t);
				} else Za$1(t.frozenWaapiAnimations);
				t.didUseCssFreeze = !1, t.frozenWaapiAnimations = [];
			} catch (e) {
				n.push(e);
			}
			try {
				r.remove(), t.styleElement = null;
			} catch (e) {
				n.push(e);
			}
		}
		return t.frozenSvgElements = Ja$1(t.frozenSvgElements, n), t.styleElement === null && t.frozenSvgElements.length === 0 && t.frozenWaapiAnimations.length === 0 && Y$1.delete(e), n;
	};
	io$1 = (e) => {
		if (Y$1.has(e.targetDocument)) return;
		let t = Fi$2(`data-react-grab-global-freeze`, ``, e.targetDocument), n = {
			styleElement: t,
			frozenSvgElements: [],
			frozenWaapiAnimations: [],
			didUseCssFreeze: !1
		};
		if (Y$1.set(e.targetDocument, n), e.runningAnimations.length > 200) t.textContent = `
*, *::before, *::after {
  animation-play-state: paused !important;
  transition: none !important;
}
`, n.didUseCssFreeze = !0;
		else for (let t of e.runningAnimations) t.pause(), n.frozenWaapiAnimations.push(t);
		let r = Ga$1(Array.from(e.targetDocument.querySelectorAll(`svg`)));
		qa$1(r, n.frozenSvgElements);
	};
	ao$1 = (e) => {
		Va$1.add(e);
		try {
			Ua$1() && io$1(no$1(e));
		} catch (t) {
			Va$1.delete(e);
			let n = ro$1(e);
			throw n.length === 0 ? t : AggregateError([t, ...n], `Freezing frame animations failed`);
		}
		return () => {
			Va$1.delete(e);
			let t = ro$1(e);
			q$1(t, `Unfreezing frame animations failed`);
		};
	};
	oo$1 = () => Ua$1() ? [] : (Va$1.add(document), [...Va$1].map(no$1));
	so$1 = (e) => {
		if (!Ua$1()) {
			for (let t of e) io$1(t);
			Pa$1();
		}
	};
	co$1 = () => {
		if (!Ua$1()) return;
		let e = [];
		for (let t of Y$1.keys()) e.push(...ro$1(t));
		try {
			Fa$1();
		} catch (t) {
			e.push(t);
		}
		q$1(e, `Unfreezing global animations failed`);
	};
	lo$1 = (e, t) => {
		for (let n of e.querySelectorAll(`:hover`)) if (!Ci$2(n) && (y(n) && t.add(n), n.shadowRoot && lo$1(n.shadowRoot, t), b(n))) {
			let e = N$1(n);
			e && lo$1(e, t);
		}
	};
	uo$1 = () => {
		let e = /* @__PURE__ */ new Set();
		return lo$1(document, e), [...e];
	};
	fo$1 = [
		`mouseenter`,
		`mouseleave`,
		`mouseover`,
		`mouseout`,
		`pointerenter`,
		`pointerleave`,
		`pointerover`,
		`pointerout`
	];
	po$1 = [
		`focus`,
		`blur`,
		`focusin`,
		`focusout`
	];
	mo$1 = [
		`background-color`,
		`color`,
		`border-color`,
		`box-shadow`,
		`transform`,
		`opacity`,
		`outline`,
		`filter`,
		`scale`,
		`visibility`,
		`display`
	];
	ho$1 = [
		`background-color`,
		`color`,
		`border-color`,
		`box-shadow`,
		`outline`,
		`outline-offset`,
		`outline-width`,
		`outline-color`,
		`outline-style`,
		`filter`,
		`opacity`,
		`ring-color`,
		`ring-width`
	];
	go$1 = /* @__PURE__ */ new Map();
	_o$1 = /* @__PURE__ */ new Map();
	Z$2 = /* @__PURE__ */ new Set();
	vo$1 = !1;
	yo$1 = (e) => {
		e.stopImmediatePropagation();
	};
	bo$1 = (e) => {
		e.preventDefault(), e.stopImmediatePropagation();
	};
	xo$1 = (e) => {
		for (let t of fo$1) e.addEventListener(t, yo$1, !0);
		for (let t of po$1) e.addEventListener(t, bo$1, !0);
	};
	So$1 = (e) => {
		for (let t of fo$1) e.removeEventListener(t, yo$1, !0);
		for (let t of po$1) e.removeEventListener(t, bo$1, !0);
	};
	Co$1 = (e) => {
		Z$2.add(e);
		let t = Vi$2(e);
		return vo$1 && xo$1(e), () => {
			vo$1 && So$1(e), Z$2.delete(e), t();
		};
	};
	wo$1 = (e, t, n) => {
		if (n?.has(e)) return null;
		let r = getComputedStyle(e), i = /* @__PURE__ */ new Map(), a = /* @__PURE__ */ new Map();
		for (let n of t) {
			let t = r.getPropertyValue(n);
			t && (i.set(n, t), a.set(n, {
				value: e.style.getPropertyValue(n),
				priority: e.style.getPropertyPriority(n)
			}));
		}
		return {
			element: e,
			frozenPropertyValues: i,
			originalPropertyValues: a
		};
	};
	To$1 = (e, t) => {
		let n = [], r = mi$2(e, t);
		for (; r && r !== document.documentElement && !Ci$2(r);) y(r) && n.push(r), r = Pt$2(r);
		return n;
	};
	Eo$1 = () => {
		let e = [], t = document.activeElement;
		for (; t && t !== document.body;) y(t) && e.push(t), t = t.shadowRoot?.activeElement ? t.shadowRoot.activeElement : b(t) ? N$1(t)?.activeElement ?? null : null;
		return e;
	};
	Do$1 = (e, t) => {
		for (let { element: n, frozenPropertyValues: r, originalPropertyValues: i } of e) {
			t.set(n, i);
			for (let [e, t] of r) n.style.setProperty(e, t, `important`);
		}
	};
	Oo$1 = (e) => {
		let t = [];
		for (let [n, r] of e) {
			for (let [e, i] of r) try {
				i.value ? n.style.setProperty(e, i.value, i.priority) : n.style.removeProperty(e), r.delete(e);
			} catch (e) {
				t.push(e);
			}
			r.size === 0 && e.delete(n);
		}
		return t;
	};
	ko$1 = (e, t) => {
		if (Bi$2()) return null;
		let n = [], r = e !== void 0 && t !== void 0 && e >= 0 && t >= 0 && e < window.innerWidth && t < window.innerHeight ? To$1(e, t) : uo$1();
		for (let e of r) {
			let t = wo$1(e, mo$1);
			t && n.push(t);
		}
		let i = [];
		for (let e of Eo$1()) {
			let t = wo$1(e, ho$1, _o$1);
			t && i.push(t);
		}
		return {
			hoverStates: n,
			focusStates: i
		};
	};
	Ao$1 = (e) => {
		if (e) {
			vo$1 = !0, Z$2.add(document);
			for (let e of Z$2) xo$1(e);
			Do$1(e.hoverStates, go$1), Do$1(e.focusStates, _o$1), Hi$2();
		}
	};
	jo$1 = () => {
		Da$1(), vo$1 = !1;
		for (let e of Z$2) So$1(e);
		let e = [...Oo$1(go$1), ...Oo$1(_o$1)];
		try {
			Ui$2();
		} catch (t) {
			e.push(t);
		}
		q$1(e, `Unfreezing pseudo states failed`);
	};
	Mo$1 = () => {
		let e = [];
		try {
			jo$1();
		} catch (t) {
			e.push(t);
		}
		try {
			co$1();
		} catch (t) {
			e.push(t);
		}
		return e;
	};
	No$1 = (e, t) => {
		let n = ko$1(e, t), r = oo$1();
		try {
			Ao$1(n), so$1(r);
		} catch (e) {
			let t = Mo$1();
			throw t.length === 0 ? e : AggregateError([e, ...t], `Freezing global interactions failed`);
		}
	};
	Po$1 = () => {
		let e = Mo$1();
		q$1(e, `Unfreezing global interactions failed`);
	};
	Q$2 = !1;
	Fo$1 = 0;
	Io$1 = (e, t, n) => {
		let r = e.get(t);
		if (r) return r;
		let i = n();
		return e.set(t, i), i;
	};
	Lo$1 = /* @__PURE__ */ new WeakMap();
	Ro$1 = /* @__PURE__ */ new WeakMap();
	zo$1 = /* @__PURE__ */ new WeakMap();
	Bo$1 = /* @__PURE__ */ new Set();
	Vo$1 = [];
	$$2 = [];
	Ho$1 = /* @__PURE__ */ new WeakMap();
	Uo$1 = /* @__PURE__ */ new WeakMap();
	Wo$1 = /* @__PURE__ */ new WeakSet();
	Go$1 = O$2;
	Ko$1 = /* @__PURE__ */ new Set();
	qo$1 = (e) => e.rendererPackageName.startsWith(`react-dom`);
	Jo$1 = (e) => {
		let t = e;
		for (; t.return;) t = t.return;
		return t.stateNode ?? null;
	};
	Yo$1 = (e) => {
		let t = e.current?.stateNode;
		if (!t || typeof t != `object`) return !1;
		let n = Reflect.get(t, `containerInfo`);
		return !!(n && typeof n == `object` && typeof Reflect.get(n, `nodeType`) == `number`);
	};
	Xo$1 = () => {
		if (Go$1.size > 0) {
			let e = /* @__PURE__ */ new Set();
			for (let t of Go$1) Yo$1(t) && e.add(t);
			return e;
		}
		let e = /* @__PURE__ */ new Set(), t = (r) => {
			let i = Xe$3(r);
			if (i) {
				let t = Jo$1(i);
				t && e.add(t);
			} else for (let n of Array.from(r.children)) if (t(n), e.size > 0) return;
		};
		return t(document.body), e;
	};
	Zo$1 = (e, t) => {
		if (!e) return t;
		if (!t) return e;
		if (!e.next || !t.next) return t;
		let n = e.next, r = t.next, i = e === n, a = t === r;
		return i && a ? (e.next = t, t.next = e) : i ? (e.next = r, t.next = e) : a ? (t.next = n, e.next = t) : (e.next = r, t.next = n), t;
	};
	Qo$1 = (e) => {
		if (!e || Ho$1.has(e)) return;
		let t = {
			originalPendingDescriptor: Object.getOwnPropertyDescriptor(e, `pending`),
			pendingValueAtPause: e.pending,
			bufferedPending: null
		};
		typeof e.getSnapshot == `function` && (t.originalGetSnapshot = e.getSnapshot, t.snapshotValueAtPause = e.getSnapshot(), e.getSnapshot = () => Q$2 ? t.snapshotValueAtPause : t.originalGetSnapshot());
		let n = t.pendingValueAtPause;
		Object.defineProperty(e, "pending", {
			configurable: !0,
			enumerable: !0,
			get: () => Q$2 ? null : n,
			set: (e) => {
				Q$2 ? e !== null && (t.bufferedPending = Zo$1(t.bufferedPending ?? null, e)) : n = e;
			}
		}), Ho$1.set(e, t);
	};
	$o$1 = (e) => {
		if (!e) return [];
		let t = [], n = e.next;
		if (!n) return [];
		let r = n;
		do
			r &&= (t.push(r.action), r.next);
		while (r && r !== n);
		return t;
	};
	es$1 = (e) => {
		let t = Ho$1.get(e);
		if (!t) return;
		t.originalGetSnapshot && (e.getSnapshot = t.originalGetSnapshot), t.originalPendingDescriptor ? Object.defineProperty(e, "pending", t.originalPendingDescriptor) : delete e.pending, e.pending = null;
		let n = e.dispatch;
		if (typeof n == `function`) {
			let e = $o$1(t.pendingValueAtPause ?? null), r = $o$1(t.bufferedPending ?? null);
			for (let t of [...e, ...r]) $$2.push(() => n(t));
		}
		Ho$1.delete(e);
	};
	ts$1 = (e) => {
		if (Uo$1.has(e)) return;
		let t = {
			originalDescriptor: Object.getOwnPropertyDescriptor(e, `memoizedValue`),
			frozenValue: e.memoizedValue
		};
		Object.defineProperty(e, "memoizedValue", {
			configurable: !0,
			enumerable: !0,
			get() {
				return Q$2 ? t.frozenValue : t.originalDescriptor?.get ? t.originalDescriptor.get.call(this) : this._memoizedValue;
			},
			set(e) {
				Q$2 ? (t.pendingValue = e, t.didReceivePendingValue = !0) : t.originalDescriptor?.set ? t.originalDescriptor.set.call(this, e) : this._memoizedValue = e;
			}
		}), t.originalDescriptor?.get || (e._memoizedValue = t.frozenValue), Uo$1.set(e, t);
	};
	ns$1 = (e) => {
		let t = Uo$1.get(e);
		t && (t.originalDescriptor ? Object.defineProperty(e, "memoizedValue", t.originalDescriptor) : delete e.memoizedValue, t.didReceivePendingValue && (e.memoizedValue = t.pendingValue), Uo$1.delete(e));
	};
	rs$1 = (e) => {
		let t = e.memoizedState;
		for (; t;) t.queue && typeof t.queue == `object` && Qo$1(t.queue), t = t.next;
		let n = e.dependencies?.firstContext;
		for (; n && typeof n == `object` && `memoizedValue` in n;) ts$1(n), n = n.next;
	};
	is$1 = (e) => {
		let t = e.memoizedState;
		for (; t;) t.queue && typeof t.queue == `object` && es$1(t.queue), t = t.next;
		let n = e.dependencies?.firstContext;
		for (; n && typeof n == `object` && `memoizedValue` in n;) ns$1(n), n = n.next;
	};
	as$1 = (e) => {
		let t = e;
		for (; t;) if (Oe$1(t) && rs$1(t), t.child) t = t.child;
		else {
			for (; t !== e && !t.sibling;) if (t = t.return, !t) return;
			if (t === e) return;
			t = t.sibling;
		}
	};
	os$1 = (e) => {
		let t = e;
		for (; t;) if (Oe$1(t) && is$1(t), t.child) t = t.child;
		else {
			for (; t !== e && !t.sibling;) if (t = t.return, !t) return;
			if (t === e) return;
			t = t.sibling;
		}
	};
	ss$1 = (e) => {
		if (Lo$1.has(e)) return;
		let t = e, n = {
			useState: t.useState,
			useReducer: t.useReducer,
			useTransition: t.useTransition,
			useSyncExternalStore: t.useSyncExternalStore
		};
		Lo$1.set(e, n), t.useState = (...t) => {
			let r = n.useState.apply(e, t);
			if (!Q$2 || !Array.isArray(r) || typeof r[1] != `function`) return r;
			let [i, a] = r;
			return [i, Io$1(Ro$1, a, () => (...e) => {
				Q$2 ? $$2.push(() => a(...e)) : a(...e);
			})];
		}, t.useReducer = (...t) => {
			let r = n.useReducer.apply(e, t);
			if (!Q$2 || !Array.isArray(r) || typeof r[1] != `function`) return r;
			let [i, a] = r;
			return [i, Io$1(Ro$1, a, () => (...e) => {
				Q$2 ? $$2.push(() => a(...e)) : a(...e);
			})];
		}, t.useTransition = (...t) => {
			let r = n.useTransition.apply(e, t);
			if (!Q$2 || !Array.isArray(r) || typeof r[1] != `function`) return r;
			let [i, a] = r;
			return [i, Io$1(zo$1, a, () => (e) => {
				Q$2 ? Vo$1.push(() => a(e)) : a(e);
			})];
		}, t.useSyncExternalStore = ((e, t, r) => Q$2 ? n.useSyncExternalStore((t) => e(() => {
			Q$2 ? Bo$1.add(t) : t();
		}), t, r) : n.useSyncExternalStore(e, t, r));
	};
	cs$1 = (e) => {
		let t = e.currentDispatcherRef;
		if (!t || typeof t != `object`) return;
		let n = `H` in t ? `H` : `current`, r = t[n];
		Object.defineProperty(t, n, {
			configurable: !0,
			enumerable: !0,
			get: () => (r && typeof r == `object` && ss$1(r), r),
			set: (e) => {
				r = e;
			}
		});
	};
	ls$1 = (e) => {
		queueMicrotask(() => {
			try {
				for (let t of x$3().renderers.values()) if (qo$1(t) && typeof t.scheduleUpdate == `function`) {
					for (let n of e) if (n.current) try {
						t.scheduleUpdate(n.current);
					} catch (e) {
						H$1(new E(`scheduleUpdate failed during unfreeze`, e));
					}
				}
			} catch (e) {
				H$1(new E(`scheduleReactUpdate failed`, e));
			}
		});
	};
	us$1 = (e) => {
		for (let t of e) try {
			t();
		} catch (e) {
			H$1(new E(`Callback failed during state replay`, e));
		}
	};
	ds$1 = () => {
		for (let e of x$3().renderers.values()) qo$1(e) && (Wo$1.has(e) || (cs$1(e), Wo$1.add(e)));
	};
	fs$1 = () => {
		Bo$1.clear(), Vo$1.length = 0, $$2.length = 0;
	};
	ps$1 = () => {
		let e = new Set(Ko$1);
		try {
			for (let t of Xo$1()) e.add(t);
		} catch (e) {
			H$1(new E(`Collecting fiber roots failed during unfreeze`, e));
		}
		Ko$1.clear(), Q$2 = !1;
		for (let t of e) try {
			os$1(t.current);
		} catch (e) {
			H$1(new E(`Resuming a fiber root failed during unfreeze`, e));
		}
		let t = Array.from(Bo$1), n = Vo$1.slice(), r = $$2.slice();
		fs$1(), us$1(t), us$1(n), us$1(r), Q$2 || ls$1(e);
	};
	ms$1 = () => {
		let e = Fo$1 === 0;
		if (Fo$1 += 1, e) try {
			ds$1(), Q$2 = !0;
			let e = Xo$1();
			for (let t of e) Ko$1.add(t), as$1(t.current);
		} catch (e) {
			throw --Fo$1, Q$2 && ps$1(), e;
		}
		let t = !1;
		return () => {
			t || (t = !0, --Fo$1, Fo$1 === 0 && ps$1());
		};
	};
	hs$1 = () => {
		try {
			return ms$1();
		} catch (e) {
			return H$1(new E(`Pausing React updates failed`, e)), () => {};
		}
	};
}));
// ../ui/dist/react-grab/renderer-CyobGvMH.js
var renderer_CyobGvMH_exports = /* @__PURE__ */ __exportAll({ ReactGrabRenderer: () => ua$1 });
var We, Ge, Ke, qe, Je, Ye, Xe$1, Ze$1, Qe$1, $e$1, et$1, tt$1, nt$1, rt$1, it$1, Z$1, at$1, ot$1, st$1, ct$1, lt$1, ut$1, dt$1, ft$1, pt$1, mt$1, ht$1, gt$1, _t$1, vt$1, yt$1, bt$1, xt$1, St$1, Ct$1, wt$1, Tt$1, Et$1, Dt$1, Ot$1, kt$1, At$1, jt$1, Mt$1, Nt$1, Pt$1, Ft$1, It$1, Lt$1, Rt$1, zt$1, Bt$1, Vt$1, Ht$1, Ut$1, Wt$1, Gt$1, Kt$1, qt$1, Jt$1, Yt$1, Xt$1, Zt$1, Qt$1, $t$1, en$1, tn$1, nn$1, rn$1, an$1, on$1, sn$1, cn$1, ln$1, un$1, dn$1, fn$1, pn$1, mn$1, hn$1, gn$1, _n$1, vn$1, yn$1, bn$1, xn$1, Sn$1, Cn$1, wn$1, Tn$1, En$1, Dn$1, On$1, kn$1, An$1, jn$1, Mn$1, Nn$1, Pn$1, Fn$1, In$1, Ln$1, Rn$1, zn$1, Bn$1, Vn$1, Hn$1, Un$1, Wn$1, Gn$1, Kn$1, qn$1, Jn$1, Yn$1, Xn$1, Zn$1, Qn$1, $n$1, er$1, tr$1, nr$1, rr$1, ir$1, ar$1, or$1, sr$1, Q$1, $$1, cr$1, lr$1, ur$1, dr$1, fr$1, pr$1, mr$1, hr$1, gr$1, _r$1, vr$1, yr$1, br$1, xr$1, Sr$1, Cr$1, wr$1, Tr$1, Er$1, Dr$1, Or$1, kr$1, Ar$1, jr$1, Mr$1, Nr$1, Pr$1, Fr$1, Ir$1, Lr$1, Rr$1, zr$1, Br$1, Vr$1, Hr$1, Ur$1, Wr$1, Gr$1, Kr$1, qr$1, Jr$1, Yr$1, Xr$1, Zr$1, Qr$1, $r$1, ei$1, ti$1, ni$1, ri$1, ii$1, ai$1, oi$1, si$1, ci$1, li$1, ui$1, di$1, fi$1, pi$1, mi$1, hi$1, gi$1, _i$1, vi$1, yi$1, bi$1, xi$1, Si$1, Ci$1, wi$1, Ti$1, Ei$1, Di$1, Oi$1, ki$1, Ai$1, ji$1, Mi$1, Ni$1, Pi$1, Fi$1, Ii$1, Li$1, Ri$1, zi$1, Bi$1, Vi$1, Hi$1, Ui$1, Wi$1, Gi$1, Ki$1, qi$1, Ji$1, Yi$1, Xi$1, Zi$1, Qi$1, $i$1, ea$1, ta$1, na$1, ra$1, ia$1, aa$1, oa$1, sa$1, ca$1, la$1, ua$1;
var init_renderer_CyobGvMH = __esmMin((() => {
	init_core_DT_3yIQi();
	init_freeze_updates_Cd2QP_tz();
	We = (e, t, n) => e + (t - e) * n;
	Ge = (e, t) => {
		/** @license MIT
		* Copyright (c) 2025 Aiden Bai
		* See LICENSE in the package root.
		*/
		let n = Math.max(t, 1);
		return 1 - (1 - e) ** (n / de);
	};
	Ke = (e) => {
		let t = e.scrollingElement ?? e.documentElement;
		return {
			width: t.clientWidth,
			height: t.clientHeight
		};
	};
	qe = /*#__PURE__*/ mn(`<canvas data-react-grab-overlay-canvas style=position:fixed;top:0;left:0;pointer-events:none>`);
	Je = {
		borderColor: ge,
		fillColor: _e,
		lerpFactor: ae
	};
	Ye = {
		drag: {
			borderColor: me,
			fillColor: he,
			lerpFactor: ue
		},
		selection: Je,
		grabbed: Je
	};
	Xe$1 = (e) => {
		let t, n = null, r = 0, i = 0, a = 1, o = null, s = null, l = null, u = [], f = null, p = [], m = ee() ? `display-p3` : `srgb`, h = () => {
			if (!t) return;
			a = Math.max(window.devicePixelRatio || 1, 2);
			let e = Ke(document);
			r = e.width || window.innerWidth, i = e.height || window.innerHeight, t.width = r * a, t.height = i * a, t.style.width = `${r}px`, t.style.height = `${i}px`, n = t.getContext(`2d`, { colorSpace: m }), n && n.scale(a, a);
		}, g = (e) => {
			if (!e) return 0;
			let t = e.match(/^(\d+(?:\.\d+)?)/);
			return t ? parseFloat(t[1]) : 0;
		}, _ = (e, t, n) => ({
			id: e,
			current: {
				x: t.x,
				y: t.y,
				width: t.width,
				height: t.height
			},
			target: {
				x: t.x,
				y: t.y,
				width: t.width,
				height: t.height
			},
			borderRadius: g(t.borderRadius),
			opacity: n?.opacity ?? 1,
			targetOpacity: n?.targetOpacity ?? n?.opacity ?? 1,
			createdAt: n?.createdAt,
			fadeStartTimestamp: null,
			isInitialized: !0
		}), v = (e, t, n) => {
			let r = e.target;
			r.x = t.x, r.y = t.y, r.width = t.width, r.height = t.height, e.borderRadius = g(t.borderRadius), n !== void 0 && (n > e.targetOpacity && (e.opacity = n), e.targetOpacity = n);
		}, y = (e) => e.boundsMultiple ?? [e.bounds], b = (e, t, n, r, i, a, o, s, c = 1) => {
			if (r <= 0 || i <= 0) return;
			let l = Math.min(r / 2, i / 2), u = Math.min(a, l), d = c !== 1;
			d && (e.globalAlpha = c), e.beginPath(), u > 0 ? e.roundRect(t, n, r, i, u) : e.rect(t, n, r, i), e.fillStyle = o, e.fill(), e.strokeStyle = s, e.lineWidth = 1, e.stroke(), d && (e.globalAlpha = 1);
		}, x = () => {
			if (!n || !e.dragVisible || !f) return;
			let t = Ye.drag;
			b(n, f.current.x, f.current.y, f.current.width, f.current.height, f.borderRadius, t.fillColor, t.borderColor);
		}, S = () => {
			if (!n || !e.selectionVisible) return;
			let t = Ye.selection;
			for (let e of u) b(n, e.current.x, e.current.y, e.current.width, e.current.height, e.borderRadius, t.fillColor, t.borderColor, e.opacity);
		}, C = (e) => {
			if (!n) return;
			let t = Ye.grabbed;
			for (let r of e) b(n, r.current.x, r.current.y, r.current.width, r.current.height, r.borderRadius, t.fillColor, t.borderColor, r.opacity);
		}, w = () => {
			n && t && (r <= 0 || i <= 0 || (n.setTransform(1, 0, 0, 1, 0, 0), n.clearRect(0, 0, t.width, t.height), n.setTransform(a, 0, 0, a, 0, 0), x(), S(), C(p)));
		}, T = (e, t, n) => {
			let r = We(e.current.x, e.target.x, t), i = We(e.current.y, e.target.y, t), a = We(e.current.width, e.target.width, t), o = We(e.current.height, e.target.height, t), s = Math.abs(r - e.target.x) < .5 && Math.abs(i - e.target.y) < .5 && Math.abs(a - e.target.width) < .5 && Math.abs(o - e.target.height) < .5;
			e.current.x = s ? e.target.x : r, e.current.y = s ? e.target.y : i, e.current.width = s ? e.target.width : a, e.current.height = s ? e.target.height : o;
			let c = !0;
			if (n?.interpolateOpacity) {
				let n = We(e.opacity, e.targetOpacity, t);
				c = Math.abs(n - e.targetOpacity) < pe, e.opacity = c ? e.targetOpacity : n;
			}
			return !s || !c;
		}, E = () => {
			let e = performance.now(), t = l === null ? de : e - l;
			l = e;
			let n = Ge(Ye.drag.lerpFactor, t), r = Ge(Ye.selection.lerpFactor, t), i = Ge(Ye.grabbed.lerpFactor, t), a = !1, c = null;
			f?.isInitialized && T(f, n) && (a = !0);
			for (let e of u) e.isInitialized && T(e, r) && (a = !0);
			let d = Date.now();
			p = p.filter((t) => {
				let n = t.id.startsWith(`label-`);
				if (t.isInitialized && T(t, i) && (a = !0), n && t.targetOpacity === 0) {
					t.fadeStartTimestamp === null && (t.fadeStartTimestamp = e);
					let n = e - t.fadeStartTimestamp, r = Math.min(1, n / 125), i = 1 - (1 - r) * (1 - r);
					return t.opacity = Math.max(0, 1 - i), r >= 1 ? !1 : (a = !0, !0);
				}
				if (n && (t.fadeStartTimestamp = null), t.createdAt !== void 0) {
					let e = d - t.createdAt;
					if (e >= 1625) return !1;
					if (e > 1500) {
						let n = Math.min(1, (e - oe) / 125);
						t.opacity = 1 - (1 - (1 - n) * (1 - n)), a = !0;
					} else {
						let t = oe - e;
						c = c === null ? t : Math.min(c, t);
					}
					return !0;
				}
				return n ? !0 : t.opacity > 0;
			}), w(), a ? o = M$1(E) : (o = null, l = null, c !== null && (s = window.setTimeout(() => {
				s = null, D();
			}, Math.max(0, c))));
		}, D = () => {
			s !== null && (window.clearTimeout(s), s = null), o === null && (o = M$1(E));
		}, O = () => {
			h(), D();
		};
		return B(ft(() => [
			e.selectionVisible,
			e.selectionBounds,
			e.selectionBoundsMultiple,
			e.selectionShouldSnap
		], ([e, t, n, r]) => {
			if (!e || !t && (!n || n.length === 0)) {
				u = [], D();
				return;
			}
			let i;
			i = n && n.length > 0 ? n : t ? [t] : [];
			let a = /* @__PURE__ */ new Map();
			for (let e of u) a.set(e.id, e);
			u = i.map((e, t) => {
				let n = `selection-${t}`, i = a.get(n);
				return i ? (v(i, e), r && (i.current.x = i.target.x, i.current.y = i.target.y, i.current.width = i.target.width, i.current.height = i.target.height), i) : _(n, e);
			}), D();
		})), B(ft(() => [e.dragVisible, e.dragBounds], ([e, t]) => {
			!e || !t ? (f = null, D()) : (f ? v(f, t) : f = _(`drag`, t), D());
		})), B(ft(() => [e.grabbedBoxes, e.labelInstances], ([e, t]) => {
			let n = e ?? [], r = t ?? [], i = /* @__PURE__ */ new Map();
			for (let e of n) i.set(e.id, e);
			let a = /* @__PURE__ */ new Map();
			for (let e of p) a.set(e.id, e);
			for (let e of n) if (!a.has(e.id)) {
				let t = _(e.id, e.bounds, { createdAt: e.createdAt });
				p.push(t), a.set(e.id, t);
			}
			for (let e of p) {
				let t = i.get(e.id);
				t && v(e, t.bounds);
			}
			let o = /* @__PURE__ */ new Set();
			for (let e of r) {
				let t = y(e), n = e.status === `fading` ? 0 : 1;
				for (let r = 0; r < t.length; r++) {
					let i = t[r], s = `label-${e.id}-${r}`;
					o.add(s);
					let c = a.get(s);
					if (c) v(c, i, n);
					else {
						let e = _(s, i, {
							opacity: 1,
							targetOpacity: n
						});
						p.push(e), a.set(s, e);
					}
				}
			}
			p = p.filter((e) => e.id.startsWith(`label-`) ? o.has(e.id) : i.has(e.id)), D();
		})), pt(() => {
			h(), D(), window.addEventListener(`resize`, O);
			let e = null, t = () => {
				Math.max(window.devicePixelRatio || 1, 2) !== a && (O(), n());
			}, n = () => {
				e && e.removeEventListener(`change`, t), e = window.matchMedia(`(resolution: ${window.devicePixelRatio}dppx)`), e.addEventListener(`change`, t);
			};
			n(), W(() => {
				window.removeEventListener(`resize`, O), e && e.removeEventListener(`change`, t), o !== null && ci$2(o), s !== null && window.clearTimeout(s);
			});
		}), (() => {
			var e = qe(), n = t;
			return typeof n == `function` ? Tn(n, e) : t = e, lt((t) => Cn(e, `z-index`, String(le))), e;
		})();
	};
	Ze$1 = /*#__PURE__*/ mn(`<div style=position:fixed;pointer-events:none;will-change:opacity;contain:strict;transform:translateZ(0)>`);
	Qe$1 = (e) => {
		let t = Lt$2(), n = t ? getComputedStyle(t).borderRadius : `0px`, r = () => t?.getBoundingClientRect() ?? null, [i, a] = z(r());
		if (t) {
			let n = () => a(r());
			pt(() => {
				let e = new ResizeObserver(n);
				e.observe(t), window.addEventListener(`scroll`, n, {
					capture: !0,
					passive: !0
				}), window.addEventListener(`resize`, n), W(() => {
					e.disconnect(), window.removeEventListener(`scroll`, n, { capture: !0 }), window.removeEventListener(`resize`, n);
				});
			}), B(() => {
				e.visible && n();
			});
		}
		let o = () => {
			let e = i();
			return e ? `${e.top}px` : `0`;
		}, s = () => {
			let e = i();
			return e ? `${e.left}px` : `0`;
		}, l = () => {
			let e = i();
			return e ? `${e.width}px` : `100%`;
		}, u = () => {
			let e = i();
			return e ? `${e.height}px` : `100%`;
		};
		return (() => {
			var t = Ze$1();
			return Cn(t, `border-radius`, n), Cn(t, `z-index`, le), Cn(t, `transition`, `opacity 125ms ease-out`), Cn(t, `box-shadow`, `inset 0 0 50px ${ve}`), lt((n) => {
				var r = o(), i = s(), a = l(), d = u(), f = +!!e.visible;
				return r !== n.e && Cn(t, `top`, n.e = r), i !== n.t && Cn(t, `left`, n.t = i), a !== n.a && Cn(t, `width`, n.a = a), d !== n.o && Cn(t, `height`, n.o = d), f !== n.i && Cn(t, `opacity`, n.i = f), n;
			}, {
				e: void 0,
				t: void 0,
				a: void 0,
				o: void 0,
				i: void 0
			}), t;
		})();
	};
	$e$1 = (e, t) => {
		e.style.height = `auto`, e.style.height = `${Math.min(e.scrollHeight, t)}px`;
	};
	et$1 = (e, t) => {
		e && e.focus(t);
	};
	tt$1 = (e) => {
		if (e <= 0) return 8;
		let t = e * ye;
		return Math.round(Math.max(4, Math.min(8, t)));
	};
	nt$1 = () => {
		let e = Lt$2();
		if (e) {
			let t = e.getBoundingClientRect();
			return {
				width: t.width,
				height: t.height,
				offsetLeft: t.left,
				offsetTop: t.top
			};
		}
		let t = window.visualViewport;
		return t ? {
			width: t.width,
			height: t.height,
			offsetLeft: t.offsetLeft,
			offsetTop: t.offsetTop
		} : {
			width: window.innerWidth,
			height: window.innerHeight,
			offsetLeft: 0,
			offsetTop: 0
		};
	};
	rt$1 = (e) => e.isComposing || e.keyCode === 229;
	it$1 = (e) => {
		if (typeof e == `string` || typeof e == `number`) return String(e);
		if (!e || typeof e != `object`) return ``;
		if (Array.isArray(e)) {
			let t = ``;
			for (let n of e) {
				if (!n) continue;
				let e = it$1(n);
				e && (t = t ? `${t} ${e}` : e);
			}
			return t;
		}
		let t = ``;
		for (let n in e) e[n] && (t = t ? `${t} ${n}` : n);
		return t;
	};
	Z$1 = (...e) => {
		let t = ``;
		for (let n of e) {
			if (!n) continue;
			let e = it$1(n);
			e && (t = t ? `${t} ${e}` : e);
		}
		return t;
	};
	at$1 = (e) => e.elementsCount && e.elementsCount > 1 ? {
		tagName: `${e.elementsCount} elements`,
		componentName: void 0
	} : {
		tagName: e.tagName || e.componentName || `element`,
		componentName: e.tagName ? e.componentName : void 0
	};
	ot$1 = /*#__PURE__*/ mn(`<svg xmlns=http://www.w3.org/2000/svg viewBox="0 0 24 24"fill=currentColor><path fill-rule=evenodd clip-rule=evenodd d="M19.2929 12.7071C19.6834 13.0976 20.3166 13.0976 20.7071 12.7071C21.0976 12.3166 21.0976 11.6834 20.7071 11.2929L12.7071 3.29289C12.5196 3.10536 12.2652 3 12 3C11.7348 3 11.4804 3.10536 11.2929 3.29289L3.29289 11.2929C2.90237 11.6834 2.90237 12.3166 3.29289 12.7071C3.68342 13.0976 4.31658 13.0976 4.70711 12.7071L11 6.41421L11 20C11 20.5523 11.4477 21 12 21C12.5523 21 13 20.5523 13 20L13 6.41422L19.2929 12.7071Z">`);
	st$1 = (e) => {
		let t = () => e.size ?? 10;
		return (() => {
			var n = ot$1();
			return lt((r) => {
				var i = t(), a = t(), o = e.class;
				return i !== r.e && gn(n, `width`, r.e = i), a !== r.t && gn(n, `height`, r.t = a), o !== r.a && gn(n, `class`, r.a = o), r;
			}, {
				e: void 0,
				t: void 0,
				a: void 0
			}), n;
		})();
	};
	ct$1 = /*#__PURE__*/ mn(`<svg xmlns=http://www.w3.org/2000/svg viewBox="0 0 24 24"fill=none stroke=currentColor stroke-width=2 stroke-linecap=round stroke-linejoin=round><path class=icon-loader-bar d="M12 2v4"style=animation-delay:0ms></path><path class=icon-loader-bar d="M15 6.8l2-3.5"style=animation-delay:-42ms></path><path class=icon-loader-bar d="M17.2 9l3.5-2"style=animation-delay:-83ms></path><path class=icon-loader-bar d="M18 12h4"style=animation-delay:-125ms></path><path class=icon-loader-bar d="M17.2 15l3.5 2"style=animation-delay:-167ms></path><path class=icon-loader-bar d="M15 17.2l2 3.5"style=animation-delay:-208ms></path><path class=icon-loader-bar d="M12 18v4"style=animation-delay:-250ms></path><path class=icon-loader-bar d="M9 17.2l-2 3.5"style=animation-delay:-292ms></path><path class=icon-loader-bar d="M6.8 15l-3.5 2"style=animation-delay:-333ms></path><path class=icon-loader-bar d="M2 12h4"style=animation-delay:-375ms></path><path class=icon-loader-bar d="M6.8 9l-3.5-2"style=animation-delay:-417ms></path><path class=icon-loader-bar d="M9 6.8l-2-3.5"style=animation-delay:-458ms>`);
	lt$1 = (e) => {
		let t = () => e.size ?? 16;
		return (() => {
			var n = ct$1();
			return n.firstChild.nextSibling.nextSibling.nextSibling.nextSibling.nextSibling.nextSibling.nextSibling.nextSibling.nextSibling.nextSibling.nextSibling, lt((r) => {
				var i = t(), a = t(), o = e.class;
				return i !== r.e && gn(n, `width`, r.e = i), a !== r.t && gn(n, `height`, r.t = a), o !== r.a && gn(n, `class`, r.a = o), r;
			}, {
				e: void 0,
				t: void 0,
				a: void 0
			}), n;
		})();
	};
	ut$1 = /*#__PURE__*/ mn(`<svg data-react-grab-arrow class="absolute block z-10"><path>`);
	dt$1 = (e) => {
		let t = () => e.position === `bottom`, n = () => tt$1(e.labelWidth ?? 0), r = () => n() * 2, i = () => n(), a = () => {
			let e = r(), n = i(), a = 1 * Math.SQRT1_2, o = e / 2, s = t() ? n : 0, c = t() ? a : n - a, l = +!!t();
			return `M0 ${s} L${o - a} ${c} A1 1 0 0 ${l} ${o + a} ${c} L${e} ${s} Z`;
		};
		return (() => {
			var n = ut$1(), o = n.firstChild;
			return gn(o, `fill`, ne), lt((s) => {
				var l = r(), u = i(), d = `0 0 ${r()} ${i()}`, f = `calc(${e.leftPercent}% + ${e.leftOffsetPx}px)`, p = t() ? `0` : void 0, m = t() ? void 0 : `0`, h = t() ? `translateX(-50%) translateY(calc(-100% + 1px))` : `translateX(-50%) translateY(calc(100% - 1px))`, g = a();
				return l !== s.e && gn(n, `width`, s.e = l), u !== s.t && gn(n, `height`, s.t = u), d !== s.a && gn(n, `viewBox`, s.a = d), f !== s.o && Cn(n, `left`, s.o = f), p !== s.i && Cn(n, `top`, s.i = p), m !== s.n && Cn(n, `bottom`, s.n = m), h !== s.s && Cn(n, `transform`, s.s = h), g !== s.h && gn(o, `d`, s.h = g), s;
			}, {
				e: void 0,
				t: void 0,
				a: void 0,
				o: void 0,
				i: void 0,
				n: void 0,
				s: void 0,
				h: void 0
			}), n;
		})();
	};
	ft$1 = /*#__PURE__*/ mn(`<span>`);
	pt$1 = /*#__PURE__*/ mn(`<span class=text-[var(--rg-text-secondary)]>.`);
	mt$1 = /*#__PURE__*/ mn(`<span class=text-[var(--rg-text-primary)]>`);
	ht$1 = /*#__PURE__*/ mn(`<span class="text-[var(--rg-text-primary)] text-[13px] leading-4 h-fit font-medium overflow-hidden text-ellipsis whitespace-nowrap min-w-0">`);
	gt$1 = /*#__PURE__*/ mn(`<button type=button>`);
	_t$1 = /*#__PURE__*/ mn(`<div>`);
	vt$1 = (t) => {
		let n = () => {
			t.onHoverChange?.(!0);
		}, r = () => {
			t.onHoverChange?.(!1);
		}, i = () => t.componentName ? `${t.componentName}.${t.tagName}` : t.tagName, a = () => (() => {
			var n = ht$1();
			return En(n, Gt(en, {
				get when() {
					return t.componentName;
				},
				get children() {
					return [(() => {
						var e = ft$1();
						return En(e, () => t.componentName), e;
					})(), (() => {
						var e = pt$1();
						return e.firstChild, En(e, () => t.tagName, null), e;
					})()];
				}
			}), null), En(n, Gt(en, {
				get when() {
					return !t.componentName;
				},
				get children() {
					var e = mt$1();
					return En(e, () => t.tagName), e;
				}
			}), null), n;
		})();
		return Gt(en, {
			get when() {
				return t.isClickable;
			},
			get fallback() {
				return (() => {
					var e = _t$1();
					return bn(e, `click`, t.onClick, !0), e.addEventListener(`mouseleave`, r), e.addEventListener(`mouseenter`, n), En(e, a), lt(() => yn(e, Z$1(`contain-layout flex items-center gap-1 max-w-[280px] overflow-hidden`, t.shrink && `shrink-0`))), e;
				})();
			},
			get children() {
				var e = gt$1();
				return bn(e, `click`, t.onClick, !0), e.addEventListener(`mouseleave`, r), e.addEventListener(`mouseenter`, n), En(e, a), lt((n) => {
					var r = `Open source for ${i()}`, a = Z$1(`contain-layout flex items-center gap-1 max-w-[280px] overflow-hidden cursor-pointer bg-transparent border-none p-0 m-0 text-left`, t.shrink && `shrink-0`);
					return r !== n.e && gn(e, `aria-label`, n.e = r), a !== n.t && yn(e, n.t = a), n;
				}, {
					e: void 0,
					t: void 0
				}), e;
			}
		});
	};
	hn([`click`]);
	yt$1 = /*#__PURE__*/ mn(`<div class="[font-synthesis:none] contain-layout shrink-0 flex flex-col items-start px-2 py-1.5 w-auto h-fit self-stretch [border-top-width:0.5px] border-t-[var(--rg-border-subtle)] antialiased">`);
	bt$1 = (e) => (() => {
		var t = yt$1();
		return En(t, () => e.children), t;
	})();
	xt$1 = (e, t = {}) => (n = {}) => {
		let r = t.variants;
		if (!r) return e;
		let i = [e];
		for (let e in r) {
			let a = n[e] ?? t.defaultVariants?.[e];
			if (a === void 0) continue;
			let o = r[e][String(a)];
			o && i.push(o);
		}
		return Z$1(...i);
	};
	St$1 = /*#__PURE__*/ mn(`<div>`);
	Ct$1 = xt$1(`contain-layout antialiased [font-synthesis:none] bg-[var(--rg-panel-bg)]`, {
		variants: { shape: {
			panel: `rounded-[14px] [corner-shape:superellipse(1.25)]`,
			pill: `rounded-full`
		} },
		defaultVariants: { shape: `panel` }
	});
	wt$1 = (e) => {
		let [t, n] = Xt(e, [`shape`, `class`]);
		return (() => {
			var e = St$1();
			return wn(e, Yt({ get class() {
				return Z$1(Ct$1({ shape: t.shape }), t.class);
			} }, n), !1, !1), e;
		})();
	};
	Tt$1 = null;
	Et$1 = {
		claim: (e) => {
			Tt$1 = e;
		},
		release: (e) => {
			Tt$1 === e && (Tt$1 = null);
		},
		isActive: (e) => Tt$1 === e
	};
	Dt$1 = (e) => {
		let t = Symbol(), n = zt$2((n) => {
			Et$1.isActive(t) && (yr(n) || (n.code === `Enter` ? e.onEnter?.(n) : n.code === `Escape` && e.onEscape?.(n)));
		});
		return pt(() => {
			Et$1.claim(t), window.addEventListener(`keydown`, n, { capture: !0 });
		}), W(() => {
			Et$1.release(t), window.removeEventListener(`keydown`, n, { capture: !0 });
		}), { claimFocus: () => Et$1.claim(t) };
	};
	Ot$1 = /*#__PURE__*/ mn(`<svg xmlns=http://www.w3.org/2000/svg viewBox="0 0 24 24"fill=currentColor><path fill-rule=evenodd clip-rule=evenodd d="M10.7071 21.2071C10.3166 21.5976 9.68342 21.5976 9.29289 21.2071L3.29289 15.2071C2.90237 14.8166 2.90237 14.1834 3.29289 13.7929L9.29289 7.79289C9.68342 7.40237 10.3166 7.40237 10.7071 7.79289C11.0976 8.18342 11.0976 8.81658 10.7071 9.20711L6.41421 13.5H16C16.5523 13.5 17 13.0523 17 12.5V3.5C17 2.94772 17.4477 2.5 18 2.5C18.5523 2.5 19 2.94772 19 3.5V12.5C19 14.1569 17.6569 15.5 16 15.5H6.41421L10.7071 19.7929C11.0976 20.1834 11.0976 20.8166 10.7071 21.2071Z">`);
	kt$1 = (e) => {
		let t = () => e.size ?? 10;
		return (() => {
			var n = Ot$1();
			return lt((r) => {
				var i = t(), a = t(), o = e.class;
				return i !== r.e && gn(n, `width`, r.e = i), a !== r.t && gn(n, `height`, r.t = a), o !== r.a && gn(n, `class`, r.a = o), r;
			}, {
				e: void 0,
				t: void 0,
				a: void 0
			}), n;
		})();
	};
	At$1 = /*#__PURE__*/ mn(`<button>`);
	jt$1 = xt$1(`contain-layout shrink-0 flex items-center justify-center cursor-pointer transition-all press-scale`, {
		variants: { variant: {
			chip: `px-[3px] py-px h-[17px] rounded-sm bg-[var(--rg-surface-hover)] [border-width:0.5px] border-solid border-[var(--rg-border-button)] hover:bg-[var(--rg-surface-active)] text-[var(--rg-text-primary)]`,
			destructive: `px-[3px] py-px h-[17px] rounded-sm bg-[var(--rg-error-bg)] hover:bg-[var(--rg-error-bg-hover)] text-[var(--rg-error-text)]`,
			ghost: `rounded-sm bg-transparent hover:bg-[var(--rg-surface-hover)] border-none outline-none p-0`
		} },
		defaultVariants: { variant: `chip` }
	});
	Mt$1 = (e) => {
		let [t, n] = Xt(e, [
			`variant`,
			`class`,
			`type`
		]);
		return (() => {
			var e = At$1();
			return wn(e, Yt({
				get type() {
					return t.type ?? `button`;
				},
				get class() {
					return Z$1(jt$1({ variant: t.variant }), t.class);
				}
			}, n), !1, !1), e;
		})();
	};
	Nt$1 = /*#__PURE__*/ mn(`<span class="text-[var(--rg-text-primary)] text-[13px] leading-3.5 font-sans font-medium">No`);
	Pt$1 = /*#__PURE__*/ mn(`<span class="text-[var(--rg-text-primary)] text-[13px] leading-3.5 font-sans font-medium">Copy`);
	Ft$1 = /*#__PURE__*/ mn(`<span class="text-[var(--rg-error-text)] text-[13px] leading-3.5 font-sans font-medium">Yes`);
	It$1 = /*#__PURE__*/ mn(`<div class="contain-layout shrink-0 flex items-center justify-end gap-[5px] w-full h-fit">`);
	Lt$1 = /*#__PURE__*/ mn(`<div data-react-grab-discard-prompt class="contain-layout shrink-0 flex flex-col justify-center items-end w-fit h-fit"><div class="contain-layout shrink-0 flex items-center gap-1 pt-1.5 pb-1 px-2 w-full h-fit"><span class="text-[var(--rg-text-primary)] text-[13px] leading-4 shrink-0 font-sans font-medium w-fit h-fit">`);
	Rt$1 = (t) => {
		let n = () => t.showCancel ?? !0, { claimFocus: r } = Dt$1({
			onEnter: (e) => {
				e.preventDefault(), e.stopPropagation();
				let n = e.composedPath()[0], r = n instanceof HTMLElement ? n : null;
				r?.closest(`[data-react-grab-discard-copy]`) ? t.onCopy?.() : r?.closest(`[data-react-grab-discard-no]`) ? t.onCancel?.() : t.onConfirm?.();
			},
			onEscape: (e) => {
				e.preventDefault(), e.stopPropagation(), t.cancelOnEscape ? t.onCancel?.() : t.onConfirm?.();
			}
		});
		return (() => {
			var i = Lt$1(), a = i.firstChild.firstChild;
			return bn(i, `click`, r, !0), bn(i, `pointerdown`, r, !0), En(a, () => t.label ?? `Discard?`), En(i, Gt(bt$1, { get children() {
				var r = It$1();
				return En(r, Gt(en, {
					get when() {
						return n();
					},
					get children() {
						return Gt(Mt$1, {
							"data-react-grab-discard-no": !0,
							get onClick() {
								return t.onCancel;
							},
							get children() {
								return Nt$1();
							}
						});
					}
				}), null), En(r, Gt(en, {
					get when() {
						return t.onCopy;
					},
					get children() {
						return Gt(Mt$1, {
							"data-react-grab-discard-copy": !0,
							get onClick() {
								return t.onCopy;
							},
							get children() {
								return Pt$1();
							}
						});
					}
				}), null), En(r, Gt(Mt$1, {
					variant: `destructive`,
					class: `gap-0.5`,
					"data-react-grab-discard-yes": !0,
					get onClick() {
						return t.onConfirm;
					},
					get children() {
						return [Ft$1(), Gt(kt$1, {
							size: 10,
							class: `text-[var(--rg-error-text)] opacity-50`
						})];
					}
				}), null), r;
			} }), null), i;
		})();
	};
	hn([`pointerdown`, `click`]);
	zt$1 = /*#__PURE__*/ mn(`<svg xmlns=http://www.w3.org/2000/svg viewBox="0 0 24 24"fill=currentColor><path fill-rule=evenodd clip-rule=evenodd d="M7 2C4.23858 2 2 4.23858 2 7V15C2 17.7614 4.23858 20 7 20H8C8.55228 20 9 19.5523 9 19C9 18.4477 8.55228 18 8 18H7C5.34315 18 4 16.6569 4 15V7C4 5.34315 5.34315 4 7 4H17C18.6569 4 20 5.34315 20 7V15C20 16.6569 18.6569 18 17 18H14.4142L16.2071 16.2071C16.5976 15.8166 16.5976 15.1834 16.2071 14.7929C15.8166 14.4024 15.1834 14.4024 14.7929 14.7929L11.2929 18.2929C11.197 18.3888 11.1247 18.4993 11.0759 18.6172C11.027 18.7351 11 18.8644 11 19C11 19.2728 11.1093 19.5201 11.2864 19.7005C11.2889 19.7031 11.2914 19.7056 11.2939 19.7081L14.7929 23.2071C15.1834 23.5976 15.8166 23.5976 16.2071 23.2071C16.5976 22.8166 16.5976 22.1834 16.2071 21.7929L14.4142 20H17C19.7614 20 22 17.7614 22 15V7C22 4.23858 19.7614 2 17 2H7Z">`);
	Bt$1 = (e) => {
		let t = () => e.size ?? 12;
		return (() => {
			var n = zt$1();
			return lt((r) => {
				var i = t(), a = t(), o = e.class;
				return i !== r.e && gn(n, `width`, r.e = i), a !== r.t && gn(n, `height`, r.t = a), o !== r.a && gn(n, `class`, r.a = o), r;
			}, {
				e: void 0,
				t: void 0,
				a: void 0
			}), n;
		})();
	};
	Vt$1 = /*#__PURE__*/ mn(`<span class="text-[var(--rg-text-primary)] text-[13px] leading-3.5 font-sans font-medium">Retry`);
	Ht$1 = /*#__PURE__*/ mn(`<span class="text-[var(--rg-text-primary)] text-[13px] leading-3.5 font-sans font-medium">Ok`);
	Ut$1 = /*#__PURE__*/ mn(`<div class="contain-layout shrink-0 flex items-center justify-end gap-[5px] w-full h-fit">`);
	Wt$1 = /*#__PURE__*/ mn(`<div data-react-grab-error role=alert aria-live=assertive class="contain-layout shrink-0 flex flex-col justify-center items-end w-fit h-fit max-w-[280px]"><div class="contain-layout shrink-0 flex items-start gap-1 px-2 w-full h-fit"><span class="text-[var(--rg-error-text)] text-[13px] leading-4 font-sans font-medium overflow-hidden line-clamp-5">`);
	Gt$1 = (t) => {
		let { claimFocus: n } = Dt$1({
			onEnter: (e) => {
				$i(e, `data-react-grab-error-ok`) ? (e.preventDefault(), e.stopPropagation(), t.onAcknowledge?.()) : (e.preventDefault(), e.stopPropagation(), t.onRetry?.());
			},
			onEscape: (e) => {
				e.preventDefault(), e.stopPropagation(), t.onAcknowledge?.();
			}
		}), r = () => !!(t.onRetry || t.onAcknowledge);
		return (() => {
			var i = Wt$1(), a = i.firstChild, s = a.firstChild;
			return bn(i, `click`, n, !0), bn(i, `pointerdown`, n, !0), En(s, () => t.error), En(i, Gt(en, {
				get when() {
					return r();
				},
				get children() {
					return Gt(bt$1, { get children() {
						var n = Ut$1();
						return En(n, Gt(en, {
							get when() {
								return t.onRetry;
							},
							get children() {
								return Gt(Mt$1, {
									"data-react-grab-retry": !0,
									class: `gap-1`,
									"aria-keyshortcuts": `Enter`,
									get onClick() {
										return t.onRetry;
									},
									get children() {
										return [Vt$1(), Gt(Bt$1, {
											size: 10,
											"aria-hidden": `true`,
											class: `text-[var(--rg-text-secondary)]`
										})];
									}
								});
							}
						}), null), En(n, Gt(en, {
							get when() {
								return t.onAcknowledge;
							},
							get children() {
								return Gt(Mt$1, {
									"data-react-grab-error-ok": !0,
									class: `gap-1`,
									"aria-keyshortcuts": `Escape`,
									get onClick() {
										return t.onAcknowledge;
									},
									get children() {
										return Ht$1();
									}
								});
							}
						}), null), n;
					} });
				}
			}), null), lt((e) => {
				var n = {
					"pt-1.5 pb-1": r(),
					"py-1.5": !r()
				}, i = t.error;
				return e.e = xn(a, n, e.e), i !== e.t && gn(s, `title`, e.t = i), e;
			}, {
				e: void 0,
				t: void 0
			}), i;
		})();
	};
	hn([`pointerdown`, `click`]);
	Kt$1 = /*#__PURE__*/ mn(`<svg xmlns=http://www.w3.org/2000/svg viewBox="0 0 24 24"fill=currentColor><path fill-rule=evenodd clip-rule=evenodd d="M7.5 12C7.5 13.3807 6.38071 14.5 5 14.5C3.61929 14.5 2.5 13.3807 2.5 12C2.5 10.6193 3.61929 9.5 5 9.5C6.38071 9.5 7.5 10.6193 7.5 12ZM14.5 12C14.5 13.3807 13.3807 14.5 12 14.5C10.6193 14.5 9.5 13.3807 9.5 12C9.5 10.6193 10.6193 9.5 12 9.5C13.3807 9.5 14.5 10.6193 14.5 12ZM19 14.5C20.3807 14.5 21.5 13.3807 21.5 12C21.5 10.6193 20.3807 9.5 19 9.5C17.6193 9.5 16.5 10.6193 16.5 12C16.5 13.3807 17.6193 14.5 19 14.5Z">`);
	qt$1 = (e) => {
		let t = () => e.size ?? 14;
		return (() => {
			var n = Kt$1();
			return lt((r) => {
				var i = t(), a = t(), o = e.class;
				return i !== r.e && gn(n, `width`, r.e = i), a !== r.t && gn(n, `height`, r.t = a), o !== r.a && gn(n, `class`, r.a = o), r;
			}, {
				e: void 0,
				t: void 0,
				a: void 0
			}), n;
		})();
	};
	Jt$1 = /*#__PURE__*/ mn(`<svg xmlns=http://www.w3.org/2000/svg viewBox="0 0 24 24"fill=currentColor><path fill-rule=evenodd clip-rule=evenodd d="M2 12C2 6.47715 6.47715 2 12 2C17.5228 2 22 6.47715 22 12C22 17.5228 17.5228 22 12 22C6.47715 22 2 17.5228 2 12ZM17.7071 10.7071C18.0976 10.3166 18.0976 9.68342 17.7071 9.29289C17.3166 8.90237 16.6834 8.90237 16.2929 9.29289L11 14.5858L8.70711 12.2929C8.31658 11.9024 7.68342 11.9024 7.29289 12.2929C6.90237 12.6834 6.90237 13.3166 7.29289 13.7071L10.2929 16.7071C10.6834 17.0976 11.3166 17.0976 11.7071 16.7071L17.7071 10.7071Z">`);
	Yt$1 = (e) => {
		let t = () => e.size ?? 14;
		return (() => {
			var n = Jt$1();
			return lt((r) => {
				var i = t(), a = t(), o = e.class;
				return i !== r.e && gn(n, `width`, r.e = i), a !== r.t && gn(n, `height`, r.t = a), o !== r.a && gn(n, `class`, r.a = o), r;
			}, {
				e: void 0,
				t: void 0,
				a: void 0
			}), n;
		})();
	};
	Xt$1 = /*#__PURE__*/ mn(`<button type=button data-react-grab-ignore-events data-react-grab-more-options aria-label="More options">`);
	Zt$1 = /*#__PURE__*/ mn(`<span class="text-[var(--rg-text-primary)] text-[13px] leading-3.5 font-sans font-medium">Keep`);
	Qt$1 = /*#__PURE__*/ mn(`<div class="contain-layout shrink-0 flex items-center justify-between gap-2 pt-1.5 pb-1 px-2 w-full h-fit"><span class="text-[var(--rg-text-primary)] text-[13px] leading-4 font-sans font-medium h-fit tabular-nums overflow-hidden text-ellipsis whitespace-nowrap min-w-0"></span><div class="contain-layout shrink-0 flex items-center gap-2 h-fit">`);
	$t$1 = /*#__PURE__*/ mn(`<div class="contain-layout shrink-0 flex items-center gap-0.5 py-1.5 px-2 w-full h-fit"><span class="text-[var(--rg-text-primary)] text-[13px] leading-4 font-sans font-medium h-fit tabular-nums overflow-hidden text-ellipsis whitespace-nowrap min-w-0">`);
	en$1 = (t) => (() => {
		var n = Xt$1();
		return bn(n, `click`, (e) => {
			e.stopImmediatePropagation(), t.onClick();
		}), bn(n, `pointerdown`, (e) => {
			e.stopImmediatePropagation();
		}), En(n, Gt(qt$1, {
			size: 14,
			"aria-hidden": `true`,
			class: `opacity-50 group-hover:opacity-100 transition-opacity`
		})), lt(() => yn(n, Z$1(jt$1({ variant: `ghost` }), `group size-4 text-[var(--rg-text-secondary)] hover:text-[var(--rg-text-primary)]`))), n;
	})();
	tn$1 = (t) => {
		let r, i, [a, o] = z(!1), [s, c] = z(!1), l = () => a() ? `Copied` : t.statusText, u = () => {
			r !== void 0 && window.clearTimeout(r), i !== void 0 && window.clearTimeout(i), c(!0), t.onFadingChange?.(!0), t.onShowContextMenu?.();
		}, d = () => {
			a() || (o(!0), r = window.setTimeout(() => {
				c(!0), t.onFadingChange?.(!0), i = window.setTimeout(() => {
					t.onDismiss?.();
				}, 125);
			}, 1375));
		}, { claimFocus: f } = Dt$1({
			onEnter: (e) => {
				$i(e, `data-react-grab-more-options`) ? (e.preventDefault(), e.stopPropagation(), u()) : $i(e, `data-react-grab-context-menu`) || (e.preventDefault(), e.stopPropagation(), d());
			},
			onEscape: (e) => {
				e.preventDefault(), e.stopPropagation(), t.onDismiss?.();
			}
		});
		return W(() => {
			r !== void 0 && window.clearTimeout(r), i !== void 0 && window.clearTimeout(i);
		}), Gt(wt$1, {
			shape: `pill`,
			"data-react-grab-completion": !0,
			role: `status`,
			"aria-live": `polite`,
			"aria-atomic": `true`,
			class: `shrink-0 flex flex-col justify-center items-end w-fit h-fit max-w-[280px] transition-opacity duration-100 ease-out`,
			get style() {
				return { opacity: +!s() };
			},
			onPointerDown: f,
			onClick: f,
			get children() {
				return [Gt(en, {
					get when() {
						return dn(() => !a())() && t.onDismiss;
					},
					get children() {
						var n = Qt$1(), r = n.firstChild, i = r.nextSibling;
						return En(r, l), En(i, Gt(en, {
							get when() {
								return t.onShowContextMenu;
							},
							get children() {
								return Gt(en$1, { onClick: u });
							}
						}), null), En(i, Gt(en, {
							get when() {
								return t.onDismiss;
							},
							get children() {
								return Gt(Mt$1, {
									"data-react-grab-dismiss": !0,
									class: `gap-1`,
									"aria-keyshortcuts": `Enter`,
									onClick: d,
									get disabled() {
										return a();
									},
									get "aria-disabled"() {
										return a();
									},
									get children() {
										return [Zt$1(), Gt(en, {
											get when() {
												return !a();
											},
											get children() {
												return Gt(kt$1, {
													size: 10,
													class: `text-[var(--rg-text-secondary)]`
												});
											}
										})];
									}
								});
							}
						}), null), n;
					}
				}), Gt(en, {
					get when() {
						return a() || !t.onDismiss;
					},
					get children() {
						var n = $t$1(), r = n.firstChild;
						return En(n, Gt(Yt$1, {
							size: 14,
							"aria-hidden": `true`,
							class: `text-[var(--rg-text-primary-85)] shrink-0`
						}), r), En(r, l), En(n, Gt(en, {
							get when() {
								return t.onShowContextMenu;
							},
							get children() {
								return Gt(en$1, { onClick: u });
							}
						}), null), n;
					}
				})];
			}
		});
	};
	nn$1 = /*#__PURE__*/ mn(`<div class="contain-layout shrink-0 flex flex-col justify-center items-start w-fit h-fit max-w-[280px]"><div class="contain-layout shrink-0 flex items-center gap-1 py-1.5 px-2 w-full h-fit"><span class="shimmer-text text-[13px] leading-4 font-sans font-medium h-fit tabular-nums overflow-hidden text-ellipsis whitespace-nowrap">`);
	rn$1 = /*#__PURE__*/ mn(`<div class="contain-layout shrink-0 flex flex-col items-start w-fit h-fit"><div class="contain-layout shrink-0 flex items-center gap-1 w-fit h-fit px-2 py-1.5">`);
	an$1 = /*#__PURE__*/ mn(`<button data-react-grab-submit type=button aria-label="Submit context"class="contain-layout shrink-0 flex items-center justify-center size-4 rounded-full bg-[var(--rg-submit-bg)] cursor-pointer ml-1 interactive-scale a11y-hitbox">`);
	on$1 = /*#__PURE__*/ mn(`<div class="shrink-0 flex justify-between items-end w-full min-h-4"><textarea data-react-grab-ignore-events data-react-grab-input aria-label="Add context for selected element"aria-keyshortcuts="Enter Escape"class="text-[var(--rg-text-primary)] text-[13px] leading-4 font-medium bg-transparent border-none resize-none flex-1 p-0 m-0 wrap-break-word overflow-y-auto"placeholder="Add context"rows=1 style=field-sizing:content;min-height:16px;scrollbar-width:none>`);
	sn$1 = /*#__PURE__*/ mn(`<div class="contain-layout shrink-0 flex flex-col justify-center items-start w-fit h-fit min-w-[150px] max-w-[280px]"><div class="contain-layout shrink-0 flex items-center gap-1 pt-1.5 pb-1 w-fit h-fit px-2 max-w-full">`);
	cn$1 = /*#__PURE__*/ mn(`<div data-react-grab-ignore-events data-react-grab-selection-label class="fixed font-sans text-[13px] antialiased select-none">`);
	ln$1 = {
		left: tt$2,
		top: tt$2,
		arrowLeftPercent: 50,
		arrowLeftOffset: 0,
		edgeOffsetX: 0
	};
	un$1 = (t) => {
		let r, i, a, o = !1, [s, l] = z(0), [u, f] = z(0), [p, h] = z(0), [g, _] = z(0), [v, y] = z(!1), [b, x] = z(!1), S = () => t.status !== `copying` && t.status !== `copied` && t.status !== `fading` && t.status !== `error`, C = () => t.status === `copied` || t.status === `fading`, T = () => !!t.shouldToggleExpandOnClick && !!t.onToggleExpand && S() && !t.discardPrompt && !t.isPromptMode, E = () => !!(t.isPromptMode || t.discardPrompt || T() || C() && (t.onDismiss || t.onShowContextMenu) || t.status === `error` && (t.onAcknowledgeError || t.onRetry)), D, O = (e) => {
			o = e;
		}, k = () => {
			_((e) => e + 1);
		}, ee = zt$2((e) => {
			yr(e) || e.code === `Enter` && !t.isPromptMode && !t.discardPrompt && S() && (e.preventDefault(), e.stopImmediatePropagation(), t.onToggleExpand?.());
		});
		pt(() => {
			let e = Lt$2();
			if (D = new ResizeObserver((t) => {
				for (let n of t) {
					let t = n.target.getBoundingClientRect();
					n.target === r && !o ? (l(t.width), f(t.height)) : n.target === i ? h(t.width) : n.target === e && k();
				}
			}), e && D.observe(e), r) {
				let e = r.getBoundingClientRect();
				l(e.width), f(e.height), D.observe(r);
			}
			i && (h(i.getBoundingClientRect().width), D.observe(i)), window.addEventListener(`scroll`, k, !0), window.addEventListener(`resize`, k), window.visualViewport?.addEventListener(`resize`, k), window.visualViewport?.addEventListener(`scroll`, k), t.onToggleExpand && window.addEventListener(`keydown`, ee, { capture: !0 });
		}), W(() => {
			D?.disconnect(), window.removeEventListener(`scroll`, k, !0), window.removeEventListener(`resize`, k), window.visualViewport?.removeEventListener(`resize`, k), window.visualViewport?.removeEventListener(`scroll`, k), window.removeEventListener(`keydown`, ee, { capture: !0 });
		});
		let te = () => `${t.tagName ?? ``}:${t.componentName ?? ``}`, j = V((e) => {
			g();
			let n = te(), r = n === e.elementIdentity ? e : {
				position: ln$1,
				computedArrowPosition: null,
				hadValidBounds: !1,
				elementIdentity: n
			}, i = t.selectionBounds, a = s(), o = u(), c = a > 0 && o > 0, l = i && i.width > 0 && i.height > 0;
			if (!c || !l) return {
				position: r.hadValidBounds ? r.position : ln$1,
				computedArrowPosition: r.computedArrowPosition,
				hadValidBounds: r.hadValidBounds,
				elementIdentity: n
			};
			let d = nt$1(), f = d.offsetLeft, m = d.offsetTop, h = f + d.width, _ = m + d.height;
			if (!(i.x + i.width > f && i.x < h && i.y + i.height > m && i.y < _)) return {
				position: ln$1,
				computedArrowPosition: r.computedArrowPosition,
				hadValidBounds: r.hadValidBounds,
				elementIdentity: n
			};
			let v = i.x + i.width / 2, y = t.mouseX ?? v, b = i.y + i.height, x = i.y, S = t.hideArrow ? 0 : tt$1(p()), C = y, w = 0, T = b + S + 4, E = C - a / 2, D = C + a / 2;
			D > h - 8 && (w = h - 8 - D), E + w < f + 8 && (w = f + 8 - E);
			let O = o + S + 4, k = T + o <= _ - 8;
			k || (T = x - O), T < m + 8 && (T = m + 8);
			let A = a / 2, ee = A - w, j = Math.min(16, A), ne = Math.max(a - 16, A), M = Math.max(j, Math.min(ne, ee)) - A;
			return {
				position: {
					left: C,
					top: T,
					arrowLeftPercent: 50,
					arrowLeftOffset: M,
					edgeOffsetX: w
				},
				computedArrowPosition: k ? `bottom` : `top`,
				hadValidBounds: !0,
				elementIdentity: n
			};
		}, {
			position: ln$1,
			computedArrowPosition: null,
			hadValidBounds: !1,
			elementIdentity: ``
		}), ne = () => j().computedArrowPosition ?? `bottom`, N = () => j().hadValidBounds;
		B(ft(() => t.selectionLabelShakeCount, () => x(!0), { defer: !0 }));
		let re$3 = (e) => {
			if (rt$1(e)) return;
			e.stopImmediatePropagation();
			let n = e.code === `Enter` && !e.shiftKey, r = e.code === `Escape`;
			n ? (e.preventDefault(), t.onSubmit?.()) : r && (e.preventDefault(), t.onConfirmDismiss?.());
		}, P = (e) => {
			let n = e.target;
			n instanceof HTMLTextAreaElement && ($e$1(n, 95), t.onInputChange?.(n.value));
		}, ie = () => at$1({
			tagName: t.tagName,
			componentName: t.componentName,
			elementsCount: t.elementsCount
		}), ae = V(() => !(t.error || t.discardPrompt || S() && t.isPromptMode)), oe = (e) => {
			e.stopImmediatePropagation(), t.filePath && t.onOpen ? t.onOpen() : T() && t.onToggleExpand?.();
		}, B$5 = (e) => {
			e.stopImmediatePropagation(), S() && t.isPromptMode && !t.discardPrompt && t.onSubmit && a && et$1(a, { preventScroll: !0 });
		}, se = () => N() && (C() || t.status === `error`);
		return Gt(en, {
			get when() {
				return dn(() => t.visible !== !1)() && (t.selectionBounds || se());
			},
			get children() {
				var o = cn$1();
				o.addEventListener(`mouseleave`, () => t.onHoverChange?.(!1)), o.addEventListener(`mouseenter`, () => t.onHoverChange?.(!0)), o.$$click = (e) => {
					e.stopImmediatePropagation();
				}, o.$$pointerdown = B$5;
				var s = r;
				return typeof s == `function` ? Tn(s, o) : r = o, Cn(o, `z-index`, `${ce}`), Cn(o, `transition`, `opacity 125ms ease-out, filter 125ms ease-out`), En(o, Gt(en, {
					get when() {
						return !t.hideArrow;
					},
					get children() {
						return Gt(dt$1, {
							get position() {
								return ne();
							},
							get leftPercent() {
								return j().position.arrowLeftPercent;
							},
							get leftOffsetPx() {
								return j().position.arrowLeftOffset;
							},
							get labelWidth() {
								return p();
							}
						});
					}
				}), null), En(o, Gt(en, {
					get when() {
						return dn(() => !!C())() && !t.error;
					},
					get children() {
						return Gt(tn$1, {
							get statusText() {
								return t.statusText ?? `Copied`;
							},
							get onDismiss() {
								return t.onDismiss;
							},
							onFadingChange: y,
							get onShowContextMenu() {
								return t.onShowContextMenu;
							}
						});
					}
				}), null), En(o, Gt(wt$1, {
					ref: (e) => i = e,
					get shape() {
						return ae() ? `pill` : `panel`;
					},
					get class() {
						return Z$1(`flex items-center gap-[5px] w-fit h-fit p-0`, b() && `animate-shake`);
					},
					get style() {
						return { display: C() && !t.error ? `none` : void 0 };
					},
					onAnimationEnd: () => x(!1),
					get children() {
						return [
							Gt(en, {
								get when() {
									return t.status === `copying`;
								},
								get children() {
									var n = nn$1(), r = n.firstChild, i = r.firstChild;
									return En(r, Gt(lt$1, {
										size: 13,
										class: `text-[var(--rg-text-secondary)] shrink-0`
									}), i), En(i, () => t.statusText ?? `Grabbing…`), n;
								}
							}),
							Gt(en, {
								get when() {
									return dn(() => !(!S() || t.isPromptMode))() && !t.discardPrompt;
								},
								get children() {
									var n = rn$1(), r = n.firstChild;
									return En(r, Gt(vt$1, {
										get tagName() {
											return ie().tagName;
										},
										get componentName() {
											return ie().componentName;
										},
										get isClickable() {
											return !!(t.filePath && t.onOpen);
										},
										onClick: oe,
										onHoverChange: O,
										shrink: !0
									})), n;
								}
							}),
							Gt(en, {
								get when() {
									return dn(() => !!(S() && t.isPromptMode))() && !t.discardPrompt;
								},
								get children() {
									var n = sn$1(), r = n.firstChild;
									return En(r, Gt(vt$1, {
										get tagName() {
											return ie().tagName;
										},
										get componentName() {
											return ie().componentName;
										},
										get isClickable() {
											return !!(t.filePath && t.onOpen);
										},
										onClick: oe,
										onHoverChange: O
									})), En(n, Gt(bt$1, { get children() {
										var n = on$1(), r = n.firstChild;
										return r.$$keydown = re$3, r.$$input = P, Tn((e) => {
											a = e, t.onSubmit && queueMicrotask(() => {
												et$1(e, { preventScroll: !0 }), $e$1(e, 95);
											});
										}, r), Cn(r, `max-height`, `95px`), En(n, Gt(en, {
											get when() {
												return t.onSubmit;
											},
											get children() {
												var n = an$1();
												return n.$$click = () => t.onSubmit?.(), En(n, Gt(st$1, {
													size: 10,
													"aria-hidden": `true`,
													class: `text-[var(--rg-submit-fg)]`
												})), n;
											}
										}), null), lt(() => r.readOnly = !t.onSubmit), lt(() => r.value = t.inputValue ?? ``), n;
									} }), null), n;
								}
							}),
							Gt(en, {
								get when() {
									return t.discardPrompt;
								},
								keyed: !0,
								children: (t) => Gt(Rt$1, {
									get label() {
										return dn(() => !!t.isKeyboardSelection)() ? `Discard selection?` : t.label;
									},
									get showCancel() {
										return !t.isKeyboardSelection;
									},
									get cancelOnEscape() {
										return t.cancelOnEscape;
									},
									get onConfirm() {
										return t.onConfirm;
									},
									get onCopy() {
										return t.onCopy;
									},
									onCancel: () => {
										t.isKeyboardSelection || t.onCancel?.(), et$1(a, { preventScroll: !0 });
									}
								})
							}),
							Gt(en, {
								get when() {
									return t.error;
								},
								get children() {
									return Gt(Gt$1, {
										get error() {
											return t.error;
										},
										get onAcknowledge() {
											return t.onAcknowledgeError;
										},
										get onRetry() {
											return t.onRetry;
										}
									});
								}
							})
						];
					}
				}), null), lt((e) => {
					var n = `${j().position.top}px`, r = `${j().position.left}px`, i = `translateX(calc(-50% + ${j().position.edgeOffsetX}px))`, a = E() ? `auto` : `none`, s = t.status === `fading` || v() ? 0 : 1, l = `drop-shadow(${re}) blur(${t.status === `fading` || v() ? `3px` : `0`})`;
					return n !== e.e && Cn(o, `top`, e.e = n), r !== e.t && Cn(o, `left`, e.t = r), i !== e.a && Cn(o, `transform`, e.a = i), a !== e.o && Cn(o, `pointer-events`, e.o = a), s !== e.i && Cn(o, `opacity`, e.i = s), l !== e.n && Cn(o, `filter`, e.n = l), e;
				}, {
					e: void 0,
					t: void 0,
					a: void 0,
					o: void 0,
					i: void 0,
					n: void 0
				}), o;
			}
		});
	};
	hn([
		`pointerdown`,
		`click`,
		`input`,
		`keydown`
	]);
	dn$1 = /*#__PURE__*/ mn(`<span style="transition-property:transform;transition-timing-function:cubic-bezier(0.32, 0.72, 0, 1)"><svg xmlns=http://www.w3.org/2000/svg viewBox="0 0 24 24"fill=currentColor><path fill-rule=evenodd clip-rule=evenodd d="M20.8977 4.02356L21.8277 4.39121C22.3784 2.99813 21.0382 1.60206 19.6238 2.09546L3.47334 7.72936C1.38661 8.45728 1.49021 11.443 3.6224 12.0245L10.1289 13.799L11.2331 19.8724C11.638 22.0991 14.7072 22.4019 15.5393 20.2972L21.8277 4.39121L20.8977 4.02356Z">`);
	fn$1 = (e) => {
		let t = () => e.size ?? 14, n = () => e.rotationDeg ?? 0;
		return (() => {
			var r = dn$1(), i = r.firstChild;
			return Cn(r, `transition-duration`, `180ms`), lt((a) => {
				var o = Z$1(`inline-flex items-center justify-center will-change-transform`, e.class), s = `rotate(${n()}deg)`, l = t(), u = t();
				return o !== a.e && yn(r, a.e = o), s !== a.t && Cn(r, `transform`, a.t = s), l !== a.a && gn(i, `width`, a.a = l), u !== a.o && gn(i, `height`, a.o = u), a;
			}, {
				e: void 0,
				t: void 0,
				a: void 0,
				o: void 0
			}), r;
		})();
	};
	pn$1 = /*#__PURE__*/ mn(`<svg xmlns=http://www.w3.org/2000/svg viewBox="0 0 12 12"fill=currentColor><path fill-rule=evenodd clip-rule=evenodd d="M1 6C1 3.239 3.239 1 6 1C8.761 1 11 3.239 11 6C11 8.761 8.761 11 6 11C5.103 11 4.261 10.764 3.532 10.35C2.959 10.759 2.257 11 1.5 11C1.376 11 1.253 10.993 1.132 10.981C0.923 10.959 0.75 10.809 0.699 10.605C0.648 10.402 0.729 10.188 0.903 10.07C1.347 9.768 1.688 9.328 1.865 8.812C1.319 8.011 1 7.042 1 6Z">`);
	mn$1 = (e) => {
		let t = () => e.size ?? 14;
		return (() => {
			var n = pn$1();
			return lt((r) => {
				var i = t(), a = t(), o = e.class;
				return i !== r.e && gn(n, `width`, r.e = i), a !== r.t && gn(n, `height`, r.t = a), o !== r.a && gn(n, `class`, r.a = o), r;
			}, {
				e: void 0,
				t: void 0,
				a: void 0
			}), n;
		})();
	};
	hn$1 = /*#__PURE__*/ mn(`<svg xmlns=http://www.w3.org/2000/svg viewBox="0 0 24 24"fill=currentColor><path fill-rule=evenodd clip-rule=evenodd d="M2 12C2 6.477 6.477 2 12 2C17.523 2 22 6.477 22 12C22 13.867 20.345 14.992 18.77 14.992H16.029C14.739 14.992 13.823 16.248 14.218 17.476L14.48 18.295C14.751 19.139 14.643 20.041 14.211 20.751C13.769 21.479 12.985 22 12 22C6.477 22 2 17.523 2 12ZM12.21 6.783C12.416 7.585 11.933 8.403 11.131 8.609C10.328 8.815 9.511 8.332 9.305 7.529C9.099 6.727 9.582 5.909 10.384 5.703C11.187 5.497 12.004 5.981 12.21 6.783ZM7.59 9.207C8.388 9.43 8.853 10.258 8.63 11.055C8.407 11.853 7.579 12.319 6.782 12.096C5.984 11.873 5.518 11.045 5.741 10.247C5.964 9.449 6.792 8.983 7.59 9.207ZM9.497 14.446C8.918 13.854 7.968 13.844 7.376 14.423C6.784 15.002 6.773 15.952 7.353 16.544C7.932 17.136 8.882 17.147 9.474 16.567C10.066 15.988 10.076 15.039 9.497 14.446ZM16.624 9.575C16.032 10.155 15.082 10.144 14.503 9.552C13.923 8.96 13.934 8.01 14.526 7.431C15.118 6.852 16.068 6.862 16.647 7.454C17.226 8.046 17.216 8.996 16.624 9.575Z">`);
	gn$1 = (e) => {
		let t = () => e.size ?? 14;
		return (() => {
			var n = hn$1();
			return lt((r) => {
				var i = t(), a = t(), o = e.class;
				return i !== r.e && gn(n, `width`, r.e = i), a !== r.t && gn(n, `height`, r.t = a), o !== r.a && gn(n, `class`, r.a = o), r;
			}, {
				e: void 0,
				t: void 0,
				a: void 0
			}), n;
		})();
	};
	_n$1 = /*#__PURE__*/ mn(`<div>`);
	vn$1 = 0;
	yn$1 = () => Date.now() - vn$1 < 800;
	bn$1 = (t) => {
		let [n, r] = z(!1), [i, a] = z(!0), o;
		B(ft(() => t.visible, (e) => {
			o !== void 0 && (clearTimeout(o), o = void 0), e ? yn$1() ? (a(!1), r(!0)) : (a(!0), o = setTimeout(() => {
				r(!0);
			}, 400)) : (n() && (vn$1 = Date.now()), r(!1));
		})), W(() => {
			o !== void 0 && clearTimeout(o), n() && (vn$1 = Date.now());
		});
		let s = () => t.position === `top` || t.position === `bottom` ? {
			left: `50%`,
			translate: `-50%`,
			"z-index": `${ce}`
		} : {
			top: `50%`,
			translate: `0 -50%`,
			"z-index": `${ce}`
		};
		return Gt(en, {
			get when() {
				return n();
			},
			get children() {
				var e = _n$1();
				return En(e, () => t.children), lt((n) => {
					var r = Z$1(`absolute whitespace-nowrap px-2 py-0.5 rounded-full text-[10px] font-sans font-medium leading-4 pointer-events-none`, `bg-[var(--rg-panel-bg)] text-[var(--rg-text-primary)] [box-shadow:var(--rg-shadow)]`, t.position === `top` && `bottom-full mb-2.5`, t.position === `bottom` && `top-full mt-2.5`, t.position === `left` && `right-full mr-2.5`, t.position === `right` && `left-full ml-2.5`, i() && `animate-tooltip-fade-in`), a = s();
					return r !== n.e && yn(e, n.e = r), n.t = Sn(e, a, n.t), n;
				}, {
					e: void 0,
					t: void 0
				}), e;
			}
		});
	};
	xn$1 = /*#__PURE__*/ mn(`<div><button data-react-grab-ignore-events type=button>`);
	Sn$1 = (t) => (() => {
		var n = xn$1(), r = n.firstChild;
		bn(r, `mouseleave`, t.onMouseLeave), bn(r, `mouseenter`, t.onMouseEnter), bn(r, `contextmenu`, (e) => t.onContextMenu?.(e)), bn(r, `click`, t.onClick, !0);
		var i = t.ref;
		return typeof i == `function` ? Tn(i, r) : t.ref = r, En(r, () => t.icon), En(n, Gt(en, {
			get when() {
				return t.tooltip;
			},
			get children() {
				return Gt(bn$1, {
					get visible() {
						return !!t.tooltipVisible;
					},
					get position() {
						return t.tooltipPosition ?? `top`;
					},
					get children() {
						return t.tooltip;
					}
				});
			}
		}), null), lt((e) => {
			var i = t.wrapperClass, a = t.isToggle ? `` : void 0, o = t.actionId, s = t.label, c = !!t.isActive, l = t.class;
			return i !== e.e && yn(n, e.e = i), a !== e.t && gn(r, `data-react-grab-toolbar-toggle`, e.t = a), o !== e.a && gn(r, `data-react-grab-toolbar-action`, e.a = o), s !== e.o && gn(r, `aria-label`, e.o = s), c !== e.i && gn(r, `aria-pressed`, e.i = c), l !== e.n && yn(r, e.n = l), e;
		}, {
			e: void 0,
			t: void 0,
			a: void 0,
			o: void 0,
			i: void 0,
			n: void 0
		}), n;
	})();
	hn([`click`]);
	Cn$1 = (e) => e === `top` || e === `bottom`;
	wn$1 = (e) => {
		let t = Cn$1(e);
		return {
			width: t ? 30 : 16,
			height: t ? 16 : 30
		};
	};
	Tn$1 = (e, t, n, r) => {
		let i = nt$1(), a = i.width, o = i.height, s = i.offsetLeft + 16, c = Math.max(s, i.offsetLeft + a - n - 16), l = i.offsetTop + 16, u = Math.max(l, i.offsetTop + o - r - 16);
		if (Cn$1(e)) {
			let r = Math.max(0, a - n - 32);
			return {
				x: Math.min(c, Math.max(s, i.offsetLeft + 16 + r * t)),
				y: e === `top` ? l : u
			};
		}
		let d = Math.max(0, o - r - 32), f = Math.min(u, Math.max(l, i.offsetTop + 16 + d * t));
		return {
			x: e === `left` ? s : c,
			y: f
		};
	};
	En$1 = (e, t, n, r, i) => {
		let a = nt$1(), o = a.width, s = a.height;
		if (Cn$1(e)) {
			let e = o - r - 32;
			return e <= 0 ? Me : Math.max(0, Math.min(1, (t - a.offsetLeft - 16) / e));
		}
		let c = s - i - 32;
		return c <= 0 ? Me : Math.max(0, Math.min(1, (n - a.offsetTop - 16) / c));
	};
	Dn$1 = (e, t, n, r, i) => {
		let a = nt$1(), o = a.width, s = a.height, { width: c, height: l } = n, u;
		if (Cn$1(t)) {
			let n = (c - r) / 2, i = e.x - n;
			u = {
				x: Ci(i, a.offsetLeft + 16, a.offsetLeft + o - c - 16),
				y: t === `top` ? a.offsetTop + 16 : a.offsetTop + s - l - 16
			};
		} else {
			let n = (l - i) / 2, r = e.y - n, d = Ci(r, a.offsetTop + 16, a.offsetTop + s - l - 16);
			u = {
				x: t === `left` ? a.offsetLeft + 16 : a.offsetLeft + o - c - 16,
				y: d
			};
		}
		let d = En$1(t, u.x, u.y, c, l);
		return {
			position: u,
			ratio: d
		};
	};
	On$1 = (e, t, n, r) => {
		let i = nt$1(), { width: a, height: o } = n, { width: s, height: c } = r;
		switch (e) {
			case `top`:
			case `bottom`: {
				let n = (a - s) / 2, r = t.x + n;
				return {
					x: Ci(r, i.offsetLeft, i.offsetLeft + i.width - s),
					y: e === `top` ? i.offsetTop : i.offsetTop + i.height - c
				};
			}
			case `left`:
			case `right`: {
				let n = (o - c) / 2, r = t.y + n, a = Ci(r, i.offsetTop, i.offsetTop + i.height - c);
				return {
					x: e === `left` ? i.offsetLeft : i.offsetLeft + i.width - s,
					y: a
				};
			}
		}
	};
	kn$1 = (e, t, n, r, i, a) => {
		let o = nt$1(), s = o.width, c = o.height, l = e + i * 150, u = t + a * 150, d = u - o.offsetTop + r / 2, f = o.offsetTop + c - u - r / 2, p = l - o.offsetLeft + n / 2, m = o.offsetLeft + s - l - n / 2, h = Math.min(d, f, p, m), g = (e) => Ci(e, o.offsetLeft + 16, o.offsetLeft + s - n - 16), _ = (e) => Ci(e, o.offsetTop + 16, o.offsetTop + c - r - 16);
		return h === d ? {
			edge: `top`,
			x: g(l),
			y: o.offsetTop + 16
		} : h === p ? {
			edge: `left`,
			x: o.offsetLeft + 16,
			y: _(u)
		} : h === m ? {
			edge: `right`,
			x: o.offsetLeft + s - n - 16,
			y: _(u)
		} : {
			edge: `bottom`,
			x: g(l),
			y: o.offsetTop + c - r - 16
		};
	};
	An$1 = /*#__PURE__*/ mn(`<svg xmlns=http://www.w3.org/2000/svg viewBox="0 0 24 24"fill=currentColor><path fill-rule=evenodd clip-rule=evenodd d="M9.29289 16.7071C8.90237 16.3166 8.90237 15.6834 9.29289 15.2929L12.5858 12L9.29289 8.70711C8.90237 8.31658 8.90237 7.68342 9.29289 7.29289C9.68342 6.90237 10.3166 6.90237 10.7071 7.29289L14.7071 11.2929C14.8946 11.4804 15 11.7348 15 12C15 12.2652 14.8946 12.5196 14.7071 12.7071L10.7071 16.7071C10.3166 17.0976 9.68342 17.0976 9.29289 16.7071Z">`);
	jn$1 = (e) => {
		let t = () => e.size ?? 14;
		return (() => {
			var n = An$1();
			return lt((r) => {
				var i = t(), a = t(), o = e.class;
				return i !== r.e && gn(n, `width`, r.e = i), a !== r.t && gn(n, `height`, r.t = a), o !== r.a && gn(n, `class`, r.a = o), r;
			}, {
				e: void 0,
				t: void 0,
				a: void 0
			}), n;
		})();
	};
	Mn$1 = /*#__PURE__*/ mn(`<div data-react-grab-toolbar-panel><div><div><div></div></div></div><button data-react-grab-ignore-events data-react-grab-toolbar-collapse type=button class="group contain-layout shrink-0 flex items-center justify-center cursor-pointer interactive-scale a11y-hitbox">`);
	Nn$1 = (t) => {
		let n = () => t.snapEdge ?? `bottom`, r = () => !Cn$1(n()), i = () => t.isCollapsed ? `duration-140` : `duration-220`, a = () => r() ? `transition-[grid-template-rows] ${i()} ease-drawer` : `transition-[grid-template-columns] ${i()} ease-drawer`, s = () => r() ? `min-h-0` : `min-w-0`, l = () => t.isCollapsed ? `${{
			top: `rounded-t-none rounded-b-[10px]`,
			bottom: `rounded-b-none rounded-t-[10px]`,
			left: `rounded-l-none rounded-r-[10px]`,
			right: `rounded-r-none rounded-l-[10px]`
		}[n()]} ${r() ? `px-0.25 py-2` : `px-2 py-0.25`}` : ``, u = () => {
			let e = t.isCollapsed;
			switch (n()) {
				case `top`: return e ? `rotate-90` : `-rotate-90`;
				case `bottom`: return e ? `-rotate-90` : `rotate-90`;
				case `left`: return e ? `rotate-0` : `rotate-180`;
				case `right`: return e ? `rotate-180` : `rotate-0`;
				default: return `-rotate-90`;
			}
		}, d = () => {
			if (t.isChevronPressed) return r() ? `scale(0.97, 1)` : `scale(1, 0.97)`;
		}, p = () => t.isChevronPressed ? `transition-[padding,border-radius,transform] duration-60 ease-[cubic-bezier(0,0,0.2,1)]` : `transition-[padding,border-radius,transform] ${i()} ease-drawer`;
		return (() => {
			var n = Mn$1(), i = n.firstChild, m = i.firstChild, h = m.firstChild, g = i.nextSibling;
			return bn(n, `click`, t.onPanelClick, !0), bn(n, `animationend`, t.onAnimationEnd), En(h, () => t.actionButtons), bn(g, `pointercancel`, t.onCollapsePointerLeave), bn(g, `pointerleave`, t.onCollapsePointerLeave), bn(g, `pointerup`, t.onCollapsePointerUp, !0), bn(g, `pointerdown`, t.onCollapsePointerDown), bn(g, `click`, t.onCollapseClick, !0), En(g, Gt(jn$1, {
				size: 18,
				get class() {
					return Z$1(`text-[var(--rg-text-secondary)] group-hover:text-[var(--rg-text-primary)] transition-[transform,color] duration-150 ease-drawer -m-0.5`, u());
				}
			})), lt((e) => {
				var o = Z$1(`flex items-center justify-center rounded-[13px] antialiased relative overflow-visible [font-synthesis:none]`, p(), r() && `flex-col`, `bg-[var(--rg-panel-bg)] [box-shadow:var(--rg-shadow)]`, !t.isCollapsed && (r() ? `px-1.5 gap-0 py-2` : `py-1.5 gap-0 px-2`), l(), t.isShaking && (r() ? `animate-shake-vertical` : `animate-shake`)), u = t.transformOrigin, _ = d(), v = Z$1(`grid relative overflow-visible`, a(), t.isCollapsed ? r() ? `grid-rows-[0fr] pointer-events-none` : `grid-cols-[0fr] pointer-events-none` : r() ? `grid-rows-[1fr]` : `grid-cols-[1fr]`), y = Z$1(`flex`, r() ? `flex-col items-center min-h-0` : `items-center min-w-0`, t.isCollapsed ? `opacity-0` : `opacity-100`, t.isCollapsed ? `transition-opacity duration-100 ease-drawer` : `transition-opacity duration-180 ease-drawer delay-[80ms]`), b = Z$1(`relative overflow-visible flex`, r() ? `flex-col items-center` : `items-center`, s()), x = t.isCollapsed ? `Expand toolbar` : `Collapse toolbar`, C = !t.isCollapsed;
				return o !== e.e && yn(n, e.e = o), u !== e.t && Cn(n, `transform-origin`, e.t = u), _ !== e.a && Cn(n, `transform`, e.a = _), v !== e.o && yn(i, e.o = v), y !== e.i && yn(m, e.i = y), b !== e.n && yn(h, e.n = b), x !== e.s && gn(g, `aria-label`, e.s = x), C !== e.h && gn(g, `aria-expanded`, e.h = C), e;
			}, {
				e: void 0,
				t: void 0,
				a: void 0,
				o: void 0,
				i: void 0,
				n: void 0,
				s: void 0,
				h: void 0
			}), n;
		})();
	};
	hn([`click`, `pointerup`]);
	Pn$1 = (e) => {
		let [t, n] = z(!1), [r, i] = z(!1), [a, o] = z(!1), [s, c] = z({
			x: 0,
			y: 0
		}), l = {
			x: 0,
			y: 0
		}, u = {
			x: 0,
			y: 0,
			time: 0
		}, d = {
			x: 0,
			y: 0
		}, f = !1, p, m, h = null, g = () => {
			h?.abort(), h = null;
		}, _ = () => {
			p !== void 0 && (ci$2(p), p = void 0);
		}, v = (t) => {
			if (!a()) {
				if (Math.hypot(t.clientX - d.x, t.clientY - d.y) <= 5) return;
				o(!0), e.onDragStart();
			}
			let n = performance.now(), r = n - u.time;
			if (r > 0) {
				let e = (t.clientX - u.x) / r, n = (t.clientY - u.y) / r;
				c({
					x: e,
					y: n
				});
			}
			u = {
				x: t.clientX,
				y: t.clientY,
				time: n
			};
			let i = t.clientX - l.x, s = t.clientY - l.y;
			e.onPositionUpdate({
				x: i,
				y: s
			});
		}, y = () => {
			g();
			let t = a();
			if (n(!1), !t) return;
			f = !0;
			let r = e.getContainerRef(), o = r?.getBoundingClientRect();
			if (!o) return;
			let c = s(), l = kn$1(o.left, o.top, o.width, o.height, c.x, c.y), u = En$1(l.edge, l.x, l.y, o.width, o.height);
			e.onSnapEdgeChange(l.edge, u), i(!0), _(), p = M$1(() => {
				let t = r?.getBoundingClientRect(), n = t ? {
					width: t.width,
					height: t.height
				} : e.getExpandedDimensions();
				p = M$1(() => {
					p = void 0;
					let t = Tn$1(l.edge, u, n.width, n.height);
					e.onSnapComplete({
						edge: l.edge,
						ratio: u,
						position: t,
						expandedDimensions: n
					}), m = setTimeout(() => {
						i(!1), e.onSnapAnimationEnd?.();
					}, 300);
				});
			});
		}, b = zt$2((t) => {
			if (t.button !== 0 || e.isCollapsed() || r()) return;
			let i = e.getContainerRef()?.getBoundingClientRect();
			if (!i) return;
			d = {
				x: t.clientX,
				y: t.clientY
			}, l = {
				x: t.clientX - i.left,
				y: t.clientY - i.top
			}, n(!0), o(!1), c({
				x: 0,
				y: 0
			}), u = {
				x: t.clientX,
				y: t.clientY,
				time: performance.now()
			}, g(), h = new AbortController();
			let { signal: a } = h;
			window.addEventListener(`pointermove`, v, { signal: a }), window.addEventListener(`pointerup`, y, { signal: a }), window.addEventListener(`pointercancel`, y, { signal: a });
		});
		return W(() => {
			g(), _(), clearTimeout(m);
		}), {
			isDragging: t,
			isSnapping: r,
			handlePointerDown: b,
			createDragAwareHandler: (e) => (t) => {
				t.stopImmediatePropagation(), f ? f = !1 : e();
			}
		};
	};
	Fn$1 = (e, t) => {
		let n = t - e;
		return e + (n - Math.round(n / 360) * 360);
	};
	In$1 = /*#__PURE__*/ mn(`<div data-react-grab-ignore-events data-react-grab-toolbar>`);
	Ln$1 = /*#__PURE__*/ mn(`<span>`);
	Rn$1 = (e, t) => {
		e.innerHTML = t;
		let n = e.querySelector(`svg`);
		if (!n) throw Error(`toolbar button icon must be SVG markup`);
		n.setAttribute(`width`, `14`), n.setAttribute(`height`, `14`), n.setAttribute(`aria-hidden`, `true`);
	};
	zn$1 = (t) => {
		let n, r, i = null, a = fs(), [s, l] = z(!1), [u, p] = z(!1), [m, h] = z(!1), [g, _] = z(a?.edge ?? `bottom`), [v, y] = z(a?.ratio ?? .5), [b, x] = z({
			x: 0,
			y: 0
		}), [S, C] = z(!1), [w, T] = z(!1), [E, O] = z(!1), [k, ee] = z(!1), [te, N] = z(0), [re, P] = z(null), [ie, ae] = z(0), F = () => {
			i?.(), i = null, t.isActive || Po$1();
		}, B$6 = Pn$1({
			getContainerRef: () => n,
			isCollapsed: u,
			getExpandedDimensions: () => K,
			onDragStart: () => {
				P(null), i && F();
			},
			onPositionUpdate: (e) => x(e),
			onSnapEdgeChange: (e, t) => {
				he(g(), e), _(e), y(t);
			},
			onSnapComplete: (e) => {
				K = e.expandedDimensions, x(e.position), He({
					edge: e.edge,
					ratio: e.ratio,
					collapsed: u(),
					enabled: !u()
				});
			}
		}), se = () => !Cn$1(g()), ce$3 = () => se() ? `mb-1.5` : `mr-1.5`, V = (e) => t.activeActionId === e, le = () => !!t.isActive && (t.activeActionId ?? `copy`) === `copy`, H = (e) => re() === e && !t.isActive && !u() && !B$6.isDragging() && !B$6.isSnapping() && !t.isContextMenuOpen, ue = () => {
			switch (g()) {
				case `top`: return `bottom`;
				case `bottom`: return `top`;
				case `left`: return `right`;
				case `right`: return `left`;
				default: return `top`;
			}
		}, de = (e) => {
			e.stopImmediatePropagation();
		}, G = (e, n) => ({
			onMouseEnter: (t) => {
				B$6.isDragging() || (P(e), n?.shouldFreezeInteractions !== !1 && !i && (i = hs$1(), No$1(t.clientX, t.clientY)), n?.onHoverChange?.(!0));
			},
			onMouseLeave: () => {
				P((t) => t === e ? null : t), n?.shouldFreezeInteractions !== !1 && !t.isActive && !t.isContextMenuOpen && F(), n?.onHoverChange?.(!1);
			}
		});
		B(ft(() => t.shakeCount, (e) => {
			e && !t.enabled && C(!0);
		})), B(ft(() => [t.isActive, t.isContextMenuOpen], ([e, t]) => {
			!e && !t && i && F();
		})), B(ft(() => le(), (e) => {
			if (!e) {
				N((e) => Fn$1(e, 0));
				return;
			}
			let t = null, n = 0, i = 0, a = () => {
				if (t = null, !r) return;
				let e = r.getBoundingClientRect(), a = e.left + e.width / 2, o = e.top + e.height / 2, s = n - a, c = i - o;
				if (Math.hypot(s, c) < 4) return;
				let l = Math.atan2(c, s) * 180 / Math.PI - -45;
				N((e) => Fn$1(e, l));
			}, o = zt$2((e) => {
				n = e.clientX, i = e.clientY, t === null && (t = M$1(a));
			});
			window.addEventListener(`pointermove`, o, { passive: !0 }), W(() => {
				window.removeEventListener(`pointermove`, o), t !== null && ci$2(t);
			});
		}));
		let K = {
			width: 78,
			height: 28
		}, [pe, me] = z(wn$1(g())), he = (e, t) => {
			Cn$1(e) !== Cn$1(t) && me(wn$1(t));
		}, ge = (e, t) => {
			let r = n?.getBoundingClientRect(), i = wn$1(t);
			return Dn$1(e, t, K, r?.width ?? i.width, r?.height ?? i.height);
		}, _e = () => {
			let e = Tn$1(g(), v(), K.width, K.height);
			x(e);
		}, ye = B$6.createDragAwareHandler(() => t.onToggle?.()), be = B$6.createDragAwareHandler(() => t.onActivateAction?.(Pe)), xe = B$6.createDragAwareHandler(() => t.onActivateAction?.(Fe)), Ce = `group contain-layout flex items-center justify-center cursor-pointer interactive-scale a11y-hitbox`, Te = () => Z$1(`relative contain-layout flex items-center justify-center`, ce$3()), Ee = (e) => e ? `text-[var(--rg-text-primary)]` : `text-[var(--rg-text-secondary)] group-hover:text-[var(--rg-text-primary)]`, De = B$6.createDragAwareHandler(() => {
			if (u()) {
				Fe$2();
				return;
			}
			let e = n?.getBoundingClientRect(), t = v();
			e && (K = {
				width: e.width,
				height: e.height
			}), T(!0), p(!0), He({
				edge: g(),
				ratio: t,
				collapsed: !0,
				enabled: !1
			}), Pe$1();
		}), Oe = () => On$1(g(), b(), K, pe()), J, Ae, Y = null, Me$1 = () => {
			let e = n?.getBoundingClientRect();
			e && e.width !== 0 && e.height !== 0 && (u() ? me({
				width: e.width,
				height: e.height
			}) : (K = {
				width: e.width,
				height: e.height
			}, Y = {
				width: e.width,
				height: e.height
			}));
		}, Pe$1 = () => {
			Ae && clearTimeout(Ae), Ae = setTimeout(() => {
				T(!1), Me$1();
			}, 260);
		}, Fe$2 = () => {
			let { position: e, ratio: t } = ge(Ue(), g());
			x(e), y(t), T(!0), p(!1), He({
				edge: g(),
				ratio: t,
				collapsed: !1,
				enabled: !0
			}), Pe$1();
		}, Le = (e, t) => {
			if (e !== 0 && t !== 0 && !(B$6.isDragging() || B$6.isSnapping()) && !w()) {
				if (u()) {
					let n = pe();
					if (n.width === e && n.height === t) return;
					me({
						width: e,
						height: t
					});
				} else Y && Y.width === e && Y.height === t || (Y = {
					width: e,
					height: t
				}, K = {
					width: e,
					height: t
				}, x(Tn$1(g(), v(), e, t)));
			}
		}, X = null, ze = () => {
			B$6.isDragging() || B$6.isSnapping() || X === null && (X = M$1(() => {
				X = null, _e();
			}));
		}, Be = () => {
			B$6.isDragging() || (h(!0), _e(), J && clearTimeout(J), J = setTimeout(() => {
				h(!1);
				let e = En$1(g(), b().x, b().y, K.width, K.height);
				y(e), He({
					edge: g(),
					ratio: e,
					collapsed: u(),
					enabled: !u()
				});
			}, 500));
		}, He = (e) => {
			ps(e), t.onStateChange?.(e);
		};
		pt(() => {
			n && t.onContainerRef?.(n);
			let e = n?.getBoundingClientRect(), r = nt$1(), i = !!(e && e.width > 0 && e.height > 0);
			if (a) {
				i && e && (K = {
					width: e.width,
					height: e.height
				}), p(a.collapsed);
				let t = Tn$1(a.edge, a.ratio, K.width, K.height);
				x(t);
			} else if (i && e) K = {
				width: e.width,
				height: e.height
			}, x({
				x: r.offsetLeft + (r.width - e.width) / 2,
				y: r.offsetTop + r.height - e.height - 16
			}), y(Me);
			else {
				let e = Tn$1(`bottom`, Me, K.width, K.height);
				x(e), y(Me);
			}
			if (t.onSubscribeToStateChanges) W(t.onSubscribeToStateChanges((e) => {
				if (w() || !n?.getBoundingClientRect()) return;
				let t = u() !== e.collapsed;
				if (he(g(), e.edge), _(e.edge), t && !e.collapsed) {
					let t = Ue();
					T(!0), p(e.collapsed);
					let { position: n, ratio: r } = ge(t, e.edge);
					x(n), y(r), Pe$1();
				} else {
					t && (T(!0), Pe$1()), p(e.collapsed);
					let n = Tn$1(e.edge, e.ratio, K.width, K.height);
					x(n), y(e.ratio);
				}
			}));
			window.addEventListener(`resize`, Be), window.visualViewport?.addEventListener(`resize`, Be), window.visualViewport?.addEventListener(`scroll`, Be);
			let o = Lt$2();
			if (o && (window.addEventListener(`scroll`, ze, {
				passive: !0,
				capture: !0
			}), typeof ResizeObserver < `u`)) {
				let e = new ResizeObserver(ze);
				e.observe(o), W(() => e.disconnect());
			}
			if (typeof ResizeObserver < `u` && n) {
				let e = new ResizeObserver((e) => {
					let t = e[0];
					if (!t) return;
					let r = t.borderBoxSize?.[0], i, a;
					if (r) i = r.inlineSize, a = r.blockSize;
					else {
						let e = n?.getBoundingClientRect();
						if (!e) return;
						i = e.width, a = e.height;
					}
					Le(i, a);
				});
				e.observe(n), W(() => e.disconnect());
			}
			let s = setTimeout(() => {
				l(!0);
			}, 500);
			W(() => {
				clearTimeout(s);
			});
		}), W(() => {
			window.removeEventListener(`resize`, Be), window.visualViewport?.removeEventListener(`resize`, Be), window.visualViewport?.removeEventListener(`scroll`, Be), window.removeEventListener(`scroll`, ze, { capture: !0 }), X !== null && ci$2(X), clearTimeout(J), clearTimeout(Ae), i && F();
		});
		let Ue = () => u() ? Oe() : b(), We = () => u() ? `cursor-pointer` : B$6.isDragging() ? `cursor-grabbing` : `cursor-grab`, Ge = () => k() || !!t.isContextMenuOpen || B$6.isDragging() || B$6.isSnapping() || w() || E(), Ke = () => !!t.isActive && !Ge(), qe = () => m() || B$6.isDragging() ? `` : B$6.isSnapping() ? `transition-[transform,opacity] duration-300 ease-out` : w() ? `transition-[transform,opacity] ${u() ? `duration-140` : `duration-220`} ease-drawer` : `transition-[transform,opacity] duration-400 ease-drawer`, Je = () => {
			switch (g()) {
				case `top`: return `center top`;
				case `bottom`: return `center bottom`;
				case `left`: return `left center`;
				case `right`: return `right center`;
				default: return `center center`;
			}
		};
		return (() => {
			var i = In$1();
			i.addEventListener(`mouseleave`, () => {
				ee(!1), t.onSelectHoverChange?.(!1);
			}), i.addEventListener(`mouseenter`, () => {
				ee(!0), u() || t.onSelectHoverChange?.(!0);
			}), bn(i, `mousedown`, de), bn(i, `pointerdown`, (e) => {
				de(e), B$6.handlePointerDown(e);
			});
			var a = n;
			return typeof a == `function` ? Tn(a, i) : n = i, En(i, Gt(Nn$1, {
				get isCollapsed() {
					return u();
				},
				get snapEdge() {
					return g();
				},
				get isShaking() {
					return S();
				},
				get isChevronPressed() {
					return E();
				},
				get transformOrigin() {
					return Je();
				},
				onAnimationEnd: () => C(!1),
				onCollapseClick: De,
				onCollapsePointerDown: () => O(!0),
				onCollapsePointerUp: () => O(!1),
				onCollapsePointerLeave: () => O(!1),
				onPanelClick: (e) => {
					u() && (e.stopPropagation(), Fe$2());
				},
				get actionButtons() {
					return [
						Gt(Sn$1, Yt({
							actionId: Ne,
							isToggle: !0,
							ref: (e) => r = e,
							get label() {
								return le() ? `Stop selecting element` : `Copy element`;
							},
							get isActive() {
								return le();
							},
							class: Ce,
							get wrapperClass() {
								return Te();
							},
							onClick: ye,
							onContextMenu: (e) => {
								e.preventDefault(), e.stopPropagation(), P(null), t.onToggleToolbarMenu?.();
							}
						}, () => G(Ne), {
							get icon() {
								return Gt(fn$1, {
									size: 14,
									get rotationDeg() {
										return te();
									},
									get class() {
										return Ee(le());
									}
								});
							},
							get tooltipVisible() {
								return H(Ne);
							},
							get tooltipPosition() {
								return ue();
							},
							tooltip: `Copy`
						})),
						Gt(Sn$1, Yt({
							actionId: Pe,
							label: `Comment on element`,
							get isActive() {
								return V(Pe);
							},
							class: Ce,
							get wrapperClass() {
								return Te();
							},
							onClick: be
						}, () => G(Pe), {
							get icon() {
								return Gt(mn$1, {
									size: 14,
									get class() {
										return Ee(V(Pe));
									}
								});
							},
							get tooltipVisible() {
								return H(Pe);
							},
							get tooltipPosition() {
								return ue();
							},
							tooltip: `Comment`
						})),
						Gt(Sn$1, Yt({
							actionId: Fe,
							label: `Style element`,
							get isActive() {
								return V(Fe);
							},
							class: Ce,
							get wrapperClass() {
								return Te();
							},
							onClick: xe
						}, () => G(Fe), {
							get icon() {
								return Gt(gn$1, {
									size: 14,
									get class() {
										return Ee(V(Fe));
									}
								});
							},
							get tooltipVisible() {
								return H(Fe);
							},
							get tooltipPosition() {
								return ue();
							},
							tooltip: `Style`
						})),
						Gt(Qt, {
							get each() {
								return t.buttons ?? [];
							},
							children: (t) => {
								let n = () => (ie(), t.pressed?.() ?? !1), r = B$6.createDragAwareHandler(() => {
									t.onClick(), ae((e) => e + 1);
								});
								return Gt(Sn$1, Yt({
									get actionId() {
										return t.id;
									},
									get label() {
										return t.label;
									},
									get isActive() {
										return n();
									},
									class: Ce,
									get wrapperClass() {
										return Te();
									},
									onClick: r
								}, () => G(t.id), {
									get icon() {
										return (() => {
											var e = Ln$1();
											return Tn((e) => Rn$1(e, t.icon), e), lt(() => yn(e, Z$1(`flex items-center justify-center`, Ee(n())))), e;
										})();
									},
									get tooltipVisible() {
										return H(t.id);
									},
									get tooltipPosition() {
										return ue();
									},
									get tooltip() {
										return t.label;
									}
								}));
							}
						})
					];
				}
			})), lt((e) => {
				var t = Z$1(`fixed left-0 top-0 font-sans text-[13px] antialiased select-none`, We(), qe(), s() ? `pointer-events-auto` : `pointer-events-none`), n = String(ce), r = `translate(${Ue().x}px, ${Ue().y}px) scale(${Ke() ? .97 : 1})`, a = Je(), o = s() ? Ke() ? .55 : 1 : 0;
				return t !== e.e && yn(i, e.e = t), n !== e.t && Cn(i, `z-index`, e.t = n), r !== e.a && Cn(i, `transform`, e.a = r), a !== e.o && Cn(i, `transform-origin`, e.o = a), o !== e.i && Cn(i, `opacity`, e.i = o), e;
			}, {
				e: void 0,
				t: void 0,
				a: void 0,
				o: void 0,
				i: void 0
			}), i;
		})();
	};
	Bn$1 = yt();
	Vn$1 = () => {
		let e = G(Bn$1);
		if (!e) throw Error(`Menu subcomponents must be rendered inside <Menu.Provider>`);
		return e;
	};
	Hn$1 = (t) => Gt(Bn$1.Provider, {
		get value() {
			return t.store;
		},
		get children() {
			return t.children;
		}
	});
	Un$1 = (t) => Gt(wt$1, {
		get class() {
			return Z$1(`flex flex-col w-fit h-fit`, t.class);
		},
		get style() {
			return t.style;
		},
		get children() {
			return t.children;
		}
	});
	Wn$1 = /*#__PURE__*/ mn(`<div role=menu aria-orientation=vertical><div aria-hidden=true class="pointer-events-none absolute opacity-0 transition-[top,left,width,height,opacity,border-radius] duration-75 ease-out bg-[var(--rg-surface-hover)]">`);
	Gn$1 = (e) => {
		let t = Vn$1();
		return (() => {
			var n = Wn$1(), r = n.firstChild;
			n.$$pointermove = () => t.notePointerMove(), Tn((n) => {
				t.setHighlightContainer(n), e.ref?.(n);
			}, n);
			var i = t.setHighlightRail;
			return typeof i == `function` ? Tn(i, r) : t.setHighlightRail = r, En(n, () => e.children, null), lt((r) => {
				var i = e.label, a = t.keyboardNavigation ? t.activeDescendantId() : void 0, o = t.keyboardNavigation ? -1 : void 0, s = Z$1(`relative flex flex-col`, e.class);
				return i !== r.e && gn(n, `aria-label`, r.e = i), a !== r.t && gn(n, `aria-activedescendant`, r.t = a), o !== r.a && gn(n, `tabindex`, r.a = o), s !== r.o && yn(n, r.o = s), r;
			}, {
				e: void 0,
				t: void 0,
				a: void 0,
				o: void 0
			}), n;
		})();
	};
	hn([`pointermove`]);
	Kn$1 = /*#__PURE__*/ mn(`<button data-react-grab-ignore-events type=button>`);
	qn$1 = (e) => {
		let t = Vn$1(), n = t.createItemId(), r = e.value, i = () => e.role ?? `menuitem`, a = () => !e.disabled, o = () => t.activeValue() === r, s;
		return pt(() => {
			s && (t.registerItem({
				value: r,
				domId: n,
				element: s,
				isEnabled: a,
				onSelect: () => e.onSelect?.()
			}), W(() => t.unregisterItem(r)));
		}), (() => {
			var c = Kn$1();
			c.$$click = (t) => {
				t.stopPropagation(), a() && e.onSelect?.();
			}, c.addEventListener(`pointerleave`, () => {
				t.clearActiveOnPointerLeave && t.setActiveItem(null);
			}), c.addEventListener(`pointerenter`, () => {
				a() && t.canActivateOnHover() && t.setActiveItem(r);
			}), c.$$pointerdown = (e) => e.stopPropagation();
			var l = s;
			return typeof l == `function` ? Tn(l, c) : s = c, gn(c, `id`, n), En(c, () => e.children), lt((n) => {
				var a = e.dataId ?? r, s = i(), l = i() === `menuitemradio` ? !!e.checked : void 0, u = !!e.disabled, d = t.keyboardNavigation ? o() ? 0 : -1 : void 0, p = e.disabled, m = Z$1(`relative z-1 contain-layout flex items-center justify-between w-full px-2 py-1 cursor-pointer text-left border-none bg-transparent disabled:opacity-40 disabled:cursor-default`, e.class);
				return a !== n.e && gn(c, `data-react-grab-menu-item`, n.e = a), s !== n.t && gn(c, `role`, n.t = s), l !== n.a && gn(c, `aria-checked`, n.a = l), u !== n.o && gn(c, `aria-disabled`, n.o = u), d !== n.i && gn(c, `tabindex`, n.i = d), p !== n.n && (c.disabled = n.n = p), m !== n.s && yn(c, n.s = m), n;
			}, {
				e: void 0,
				t: void 0,
				a: void 0,
				o: void 0,
				i: void 0,
				n: void 0,
				s: void 0
			}), c;
		})();
	};
	hn([`pointerdown`, `click`]);
	Jn$1 = /*#__PURE__*/ mn(`<span>`);
	Yn$1 = (e) => (() => {
		var t = Jn$1();
		return En(t, () => e.children), lt(() => yn(t, Z$1(`text-[13px] leading-4 font-sans font-medium`, e.class))), t;
	})();
	Xn$1 = /*#__PURE__*/ mn(`<svg xmlns=http://www.w3.org/2000/svg viewBox="0 0 24 24"fill=currentColor><path fill-rule=evenodd clip-rule=evenodd d="M6.5 5C5.67157 5 5 5.67157 5 6.5C5 7.32843 5.67157 8 6.5 8H8V6.5C8 5.67157 7.32843 5 6.5 5ZM10 8V6.5C10 4.567 8.433 3 6.5 3C4.567 3 3 4.567 3 6.5C3 8.433 4.567 10 6.5 10H8V14H6.5C4.567 14 3 15.567 3 17.5C3 19.433 4.567 21 6.5 21C8.433 21 10 19.433 10 17.5V16H14V17.5C14 19.433 15.567 21 17.5 21C19.433 21 21 19.433 21 17.5C21 15.567 19.433 14 17.5 14H16V10H17.5C19.433 10 21 8.433 21 6.5C21 4.567 19.433 3 17.5 3C15.567 3 14 4.567 14 6.5V8H10ZM10 10V14H14V10H10ZM16 8H17.5C18.3284 8 19 7.32843 19 6.5C19 5.67157 18.3284 5 17.5 5C16.6716 5 16 5.67157 16 6.5V8ZM16 16V17.5C16 18.3284 16.6716 19 17.5 19C18.3284 19 19 18.3284 19 17.5C19 16.6716 18.3284 16 17.5 16H16ZM8 16H6.5C5.67157 16 5 16.6716 5 17.5C5 18.3284 5.67157 19 6.5 19C7.32843 19 8 18.3284 8 17.5V16Z">`);
	Zn$1 = (e) => {
		let t = () => e.size ?? 11;
		return (() => {
			var n = Xn$1();
			return lt((r) => {
				var i = t(), a = t(), o = e.class;
				return i !== r.e && gn(n, `width`, r.e = i), a !== r.t && gn(n, `height`, r.t = a), o !== r.a && gn(n, `class`, r.a = o), r;
			}, {
				e: void 0,
				t: void 0,
				a: void 0
			}), n;
		})();
	};
	Qn$1 = /*#__PURE__*/ mn(`<span>`);
	$n$1 = /*#__PURE__*/ mn(`<span style=display:inline-flex;align-items:center;gap:2px>`);
	er$1 = /*#__PURE__*/ mn(`<span>Ctrl+`);
	tr$1 = (n) => {
		let r = () => n.shortcut === `Enter`, i = () => n.modifier !== !1, a = Ki();
		return (() => {
			var t = $n$1();
			return En(t, Gt(en, {
				get when() {
					return r();
				},
				get children() {
					return Gt(kt$1, { size: 8 });
				}
			}), null), En(t, Gt(en, {
				get when() {
					return !r();
				},
				get children() {
					return Gt(en, {
						get when() {
							return i();
						},
						get fallback() {
							return (() => {
								var e = Qn$1();
								return En(e, () => n.shortcut), e;
							})();
						},
						get children() {
							return Gt(en, {
								when: a,
								get fallback() {
									return (() => {
										var e = er$1();
										return e.firstChild, En(e, () => n.shortcut, null), e;
									})();
								},
								get children() {
									return [Gt(Zn$1, { size: 9 }), (() => {
										var e = Qn$1();
										return En(e, () => n.shortcut), e;
									})()];
								}
							});
						}
					});
				}
			}), null), lt(() => yn(t, n.class)), t;
		})();
	};
	nr$1 = (t) => Gt(tr$1, {
		get shortcut() {
			return t.shortcut;
		},
		get modifier() {
			return t.modifier;
		},
		get class() {
			return Z$1(`text-[11px] font-sans text-[var(--rg-text-secondary)] ml-4`, t.class);
		}
	});
	rr$1 = ({ hiddenOpacity: e = `0`, visibleOpacity: t = `1` } = {}) => {
		let n, r, i, a = (e) => {
			r && (r.style.opacity = t, r.style.top = `${e.offsetTop}px`, r.style.left = `${e.offsetLeft}px`, r.style.width = `${e.offsetWidth}px`, r.style.height = `${e.offsetHeight}px`);
		}, o = typeof ResizeObserver > `u` ? void 0 : new ResizeObserver(() => {
			i && a(i);
		});
		W(() => o?.disconnect());
		let s = () => {
			i = void 0, o?.disconnect(), r && (r.style.opacity = e);
		};
		return {
			containerRef: (e) => {
				n = e;
			},
			followerRef: (e) => {
				r = e;
			},
			followElement: (e) => {
				r && n && (e ? (i = e, a(e), o?.disconnect(), o?.observe(e), o?.observe(n)) : s());
			},
			hideFollower: s
		};
	};
	ir$1 = (e, t) => e !== t && e instanceof HTMLElement;
	ar$1 = (e, t) => {
		let n = [];
		for (let r of Array.from(e.children)) ir$1(r, t) && n.push(r);
		return n;
	};
	or$1 = (e = {}) => {
		let { topCornerRadiusPx: t, bottomCornerRadiusPx: n, cornerShape: r } = e, i = t !== void 0 || n !== void 0, a, o = !1, s = (e) => {
			if (!a || !i) return;
			let s = e.parentElement;
			if (!s) return;
			let c = ar$1(s, a), l = c[0] === e, u = c[c.length - 1] === e, d = l && t !== void 0 ? `${t}px` : `0px`, f = u && n !== void 0 ? `${n}px` : `0px`;
			a.style.borderTopLeftRadius = d, a.style.borderTopRightRadius = d, a.style.borderBottomLeftRadius = f, a.style.borderBottomRightRadius = f, r && !o && (a.style.setProperty(`corner-shape`, r), o = !0);
		}, { containerRef: c, followerRef: l, followElement: u, hideFollower: d } = rr$1();
		return {
			containerRef: c,
			highlightRef: (e) => {
				a = e, o = !1, l(e);
			},
			updateHighlight: (e) => {
				u(e), e && s(e);
			},
			clearHighlight: d
		};
	};
	sr$1 = (e = {}) => {
		let t = /* @__PURE__ */ new Map(), n = [], r = `react-grab-menu-${Math.random().toString(36).slice(2, 8)}`, i = 0, a = !1, o = e.value !== void 0, [s, c] = z(null), [l, u] = z(0), d = V(() => o ? e.value?.() ?? null : s()), f = (t) => {
			o ? e.onValueChange?.(t) : c(t);
		}, p = or$1(e.highlight ?? {});
		B(ft([d, l], ([e]) => {
			if (e === null) {
				p.clearHighlight();
				return;
			}
			let n = t.get(e);
			n ? p.updateHighlight(n.element) : p.clearHighlight();
		}));
		let m = () => n.filter((e) => t.get(e)?.isEnabled()), h = () => {
			let e = m();
			e.length > 0 && f(e[0]);
		}, g = () => {
			let e = m();
			e.length > 0 && f(e[e.length - 1]);
		}, _ = () => {
			let e = m();
			if (e.length === 0) return;
			let t = e.indexOf(d() ?? ``), n = t === -1 ? 0 : (t + 1) % e.length;
			f(e[n]);
		}, v = () => {
			let e = m();
			if (e.length === 0) return;
			let t = e.indexOf(d() ?? ``), n = t === -1 ? e.length - 1 : (t - 1 + e.length) % e.length;
			f(e[n]);
		}, y = V(() => {
			l();
			let e = d();
			if (e !== null) return t.get(e)?.domId;
		});
		return {
			keyboardNavigation: e.keyboardNavigation ?? !1,
			clearActiveOnPointerLeave: e.clearActiveOnPointerLeave ?? !1,
			activeValue: d,
			activeDescendantId: y,
			setActiveItem: f,
			createItemId: () => `${r}-item-${i++}`,
			canActivateOnHover: () => !(e.requirePointerMove ?? !1) || a,
			notePointerMove: () => {
				a = !0;
			},
			resetPointerMove: () => {
				a = !1;
			},
			registerItem: (e) => {
				t.set(e.value, e), n.includes(e.value) || n.push(e.value), u((e) => e + 1);
			},
			unregisterItem: (e) => {
				t.delete(e);
				let r = n.indexOf(e);
				r !== -1 && n.splice(r, 1), o || c((t) => t === e ? null : t), u((e) => e + 1);
			},
			getActiveItem: () => {
				let e = d();
				return e === null ? void 0 : t.get(e);
			},
			selectFirst: h,
			selectLast: g,
			selectNext: _,
			selectPrevious: v,
			setHighlightContainer: p.containerRef,
			setHighlightRail: p.highlightRef
		};
	};
	Q$1 = {
		Provider: Hn$1,
		Panel: Un$1,
		List: Gn$1,
		Item: qn$1,
		Label: Yn$1,
		Shortcut: nr$1
	};
	$$1 = (e) => {
		e.type === `contextmenu` && e.preventDefault(), e.stopImmediatePropagation();
	};
	cr$1 = (e) => {
		let t = zt$2((t) => {
			e.isOpen() && (e.shouldIgnoreKeyboardEvent?.(t) || e.shouldIgnoreInputEvents && yr(t) || t.code === `Escape` && (t.preventDefault(), t.stopImmediatePropagation(), e.onDismiss(`keyboard`)));
		}), n = zt$2((t) => {
			e.isOpen() && ($i(t, `data-react-grab-ignore-events`) || e.shouldIgnoreRightClick && t instanceof MouseEvent && t.button === 2 || e.onDismiss(`pointer`));
		}), r = M$1(() => {
			window.addEventListener(`mousedown`, n, { capture: !0 }), window.addEventListener(`touchstart`, n, { capture: !0 });
		});
		return window.addEventListener(`keydown`, t, { capture: !0 }), () => {
			ci$2(r), window.removeEventListener(`keydown`, t, { capture: !0 }), window.removeEventListener(`mousedown`, n, { capture: !0 }), window.removeEventListener(`touchstart`, n, { capture: !0 });
		};
	};
	lr$1 = /*#__PURE__*/ mn(`<div class="contain-layout shrink-0 flex items-center gap-1 pt-1.5 pb-1 w-fit h-fit px-2">`);
	ur$1 = /*#__PURE__*/ mn(`<div data-react-grab-ignore-events data-react-grab-context-menu class="fixed font-sans text-[13px] antialiased [filter:var(--rg-drop-shadow)] select-none"style=pointer-events:auto>`);
	dr$1 = (t) => {
		let n, r, i = null, a = sr$1({
			keyboardNavigation: !0,
			highlight: {
				bottomCornerRadiusPx: 14,
				cornerShape: Ie
			}
		}), [s, l] = z(0), [u, f] = z(0), p = V(() => t.position !== null), m = V(() => at$1({
			tagName: t.tagName,
			componentName: t.componentName
		})), h = () => {
			if (n) {
				let e = n.getBoundingClientRect();
				l(e.width), f(e.height);
			}
		};
		B(() => {
			p() && M$1(h);
		});
		let g = V(() => {
			let e = t.selectionBounds, n = t.position, r = s(), i = u();
			if (r === 0 || i === 0 || !e || !n) return {
				left: Be.left,
				top: Be.top,
				arrowLeft: 0,
				arrowPosition: `bottom`
			};
			let a = n.x ?? e.x + e.width / 2, o = Math.max(4, Math.min(a - r / 2, window.innerWidth - r - 4)), c = Math.max(8, Math.min(a - o, r - 8)), l = e.y + e.height + 8 + 4, d = e.y - i - 8 - 4, f = l + i > window.innerHeight, p = d >= 0, m = f && p, h = m ? d : l, g = m ? `top` : `bottom`;
			if (f && !p) {
				let t = n.y ?? e.y + e.height / 2;
				h = Math.max(4, Math.min(t + 4, window.innerHeight - i - 4)), g = `top`;
			}
			return {
				left: o,
				top: h,
				arrowLeft: c,
				arrowPosition: g
			};
		}), _ = V(() => {
			let e = t.actions ?? [], n = t.actionContext;
			return e.map((e) => ({
				id: e.id,
				label: e.label,
				action: () => {
					n && Ds(e, n);
				},
				enabled: Es(e, n),
				shortcut: e.shortcut,
				shortcutModifier: e.shortcutModifier
			}));
		});
		B(ft(p, (e) => {
			if (e) {
				let e = n?.getRootNode(), t = e instanceof ShadowRoot ? e.activeElement : null, a = document.activeElement;
				i = a instanceof HTMLElement && t === null && !(n instanceof Element && n.contains(a)) ? a : null, r?.focus({ preventScroll: !0 });
				return;
			}
			a.setActiveItem(null);
			let t = i;
			i = null, t instanceof HTMLElement && document.contains(t) && M$1(() => {
				let e = document.activeElement;
				(e === null || e === document.body) && t.focus({ preventScroll: !0 });
			});
		})), pt(() => {
			h();
			let e = (e) => {
				if (!p()) return;
				let n = e.key === `ArrowDown`, r = e.key === `ArrowUp`, i = e.key === `Home`, o = e.key === `End`, s = e.key === `Tab`;
				if (n || r || i || o || s) {
					e.preventDefault(), e.stopPropagation();
					let t = n || s && !e.shiftKey;
					i ? a.selectFirst() : o ? a.selectLast() : t ? a.selectNext() : a.selectPrevious();
					return;
				}
				let c = t.actions ?? [], l = t.actionContext, u = (n) => {
					l && Ds(n, l) && (e.preventDefault(), e.stopPropagation(), t.onHide());
				};
				if (e.key === `Enter`) {
					let t = a.getActiveItem();
					if (t) {
						e.preventDefault(), e.stopPropagation(), t.isEnabled() && t.onSelect();
						return;
					}
				}
				let d = Ss(c, e, { includeModifierShortcuts: !0 });
				d && u(d);
			}, n = cr$1({
				isOpen: p,
				onDismiss: t.onDismiss,
				shouldIgnoreRightClick: !0
			}), r = zt$2(e);
			window.addEventListener(`keydown`, r, { capture: !0 }), W(() => {
				n(), window.removeEventListener(`keydown`, r, { capture: !0 });
			});
		});
		let v = V(() => {
			let { tagName: e, componentName: t } = m();
			return `Actions for ${t ? `${t}.${e}` : e}`;
		});
		return Gt(en, {
			get when() {
				return p();
			},
			get children() {
				var i = ur$1();
				bn(i, `contextmenu`, $$1, !0), bn(i, `click`, $$1, !0), bn(i, `mousedown`, $$1, !0), bn(i, `pointerdown`, $$1, !0);
				var s = n;
				return typeof s == `function` ? Tn(s, i) : n = i, Cn(i, `z-index`, `${ce}`), En(i, Gt(dt$1, {
					get position() {
						return g().arrowPosition;
					},
					leftPercent: 0,
					get leftOffsetPx() {
						return g().arrowLeft;
					}
				}), null), En(i, Gt(Q$1.Panel, {
					class: `justify-center items-start min-w-[100px]`,
					get children() {
						return [(() => {
							var n = lr$1();
							return En(n, Gt(vt$1, {
								get tagName() {
									return m().tagName;
								},
								get componentName() {
									return m().componentName;
								},
								get isClickable() {
									return t.hasFilePath;
								},
								onClick: (e) => {
									if (e.stopPropagation(), t.hasFilePath && t.actionContext) {
										let e = t.actions?.find((e) => e.id === `open`);
										e && Ds(e, t.actionContext);
									}
								},
								shrink: !0
							})), n;
						})(), Gt(bt$1, { get children() {
							return Gt(Q$1.Provider, {
								store: a,
								get children() {
									return Gt(Q$1.List, {
										ref: (e) => r = e,
										get label() {
											return v();
										},
										class: `w-[calc(100%+16px)] -mx-2 -my-1.5 outline-none`,
										get children() {
											return Gt(Qt, {
												get each() {
													return _();
												},
												children: (n) => Gt(Q$1.Item, {
													get value() {
														return n.id;
													},
													get dataId() {
														return n.label.toLowerCase();
													},
													get disabled() {
														return !n.enabled;
													},
													onSelect: () => {
														n.action(), t.onHide();
													},
													get children() {
														return [Gt(Q$1.Label, {
															class: `text-[var(--rg-text-primary)]`,
															get children() {
																return n.label;
															}
														}), Gt(en, {
															get when() {
																return n.shortcut;
															},
															children: (t) => Gt(Q$1.Shortcut, {
																get shortcut() {
																	return t();
																},
																get modifier() {
																	return n.shortcutModifier;
																}
															})
														})];
													}
												})
											});
										}
									});
								}
							});
						} })];
					}
				}), null), lt((e) => {
					var t = `${g().top}px`, n = `${g().left}px`;
					return t !== e.e && Cn(i, `top`, e.e = t), n !== e.t && Cn(i, `left`, e.t = n), e;
				}, {
					e: void 0,
					t: void 0
				}), i;
			}
		});
	};
	hn([
		`pointerdown`,
		`mousedown`,
		`click`,
		`contextmenu`
	]);
	fr$1 = (e, t, n, r) => Math.max(r, Math.min(e, n - t - r));
	pr$1 = ({ anchor: e, measuredWidth: t, measuredHeight: n, viewportLeft: r = 0, viewportTop: i = 0, viewportWidth: a, viewportHeight: o, anchorGapPx: s, viewportPaddingPx: c, offscreenPosition: l }) => {
		if (!e || t === 0 || n === 0) return l;
		let u, d;
		return e.edge === `left` || e.edge === `right` ? (u = e.edge === `left` ? e.x + s : e.x - t - s, d = e.y - n / 2) : (u = e.x - t / 2, d = e.edge === `top` ? e.y + s : e.y - n - s), {
			left: r + fr$1(u - r, t, a, c),
			top: i + fr$1(d - i, n, o, c)
		};
	};
	mr$1 = (e, t) => {
		let [n, r] = z(0), [i, a] = z(0), [o, s] = z(!1), [c, l] = z(!1), [u, d] = z(0), [f, p] = z(`bottom`), m, h, g = () => {
			clearTimeout(m), h !== void 0 && (ci$2(h), h = void 0);
		}, v = () => {
			let t = e();
			t && (r(t.offsetWidth), a(t.offsetHeight));
		}, y = () => {
			H(() => {
				d((e) => e + 1), v();
			});
		}, b = V(() => t() !== null);
		return B(() => {
			let n = t();
			n ? (p(n.edge), clearTimeout(m), s(!0), h !== void 0 && ci$2(h), h = M$1(() => {
				v(), e()?.offsetHeight, l(!0);
			})) : (h !== void 0 && ci$2(h), l(!1), m = setTimeout(() => {
				s(!1);
			}, 120)), W(g);
		}), B(() => {
			if (!b()) return;
			window.addEventListener(`resize`, y), window.visualViewport?.addEventListener(`resize`, y), window.visualViewport?.addEventListener(`scroll`, y);
			let e = Lt$2(), t;
			e && typeof ResizeObserver < `u` && (t = new ResizeObserver(y), t.observe(e)), W(() => {
				window.removeEventListener(`resize`, y), window.visualViewport?.removeEventListener(`resize`, y), window.visualViewport?.removeEventListener(`scroll`, y), t?.disconnect();
			});
		}), {
			shouldMount: o,
			isAnimatedIn: c,
			lastAnchorEdge: f,
			displayPosition: V((e) => {
				u();
				let r = nt$1(), a = pr$1({
					anchor: t(),
					measuredWidth: n(),
					measuredHeight: i(),
					viewportLeft: r.offsetLeft,
					viewportTop: r.offsetTop,
					viewportWidth: r.width,
					viewportHeight: r.height,
					anchorGapPx: 8,
					viewportPaddingPx: 8,
					offscreenPosition: Be
				});
				return a.left === Be.left ? e : a;
			}, Be),
			measure: v,
			clearAnimationHandles: g
		};
	};
	hr$1 = {
		padding: {
			prefix: `p`,
			scale: `spacing`
		},
		"padding-top,padding-bottom": {
			prefix: `py`,
			scale: `spacing`
		},
		"padding-left,padding-right": {
			prefix: `px`,
			scale: `spacing`
		},
		"padding-top": {
			prefix: `pt`,
			scale: `spacing`
		},
		"padding-right": {
			prefix: `pr`,
			scale: `spacing`
		},
		"padding-bottom": {
			prefix: `pb`,
			scale: `spacing`
		},
		"padding-left": {
			prefix: `pl`,
			scale: `spacing`
		},
		margin: {
			prefix: `m`,
			scale: `spacing`
		},
		"margin-top,margin-bottom": {
			prefix: `my`,
			scale: `spacing`
		},
		"margin-left,margin-right": {
			prefix: `mx`,
			scale: `spacing`
		},
		"margin-top": {
			prefix: `mt`,
			scale: `spacing`
		},
		"margin-right": {
			prefix: `mr`,
			scale: `spacing`
		},
		"margin-bottom": {
			prefix: `mb`,
			scale: `spacing`
		},
		"margin-left": {
			prefix: `ml`,
			scale: `spacing`
		},
		gap: {
			prefix: `gap`,
			scale: `spacing`
		},
		"column-gap": {
			prefix: `gap-x`,
			scale: `spacing`
		},
		"row-gap": {
			prefix: `gap-y`,
			scale: `spacing`
		},
		"width,height": {
			prefix: `size`,
			scale: `spacing`
		},
		width: {
			prefix: `w`,
			scale: `spacing`
		},
		height: {
			prefix: `h`,
			scale: `spacing`
		},
		"min-width": {
			prefix: `min-w`,
			scale: `spacing`
		},
		"max-width": {
			prefix: `max-w`,
			scale: `spacing`
		},
		"min-height": {
			prefix: `min-h`,
			scale: `spacing`
		},
		"max-height": {
			prefix: `max-h`,
			scale: `spacing`
		},
		"top,right,bottom,left": {
			prefix: `inset`,
			scale: `spacing`
		},
		"top,bottom": {
			prefix: `inset-y`,
			scale: `spacing`
		},
		"left,right": {
			prefix: `inset-x`,
			scale: `spacing`
		},
		top: {
			prefix: `top`,
			scale: `spacing`
		},
		right: {
			prefix: `right`,
			scale: `spacing`
		},
		bottom: {
			prefix: `bottom`,
			scale: `spacing`
		},
		left: {
			prefix: `left`,
			scale: `spacing`
		},
		"border-width": {
			prefix: `border`,
			scale: `border-width`
		},
		"border-top-width": {
			prefix: `border-t`,
			scale: `border-width`
		},
		"border-right-width": {
			prefix: `border-r`,
			scale: `border-width`
		},
		"border-bottom-width": {
			prefix: `border-b`,
			scale: `border-width`
		},
		"border-left-width": {
			prefix: `border-l`,
			scale: `border-width`
		},
		"z-index": {
			prefix: `z`,
			scale: `z-index`
		},
		opacity: {
			prefix: `opacity`,
			scale: `opacity`
		}
	};
	gr$1 = (e, t) => {
		let n = hr$1[e];
		if (!n) return null;
		let { prefix: r, scale: i } = n;
		if (i === `spacing`) {
			if (t < 0) return null;
			let e = t / 4;
			return !Number.isInteger(e) || e > 96 ? null : `${r}-${e}`;
		}
		return i === `border-width` ? !Number.isInteger(t) || t < 0 || t > 8 ? null : t === 1 ? r : `${r}-${t}` : i === `z-index` ? !Number.isInteger(t) || t < 0 || t > 50 ? null : `${r}-${t}` : !Number.isFinite(t) || t < 0 || t > 100 || t % 5 != 0 ? null : `${r}-${t}`;
	};
	_r$1 = (e) => {
		let t = e.getRootNode();
		return w(t) ? t.activeElement : e.ownerDocument?.activeElement ?? document.activeElement;
	};
	vr$1 = (e) => e.toLowerCase() === `#00000000` ? We$1 : e.toUpperCase();
	yr$1 = /*#__PURE__*/ mn(`<input data-react-grab-ignore-events data-react-grab-input type=text autocapitalize=none autocorrect=off autocomplete=off>`);
	br$1 = (e) => {
		let [t, n] = Xt(e, [
			`class`,
			`ref`,
			`autoFocusSelect`
		]);
		return (() => {
			var e = yr$1();
			return Tn((e) => {
				t.autoFocusSelect && queueMicrotask(() => {
					e.focus(), e.select();
				}), t.ref?.(e);
			}, e), gn(e, `spellcheck`, !1), wn(e, Yt({ get class() {
				return Z$1(`bg-transparent border-none outline-none text-[var(--rg-text-primary)] p-0 m-0`, t.class);
			} }, n), !1, !1), e;
		})();
	};
	xr$1 = `text-[13px] leading-4 font-medium`;
	Sr$1 = `text-[12px] leading-4 font-medium tabular-nums`;
	Cr$1 = {
		position: `absolute`,
		opacity: 0,
		"pointer-events": `none`,
		width: `0`,
		height: `0`,
		margin: `0`,
		padding: `0`,
		overflow: `hidden`
	};
	wr$1 = /*#__PURE__*/ mn(`<div class="flex items-center gap-2 w-full px-2 h-[20px]"><div class="flex items-center gap-1 ml-auto shrink-0"><button type=button data-react-grab-ignore-events aria-label="Pick color"class="size-[16px] rounded-[4px] border-[var(--rg-border-button)] [border-width:0.5px] border-solid p-0 m-0 cursor-pointer shrink-0"></button><input data-react-grab-ignore-events type=color aria-hidden=true tabindex=-1 class="absolute opacity-0 pointer-events-none size-0">`);
	Tr$1 = /*#__PURE__*/ mn(`<span>`);
	Er$1 = (e) => e.length === 9 ? e.slice(0, 7) : e;
	Dr$1 = (t) => {
		let [n, r] = z(null), i = () => n() !== null, a = () => vr$1(t.value), o, s = !0;
		W(() => {
			s = !1;
		}), pt(() => {
			let e = t.onRegisterTrigger?.(() => o?.click());
			W(() => e?.());
		});
		let l = () => {
			let e = n();
			if (e === null) return;
			r(null);
			let i = Gr(e);
			i ? i.toLowerCase() !== t.value.toLowerCase() && t.onCommit(i, `keyboard`) : t.onInvalidCommit?.(), t.onEditComplete?.();
		}, u = () => {
			i() && (r(null), t.onEditComplete?.());
		}, m = (e) => {
			rt$1(e) || (e.stopImmediatePropagation(), e.key === `Enter` ? (e.preventDefault(), l()) : e.key === `Escape` && (e.preventDefault(), u()));
		}, h = (e) => {
			e.preventDefault(), e.stopPropagation(), t.onInteract?.(), o?.click();
		};
		return (() => {
			var u = wr$1(), p = u.firstChild, g = p.firstChild, _ = g.nextSibling;
			En(u, Gt(en, {
				get when() {
					return t.label;
				},
				children: (e) => (() => {
					var t = Tr$1();
					return yn(t, `${xr$1} text-[var(--rg-text-primary)] truncate min-w-0`), En(t, e), t;
				})()
			}), p), En(p, Gt(en, {
				get when() {
					return i();
				},
				get fallback() {
					return (() => {
						var e = Tr$1();
						return e.$$click = (e) => {
							e.preventDefault(), e.stopPropagation(), r(a());
						}, e.$$pointerdown = (e) => e.stopPropagation(), yn(e, `${Sr$1} text-[var(--rg-text-primary)] cursor-text`), En(e, a), lt(() => gn(e, `data-react-grab-value`, a())), e;
					})();
				},
				get children() {
					return Gt(br$1, {
						autoFocusSelect: !0,
						inputmode: `text`,
						"aria-label": `Style color hex`,
						class: `${Sr$1} text-right`,
						get style() {
							return {
								"field-sizing": `content`,
								"min-width": `32px`,
								"max-width": t.emphasized ? `100px` : `72px`
							};
						},
						get value() {
							return n() ?? ``;
						},
						onInput: (e) => r(e.currentTarget.value),
						onKeyDown: m,
						onBlur: l,
						onPointerDown: (e) => e.stopPropagation(),
						onMouseDown: (e) => e.stopPropagation(),
						onClick: (e) => e.stopPropagation()
					});
				}
			}), g), g.$$click = h, g.$$mousedown = (e) => e.preventDefault(), g.$$pointerdown = (e) => e.stopPropagation(), _.$$input = (e) => {
				if (!s) return;
				let n = e.currentTarget.value, r = t.value.length === 9 ? t.value.slice(7) : ``, i = n + (r.toLowerCase() === `00` ? `` : r);
				t.onInteract?.(), i && i.toLowerCase() !== t.value.toLowerCase() && t.onCommit(i, `pointer`);
			};
			var v = o;
			return typeof v == `function` ? Tn(v, _) : o = _, lt((e) => Cn(g, `background-color`, t.value)), lt(() => _.value = Er$1(t.value)), u;
		})();
	};
	hn([
		`pointerdown`,
		`mousedown`,
		`click`,
		`input`
	]);
	Or$1 = /*#__PURE__*/ mn(`<svg width=1em height=1em viewBox="0 0 24 24"fill=none xmlns=http://www.w3.org/2000/svg aria-hidden=true style=width:16px;height:16px;flex-shrink:0;cursor:pointer><path fill-rule=evenodd clip-rule=evenodd>`);
	kr$1 = (e) => {
		let t = () => e.active ? `var(--rg-text-primary)` : `var(--rg-text-secondary)`, n = () => e.active ? e.direction === `left` ? -2 : 2 : 0, r = () => e.active ? 1.12 : 1, i = () => e.direction === `left` ? `M2.54282 10.5427C1.76177 11.3238 1.76177 12.5901 2.54282 13.3712L9.70183 20.5302C11.1193 21.9476 13.5428 20.9437 13.5428 18.9392V16.9573H21.0428C22.1474 16.9573 23.0428 16.0619 23.0428 14.9573V8.95733C23.0428 7.85276 22.1474 6.95733 21.0428 6.95733L13.5428 6.95733V4.97473C13.5428 2.9702 11.1193 1.96631 9.70183 3.38373L2.54282 10.5427Z` : `M14.2983 20.5302L21.4573 13.3712C22.2383 12.5901 22.2383 11.3238 21.4573 10.5427L14.2983 3.38374C12.8808 1.96631 10.4573 2.97019 10.4573 4.97473V6.95659L3.95728 6.95659C2.85271 6.95659 1.95728 7.85202 1.95728 8.95659V14.9566C1.95728 16.0612 2.85271 16.9566 3.95727 16.9566H10.4573V18.9392C10.4573 20.9437 12.8808 21.9476 14.2983 20.5302Z`;
		return (() => {
			var a = Or$1(), s = a.firstChild;
			return bn(a, `pointerdown`, e.onPointerDown, !0), lt((o) => {
				var l = `translateX(${n()}px) scale(${r()})`, u = e.active ? `transform 30ms cubic-bezier(0, 0, 0.2, 1)` : `transform 450ms cubic-bezier(0.34, 1.8, 0.64, 1)`, d = i(), f = t(), p = e.active ? `fill 50ms ease` : `fill 300ms ease`;
				return l !== o.e && Cn(a, `transform`, o.e = l), u !== o.t && Cn(a, `transition`, o.t = u), d !== o.a && gn(s, `d`, o.a = d), f !== o.o && gn(s, `fill`, o.o = f), p !== o.i && Cn(s, `transition`, o.i = p), o;
			}, {
				e: void 0,
				t: void 0,
				a: void 0,
				o: void 0,
				i: void 0
			}), a;
		})();
	};
	hn([`pointerdown`]);
	Ar$1 = /*#__PURE__*/ mn(`<span>`);
	jr$1 = /*#__PURE__*/ mn(`<div class="flex items-center gap-2 w-full px-2 h-[20px]"><div class="ml-auto shrink-0 flex items-center gap-1"><span data-react-grab-ignore-events role=button tabindex=-1 class="text-[12px] leading-4 font-medium tabular-nums text-[var(--rg-text-primary)] cursor-pointer select-none min-w-[36px] text-center">`);
	Mr$1 = (t) => {
		let n = () => t.options.find((e) => e.value === t.value)?.label ?? t.value;
		return (() => {
			var r = jr$1(), i = r.firstChild, a = i.firstChild;
			return En(r, Gt(en, {
				get when() {
					return t.label;
				},
				get children() {
					var e = Ar$1();
					return yn(e, `${xr$1} text-[var(--rg-text-primary)] truncate min-w-0`), En(e, () => t.label), e;
				}
			}), i), En(i, Gt(kr$1, {
				direction: `left`,
				get active() {
					return t.activeKey === `left`;
				},
				onPointerDown: () => t.onStep(-1)
			}), a), a.$$contextmenu = (e) => {
				e.preventDefault(), e.stopPropagation(), t.onStep(-1);
			}, a.$$click = (e) => {
				e.preventDefault(), e.stopPropagation(), t.onStep(1);
			}, a.$$mousedown = (e) => e.preventDefault(), a.$$pointerdown = (e) => e.stopPropagation(), En(a, n), En(i, Gt(kr$1, {
				direction: `right`,
				get active() {
					return t.activeKey === `right`;
				},
				onPointerDown: () => t.onStep(1)
			}), null), lt((e) => {
				var n = `Cycle ${t.label ?? `value`}`, r = `Click to cycle (${t.options.length} options)`;
				return n !== e.e && gn(a, `aria-label`, e.e = n), r !== e.t && gn(a, `title`, e.t = r), e;
			}, {
				e: void 0,
				t: void 0
			}), r;
		})();
	};
	hn([
		`pointerdown`,
		`mousedown`,
		`click`,
		`contextmenu`
	]);
	Nr$1 = (e) => e.kind === `numeric` ? e : null;
	Pr$1 = (e) => e.kind === `color` ? e : null;
	Fr$1 = (e) => e.kind === `enum` ? e : null;
	Ir$1 = /^[0-9]$/;
	Lr$1 = (e) => {
		let t = Array.from(e), n = t.findIndex((e) => Ir$1.test(e));
		if (n < 0) return {
			prefixLiterals: t,
			rightAlignedSegments: []
		};
		let r = t.slice(0, n), i = [], a = 0;
		for (let e = t.length - 1; e >= n; e--) {
			let n = t[e], r = Ir$1.test(n) ? `digit` : `literal`;
			i.push({
				kind: r,
				value: n,
				digitDistanceFromRight: a
			}), r === `digit` && (a += 1);
		}
		return {
			prefixLiterals: r,
			rightAlignedSegments: i
		};
	};
	Rr$1 = /*#__PURE__*/ mn(`<span class=rg-slot-column aria-hidden=true><span class=rg-slot-sizer>0</span><span class=rg-slot-cell>`);
	zr$1 = /*#__PURE__*/ mn(`<span class="rg-slot-cell rg-slot-exit">`);
	Br$1 = /*#__PURE__*/ mn(`<span><span class=inline-flex style=flex-direction:row-reverse>`);
	Vr$1 = /*#__PURE__*/ mn(`<span class="inline-block whitespace-pre"aria-hidden=true>`);
	Hr$1 = (e) => typeof e == `number` ? e : Number.parseFloat(String(e).replace(/[^0-9.-]/g, ``)) || 0;
	Ur$1 = (t) => {
		let n, [r, i] = z(null), a = null;
		B(ft(() => t.digit, (e, r) => {
			r !== void 0 && e !== r && (n && (n.classList.remove(`rg-slot-enter`), n.offsetWidth, n.classList.add(`rg-slot-enter`)), i({
				digit: r,
				direction: t.direction
			}), a !== null && clearTimeout(a), a = setTimeout(() => i(null), t.delayMs + 350));
		}, { defer: !0 })), W(() => {
			a !== null && clearTimeout(a);
		});
		let o = () => ({ "animation-delay": `${t.delayMs}ms` });
		return (() => {
			var i = Rr$1(), a = i.firstChild.nextSibling, s = n;
			return typeof s == `function` ? Tn(s, a) : n = a, En(a, () => t.digit), En(i, Gt(en, {
				get when() {
					return r();
				},
				keyed: !0,
				children: (e) => (() => {
					var t = zr$1();
					return En(t, () => e.digit), lt((n) => {
						var r = String(e.direction), i = o();
						return r !== n.e && gn(t, `data-dir`, n.e = r), n.t = Sn(t, i, n.t), n;
					}, {
						e: void 0,
						t: void 0
					}), t;
				})()
			}), null), lt((e) => {
				var n = String(t.direction), r = o();
				return n !== e.e && gn(a, `data-dir`, e.e = n), e.t = Sn(a, r, e.t), e;
			}, {
				e: void 0,
				t: void 0
			}), i;
		})();
	};
	Wr$1 = (t) => {
		let n = V(() => String(t.children ?? ``)), r = V(() => Lr$1(n())), i = V(ft(() => Hr$1(t.children), (e, t) => t === void 0 || e === t ? 0 : e > t ? 1 : -1)), a = V(() => t.direction ?? i());
		return (() => {
			var i = Br$1(), o = i.firstChild;
			return En(i, Gt($t, {
				get each() {
					return r().prefixLiterals;
				},
				children: (e) => (() => {
					var t = Vr$1();
					return En(t, e), t;
				})()
			}), o), En(o, Gt($t, {
				get each() {
					return r().rightAlignedSegments;
				},
				children: (t) => Gt(en, {
					get when() {
						return t().kind === `digit`;
					},
					get fallback() {
						return (() => {
							var e = Vr$1();
							return En(e, () => t().value), e;
						})();
					},
					get children() {
						return Gt(Ur$1, {
							get digit() {
								return Number(t().value);
							},
							get direction() {
								return a();
							},
							get delayMs() {
								return t().digitDistanceFromRight * 18;
							}
						});
					}
				})
			})), lt((e) => {
				var r = n(), a = `inline-flex ${t.class ?? ``}`, o = {
					...t.style,
					"--rg-slot-dur": `350ms`
				};
				return r !== e.e && gn(i, `aria-label`, e.e = r), a !== e.t && yn(i, e.t = a), e.a = Sn(i, o, e.a), e;
			}, {
				e: void 0,
				t: void 0,
				a: void 0
			}), i;
		})();
	};
	Gr$1 = /*#__PURE__*/ mn(`<div data-react-grab-slider-base aria-hidden=true class="absolute inset-y-0 left-0 right-0 rounded-[6px] bg-[var(--rg-surface-active)] pointer-events-none">`);
	Kr$1 = /*#__PURE__*/ mn(`<div class="flex items-center gap-1 w-full px-1"><div role=slider aria-keyshortcuts="ArrowLeft ArrowRight"tabindex=-1 class="relative flex-1 h-[20px] flex items-center overflow-hidden rounded-[6px]"><div data-react-grab-slider-fill aria-hidden=true class="absolute inset-y-0 left-0 bg-[var(--rg-surface-active)] pointer-events-none"></div><div aria-hidden=true class="absolute inset-0 pointer-events-none"></div><div data-react-grab-slider-handle aria-hidden=true class="absolute top-[2px] bottom-[2px] w-[2px] rounded-[1px] bg-[var(--rg-text-primary)] pointer-events-none"style="transition:opacity 120ms ease"></div><div class="relative z-10 flex items-center justify-between w-full px-2 pointer-events-none">`);
	qr$1 = /*#__PURE__*/ mn(`<div data-react-grab-slider-hash-mark class="absolute top-1/2 w-px h-[8px] rounded-[1px] bg-[var(--rg-text-secondary)]"style="transform:translate(-50%, -50%);transition:opacity 200ms ease">`);
	Jr$1 = /*#__PURE__*/ mn(`<span>`);
	Yr$1 = /*#__PURE__*/ mn(`<span class="ml-auto flex items-center gap-1 pointer-events-auto"><span data-react-grab-value-text><span class="text-[9px] text-[var(--rg-text-secondary)] ml-px">`);
	Xr$1 = /*#__PURE__*/ mn(`<span data-react-grab-tailwind-label aria-hidden=true class="text-[10px] leading-4 text-[var(--rg-text-secondary)] tabular-nums">`);
	Zr$1 = Array.from({ length: 9 }, (e, t) => (t + 1) * 100 / 10);
	Qr$1 = /^(-?(?:\d+\.?\d*|\.\d+))\s*([a-zA-Z%]*)$/;
	$r$1 = (t) => {
		let [n, r] = z(null), [i, a] = z(0), [o, s] = z(!1), [l, p] = z(!1), m = () => n() !== null, h, g = null, _ = () => {
			let e = t.max - t.min;
			return e <= 0 ? 0 : (Math.max(t.min, Math.min(t.max, t.value)) - t.min) / e * 100;
		}, v = () => {
			t.onCommitValue && r(Oa(t.value));
		}, y = () => !!t.onCommitValue, x = (e, n) => {
			if (n.width <= 0) return t.value;
			let r = Math.max(0, Math.min(1, (e - n.left) / n.width));
			return t.min + r * (t.max - t.min);
		}, T = (e, n) => {
			t.onCommitValue?.(x(e, n), `pointer`);
		}, E = () => {
			g = null, s(!1), a(0);
		}, D = (e, t) => {
			let n = Math.max(0, e - t.right) - Math.max(0, t.left - e), r = Math.max(0, Math.abs(n) - 32), i = Math.sqrt(Math.min(r / 200, 1));
			return Math.sign(n) * 8 * i;
		}, O = (e) => {
			if (!y() || m()) return;
			e.preventDefault(), e.stopPropagation(), e.currentTarget.setPointerCapture(e.pointerId);
			let n = h, r = n !== void 0 && e.target instanceof Node && n.contains(e.target);
			g = {
				startX: e.clientX,
				isDragging: !1,
				startedOnValueText: r,
				trackRect: e.currentTarget.getBoundingClientRect()
			}, s(!0), t.onInteract?.();
		}, k = (e) => {
			if (!g) return;
			let n = e.clientX - g.startX;
			!g.isDragging && Math.abs(n) < 3 || (g.isDragging = !0, t.onInteract?.(), T(e.clientX, g.trackRect), a(D(e.clientX, g.trackRect)));
		}, ee = (e) => {
			let t = e.currentTarget;
			t.hasPointerCapture(e.pointerId) && t.releasePointerCapture(e.pointerId), g && (g.isDragging || (g.startedOnValueText ? v() : T(g.startX, g.trackRect)), E());
		}, te = (e) => {
			let t = e.currentTarget;
			t.hasPointerCapture(e.pointerId) && t.releasePointerCapture(e.pointerId), g && E();
		};
		pt(() => {
			let e = () => {
				E(), p(!1);
			};
			window.addEventListener(`blur`, e), W(() => window.removeEventListener(`blur`, e));
		});
		let j = () => !!t.activeKey || o() || l(), ne = () => {
			let e = n();
			if (e === null) return;
			r(null);
			let i = e.trim().replace(/(\d),(\d)/g, `$1.$2`).match(Qr$1);
			if (!i) {
				t.onInvalidCommit?.(), t.onEditComplete?.();
				return;
			}
			let a = i[2].toLowerCase();
			if (a && a !== t.unit) {
				t.onInvalidCommit?.(), t.onEditComplete?.();
				return;
			}
			let o = Number.parseFloat(i[1]);
			Number.isFinite(o) ? t.onCommitValue?.(o, `keyboard`) : t.onInvalidCommit?.(), t.onEditComplete?.();
		}, N = () => {
			m() && (r(null), t.onEditComplete?.());
		}, re = (e) => {
			rt$1(e) || (e.stopImmediatePropagation(), e.key === `Enter` ? (e.preventDefault(), ne()) : e.key === `Escape` && (e.preventDefault(), N()));
		}, P = () => {
			let e = i();
			return {
				cursor: y() && !m() ? `ew-resize` : `default`,
				"touch-action": y() ? `none` : `auto`,
				"user-select": `none`,
				"-webkit-user-select": `none`,
				"min-width": t.emphasized ? `96px` : void 0,
				transform: e === 0 ? `translateX(0)` : `translateX(${e}px)`,
				transition: o() ? `none` : `transform 350ms ${Ke$1}`
			};
		};
		return (() => {
			var i = Kr$1(), a = i.firstChild, s = a.firstChild, g = s.nextSibling, v = g.nextSibling, y = v.nextSibling;
			return a.addEventListener(`lostpointercapture`, te), a.addEventListener(`pointercancel`, te), a.$$pointerup = ee, a.addEventListener(`pointerleave`, () => p(!1)), a.addEventListener(`pointerenter`, () => p(!0)), a.$$pointermove = k, a.$$pointerdown = O, En(a, Gt(en, {
				get when() {
					return t.emphasized;
				},
				get children() {
					var e = Gr$1();
					return lt((t) => Cn(e, `opacity`, m() ? 0 : .55)), e;
				}
			}), s), En(g, Gt($t, {
				each: Zr$1,
				children: (e) => (() => {
					var t = qr$1();
					return lt((n) => {
						var r = `${e()}%`, i = j() ? .4 : 0;
						return r !== n.e && Cn(t, `left`, n.e = r), i !== n.t && Cn(t, `opacity`, n.t = i), n;
					}, {
						e: void 0,
						t: void 0
					}), t;
				})()
			})), En(y, Gt(en, {
				get when() {
					return t.label;
				},
				children: (e) => (() => {
					var t = Jr$1();
					return yn(t, `${xr$1} text-[var(--rg-text-primary)] truncate min-w-0`), En(t, e), t;
				})()
			}), null), En(y, Gt(en, {
				get when() {
					return m();
				},
				get fallback() {
					return (() => {
						var n = Yr$1(), r = n.firstChild, i = r.firstChild, a = h;
						return typeof a == `function` ? Tn(a, n) : h = n, En(n, Gt(en, {
							get when() {
								return t.tailwindLabel;
							},
							children: (e) => (() => {
								var t = Xr$1();
								return En(t, e), t;
							})()
						}), r), yn(r, `${Sr$1} text-[var(--rg-text-primary)]`), En(r, Gt(Wr$1, { get children() {
							return Oa(t.value);
						} }), i), En(i, () => t.unit), lt((e) => {
							var r = `${Oa(t.value)}${t.unit}`, i = t.onCommitValue ? `text` : `default`;
							return r !== e.e && gn(n, `data-react-grab-value`, e.e = r), i !== e.t && Cn(n, `cursor`, e.t = i), e;
						}, {
							e: void 0,
							t: void 0
						}), n;
					})();
				},
				get children() {
					return Gt(br$1, {
						autoFocusSelect: !0,
						inputmode: `decimal`,
						"aria-label": `Style value`,
						class: `${Sr$1} text-right pointer-events-auto ml-auto`,
						get style() {
							return {
								"field-sizing": `content`,
								"min-width": `16px`,
								"max-width": t.emphasized ? `120px` : `60px`
							};
						},
						get value() {
							return n() ?? ``;
						},
						onInput: (e) => r(e.currentTarget.value),
						onKeyDown: re,
						onBlur: ne,
						onPointerDown: (e) => e.stopPropagation(),
						onMouseDown: (e) => e.stopPropagation(),
						onClick: (e) => e.stopPropagation()
					});
				}
			}), null), En(i, Gt(en, {
				get when() {
					return t.emphasized;
				},
				get children() {
					return [Gt(kr$1, {
						direction: `left`,
						get active() {
							return t.activeKey === `left`;
						},
						onPointerDown: () => t.onStep(-1)
					}), Gt(kr$1, {
						direction: `right`,
						get active() {
							return t.activeKey === `right`;
						},
						onPointerDown: () => t.onStep(1)
					})];
				}
			}), null), lt((e) => {
				var n = t.label ?? `Value`, r = t.min, i = t.max, u = t.value, d = `${Oa(t.value)}${t.unit}`, f = P(), p = `${_()}%`, h = +!m(), g = `calc(${_()}% - 1px)`, y = t.activeKey ? .9 : o() || l() ? .35 : 0;
				return n !== e.e && gn(a, `aria-label`, e.e = n), r !== e.t && gn(a, `aria-valuemin`, e.t = r), i !== e.a && gn(a, `aria-valuemax`, e.a = i), u !== e.o && gn(a, `aria-valuenow`, e.o = u), d !== e.i && gn(a, `aria-valuetext`, e.i = d), e.n = Sn(a, f, e.n), p !== e.s && Cn(s, `width`, e.s = p), h !== e.h && Cn(s, `opacity`, e.h = h), g !== e.r && Cn(v, `left`, e.r = g), y !== e.d && Cn(v, `opacity`, e.d = y), e;
			}, {
				e: void 0,
				t: void 0,
				a: void 0,
				o: void 0,
				i: void 0,
				n: void 0,
				s: void 0,
				h: void 0,
				r: void 0,
				d: void 0
			}), i;
		})();
	};
	hn([
		`pointerdown`,
		`pointermove`,
		`pointerup`
	]);
	ei$1 = (t) => Gt(tn, { get children() {
		return [
			Gt(nn, {
				get when() {
					return Nr$1(t.property);
				},
				children: (r) => Gt($r$1, {
					get label() {
						return dn(() => !!t.showLabel)() ? r().label : void 0;
					},
					get value() {
						return r().value;
					},
					get min() {
						return r().min;
					},
					get max() {
						return r().max;
					},
					get unit() {
						return r().unit;
					},
					get activeKey() {
						return t.activeKey;
					},
					get onStep() {
						return t.onStep;
					},
					get onCommitValue() {
						return t.onCommit;
					},
					get onEditComplete() {
						return t.onEditComplete;
					},
					get onInvalidCommit() {
						return t.onInvalidCommit;
					},
					get onInteract() {
						return t.onInteract;
					},
					get tailwindLabel() {
						return t.tailwindLabel;
					},
					get emphasized() {
						return t.emphasized;
					}
				})
			}),
			Gt(nn, {
				get when() {
					return Pr$1(t.property);
				},
				children: (r) => Gt(Dr$1, {
					get label() {
						return dn(() => !!t.showLabel)() ? r().label : void 0;
					},
					get value() {
						return r().value;
					},
					get onCommit() {
						return t.onCommit;
					},
					get onEditComplete() {
						return t.onEditComplete;
					},
					get onInvalidCommit() {
						return t.onInvalidCommit;
					},
					get onRegisterTrigger() {
						return t.onColorPickerRegister;
					},
					get onInteract() {
						return t.onInteract;
					},
					get emphasized() {
						return t.emphasized;
					}
				})
			}),
			Gt(nn, {
				get when() {
					return Fr$1(t.property);
				},
				children: (r) => Gt(Mr$1, {
					get label() {
						return dn(() => !!t.showLabel)() ? r().label : void 0;
					},
					get value() {
						return r().value;
					},
					get options() {
						return r().options;
					},
					get activeKey() {
						return t.activeKey;
					},
					get onStep() {
						return t.onStep;
					}
				})
			})
		];
	} });
	ti$1 = /*#__PURE__*/ mn(`<span class="text-[var(--rg-text-primary)] text-[13px] leading-3.5 font-sans font-medium">Copy`);
	ni$1 = (t) => Gt(Mt$1, {
		"data-react-grab-ignore-events": !0,
		"data-react-grab-copy-button": !0,
		onClick: (e) => {
			e.preventDefault(), e.stopPropagation(), t.onCopy();
		},
		get children() {
			return ti$1();
		}
	});
	ri$1 = () => {
		let [e, t] = z(!1), n, r = () => {
			clearTimeout(n), t(!1);
		}, i = () => {
			t(!0), clearTimeout(n), n = setTimeout(() => {
				t(!1);
			}, Ge$1);
		}, a = () => {
			clearTimeout(n);
		};
		return W(a), {
			isPending: e,
			show: i,
			hide: r,
			cleanup: a
		};
	};
	ii$1 = /*#__PURE__*/ mn(`<div role=menu aria-orientation=vertical aria-label="Editable properties"class="relative flex flex-col w-full max-h-[var(--rg-edit-list-max-h)] overflow-y-auto outline-none no-scrollbar"><div aria-hidden=true class="pointer-events-none absolute opacity-0 transition-[top,left,width,height,opacity,border-radius] duration-75 ease-out bg-[var(--rg-surface-hover)]">`);
	ai$1 = /*#__PURE__*/ mn(`<div class="flex items-center w-full h-[24px]">`);
	oi$1 = /*#__PURE__*/ mn(`<button data-react-grab-ignore-events type=button role=menuitem tabindex=-1 class="relative z-1 contain-layout block w-full h-[24px] px-0 py-0 cursor-pointer text-left border-none bg-transparent">`);
	si$1 = /*#__PURE__*/ mn(`<div class="flex items-center justify-between w-full h-[24px] px-2 gap-2"><span class="text-[13px] leading-4 font-sans font-medium text-[var(--rg-text-primary)] truncate min-w-0">`);
	ci$1 = /*#__PURE__*/ mn(`<span class="font-sans text-[var(--rg-text-secondary)] tabular-nums shrink-0"><span class=text-[11px]></span><span class="text-[9px] ml-px">`);
	li$1 = /*#__PURE__*/ mn(`<span aria-hidden=true class="size-[12px] rounded-[3px] border-[var(--rg-border-button)] [border-width:0.5px] border-solid shrink-0">`);
	ui$1 = /*#__PURE__*/ mn(`<span class="text-[11px] font-sans text-[var(--rg-text-secondary)] shrink-0">`);
	di$1 = (e) => {
		let t = Fr$1(e);
		return t ? t.options.find((e) => e.value === t.value)?.label ?? t.value : ``;
	};
	fi$1 = (t) => {
		let n = [], r, i = !1, a = null, o = () => {
			if (!r) return !1;
			let e = _r$1(r);
			return e instanceof HTMLElement && e.matches(`input[data-react-grab-input]`);
		}, s = (e, n) => {
			n === `move` && (i = !0);
			let r = o();
			i && (r || (t.isAdjusting() ? a = e : (a = null, e !== t.activeIndex && t.onHoverIndex(e))));
		};
		B(() => {
			if (t.isAdjusting()) return;
			let e = a;
			if (e === null) return;
			a = null;
			let r = n[e], s = o();
			i && (s || r?.matches(`:hover`) && e !== t.activeIndex && t.onHoverIndex(e));
		});
		let { containerRef: l, highlightRef: f, updateHighlight: p, clearHighlight: m } = or$1({});
		B(() => {
			n.length = t.properties.length;
		});
		let h;
		return B(() => {
			let e = t.activeIndex, i = n[e];
			if (!i || e < 0) {
				m();
				return;
			}
			if (p(i), h !== void 0 && cancelAnimationFrame(h), h = requestAnimationFrame(() => {
				h = void 0;
				let t = n[e];
				t && p(t);
			}), !r) return;
			let a = r.getBoundingClientRect(), o = i.getBoundingClientRect();
			(o.top < a.top || o.bottom > a.bottom) && i.scrollIntoView({ block: `nearest` });
		}), W(() => {
			h !== void 0 && cancelAnimationFrame(h);
		}), (() => {
			var o = ii$1(), p = o.firstChild;
			return o.$$pointermove = () => {
				i = !0;
			}, Tn((e) => {
				r = e, l(e);
			}, o), Tn(f, p), En(o, Gt($t, {
				get each() {
					return t.properties;
				},
				children: (i, o) => {
					let l = () => o === t.activeIndex;
					return (() => {
						var u = oi$1();
						return u.$$click = (e) => {
							e.stopPropagation(), t.onSelect(o);
						}, u.$$mousedown = (e) => {
							let t = r ? _r$1(r) : null;
							t instanceof HTMLElement && t.matches(`input[data-react-grab-input]`) && t.blur(), e.preventDefault();
						}, u.addEventListener(`pointerleave`, () => {
							a === o && (a = null);
						}), u.$$pointermove = () => {
							s(o, `move`);
						}, u.addEventListener(`pointerenter`, () => {
							s(o, `enter`);
						}), Tn((e) => {
							n[o] = e, W(() => {
								n[o] === e && (n[o] = void 0);
							});
						}, u), En(u, Gt(en, {
							get when() {
								return l();
							},
							get fallback() {
								return (() => {
									var t = si$1(), n = t.firstChild;
									return En(n, () => i().label), En(t, Gt(tn, { get children() {
										return [
											Gt(nn, {
												get when() {
													return Nr$1(i());
												},
												children: (e) => (() => {
													var t = ci$1(), n = t.firstChild, r = n.nextSibling;
													return En(n, () => Oa(e().value)), En(r, () => e().unit), t;
												})()
											}),
											Gt(nn, {
												get when() {
													return Pr$1(i());
												},
												children: (e) => (() => {
													var t = li$1();
													return lt((n) => Cn(t, `background-color`, e().value)), t;
												})()
											}),
											Gt(nn, {
												get when() {
													return Fr$1(i());
												},
												children: (e) => (() => {
													var t = ui$1();
													return En(t, () => di$1(e())), t;
												})()
											})
										];
									} }), null), t;
								})();
							},
							get children() {
								var n = ai$1();
								return En(n, Gt(ei$1, {
									get property() {
										return i();
									},
									get activeKey() {
										return t.activeKey;
									},
									get onStep() {
										return t.onStep;
									},
									get onCommit() {
										return t.onCommit;
									},
									get onEditComplete() {
										return t.onEditComplete;
									},
									get onInvalidCommit() {
										return t.onInvalidCommit;
									},
									get onInteract() {
										return t.onInteract;
									},
									get onColorPickerRegister() {
										return t.onColorPickerRegister;
									},
									showLabel: !0,
									get tailwindLabel() {
										return t.activeTailwindLabel;
									}
								})), n;
							}
						})), lt((e) => {
							var t = i().key, n = l() ? `true` : void 0;
							return t !== e.e && gn(u, `data-react-grab-edit-property`, e.e = t), n !== e.t && gn(u, `aria-current`, e.t = n), e;
						}, {
							e: void 0,
							t: void 0
						}), u;
					})();
				}
			}), null), o;
		})();
	};
	hn([
		`pointermove`,
		`mousedown`,
		`click`
	]);
	pi$1 = (e, t, n) => e.kind === `color` && typeof t == `string` && typeof n == `string` ? t.toLowerCase() === n.toLowerCase() : t === n;
	mi$1 = (e) => e === `ArrowLeft` ? -1 : 1;
	hi$1 = (e) => {
		let [t, n] = z(0), r = null, i = null, a = null, o = () => {
			i !== null && clearTimeout(i), a !== null && clearInterval(a), i = null, a = null, r = null;
		}, s = () => {
			o(), n(0);
		}, c = (t) => {
			o(), r = t;
			let n = mi$1(t);
			i = setTimeout(() => {
				a = setInterval(() => {
					e.step(n, e.isShiftHeld(), e.isAltHeld(), !0);
				}, 60);
			}, 280);
		};
		return pt(() => {
			window.addEventListener(`blur`, s), W(() => window.removeEventListener(`blur`, s));
		}), W(s), {
			heldDirection: t,
			pressArrow: (t, r, i, a) => {
				if (r) return;
				let o = mi$1(t);
				c(t), n(o), e.step(o, i, a, !1);
			},
			releaseKey: (e) => {
				e === r && s();
			},
			cancelRepeat: s
		};
	};
	gi$1 = (e, t, n) => {
		if (e.length === 0) return null;
		let r = e.findIndex((e) => e.value === t), i = e[((r === -1 ? 0 : r) + n + e.length) % e.length];
		return !i || i.value === t ? null : i;
	};
	_i$1 = (e, t, n, r, a) => {
		if (e.kind === `enum`) return gi$1(e.options, e.value, t)?.value ?? null;
		if (e.kind !== `numeric`) return null;
		let o = Math.min(e.min, e.value), s = Math.max(e.max, e.value);
		if (!n && !r && e.unit === `px` && a?.hasTokens) {
			let n = a.stepLength(e.value, t, e.cssProperties[0]);
			if (n !== null) {
				let t = ka(Ci(n, o, s));
				if (t !== e.value) return t;
			}
		}
		let c = n ? 10 : 1, l = ka(Ci(e.value + t * c, o, s));
		return l === e.value ? null : l;
	};
	vi$1 = /^-?\d*\.?\d+$/;
	yi$1 = (e) => vi$1.test(e);
	bi$1 = (e) => e.toLowerCase().replace(/[^a-z0-9]+/g, ``);
	xi$1 = (e) => {
		if (e.kind !== `enum`) return [];
		let t = [];
		for (let n of e.options) {
			let e = bi$1(n.value), r = bi$1(n.label);
			e && t.push(e), r && r !== e && t.push(r);
		}
		return t;
	};
	Si$1 = (e) => {
		let t = [];
		for (let n = 0; n < e.length; n++) {
			let r = e[n];
			for (let e of r.tailwindAliases) {
				let i = bi$1(e);
				i && t.push({
					property: r,
					originalIndex: n,
					kind: `alias`,
					term: i
				});
			}
			for (let e of xi$1(r)) t.push({
				property: r,
				originalIndex: n,
				kind: `enum-option`,
				term: e
			});
			let i = bi$1(r.label);
			i && t.push({
				property: r,
				originalIndex: n,
				kind: `name`,
				term: i
			});
			let a = bi$1(r.key);
			a && a !== i && t.push({
				property: r,
				originalIndex: n,
				kind: `name`,
				term: a
			});
			let o = bi$1(`${r.label} ${r.key}`);
			o && t.push({
				property: r,
				originalIndex: n,
				kind: `label`,
				term: o
			});
		}
		return t;
	};
	Ci$1 = (e, t) => {
		if (t.kind === `alias` || t.kind === `enum-option` || t.kind === `name`) return e === t.term ? Ye$1 + t.term.length : t.term.startsWith(e) || e.startsWith(t.term) ? Xe$2 + Math.min(e.length, t.term.length) : null;
		let n = t.term.indexOf(e);
		return n === -1 ? null : n === 0 ? Ze$2 - t.term.length * 1 : Qe$2 - n * 100 - t.term.length * 1;
	};
	wi$1 = (e, t, n, r) => {
		let i = e.get(t.key);
		i && i.score >= r || e.set(t.key, {
			property: t,
			score: r,
			originalIndex: n
		});
	};
	Ti$1 = (e) => {
		let t = Si$1(e), n = /* @__PURE__ */ new Map();
		for (let t = 0; t < e.length; t++) n.set(e[t].key, t);
		let r = (t, r, i) => {
			for (let a = 0; a < r.length; a++) {
				let o = r[a], s = n.get(o);
				if (s === void 0) continue;
				let c = e[s];
				wi$1(t, c, s, i - a);
			}
		};
		return { search: (n) => {
			let i = bi$1(n);
			if (!i || yi$1(i)) return e;
			let a = /* @__PURE__ */ new Map();
			r(a, so(n), qe$1), r(a, co(n), Je$1);
			for (let e of t) {
				let t = Ci$1(i, e);
				t !== null && wi$1(a, e.property, e.originalIndex, t);
			}
			return Array.from(a.values()).sort((e, t) => t.score === e.score ? e.originalIndex - t.originalIndex : t.score - e.score).map((e) => e.property);
		} };
	};
	Ei$1 = (e, t) => e.kind === t.kind && !pi$1(e, t.value, e.original);
	Di$1 = (e) => {
		let { initialProperties: t, searchQuery: n } = e, [r, i] = z({}), a = Ti$1(t), o = V(() => {
			let e = n(), i = r(), o = e ? t : t.filter((e) => e.isPrioritized || e.isCanonical && !e.isDefault || i[e.key] !== void 0);
			return e ? a.search(e) : o.slice(0, 40);
		}), s = new Map(t.map((e) => [e.key, e])), c = V(() => {
			let e = r(), t = Object.keys(e);
			if (t.length === 0) return o();
			let n = /* @__PURE__ */ new Map();
			for (let r of t) {
				let t = e[r];
				if (t.kind !== `numeric`) continue;
				let i = s.get(r);
				if (i) for (let e of i.cssProperties) n.set(e, t.value);
			}
			return o().map((t) => {
				let r = e[t.key];
				if (r !== void 0) return l(t, r);
				if (t.kind !== `numeric`) return t;
				let i = n.get(t.cssProperties[0]);
				return i === void 0 ? t : t.cssProperties.every((e) => n.get(e) === i) ? {
					...t,
					value: i
				} : t;
			});
		}), l = (e, t) => e.kind === `numeric` && t.kind === `numeric` || e.kind === `color` && t.kind === `color` || e.kind === `enum` && t.kind === `enum` ? {
			...e,
			value: t.value
		} : e;
		return {
			filteredProperties: c,
			applyStyle: (e, t) => {
				let n;
				if (e.kind === `numeric` && typeof t == `number`) n = {
					kind: `numeric`,
					value: t
				};
				else if (e.kind === `color` && typeof t == `string`) n = {
					kind: `color`,
					value: t
				};
				else if (e.kind === `enum` && typeof t == `string`) n = {
					kind: `enum`,
					value: t
				};
				else return;
				i((t) => {
					let r = Object.keys(t);
					if (r[r.length - 1] === e.key) return {
						...t,
						[e.key]: n
					};
					let i = {};
					for (let n of r) {
						if (n === e.key) continue;
						let r = s.get(n);
						r !== void 0 && r.cssProperties.every((t) => e.cssProperties.includes(t)) || (i[n] = t[n]);
					}
					return i[e.key] = n, i;
				});
			},
			buildPendingEdits: () => {
				let e = r(), t = [], n = /* @__PURE__ */ new Set();
				for (let r of Object.keys(e)) {
					let i = e[r], a = s.get(r);
					if (!a || a.kind !== i.kind) continue;
					let o = Ei$1(a, i), c = !o && a.cssProperties.some((e) => n.has(e));
					if (o || c) {
						if (o) for (let e of a.cssProperties) n.add(e);
						a.kind === `numeric` && i.kind === `numeric` ? t.push({
							kind: `numeric`,
							key: a.key,
							cssProperties: a.cssProperties,
							value: i.value,
							unit: a.unit
						}) : a.kind === `color` && i.kind === `color` ? t.push({
							kind: `color`,
							key: a.key,
							cssProperties: a.cssProperties,
							value: i.value
						}) : a.kind === `enum` && i.kind === `enum` && t.push({
							kind: `enum`,
							key: a.key,
							cssProperties: a.cssProperties,
							value: i.value
						});
					}
				}
				return t;
			},
			hasPendingStyles: () => {
				let e = r();
				for (let n of t) {
					let t = e[n.key];
					if (t && Ei$1(n, t)) return !0;
				}
				return !1;
			},
			hasChangedStyleFor: (e) => {
				let t = r()[e], n = s.get(e);
				return !!(t && n && Ei$1(n, t));
			}
		};
	};
	Oi$1 = /* @__PURE__ */ new Map();
	for (let e of fa) for (let t of e) Oi$1.set(t.key, t.longhands);
	ki$1 = (e) => {
		if (e.includes(`,`)) return e.split(`,`);
		let t = Oi$1.get(e);
		return t ? [...t] : [e];
	};
	Ai$1 = /^-?(?:\d+\.?|\d*\.\d*)?$/;
	ji$1 = (e) => Ai$1.test(e);
	Mi$1 = (e) => {
		let t = e.startsWith(`-`);
		return {
			isNegative: t,
			base: t ? e.slice(1) : e
		};
	};
	Ni$1 = /^([a-z-]+)-(-?\d+(?:\.\d+)?)$/;
	Pi$1 = /^(.+?)-\[(.+)]$/;
	Fi$1 = /^(-?\d*\.?\d+)\s*([a-zA-Z%]*)$/;
	Ii$1 = /^(-?\d*\.?\d+)(px|rem)?$/i;
	Li$1 = /^(?:length|size):/i;
	Ri$1 = /* @__PURE__ */ new Set([
		`opacity`,
		`border-width`,
		`border-top-width`,
		`border-right-width`,
		`border-bottom-width`,
		`border-left-width`,
		`z-index`,
		`font-weight`
	]);
	zi$1 = (e, t) => {
		for (let n of e) if (n.key === t && n.kind === `numeric`) return n;
		return null;
	};
	Bi$1 = (e, t) => {
		for (let n of e) if (n.key === t && n.kind === `enum`) return n;
		return null;
	};
	Vi$1 = (e, t) => {
		for (let n of e) if (n.key === t && n.kind === `color`) return n;
		return null;
	};
	Hi$1 = (e) => {
		let t = e.replace(/^(?:color|length|size):/i, ``).trim();
		return /(?:var|url)\(/i.test(t) ? t : t.replace(/_/g, ` `);
	};
	Ui$1 = (e) => {
		let t = e.lastIndexOf(`:`), n = t >= 0 ? e.slice(t + 1) : e;
		return (n.startsWith(`!`) ? n.slice(1) : n).toLowerCase();
	};
	Wi$1 = (e, t) => {
		let n = e.match(Ii$1);
		if (!n) return null;
		let r = n[2]?.toLowerCase();
		if (!r && !t) return null;
		let i = Number.parseFloat(n[1]);
		return Number.isFinite(i) ? r === `rem` ? i * 16 : i : null;
	};
	Gi$1 = (e, t) => {
		let n = ki$1(t), r = /* @__PURE__ */ new Map();
		for (let t of n) for (let n of e) n.kind === `numeric` && n.cssProperties.includes(t) && r.set(n.key, n);
		return Array.from(r.values());
	};
	Ki$1 = (e, t) => ka(Ci(t, e.min, e.max));
	qi$1 = (e) => {
		let { initialProperties: t, searchQuery: n, isCompact: i, activeProperty: a, commit: o, setIsCompact: s } = e, c = (e) => {
			for (let n of t) if (n.key === e && (n.kind === `numeric` || n.kind === `enum`)) return !0;
			return Gi$1(t, e).length > 0;
		}, l = (e) => {
			let t = e.trim().replace(/(\d),(\d)/g, `$1.$2`).match(Fi$1);
			if (!t) return null;
			let n = Number.parseFloat(t[1]);
			return Number.isFinite(n) ? {
				value: n,
				unit: t[2].toLowerCase()
			} : null;
		}, u = (e, t) => !e || t.toLowerCase().startsWith(e), d = (e) => {
			if (!i()) return !1;
			let t = a();
			if (t?.kind !== `numeric`) return !1;
			let n = l(e);
			return n ? u(n.unit, t.unit) : ji$1(e);
		}, f = V(() => d(n())), m = (e) => {
			if (!i()) return !1;
			let t = a();
			if (t?.kind !== `numeric`) return !1;
			let n = l(e);
			if (!n) return !1;
			let r = t.unit.toLowerCase();
			if (n.unit !== `` && n.unit !== r) return u(n.unit, t.unit);
			let s = Ki$1(t, n.value);
			return s !== t.value && o(t, s), !0;
		}, h = (e, n) => {
			let r = zi$1(t, e);
			if (r) return r.unit === `px` && (s(!0), o(r, Ki$1(r, n), { shouldCompact: !0 }), !0);
			let i = Gi$1(t, e);
			return i.length !== 0 && (s(!0), H(() => {
				for (let e of i) o(e, Ki$1(e, n), { shouldCompact: !0 });
			}), !0);
		}, g = (e, n) => {
			let r = to(e);
			if (!r) return !1;
			let i = Vi$1(t, r);
			if (!i) return !1;
			let a = n === null ? ro(e) : Gr(n);
			return a ? (s(!0), o(i, a, { shouldCompact: !0 }), !0) : !1;
		}, v = (e) => {
			let t = e.trim(), n = t.match(Pi$1);
			if (!n) return !1;
			let r = n[2], i = Hi$1(r);
			if (g(t, i)) return !0;
			let a = Wi$1(i, Li$1.test(r.trim()));
			if (a === null) return !1;
			let { isNegative: o, base: s } = Mi$1(n[1]), c = fo(Ui$1(s));
			return c ? h(c, o ? -a : a) : !1;
		}, b = (e) => {
			if (v(e)) return;
			let n = Qa(e), { isNegative: i, base: a } = Mi$1(n), l = a.replace(/-\d*$/, ``).replace(/-$/, ``), u = l ? fo(l) : null;
			if (u && c(u) && s(!0), !i) {
				if (g(n, null)) return;
				let e = po(n);
				if (e) {
					let n = Bi$1(t, e.property);
					if (!n) return;
					let r = n.options.find((t) => t.value === e.value);
					r && (s(!0), o(n, r.value, { shouldCompact: !0 }));
					return;
				}
			}
			let d = a.match(Ni$1);
			if (!d) return;
			let f = fo(d[1]);
			if (!f) return;
			let p = Number.parseFloat(d[2]);
			if (!Number.isFinite(p)) return;
			let m = i ? -p : p, h = Ri$1.has(f) ? m : m * 4, b = zi$1(t, f);
			if (b) {
				o(b, Ki$1(b, h), { shouldCompact: !0 });
				return;
			}
			if (!i) {
				let e = Bi$1(t, f);
				if (e) {
					let t = String(p), n = e.options.find((e) => e.value === t);
					if (n) {
						o(e, n.value, { shouldCompact: !0 });
						return;
					}
				}
			}
			let x = Gi$1(t, f);
			x.length !== 0 && H(() => {
				for (let e of x) o(e, Ki$1(e, h), { shouldCompact: !0 });
			});
		}, x = (e) => {
			if (po(e)) return !0;
			let { base: n } = Mi$1(e), i = n.match(Ni$1);
			if (!i) return !1;
			let a = fo(i[1]);
			return a ? c(a) ? !0 : Bi$1(t, a) !== null : !1;
		};
		return {
			isInlineNumericEdit: f,
			isInlineNumericDraft: d,
			tryApplyNumericValue: m,
			applyTailwindClass: (e) => {
				let t = e.trim().replace(/^class\s*=\s*["']/, ``).replace(/["']\s*$/, ``).replace(/\s+\[/g, `-[`), n = Qa(t), r = t.split(/\s+/).filter(Boolean);
				if (r.length > 1 && !x(n)) for (let e of r) b(e);
				else b(t);
			}
		};
	};
	Ji$1 = /*#__PURE__*/ mn(`<span class="text-[var(--rg-text-primary)] text-[13px] leading-3.5 font-sans font-medium">No`);
	Yi$1 = /*#__PURE__*/ mn(`<span class="text-[var(--rg-error-text)] text-[13px] leading-3.5 font-sans font-medium">Yes`);
	Xi$1 = /*#__PURE__*/ mn(`<div role=alertdialog aria-modal=true aria-labelledby=rg-discard-prompt-title class="contain-layout shrink-0 flex items-center justify-between gap-2 pt-1.5 pb-1 px-2 w-full self-stretch"><span id=rg-discard-prompt-title class="text-[var(--rg-text-primary)] text-[13px] leading-4 font-sans font-medium">Discard edits?</span><div class="flex items-center gap-[5px]">`);
	Zi$1 = /*#__PURE__*/ mn(`<div><textarea data-react-grab-ignore-events data-react-grab-input aria-label="Search properties"aria-keyshortcuts="Enter Escape ArrowUp ArrowDown ArrowLeft ArrowRight Tab"autocapitalize=none autocorrect=off autocomplete=off class="text-[var(--rg-text-primary)] text-[13px] leading-4 font-medium bg-transparent border-none resize-none w-full p-0 m-0 outline-none"rows=1 style=caret-color:var(--rg-text-primary);field-sizing:content;min-height:16px;max-height:16px;scrollbar-width:none>`);
	Qi$1 = /*#__PURE__*/ mn(`<div>`);
	$i$1 = /*#__PURE__*/ mn(`<div role=dialog aria-modal=true aria-label="Style element"data-react-grab-ignore-events data-react-grab-edit-panel>`);
	ea$1 = /*#__PURE__*/ mn(`<div class="flex items-center justify-center w-full px-3 py-1.5 min-h-[28px]">`);
	ta$1 = (t) => Gt(en, {
		get when() {
			return t.state;
		},
		children: (n) => Gt(en, {
			keyed: !0,
			get when() {
				return n().element;
			},
			children: (r) => Gt(na$1, {
				get state() {
					return n();
				},
				position: () => t.position,
				get onDismiss() {
					return t.onDismiss;
				},
				get onSubmit() {
					return t.onSubmit;
				},
				get onPendingEditsChange() {
					return t.onPendingEditsChange;
				},
				get onInteractingChange() {
					return t.onInteractingChange;
				}
			})
		})
	});
	na$1 = (t) => {
		let r = t.state.properties, s, l = t.state.preview, [u, p] = z(t.state.initialSearchQuery ?? ``), [m, h] = z(null), [g, _] = z(null), v = Di$1({
			initialProperties: r,
			searchQuery: () => m() ?? u()
		}), y = () => {
			let e = v.filteredProperties().findIndex((e) => e.kind === `numeric`);
			return e > 0 ? e : 0;
		}, [b, x] = z(y()), T = V(() => v.hasPendingStyles()), E = V(() => T() || !!t.state.hasSessionEdits), [D, O] = z(!1), k, ee, te, j = !1, [ne, P] = z(!1), ie = V(() => ne() || T()), [I, oe] = z(!1), se = Cs(), ce$4 = V(() => at$1({
			tagName: t.state.tagName,
			componentName: t.state.componentName
		})), V$4 = V(() => {
			let e = v.filteredProperties();
			return e.length === 0 ? null : e[Math.min(Math.max(0, b()), e.length - 1)];
		}), le, H = mr$1(() => le, t.position), ue = (e) => {
			_(e), clearTimeout(k), k = setTimeout(() => {
				_((t) => t === e ? null : t);
			}, 100);
		}, de = V(() => ie() && !I());
		B(() => {
			let e = de();
			t.onInteractingChange?.(e);
		});
		let fe = () => {
			P(!0), clearTimeout(ee), ee = setTimeout(() => {
				P(!1);
			}, 150);
		}, W$4 = () => {
			queueMicrotask(() => {
				s && _r$1(s) !== s && et$1(s, { preventScroll: !0 });
			});
		}, G = () => {
			m() === null && h(u());
		}, pe = () => {
			j = !1, clearTimeout(te);
		}, me = () => {
			j = !1, clearTimeout(te), te = setTimeout(() => {
				j = !0;
			}, 500);
		}, ge = (e) => {
			let t = e.trim(), n = t.replace(/^-/, ``).replace(`.`, ``);
			/[a-z%]+$/i.test(t) || n.length > 1 ? me() : pe();
		}, _e = (e) => {
			if (!j) return e;
			let t = u();
			if (!t || !e.startsWith(t)) return pe(), e;
			let n = e.slice(t.length);
			return /^[-.\d]/.test(n) ? (pe(), n) : e;
		}, q = (e) => {
			if (!j || e.metaKey || e.ctrlKey || e.altKey || !/^[-.\d]$/.test(e.key)) return !1;
			e.preventDefault(), e.stopImmediatePropagation(), pe();
			let t = e.key;
			return s && (s.value = t), J.tryApplyNumericValue(t) ? (G(), p(t), ge(t), W$4(), !0) : (p(t), W$4(), !0);
		}, ve = () => {
			pe(), h(null), O(!1);
		}, ye = (e, n, r = {}) => {
			v.applyStyle(e, n), l.apply(e.cssProperties, Aa(e, n)), t.onPendingEditsChange?.(v.buildPendingEdits()), fe(), r.isFromKeyRepeat || Y.hide(), r.flashDirection && ue(r.flashDirection === 1 ? `right` : `left`), r.shouldFocus && W$4(), r.shouldCompact && O(!0), r.source === `keyboard` && se.arm();
		}, be = ms((e) => e.shiftKey), xe = ms((e) => e.altKey), Se = (e, n, r, i, a) => {
			let o = V$4();
			if (!o) return null;
			let s = _i$1(o, e, n, r, t.state.designTokens);
			return s === null ? (ue(e === 1 ? `right` : `left`), null) : (ye(o, s, {
				flashDirection: e,
				shouldFocus: !0,
				isFromKeyRepeat: i,
				source: a
			}), o);
		}, Ce = (e, t, n, r) => {
			Se(e, t, n, r, `keyboard`) && O(!0);
		}, Te = (e) => {
			Se(e, !1, !1, !1, `pointer`);
		}, Ee = hi$1({
			step: Ce,
			isShiftHeld: be,
			isAltHeld: xe
		}), De = (e, t) => {
			let n = V$4();
			if (n) {
				if (n.kind === `numeric` && typeof e == `number`) {
					let r = ka(Ci(e, n.min, n.max));
					r !== n.value && ye(n, r, { source: t });
				} else typeof e == `string` && (pi$1(n, e, n.value) || ye(n, e, { source: t }));
			}
		}, Oe = V(() => {
			if (!be()) return null;
			let e = V$4();
			return !e || e.kind !== `numeric` ? null : gr$1(e.key, e.value);
		}), J = qi$1({
			initialProperties: r,
			searchQuery: u,
			isCompact: D,
			activeProperty: V$4,
			commit: (e, t, n) => {
				ye(e, t, {
					...n,
					source: `keyboard`
				});
			},
			setIsCompact: O
		}), Ae = V(() => D() && u() !== `` && J.isInlineNumericEdit()), je = () => {
			Y.hide(), se.clear(), t.onSubmit(v.buildPendingEdits());
		}, Y = ri$1(), Me, Ne = () => {
			Me && (Me.classList.remove(`animate-shake`), Me.offsetWidth, Me.classList.add(`animate-shake`));
		}, Pe = (e) => {
			Y.hide(), e === `discard` && l.restore(), t.onDismiss();
		}, Fe = (e) => {
			if (Ee.cancelRepeat(), e === `pointer` && se.clear(), Y.isPending()) {
				Pe(`discard`);
				return;
			}
			if (!E()) {
				Pe(l.hasAppliedStyles() ? `discard` : `preserve`);
				return;
			}
			let t = D();
			ve(), e === `keyboard` && t || (Y.show(), Ne());
		}, Ie = (e) => {
			let t = v.filteredProperties();
			t.length !== 0 && (x((n) => (n + e + t.length) % t.length), ve());
		}, Le = () => {
			se.clear(), Y.hide();
		}, X = [], Re = (e) => (X.push(e), () => {
			X = X.filter((t) => t !== e);
		}), ze = () => X[X.length - 1] ?? null, Be = (e, t) => {
			V$4()?.kind === `color` ? t.repeat || ze()?.() : Ee.pressArrow(e, t.repeat, t.shiftKey, t.altKey);
		}, Ve$1 = {
			ArrowUp: () => Ie(-1),
			ArrowDown: () => Ie(1),
			ArrowLeft: (e) => Be(`ArrowLeft`, e),
			ArrowRight: (e) => Be(`ArrowRight`, e),
			Tab: (e) => Ie(e.shiftKey ? -1 : 1),
			Enter: () => {
				if (Y.isPending()) return;
				let e = V$4(), t = ze();
				e?.kind === `color` && !v.hasChangedStyleFor(e.key) && t && !E() ? t() : je();
			},
			Escape: () => Fe(`keyboard`)
		}, He = (e) => {
			if (rt$1(e)) return;
			if (Y.isPending()) {
				let t = e.composedPath()[0];
				if (t instanceof HTMLElement && t.closest(`[data-react-grab-discard-button]`) !== null && (e.key === `Tab` || e.key === `Enter`)) return;
				if (e.key === `Tab` || e.key === `ArrowUp` || e.key === `ArrowDown` || e.key === `ArrowLeft` || e.key === `ArrowRight`) {
					e.preventDefault(), e.stopImmediatePropagation();
					return;
				}
			}
			let t = Ve$1[e.key];
			t && (e.preventDefault(), e.stopImmediatePropagation(), t(e));
		};
		pt(() => {
			queueMicrotask(() => {
				if (et$1(s, { preventScroll: !0 }), s) {
					let e = s.value.length;
					s.setSelectionRange(e, e);
				}
			}), H.measure();
			let e = u();
			e && J.applyTailwindClass(e);
			let t = cr$1({
				isOpen: () => !0,
				onDismiss: Fe,
				shouldIgnoreKeyboardEvent: (e) => {
					let t = e.composedPath()[0];
					return Y.isPending() && t instanceof HTMLElement && t.closest(`[data-react-grab-discard-button]`) !== null;
				},
				shouldIgnoreRightClick: !0,
				shouldIgnoreInputEvents: !0
			}), n = zt$2((e) => {
				$i(e, `data-react-grab-input`) || q(e) || He(e);
			}), r = zt$2((e) => {
				Ee.releaseKey(e.key);
			}), i = zt$2((e) => {
				e.pointerType === `mouse` && (Y.isPending() || se.consume() && E() && Fe(`pointer`));
			});
			window.addEventListener(`keydown`, n, { capture: !0 }), window.addEventListener(`keyup`, r, { capture: !0 }), window.addEventListener(`pointermove`, i, { capture: !0 }), W(() => {
				t(), window.removeEventListener(`keydown`, n, { capture: !0 }), window.removeEventListener(`keyup`, r, { capture: !0 }), window.removeEventListener(`pointermove`, i, { capture: !0 }), clearTimeout(k), clearTimeout(ee), clearTimeout(te), Y.cleanup(), H.clearAnimationHandles(), P(!1);
			});
		});
		let Ue = (e) => {
			x(e), W$4();
		};
		return Gt(en, {
			get when() {
				return H.shouldMount();
			},
			get children() {
				var t = $i$1();
				bn(t, `contextmenu`, $$1, !0), bn(t, `click`, $$1, !0), bn(t, `mousedown`, $$1, !0), bn(t, `pointerdown`, $$1, !0);
				var r = le;
				return typeof r == `function` ? Tn(r, t) : le = t, Cn(t, `z-index`, `${ce}`), Cn(t, `--rg-edit-list-max-h`, `280px`), En(t, Gt(wt$1, {
					ref: (e) => Me = e,
					class: `flex flex-col justify-center items-start overflow-hidden w-fit h-fit`,
					get style() {
						return {
							"min-width": D() ? void 0 : `200px`,
							"max-width": `320px`,
							transform: `translateX(${Ee.heldDirection() * 2}px)`,
							transition: `transform 280ms ${Ke$1}`
						};
					},
					get children() {
						return [
							Gt(en, {
								get when() {
									return Y.isPending();
								},
								get fallback() {
									return Gt(en, {
										get when() {
											return !D();
										},
										get children() {
											var t = Qi$1();
											return t.addEventListener(`mouseleave`, () => oe(!1)), t.addEventListener(`mouseenter`, () => oe(!0)), En(t, Gt(vt$1, {
												get tagName() {
													return ce$4().tagName;
												},
												get componentName() {
													return ce$4().componentName;
												},
												isClickable: !1,
												onClick: () => {},
												shrink: !0
											}), null), En(t, Gt(en, {
												get when() {
													return E();
												},
												get children() {
													return Gt(ni$1, { onCopy: je });
												}
											}), null), lt(() => yn(t, Z$1(`contain-layout shrink-0 flex items-center gap-1 pt-1.5 pb-1 h-fit px-2`, E() ? `w-full self-stretch justify-between` : `w-fit`))), t;
										}
									});
								},
								get children() {
									var t = Xi$1(), n = t.firstChild.nextSibling;
									return En(n, Gt(Mt$1, {
										"data-react-grab-ignore-events": !0,
										"data-react-grab-discard-button": `cancel`,
										onClick: (e) => {
											e.preventDefault(), e.stopPropagation(), Le();
										},
										get children() {
											return Ji$1();
										}
									}), null), En(n, Gt(Mt$1, {
										variant: `destructive`,
										"data-react-grab-ignore-events": !0,
										"data-react-grab-discard-button": `confirm`,
										onClick: (e) => {
											e.preventDefault(), e.stopPropagation(), Pe(`discard`);
										},
										get children() {
											return Yi$1();
										}
									}), null), t;
								}
							}),
							(() => {
								var e = Zi$1(), t = e.firstChild;
								return t.$$keydown = He, t.$$input = (e) => {
									let t = _e(e.currentTarget.value);
									t !== e.currentTarget.value && (e.currentTarget.value = t), J.tryApplyNumericValue(t) ? (G(), p(t), ge(t), W$4()) : (pe(), J.isInlineNumericDraft(t) ? (G(), p(t), W$4()) : (p(t), h(null), x(t.trim() === `` ? y() : 0), ve(), J.applyTailwindClass(t)));
								}, Tn((e) => {
									s = e;
								}, t), gn(t, `spellcheck`, !1), lt((n) => {
									var r = Ae() ? `` : `[font-synthesis:none] contain-layout shrink-0 flex flex-col items-start px-2 py-1.5 w-full self-stretch [border-top-width:0.5px] border-t-[var(--rg-border-subtle)] antialiased`, i = Ae() ? Cr$1 : void 0, a = Ae() ? -1 : 0, o = D() ? V$4()?.label ?? `Search property` : `Search property`;
									return r !== n.e && yn(e, n.e = r), n.t = Sn(e, i, n.t), a !== n.a && gn(t, `tabindex`, n.a = a), o !== n.o && gn(t, `placeholder`, n.o = o), n;
								}, {
									e: void 0,
									t: void 0,
									a: void 0,
									o: void 0
								}), lt(() => t.value = u()), e;
							})(),
							Gt(en, {
								get when() {
									return v.filteredProperties().length > 0;
								},
								get children() {
									var t = Qi$1();
									return En(t, Gt(fi$1, {
										get properties() {
											return v.filteredProperties();
										},
										get activeIndex() {
											return b();
										},
										get activeKey() {
											return g();
										},
										onHoverIndex: x,
										onSelect: Ue,
										onStep: Te,
										onCommit: De,
										onColorPickerRegister: Re,
										onEditComplete: W$4,
										onInvalidCommit: Ne,
										onInteract: fe,
										isAdjusting: ne,
										get activeTailwindLabel() {
											return Oe();
										}
									})), lt((e) => {
										var n = D() ? `` : `[font-synthesis:none] contain-layout shrink-0 flex flex-col items-start w-full self-stretch antialiased`, r = D() ? Cr$1 : void 0;
										return n !== e.e && yn(t, e.e = n), e.t = Sn(t, r, e.t), e;
									}, {
										e: void 0,
										t: void 0
									}), t;
								}
							}),
							Gt(en, {
								get when() {
									return dn(() => !!D())() && V$4();
								},
								children: (t) => (() => {
									var n = ea$1();
									return n.$$mousedown = (e) => e.preventDefault(), En(n, Gt(ei$1, {
										get property() {
											return t();
										},
										get activeKey() {
											return g();
										},
										onStep: Te,
										onCommit: De,
										onEditComplete: W$4,
										onInvalidCommit: Ne,
										onInteract: fe,
										onColorPickerRegister: Re,
										showLabel: !1,
										get tailwindLabel() {
											return Oe();
										},
										emphasized: !0
									})), n;
								})()
							})
						];
					}
				})), lt((e) => {
					var n = D() ? `true` : `false`, r = Z$1(`fixed font-sans text-[13px] antialiased [filter:var(--rg-drop-shadow)] select-none`, H.isAnimatedIn() ? `transition-[opacity,transform] duration-220 ease-spring` : `transition-[opacity,transform] duration-120 ease-drawer`), i = `${H.displayPosition().top}px`, a = `${H.displayPosition().left}px`, o = H.isAnimatedIn() ? `auto` : `none`, s = Ve[H.lastAnchorEdge()], l = H.isAnimatedIn() ? `1` : `0`, u = H.isAnimatedIn() ? `scale(1)` : `scale(0.92)`;
					return n !== e.e && gn(t, `data-rg-compact`, e.e = n), r !== e.t && yn(t, e.t = r), i !== e.a && Cn(t, `top`, e.a = i), a !== e.o && Cn(t, `left`, e.o = a), o !== e.i && Cn(t, `pointer-events`, e.i = o), s !== e.n && Cn(t, `transform-origin`, e.n = s), l !== e.s && Cn(t, `opacity`, e.s = l), u !== e.h && Cn(t, `transform`, e.h = u), e;
				}, {
					e: void 0,
					t: void 0,
					a: void 0,
					o: void 0,
					i: void 0,
					n: void 0,
					s: void 0,
					h: void 0
				}), t;
			}
		});
	};
	hn([
		`pointerdown`,
		`mousedown`,
		`click`,
		`contextmenu`,
		`input`,
		`keydown`
	]);
	ra$1 = /*#__PURE__*/ mn(`<div data-react-grab-ignore-events>`);
	ia$1 = (t) => {
		let n, r = () => t.interactive !== !1, i = mr$1(() => n, () => t.position);
		return pt(() => {
			i.measure();
			let e = t.onDismiss ? cr$1({
				isOpen: () => !!t.position,
				onDismiss: t.onDismiss
			}) : void 0;
			W(() => {
				i.clearAnimationHandles(), e?.();
			});
		}), Gt(en, {
			get when() {
				return i.shouldMount();
			},
			get children() {
				var e = ra$1();
				return Tn((e) => n = e, e), wn(e, Yt(() => ({ [t.dataAttribute]: `` }), {
					get class() {
						return Z$1(`fixed font-sans text-[13px] antialiased [filter:var(--rg-drop-shadow)] select-none will-change-[opacity,transform]`, i.isAnimatedIn() ? `transition-[opacity,transform] duration-220 ease-spring` : `transition-[opacity,transform] duration-120 ease-drawer`);
					},
					get style() {
						return {
							top: `${i.displayPosition().top}px`,
							left: `${i.displayPosition().left}px`,
							"z-index": `${ce}`,
							"pointer-events": r() && i.isAnimatedIn() ? `auto` : `none`,
							"transform-origin": Ve[i.lastAnchorEdge()],
							opacity: i.isAnimatedIn() ? `1` : `0`,
							transform: i.isAnimatedIn() ? `scale(1)` : `scale(0.92)`
						};
					},
					onPointerDown: $$1,
					onMouseDown: $$1,
					onClick: $$1,
					onContextMenu: $$1
				}), !1, !0), En(e, () => t.children), e;
			}
		});
	};
	aa$1 = (t) => {
		let n = sr$1({
			clearActiveOnPointerLeave: !0,
			highlight: {
				topCornerRadiusPx: 14,
				bottomCornerRadiusPx: 14,
				cornerShape: Ie
			}
		});
		return Gt(ia$1, {
			get position() {
				return t.position;
			},
			dataAttribute: `data-react-grab-toolbar-menu`,
			get onDismiss() {
				return t.onDismiss;
			},
			get children() {
				return Gt(Q$1.Panel, {
					class: `overflow-hidden`,
					style: { "min-width": `100px` },
					get children() {
						return Gt(Q$1.Provider, {
							store: n,
							get children() {
								return Gt(Q$1.List, {
									label: `Default action`,
									get children() {
										return Gt(Qt, {
											get each() {
												return t.actions;
											},
											children: (n) => {
												let r = () => n.id === t.defaultActionId;
												return Gt(Q$1.Item, {
													get value() {
														return n.id;
													},
													role: `menuitemradio`,
													get checked() {
														return r();
													},
													onSelect: () => {
														t.onSetDefaultAction(n.id), t.onDismiss();
													},
													get children() {
														return [Gt(Q$1.Label, {
															get class() {
																return r() ? `text-[var(--rg-text-primary)]` : `text-[var(--rg-text-secondary)]`;
															},
															get children() {
																return n.label;
															}
														}), Gt(en, {
															get when() {
																return n.shortcut;
															},
															children: (t) => Gt(Q$1.Shortcut, {
																get shortcut() {
																	return t();
																},
																get modifier() {
																	return n.shortcutModifier;
																}
															})
														})];
													}
												});
											}
										});
									}
								});
							}
						});
					}
				});
			}
		});
	};
	oa$1 = /*#__PURE__*/ mn(`<span aria-hidden=true class="shrink-0 font-mono text-[11px] leading-4 text-[var(--rg-text-secondary)] opacity-60 mr-1">`);
	sa$1 = /*#__PURE__*/ mn(`<span class=text-[var(--rg-text-secondary)]>.`);
	ca$1 = /*#__PURE__*/ mn(`<span class="flex items-center min-w-0 w-full"><span class="text-[13px] leading-4 h-fit font-medium overflow-hidden text-ellipsis whitespace-nowrap min-w-0 transition-colors">`);
	la$1 = (t) => {
		let r = () => t.state?.activeIndex ?? 0, i = sr$1({
			value: () => String(r()),
			highlight: {
				topCornerRadiusPx: 14,
				bottomCornerRadiusPx: 14,
				cornerShape: Ie
			}
		});
		return Gt(ia$1, {
			get position() {
				return t.position;
			},
			dataAttribute: `data-react-grab-hierarchy-menu`,
			interactive: !1,
			get children() {
				return Gt(Q$1.Panel, {
					class: `overflow-hidden`,
					style: { "min-width": `160px` },
					get children() {
						return Gt(Q$1.Provider, {
							store: i,
							get children() {
								return Gt(Q$1.List, {
									label: `Navigate element hierarchy`,
									get children() {
										return Gt(Qt, {
											get each() {
												return t.state?.items ?? [];
											},
											children: (t, i) => Gt(Q$1.Item, {
												get value() {
													return String(i());
												},
												role: `menuitemradio`,
												get checked() {
													return i() === r();
												},
												get children() {
													var a = ca$1(), o = a.firstChild;
													return En(a, Gt(en, {
														get when() {
															return t.depth > 0;
														},
														get children() {
															var e = oa$1();
															return En(e, () => t.isLast ? `└─` : `├─`), lt((n) => Cn(e, `padding-left`, `${(t.depth - 1) * 12}px`)), e;
														}
													}), o), En(o, Gt(en, {
														get when() {
															return t.componentName;
														},
														get children() {
															return [dn(() => t.componentName), sa$1()];
														}
													}), null), En(o, () => t.tagName, null), lt((e) => {
														var t = i() === r(), n = i() !== r();
														return t !== e.e && o.classList.toggle(`text-[var(--rg-text-primary)]`, e.e = t), n !== e.t && o.classList.toggle(`text-[var(--rg-text-secondary)]`, e.t = n), e;
													}, {
														e: void 0,
														t: void 0
													}), a;
												}
											})
										});
									}
								});
							}
						});
					}
				});
			}
		});
	};
	ua$1 = (t) => [
		Gt(Xe$1, {
			get selectionVisible() {
				return t.selectionVisible;
			},
			get selectionBounds() {
				return t.selectionBounds;
			},
			get selectionBoundsMultiple() {
				return t.selectionBoundsMultiple;
			},
			get selectionShouldSnap() {
				return t.selectionShouldSnap;
			},
			get dragVisible() {
				return t.dragVisible;
			},
			get dragBounds() {
				return t.dragBounds;
			},
			get grabbedBoxes() {
				return t.grabbedBoxes;
			},
			get labelInstances() {
				return t.labelInstances;
			}
		}),
		Gt(Qe$1, { get visible() {
			return t.isFrozen ?? !1;
		} }),
		Gt(en, {
			get when() {
				return dn(() => !!t.selectionLabelVisible)() && (t.frozenLabelEntryAccessors?.length ?? 0) > 0;
			},
			get children() {
				return Gt(Qt, {
					get each() {
						return t.frozenLabelEntryAccessors ?? [];
					},
					children: (r, i) => Gt(en, {
						get when() {
							return r.read();
						},
						children: (r) => Gt(un$1, {
							get tagName() {
								return r().tagName;
							},
							get componentName() {
								return r().componentName;
							},
							get selectionBounds() {
								return r().bounds;
							},
							get mouseX() {
								return r().mouseX;
							},
							visible: !0,
							get shouldToggleExpandOnClick() {
								return i() === 0;
							},
							get onToggleExpand() {
								return dn(() => i() === 0)() ? t.onToggleExpand : void 0;
							}
						})
					})
				});
			}
		}),
		Gt(en, {
			get when() {
				return dn(() => !!t.selectionLabelVisible)() && t.pendingShiftPreviewEntry;
			},
			children: (t) => Gt(un$1, {
				get tagName() {
					return t().tagName;
				},
				get componentName() {
					return t().componentName;
				},
				get selectionBounds() {
					return t().bounds;
				},
				get mouseX() {
					return t().mouseX;
				},
				visible: !0
			})
		}),
		Gt(en, {
			get when() {
				return dn(() => !!(t.selectionLabelVisible && t.selectionBounds))() && (t.frozenLabelEntryAccessors?.length ?? 0) === 0;
			},
			get children() {
				return Gt(un$1, {
					get tagName() {
						return t.selectionTagName;
					},
					get componentName() {
						return t.selectionComponentName;
					},
					get elementsCount() {
						return t.selectionElementsCount;
					},
					get selectionBounds() {
						return t.selectionBounds;
					},
					get mouseX() {
						return t.mouseX;
					},
					get visible() {
						return t.selectionLabelVisible;
					},
					get isPromptMode() {
						return t.isPromptMode;
					},
					get inputValue() {
						return t.inputValue;
					},
					get status() {
						return t.selectionLabelStatus;
					},
					get filePath() {
						return t.selectionFilePath;
					},
					get onInputChange() {
						return t.onInputChange;
					},
					get onSubmit() {
						return t.onInputSubmit;
					},
					get onToggleExpand() {
						return t.onToggleExpand;
					},
					get selectionLabelShakeCount() {
						return t.selectionLabelShakeCount;
					},
					get onConfirmDismiss() {
						return t.onConfirmDismiss;
					},
					get discardPrompt() {
						return t.discardPrompt;
					},
					get onOpen() {
						return t.onOpenSelectionFile;
					}
				});
			}
		}),
		Gt(Qt, {
			get each() {
				return t.labelInstanceAccessors ?? [];
			},
			children: (n) => Gt(en, {
				get when() {
					return n.read();
				},
				children: (n) => Gt(un$1, {
					get tagName() {
						return n().tagName;
					},
					get componentName() {
						return n().componentName;
					},
					get elementsCount() {
						return n().elementsCount;
					},
					get selectionBounds() {
						return n().bounds;
					},
					get mouseX() {
						return n().mouseX;
					},
					visible: !0,
					get status() {
						return n().status;
					},
					get statusText() {
						return n().statusText;
					},
					get isPromptMode() {
						return n().isPromptMode;
					},
					get inputValue() {
						return n().inputValue;
					},
					get error() {
						return n().errorMessage;
					},
					get hideArrow() {
						return n().hideArrow;
					},
					get onShowContextMenu() {
						let e = n();
						if (!(e.status !== `copied` && e.status !== `fading` || !Z(e.element))) return () => t.onShowContextMenuInstance?.(e.id);
					},
					onRetry: () => t.onRetryInstance?.(n().id),
					onAcknowledgeError: () => t.onAcknowledgeErrorInstance?.(n().id),
					onHoverChange: (e) => t.onLabelInstanceHoverChange?.(n().id, e)
				})
			})
		}),
		Gt(en, {
			get when() {
				return t.toolbarVisible !== !1;
			},
			get children() {
				return Gt(zn$1, {
					get isActive() {
						return t.isActive;
					},
					get isContextMenuOpen() {
						return t.contextMenuPosition !== null;
					},
					get onToggle() {
						return t.onToggleActive;
					},
					get onActivateAction() {
						return t.onActivateAction;
					},
					get activeActionId() {
						return t.activeActionId;
					},
					get enabled() {
						return t.enabled;
					},
					get shakeCount() {
						return t.shakeCount;
					},
					get onStateChange() {
						return t.onToolbarStateChange;
					},
					get onSubscribeToStateChanges() {
						return t.onSubscribeToToolbarStateChanges;
					},
					get onSelectHoverChange() {
						return t.onToolbarSelectHoverChange;
					},
					get onContainerRef() {
						return t.onToolbarRef;
					},
					get onToggleToolbarMenu() {
						return t.onToggleToolbarMenu;
					},
					get buttons() {
						return t.toolbarButtons;
					}
				});
			}
		}),
		Gt(dr$1, {
			get position() {
				return t.contextMenuPosition ?? null;
			},
			get selectionBounds() {
				return t.contextMenuBounds ?? null;
			},
			get tagName() {
				return t.contextMenuTagName;
			},
			get componentName() {
				return t.contextMenuComponentName;
			},
			get hasFilePath() {
				return t.contextMenuHasFilePath ?? !1;
			},
			get actions() {
				return t.actions;
			},
			get actionContext() {
				return t.actionContext;
			},
			get onDismiss() {
				return t.onContextMenuDismiss ?? (() => {});
			},
			get onHide() {
				return t.onContextMenuHide ?? (() => {});
			}
		}),
		Gt(aa$1, {
			get position() {
				return t.toolbarMenuPosition ?? null;
			},
			get actions() {
				return t.toolbarMenuActions ?? [];
			},
			get defaultActionId() {
				return t.defaultActionId ?? `copy`;
			},
			get onSetDefaultAction() {
				return t.onSetDefaultAction ?? (() => {});
			},
			get onDismiss() {
				return t.onToolbarMenuDismiss ?? (() => {});
			}
		}),
		Gt(la$1, {
			get position() {
				return t.hierarchyMenuPosition ?? null;
			},
			get state() {
				return t.hierarchyState;
			}
		}),
		Gt(ta$1, {
			get state() {
				return t.editPanelState ?? null;
			},
			get position() {
				return t.editPanelPosition ?? null;
			},
			get onDismiss() {
				return t.onEditPanelDismiss ?? (() => {});
			},
			get onSubmit() {
				return t.onEditPanelSubmit ?? (() => {});
			},
			get onPendingEditsChange() {
				return t.onEditPanelPendingEditsChange;
			},
			get onInteractingChange() {
				return t.onEditPanelInteractingChange;
			}
		})
	];
}));
// ../ui/dist/react-grab/core-DT-3yIQi.js
/** @license MIT
* Copyright (c) 2025 Aiden Bai
* See LICENSE in the package root.
*/
function Xe(e) {
	let t = String(e), n = t.length - 1;
	return M.context.id + (n ? String.fromCharCode(96 + n) : ``) + t;
}
function Ze(e) {
	M.context = e;
}
function ot(e, t) {
	let n = I, r = P, i = e.length === 0, a = t === void 0 ? r : t, o = i ? rt : {
		owned: null,
		cleanups: null,
		context: a ? a.context : null,
		owner: a
	}, s = i ? e : () => e(st(o));
	P = o, I = null;
	try {
		return K(s, !0);
	} finally {
		I = n, P = r;
	}
}
function st(e) {
	return () => U(() => Ft(e));
}
function z(e, t) {
	t = t ? Object.assign({}, tt, t) : tt;
	let n = {
		value: e,
		observers: null,
		observerSlots: null,
		comparator: t.equals || void 0
	};
	return [Ct.bind(n), (e) => (typeof e == `function` && (e = F && F.running && F.sources.has(n) ? e(n.tValue) : e(n.value)), wt(n, e))];
}
function ct(e, t, n) {
	Tt(Dt(e, t, !0, 1));
}
function lt(e, t, n) {
	Tt(Dt(e, t, !1, 1));
}
function B(e, t, n) {
	nt = jt;
	let r = Dt(e, t, !1, 1), i = xt && G(xt);
	i && (r.suspense = i), n && n.render || (r.user = !0), R ? R.push(r) : Tt(r);
}
function V(e, t, n) {
	n = n ? Object.assign({}, tt, n) : tt;
	let r = Dt(e, t, !0, 0);
	return r.observers = null, r.observerSlots = null, r.comparator = n.equals || void 0, Tt(r), Ct.bind(r);
}
function ut(e) {
	return e && typeof e == `object` && `then` in e;
}
function dt(e, t, n) {
	let r, i, a;
	typeof t == `function` ? (r = e, i = t, a = n || {}) : (r = !0, i = e, a = t || {});
	let o = null, s = it, c = null, l = !1, u = !1, d = `initialValue` in a, f = typeof r == `function` && V(r), p = /* @__PURE__ */ new Set(), [m, h] = (a.storage || z)(a.initialValue), [g, _] = z(void 0), [ee, v] = z(void 0, { equals: !1 }), [y, b] = z(d ? `ready` : `unresolved`);
	P && W(() => {
		for (let e of p.keys()) e.decrement();
		p.clear(), F && o && F.promises.delete(o), o = null;
	}), M.context && (c = M.getNextContextId(), a.ssrLoadFrom === `initial` ? s = a.initialValue : M.load && M.has(c) && (s = M.load(c)));
	function x(e, t, n, r) {
		return o === e && (o = null, r !== void 0 && (d = !0), (e === s || t === s) && a.onHydrated && queueMicrotask(() => a.onHydrated(r, { value: t })), s = it, F && e && l ? (F.promises.delete(e), l = !1, K(() => {
			F.running = !0, S(t, n);
		}, !1)) : S(t, n)), t;
	}
	function S(e, t) {
		K(() => {
			t === void 0 && h(() => e), b(t === void 0 ? d ? `ready` : `unresolved` : `errored`), _(t);
			for (let e of p.keys()) e.decrement();
			p.clear();
		}, !1);
	}
	function te() {
		let e = xt && G(xt), t = m(), n = g();
		if (n !== void 0 && !o) throw n;
		return I && !I.user && e && ct(() => {
			ee(), o && (e.resolved && F && l ? F.promises.add(o) : p.has(e) || (e.increment(), p.add(e)));
		}), t;
	}
	function C(e = !0) {
		if (e !== !1 && u) return;
		u = !1;
		let t = f ? f() : r;
		if (l = F && F.running, t == null || t === !1) {
			x(o, U(m));
			return;
		}
		F && o && F.promises.delete(o);
		let n, a = s === it ? U(() => {
			try {
				return i(t, {
					value: m(),
					refetching: e
				});
			} catch (e) {
				n = e;
			}
		}) : s;
		if (n !== void 0) x(o, void 0, Lt(n), t);
		else return ut(a) ? (o = a, `v` in a ? (a.s === 1 ? x(o, a.v, void 0, t) : x(o, void 0, Lt(a.v), t), a) : (u = !0, queueMicrotask(() => u = !1), K(() => {
			b(d ? `refreshing` : `pending`), v();
		}, !1), a.then((e) => x(a, e, void 0, t), (e) => x(a, void 0, Lt(e), t)))) : (x(o, a, void 0, t), a);
	}
	Object.defineProperties(te, {
		state: { get: () => y() },
		error: { get: () => g() },
		loading: { get() {
			let e = y();
			return e === `pending` || e === `refreshing`;
		} },
		latest: { get() {
			if (!d) return te();
			let e = g();
			if (e && !o) throw e;
			return m();
		} }
	});
	let ne = P;
	return f ? ct(() => (ne = P, C(!1))) : C(!1), [te, {
		refetch: (e) => gt(ne, () => C(e)),
		mutate: h
	}];
}
function H(e) {
	return K(e, !1);
}
function U(e) {
	if (I === null) return e();
	let t = I;
	I = null;
	try {
		return e();
	} finally {
		I = t;
	}
}
function ft(e, t, n) {
	let r = Array.isArray(e), i, a = n && n.defer;
	return (n) => {
		let o;
		if (r) {
			o = Array(e.length);
			for (let t = 0; t < e.length; t++) o[t] = e[t]();
		} else o = e();
		if (a) return a = !1, n;
		let s = U(() => t(o, i, n));
		return i = o, s;
	};
}
function pt(e) {
	B(() => U(e));
}
function W(e) {
	return P === null || (P.cleanups === null ? P.cleanups = [e] : P.cleanups.push(e)), e;
}
function mt() {
	return I;
}
function ht() {
	return P;
}
function gt(e, t) {
	let n = P, r = I;
	P = e, I = null;
	try {
		return K(t, !0);
	} catch (e) {
		Rt(e);
	} finally {
		P = n, I = r;
	}
}
function yt(e, t) {
	let n = Symbol(`context`);
	return {
		id: n,
		Provider: Bt(n),
		defaultValue: e
	};
}
function G(e) {
	let t;
	return P && P.context && (t = P.context[e.id]) !== void 0 ? t : e.defaultValue;
}
function bt(e) {
	let t = V(e), n = V(() => zt(t()));
	return n.toArray = () => {
		let e = n();
		return Array.isArray(e) ? e : e == null ? [] : [e];
	}, n;
}
function St(e) {
	e.observerSlots = e.observers.map((t, n) => {
		let r = t.sources;
		if (!r) return -1;
		for (let i = r.length - 1; i >= 0; i--) if (r[i] === e && t.sourceSlots && t.sourceSlots[i] === n) return i;
		return r.lastIndexOf(e);
	});
}
function Ct() {
	let e = F && F.running;
	if (this.sources && (e ? this.tState : this.state)) {
		if ((e ? this.tState : this.state) === 1) Tt(this);
		else {
			let e = L;
			L = null, K(() => Mt(this), !1), L = e;
		}
	}
	if (I) {
		let e = this.observers;
		if (!e || e[e.length - 1] !== I) {
			let t = e ? e.length : 0;
			if (I.sources) I.sources.push(this), I.sourceSlots.push(t);
			else {
				let e = [this], n = [t];
				I.sources = e, I.sourceSlots = n;
			}
			if (e) this.observerSlots && this.observerSlots.length === e.length || St(this), e.push(I), this.observerSlots.push(I.sources.length - 1);
			else {
				let e = [I], t = [I.sources.length - 1];
				this.observers = e, this.observerSlots = t;
			}
		}
	}
	return e && F.sources.has(this) ? this.tValue : this.value;
}
function wt(e, t, n) {
	let r = F && F.running && F.sources.has(e) ? e.tValue : e.value;
	if (!e.comparator || !e.comparator(r, t)) {
		if (F) {
			let r = F.running;
			(r || !n && F.sources.has(e)) && (F.sources.add(e), e.tValue = t), r || (e.value = t);
		} else e.value = t;
		e.observers && e.observers.length && K(() => {
			for (let t = 0; t < e.observers.length; t += 1) {
				let n = e.observers[t], r = F && F.running;
				r && F.disposed.has(n) || ((r ? !n.tState : !n.state) && (n.pure ? L.push(n) : R.push(n), n.observers && Nt(n)), r ? n.tState = 1 : n.state = 1);
			}
			if (L.length > 1e6) throw L = [], Error();
		}, !1);
	}
	return t;
}
function Tt(e) {
	if (!e.fn) return;
	Ft(e, !0);
	let t = at;
	Et(e, F && F.running && F.sources.has(e) ? e.tValue : e.value, t), F && !F.running && F.sources.has(e) && queueMicrotask(() => {
		K(() => {
			F && (F.running = !0), I = P = e, Et(e, e.tValue, t), I = P = null;
		}, !1);
	});
}
function Et(e, t, n) {
	let r, i = P, a = I;
	I = P = e;
	try {
		r = e.fn(t);
	} catch (t) {
		return e.pure && (F && F.running ? (e.tState = 1, e.tOwned && e.tOwned.forEach((e) => Ft(e)), e.tOwned = void 0) : (e.state = 1, e.owned && e.owned.forEach((e) => Ft(e)), e.owned = null)), e.updatedAt = n + 1, Rt(t);
	} finally {
		I = a, P = i;
	}
	(!e.updatedAt || e.updatedAt <= n) && (e.updatedAt != null && `observers` in e ? wt(e, r, !0) : F && F.running && e.pure ? (F.sources.has(e) || (e.value = r), F.sources.add(e), e.tValue = r) : e.value = r, e.updatedAt = n);
}
function Dt(e, t, n, r = 1, i) {
	let a = {
		fn: e,
		state: r,
		updatedAt: null,
		owned: null,
		sources: null,
		sourceSlots: null,
		cleanups: null,
		value: t,
		owner: P,
		context: P ? P.context : null,
		pure: n
	};
	return F && F.running && (a.state = 0, a.tState = r), P === null || P !== rt && (F && F.running && P.pure ? P.tOwned ? P.tOwned.push(a) : P.tOwned = [a] : P.owned ? P.owned.push(a) : P.owned = [a]), a;
}
function Ot(e) {
	let t = F && F.running;
	if ((t ? e.tState : e.state) === 0) return;
	if ((t ? e.tState : e.state) === 2) return Mt(e);
	if (e.suspense && U(e.suspense.inFallback)) return e.suspense.effects.push(e);
	let n = [e];
	for (; (e = e.owner) && (!e.updatedAt || e.updatedAt < at);) {
		if (t && F.disposed.has(e)) return;
		(t ? e.tState : e.state) && n.push(e);
	}
	for (let r = n.length - 1; r >= 0; r--) {
		if (e = n[r], t) {
			let t = e, i = n[r + 1];
			for (; (t = t.owner) && t !== i;) if (F.disposed.has(t)) return;
		}
		if ((t ? e.tState : e.state) === 1) Tt(e);
		else if ((t ? e.tState : e.state) === 2) {
			let t = L;
			L = null, K(() => Mt(e, n[0]), !1), L = t;
		}
	}
}
function K(e, t) {
	if (L) return e();
	let n = !1;
	t || (L = []), R ? n = !0 : R = [], at++;
	try {
		let t = e();
		return kt(n), t;
	} catch (e) {
		n || (R = null), L = null, Rt(e);
	}
}
function kt(e) {
	if (L &&= (At(L), null), e) return;
	let t;
	if (F) {
		if (!F.promises.size && !F.queue.size) {
			let e = F.sources, n = F.disposed;
			R.push.apply(R, F.effects), t = F.resolve;
			for (let e of R) `tState` in e && (e.state = e.tState), delete e.tState;
			F = null, K(() => {
				for (let e of n) Ft(e);
				for (let t of e) {
					if (t.value = t.tValue, t.owned) for (let e = 0, n = t.owned.length; e < n; e++) Ft(t.owned[e]);
					t.tOwned && (t.owned = t.tOwned), delete t.tValue, delete t.tOwned, t.tState = 0;
				}
				vt(!1);
			}, !1);
		} else if (F.running) {
			F.running = !1, F.effects.push.apply(F.effects, R), R = null, vt(!0);
			return;
		}
	}
	let n = R;
	R = null, n.length && K(() => nt(n), !1), t && t();
}
function At(e) {
	for (let t = 0; t < e.length; t++) Ot(e[t]);
}
function jt(e) {
	let t, n = 0;
	for (t = 0; t < e.length; t++) {
		let r = e[t];
		r.user ? e[n++] = r : Ot(r);
	}
	if (M.context) {
		if (M.count) {
			M.effects ||= [], M.effects.push(...e.slice(0, n));
			return;
		}
		Ze();
	}
	for (M.effects && (M.done || !M.count) && (e = [...M.effects, ...e], n += M.effects.length, delete M.effects), t = 0; t < n; t++) Ot(e[t]);
}
function Mt(e, t) {
	let n = F && F.running;
	n ? e.tState = 0 : e.state = 0;
	for (let r = 0; r < e.sources.length; r += 1) {
		let i = e.sources[r];
		if (i.sources) {
			let e = n ? i.tState : i.state;
			e === 1 ? i !== t && (!i.updatedAt || i.updatedAt < at) && Ot(i) : e === 2 && Mt(i, t);
		}
	}
}
function Nt(e) {
	let t = F && F.running;
	for (let n = 0; n < e.observers.length; n += 1) {
		let r = e.observers[n];
		(t ? !r.tState : !r.state) && (t ? r.tState = 2 : r.state = 2, r.pure ? L.push(r) : R.push(r), r.observers && Nt(r));
	}
}
function Pt(e) {
	F && F.sources.has(e) && (F.sources.delete(e), delete e.tValue);
}
function Ft(e, t) {
	let n;
	if (e.sources) for (; e.sources.length;) {
		let n = e.sources.pop(), r = e.sourceSlots.pop(), i = n.observers;
		if (i && i.length) {
			n.observerSlots && n.observerSlots.length === i.length || St(n);
			let e = i.pop(), t = n.observerSlots.pop();
			r < i.length && (e.sourceSlots[t] = r, i[r] = e, n.observerSlots[r] = t);
		}
		t || i && i.length || !F || Pt(n);
	}
	if (e.tOwned) {
		for (n = e.tOwned.length - 1; n >= 0; n--) Ft(e.tOwned[n]);
		delete e.tOwned;
	}
	if (F && F.running && e.pure) It(e, !0);
	else if (e.owned) {
		for (n = e.owned.length - 1; n >= 0; n--) Ft(e.owned[n]);
		e.owned = null;
	}
	if (t || Pt(e), e.cleanups) {
		for (n = e.cleanups.length - 1; n >= 0; n--) e.cleanups[n]();
		e.cleanups = null;
	}
	F && F.running ? e.tState = 0 : e.state = 0;
}
function It(e, t) {
	if (t || (e.tState = 0, F.disposed.add(e), Pt(e)), e.owned) for (let t = 0; t < e.owned.length; t++) It(e.owned[t]);
}
function Lt(e) {
	return e instanceof Error ? e : Error(typeof e == `string` ? e : `Unknown error`, { cause: e });
}
function Rt(e, t = P) {
	throw Lt(e);
}
function zt(e) {
	if (typeof e == `function` && !e.length) return zt(e());
	if (Array.isArray(e)) {
		let t = [];
		for (let n = 0; n < e.length; n++) {
			let r = zt(e[n]);
			if (Array.isArray(r)) {
				if (r.length < 32768) t.push.apply(t, r);
				else for (let e = 0; e < r.length; e++) t.push(r[e]);
			} else t.push(r);
		}
		return t;
	}
	return e;
}
function Bt(e, t) {
	return function(t) {
		let n;
		return lt(() => n = U(() => (P.context = {
			...P.context,
			[e]: t.value
		}, bt(() => t.children))), void 0), n;
	};
}
function Ht(e) {
	for (let t = 0; t < e.length; t++) e[t]();
}
function Ut(e, t, n = {}) {
	let r = [], i = [], a = [], o = 0, s = t.length > 1 ? [] : null;
	return W(() => Ht(a)), () => {
		let c = e() || [], l = c.length, u, d;
		return c[et], U(() => {
			let e, t, p, m, h, g, _, ee, v;
			if (l === 0) o !== 0 && (Ht(a), a = [], r = [], i = [], o = 0, s &&= []), n.fallback && (r = [Vt], i[0] = ot((e) => (a[0] = e, n.fallback())), o = 1);
			else if (o === 0) {
				for (i = Array(l), d = 0; d < l; d++) r[d] = c[d], i[d] = ot(f);
				o = l;
			} else {
				for (p = Array(l), m = Array(l), s && (h = Array(l)), g = 0, _ = Math.min(o, l); g < _ && r[g] === c[g]; g++);
				for (_ = o - 1, ee = l - 1; _ >= g && ee >= g && r[_] === c[ee]; _--, ee--) p[ee] = i[_], m[ee] = a[_], s && (h[ee] = s[_]);
				for (e = /* @__PURE__ */ new Map(), t = Array(ee + 1), d = ee; d >= g; d--) v = c[d], u = e.get(v), t[d] = u === void 0 ? -1 : u, e.set(v, d);
				for (u = g; u <= _; u++) v = r[u], d = e.get(v), d !== void 0 && d !== -1 ? (p[d] = i[u], m[d] = a[u], s && (h[d] = s[u]), d = t[d], e.set(v, d)) : a[u]();
				for (d = g; d < l; d++) d in p ? (i[d] = p[d], a[d] = m[d], s && (s[d] = h[d], s[d](d))) : i[d] = ot(f);
				i = i.slice(0, o = l), r = c.slice(0);
			}
			return i;
		});
		function f(e) {
			if (a[d] = e, s) {
				let [e, n] = z(d);
				return s[d] = n, t(c[d], e);
			}
			return t(c[d]);
		}
	};
}
function Wt(e, t, n = {}) {
	let r = [], i = [], a = [], o = [], s = 0, c;
	return W(() => Ht(a)), () => {
		let l = e() || [], u = l.length;
		return l[et], U(() => {
			if (u === 0) return s !== 0 && (Ht(a), a = [], r = [], i = [], s = 0, o = []), n.fallback && (r = [Vt], i[0] = ot((e) => (a[0] = e, n.fallback())), s = 1), i;
			for (r[0] === Vt && (a[0](), a = [], r = [], i = [], s = 0), c = 0; c < u; c++) c < r.length && r[c] !== l[c] ? o[c](() => l[c]) : c >= r.length && (i[c] = ot(d));
			for (; c < r.length; c++) a[c]();
			return s = o.length = a.length = u, r = l.slice(0), i = i.slice(0, s);
		});
		function d(e) {
			a[c] = e;
			let [n, r] = z(l[c]);
			return o[c] = r, t(n, c);
		}
	};
}
function Gt(e, t) {
	return U(() => e(t || {}));
}
function Kt() {
	return !0;
}
function Jt(e) {
	return (e = typeof e == `function` ? e() : e) ? e : {};
}
function q() {
	for (let e = 0, t = this.length; e < t; ++e) {
		let t = this[e]();
		if (t !== void 0) return t;
	}
}
function Yt(...e) {
	let t = !1;
	for (let n = 0; n < e.length; n++) {
		let r = e[n];
		t ||= !!r && $e in r, e[n] = typeof r == `function` ? (t = !0, V(r)) : r;
	}
	if (N && t) return new Proxy({
		get(t) {
			for (let n = e.length - 1; n >= 0; n--) {
				let r = Jt(e[n])[t];
				if (r !== void 0) return r;
			}
		},
		has(t) {
			for (let n = e.length - 1; n >= 0; n--) if (t in Jt(e[n])) return !0;
			return !1;
		},
		keys() {
			let t = [];
			for (let n = 0; n < e.length; n++) t.push(...Object.keys(Jt(e[n])));
			return [...new Set(t)];
		}
	}, qt);
	let n = {}, r = Object.create(null);
	for (let t = e.length - 1; t >= 0; t--) {
		let i = e[t];
		if (!i) continue;
		let a = Object.getOwnPropertyNames(i);
		for (let e = a.length - 1; e >= 0; e--) {
			let t = a[e];
			if (t === `__proto__` || t === `constructor`) continue;
			let o = Object.getOwnPropertyDescriptor(i, t);
			if (!r[t]) r[t] = o.get ? {
				enumerable: !0,
				configurable: !0,
				get: q.bind(n[t] = [o.get.bind(i)])
			} : o.value === void 0 ? void 0 : o;
			else {
				let e = n[t];
				e && (o.get ? e.push(o.get.bind(i)) : o.value !== void 0 && e.push(() => o.value));
			}
		}
	}
	let i = {}, a = Object.keys(r);
	for (let e = a.length - 1; e >= 0; e--) {
		let t = a[e], n = r[t];
		n && n.get ? Object.defineProperty(i, t, n) : i[t] = n ? n.value : void 0;
	}
	return i;
}
function Xt(e, ...t) {
	let n = t.length;
	if (N && $e in e) {
		let r = n > 1 ? t.flat() : t[0], i = /* @__PURE__ */ new Set(), a = t.map((t) => {
			let n = t.filter((e) => !i.has(e) && (i.add(e), !0));
			return new Proxy({
				get(t) {
					return n.includes(t) ? e[t] : void 0;
				},
				has(t) {
					return n.includes(t) && t in e;
				},
				keys() {
					return n.filter((t) => t in e);
				}
			}, qt);
		});
		return a.push(new Proxy({
			get(t) {
				return r.includes(t) ? void 0 : e[t];
			},
			has(t) {
				return !r.includes(t) && t in e;
			},
			keys() {
				return Object.keys(e).filter((e) => !r.includes(e));
			}
		}, qt)), a;
	}
	let r = [];
	for (let e = 0; e <= n; e++) r[e] = {};
	for (let i of Object.getOwnPropertyNames(e)) {
		let a = n;
		for (let e = 0; e < t.length; e++) if (t[e].includes(i)) {
			a = e;
			break;
		}
		let o = Object.getOwnPropertyDescriptor(e, i);
		!o.get && !o.set && o.enumerable && o.writable && o.configurable ? r[a][i] = o.value : Object.defineProperty(r[a], i, o);
	}
	return r;
}
function Qt(e) {
	let t = `fallback` in e && { fallback: () => e.fallback };
	return V(Ut(() => e.each, e.children, t || void 0));
}
function $t(e) {
	let t = `fallback` in e && { fallback: () => e.fallback };
	return V(Wt(() => e.each, e.children, t || void 0));
}
function en(e) {
	let t = e.keyed, n = V(() => e.when, void 0, void 0), r = t ? n : V(n, void 0, { equals: (e, t) => !e == !t });
	return V(() => {
		let i = r();
		if (i) {
			let a = e.children;
			return typeof a == `function` && a.length > 0 ? U(() => a(t ? i : () => {
				if (!U(r)) throw Zt(`Show`);
				return n();
			})) : a;
		}
		return e.fallback;
	}, void 0, void 0);
}
function tn(e) {
	let t = bt(() => e.children), n = V(() => {
		let e = t(), n = Array.isArray(e) ? e : [e], r = () => void 0;
		for (let e = 0; e < n.length; e++) {
			let t = e, i = n[e], a = r, o = V(() => a() ? void 0 : i.when, void 0, void 0), s = i.keyed ? o : V(o, void 0, { equals: (e, t) => !e == !t });
			r = () => a() || (s() ? [
				t,
				o,
				i
			] : void 0);
		}
		return r;
	});
	return V(() => {
		let t = n()();
		if (!t) return e.fallback;
		let [r, i, a] = t, o = a.children;
		return typeof o == `function` && o.length > 0 ? U(() => o(a.keyed ? i() : () => {
			if (U(n)()?.[0] !== r) throw Zt(`Match`);
			return i();
		})) : o;
	}, void 0, void 0);
}
function nn(e) {
	return e;
}
function cn(e, t) {
	let n = sn[e];
	return typeof n == `object` ? n[t] ? n.$ : void 0 : n;
}
function J(e, t, n) {
	let r = n.length, i = t.length, a = r, o = 0, s = 0, c = t[i - 1].nextSibling, l = null;
	for (; o < i || s < a;) if (t[o] === n[s]) o++, s++;
	else {
		for (; t[i - 1] === n[a - 1];) i--, a--;
		if (i === o) {
			let t = a < r ? s ? n[s - 1].nextSibling : n[a - s] : c;
			for (; s < a;) e.insertBefore(n[s++], t);
		} else if (a === s) for (; o < i;) l && l.has(t[o]) || t[o].remove(), o++;
		else if (t[o] === n[a - 1] && n[s] === t[i - 1]) {
			let r = t[--i].nextSibling;
			e.insertBefore(n[s++], t[o++].nextSibling), e.insertBefore(n[--a], r), t[i] = n[a];
		} else {
			if (!l) {
				l = /* @__PURE__ */ new Map();
				let e = s;
				for (; e < a;) l.set(n[e], e++);
			}
			let r = l.get(t[o]);
			if (r != null) {
				if (s < r && r < a) {
					let c = o, u = 1, d;
					for (; ++c < i && c < a && (d = l.get(t[c])) != null && d === r + u;) u++;
					if (u > r - s) {
						let i = t[o];
						for (; s < r;) e.insertBefore(n[s++], i);
					} else e.replaceChild(n[s++], t[o++]);
				} else o++;
			} else t[o++].remove();
		}
	}
}
function pn(e, t, n, r = {}) {
	let i;
	return ot((r) => {
		i = r, t === document ? e() : En(t, e(), !n && t.firstChild ? null : void 0, n);
	}, r.owner), () => {
		i(), t.textContent = ``;
	};
}
function mn(e, t, n, r) {
	let i, a = () => {
		let t = r ? document.createElementNS(`http://www.w3.org/1998/Math/MathML`, `template`) : document.createElement(`template`);
		return t.innerHTML = e, n ? t.content.firstChild.firstChild : r ? t.firstChild : t.content.firstChild;
	}, o = t ? () => U(() => document.importNode(i ||= a(), !0)) : () => (i ||= a()).cloneNode(!0);
	return o.cloneNode = o, o;
}
function hn(e, t = window.document) {
	let n = t[fn] || (t[fn] = /* @__PURE__ */ new Set());
	for (let r = 0, i = e.length; r < i; r++) {
		let i = e[r];
		n.has(i) || (n.add(i), t.addEventListener(i, Mn));
	}
}
function gn(e, t, n) {
	On(e) || (n == null ? e.removeAttribute(t) : e.setAttribute(t, n));
}
function _n(e, t, n, r) {
	On(e) || (r == null ? e.removeAttributeNS(t, n) : e.setAttributeNS(t, n, r));
}
function vn(e, t, n) {
	On(e) || (n ? e.setAttribute(t, ``) : e.removeAttribute(t));
}
function yn(e, t) {
	On(e) || (t == null ? e.removeAttribute(`class`) : e.className = t);
}
function bn(e, t, n, r) {
	if (r) Array.isArray(n) ? (e[`$$${t}`] = n[0], e[`$$${t}Data`] = n[1]) : e[`$$${t}`] = n;
	else if (Array.isArray(n)) {
		let r = n[0];
		e.addEventListener(t, n[0] = (t) => r.call(e, n[1], t));
	} else e.addEventListener(t, n, typeof n != `function` && n);
}
function xn(e, t, n = {}) {
	let r = Object.keys(t || {}), i = Object.keys(n), a, o;
	for (a = 0, o = i.length; a < o; a++) {
		let r = i[a];
		r && r !== `undefined` && !t[r] && (An(e, r, !1), delete n[r]);
	}
	for (a = 0, o = r.length; a < o; a++) {
		let i = r[a], o = !!t[i];
		i && i !== `undefined` && n[i] !== o && o && (An(e, i, !0), n[i] = o);
	}
	return n;
}
function Sn(e, t, n) {
	if (!t) return n ? gn(e, `style`) : t;
	let r = e.style;
	if (typeof t == `string`) return r.cssText = t;
	typeof n == `string` && (r.cssText = n = void 0), n ||= {}, t ||= {};
	let i, a;
	for (a in n) t[a] ?? r.removeProperty(a), delete n[a];
	for (a in t) i = t[a], i !== n[a] && (r.setProperty(a, i), n[a] = i);
	return n;
}
function Cn(e, t, n) {
	n == null ? e.style.removeProperty(t) : e.style.setProperty(t, n);
}
function wn(e, t = {}, n, r) {
	let i = {};
	return r || lt(() => i.children = Nn(e, t.children, i.children)), lt(() => typeof t.ref == `function` && Tn(t.ref, e)), lt(() => Dn(e, t, n, !0, i, !0)), i;
}
function Tn(e, t, n) {
	return U(() => e(t, n));
}
function En(e, t, n, r) {
	if (n !== void 0 && !r && (r = []), typeof t != `function`) return Nn(e, t, r, n);
	lt((r) => Nn(e, t(), r, n), r);
}
function Dn(e, t, n, r, i = {}, a = !1) {
	t ||= {};
	for (let r in i) if (!(r in t)) {
		if (r === `children`) continue;
		i[r] = jn(e, r, null, i[r], n, a, t);
	}
	for (let o in t) {
		if (o === `children`) {
			r || Nn(e, t.children);
			continue;
		}
		let s = t[o];
		i[o] = jn(e, o, s, i[o], n, a, t);
	}
}
function On(e) {
	return !!M.context && !M.done && (!e || e.isConnected);
}
function kn(e) {
	return e.toLowerCase().replace(/-([a-z])/g, (e, t) => t.toUpperCase());
}
function An(e, t, n) {
	let r = t.trim().split(/\s+/);
	for (let t = 0, i = r.length; t < i; t++) e.classList.toggle(r[t], n);
}
function jn(e, t, n, r, i, a, o) {
	let s, c, l, u, d;
	if (t === `style`) return Sn(e, n, r);
	if (t === `classList`) return xn(e, n, r);
	if (n === r) return r;
	if (t === `ref`) a || n(e);
	else if (t.slice(0, 3) === `on:`) {
		let i = t.slice(3);
		r && e.removeEventListener(i, r, typeof r != `function` && r), n && e.addEventListener(i, n, typeof n != `function` && n);
	} else if (t.slice(0, 10) === `oncapture:`) {
		let i = t.slice(10);
		r && e.removeEventListener(i, r, !0), n && e.addEventListener(i, n, !0);
	} else if (t.slice(0, 2) === `on`) {
		let i = t.slice(2).toLowerCase(), a = ln.has(i);
		if (!a && r) {
			let t = Array.isArray(r) ? r[0] : r;
			e.removeEventListener(i, t);
		}
		(a || n) && (bn(e, i, n, a), a && hn([i]));
	} else if (t.slice(0, 5) === `attr:`) gn(e, t.slice(5), n);
	else if (t.slice(0, 5) === `bool:`) vn(e, t.slice(5), n);
	else if ((d = t.slice(0, 5) === `prop:`) || (l = an.has(t)) || !i && ((u = cn(t, e.tagName)) || (c = rn.has(t))) || (s = e.nodeName.includes(`-`) || `is` in o)) {
		if (d) t = t.slice(5), c = !0;
		else if (On(e)) return n;
		t === `class` || t === `className` ? yn(e, n) : s && !c && !l ? e[kn(t)] = n : (t === `value` || t === `defaultValue`) && (e.nodeName === `INPUT` || e.nodeName === `TEXTAREA`) ? e[u || t] = n ?? `` : e[u || t] = n;
	} else {
		let r = i && t.indexOf(`:`) > -1 && un[t.split(`:`)[0]];
		r ? _n(e, r, t, n) : gn(e, on[t] || t, n);
	}
	return n;
}
function Mn(e) {
	if (M.registry && M.events && M.events.find(([t, n]) => n === e)) return;
	let t = e.target, n = `$$${e.type}`, r = e.target, i = e.currentTarget, a = (t) => Object.defineProperty(e, "target", {
		configurable: !0,
		value: t
	}), o = () => {
		let r = t[n];
		if (r && !t.disabled) {
			let i = t[`${n}Data`];
			if (i === void 0 ? r.call(t, e) : r.call(t, i, e), e.cancelBubble) return;
		}
		return t.host && typeof t.host != `string` && !t.host._$host && t.contains(e.target) && a(t.host), !0;
	}, s = () => {
		for (; o() && (t = t._$host || t.parentNode || t.host););
	};
	Object.defineProperty(e, "currentTarget", {
		configurable: !0,
		get() {
			return t || document;
		}
	}), M.registry && !M.done && (M.done = _$HY.done = !0);
	try {
		if (e.composedPath) {
			let n = e.composedPath();
			a(n[0]);
			for (let e = 0; e < n.length - 2 && (t = n[e], o()); e++) {
				if (t._$host) {
					t = t._$host, s();
					break;
				}
				if (t.parentNode === i) break;
			}
		} else s();
		a(r);
	} finally {
		e = t = null;
	}
}
function Nn(e, t, n, r, i) {
	let a = On(e);
	if (a) {
		!n && (n = [...e.childNodes]);
		let t = [];
		for (let e = 0; e < n.length; e++) {
			let r = n[e];
			r.nodeType === 8 && r.data.slice(0, 2) === `!$` ? r.remove() : t.push(r);
		}
		n = t;
	}
	for (; typeof n == `function`;) n = n();
	if (t === n) return n;
	let o = typeof t, s = r !== void 0;
	if (e = s && n[0] && n[0].parentNode || e, o === `string` || o === `number`) {
		if (a || o === `number` && (t = t.toString(), t === n)) return n;
		if (s) {
			let i = n[0];
			i && i.nodeType === 3 ? i.data !== t && (i.data = t) : i = document.createTextNode(t), n = In(e, n, r, i);
		} else n = n !== `` && typeof n == `string` ? e.firstChild.data = t : e.textContent = t;
	} else if (t == null || o === `boolean`) {
		if (a) return n;
		n = In(e, n, r);
	} else if (o === `function`) return lt(() => {
		let i = t();
		for (; typeof i == `function`;) i = i();
		n = Nn(e, i, n, r);
	}), () => n;
	else if (Array.isArray(t)) {
		let o = [], c = n && Array.isArray(n);
		if (Pn(o, t, n, i)) return lt(() => n = Nn(e, o, n, r, !0)), () => n;
		if (a) {
			if (!o.length) return n;
			if (r === void 0) return n = [...e.childNodes];
			let t = o[0];
			if (t.parentNode !== e) return n;
			let i = [t];
			for (; (t = t.nextSibling) !== r;) i.push(t);
			return n = i;
		}
		if (o.length === 0) {
			if (n = In(e, n, r), s) return n;
		} else c ? n.length === 0 ? Fn(e, o, r) : J(e, n, o) : (n && In(e), Fn(e, o));
		n = o;
	} else if (t.nodeType) {
		if (a && t.parentNode) return n = s ? [t] : t;
		if (Array.isArray(n)) {
			if (s) return n = In(e, n, r, t);
			In(e, n, null, t);
		} else n == null || n === `` || !e.firstChild ? e.appendChild(t) : e.replaceChild(t, e.firstChild);
		n = t;
	}
	return n;
}
function Pn(e, t, n, r) {
	let i = !1;
	for (let a = 0, o = t.length; a < o; a++) {
		let o = t[a], s = n && n[e.length], c;
		if (o != null && o !== !0 && o !== !1) {
			if ((c = typeof o) == `object` && o.nodeType) e.push(o);
			else if (Array.isArray(o)) i = Pn(e, o, s) || i;
			else if (c === `function`) {
				if (r) {
					for (; typeof o == `function`;) o = o();
					i = Pn(e, Array.isArray(o) ? o : [o], Array.isArray(s) ? s : [s]) || i;
				} else e.push(o), i = !0;
			} else {
				let t = String(o);
				s && s.nodeType === 3 && s.data === t ? e.push(s) : e.push(document.createTextNode(t));
			}
		}
	}
	return i;
}
function Fn(e, t, n = null) {
	for (let r = 0, i = t.length; r < i; r++) e.insertBefore(t[r], n);
}
function In(e, t, n, r) {
	if (n === void 0) return e.textContent = ``;
	let i = r || document.createTextNode(``);
	if (t.length) {
		let r = !1;
		for (let a = t.length - 1; a >= 0; a--) {
			let o = t[a];
			if (i !== o) {
				let t = o.parentNode === e;
				!r && !a ? t ? e.replaceChild(i, o) : e.insertBefore(i, n) : t && o.remove();
			} else r = !0;
		}
	} else e.insertBefore(i, n);
	return [i];
}
function Hn(e) {
	let t = e[$e];
	if (!t && (Object.defineProperty(e, $e, { value: t = new Proxy(e, Jn) }), !Array.isArray(e))) {
		let n = Object.keys(e), r = Object.getOwnPropertyDescriptors(e), i = Object.getPrototypeOf(e), a = i !== null && typeof e == `object` && !!e && !Array.isArray(e) && i !== Object.prototype;
		if (a) {
			let e = Object.getOwnPropertyDescriptors(i);
			n.push(...Object.keys(e)), Object.assign(r, e);
		}
		for (let i = 0, o = n.length; i < o; i++) {
			let o = n[i];
			a && o === `constructor` || r[o].get && Object.defineProperty(e, o, {
				configurable: !0,
				enumerable: r[o].enumerable,
				get: r[o].get.bind(t)
			});
		}
	}
	return t;
}
function Un(e) {
	let t;
	return typeof e == `object` && !!e && (e[$e] || !(t = Object.getPrototypeOf(e)) || t === Object.prototype || Array.isArray(e));
}
function Y(e, t = /* @__PURE__ */ new Set()) {
	let n, r, i, a;
	if (n = e != null && e[Rn]) return n;
	if (!Un(e) || t.has(e)) return e;
	if (Array.isArray(e)) {
		Object.isFrozen(e) ? e = e.slice(0) : t.add(e);
		for (let n = 0, a = e.length; n < a; n++) i = e[n], (r = Y(i, t)) !== i && (e[n] = r);
	} else {
		Object.isFrozen(e) ? e = Object.assign({}, e) : t.add(e);
		let n = Object.keys(e), o = Object.getOwnPropertyDescriptors(e);
		for (let s = 0, c = n.length; s < c; s++) a = n[s], o[a].get || (i = e[a], (r = Y(i, t)) !== i && (e[a] = r));
	}
	return e;
}
function Wn(e, t) {
	let n = e[t];
	return n || Object.defineProperty(e, t, { value: n = Object.create(null) }), n;
}
function Gn(e, t, n) {
	if (e[t]) return e[t];
	let [r, i] = z(n, {
		equals: !1,
		internal: !0
	});
	return r.$ = i, e[t] = r;
}
function X(e, t) {
	let n = Reflect.getOwnPropertyDescriptor(e, t);
	return !n || n.get || !n.configurable || t === $e || t === zn || (delete n.value, delete n.writable, n.get = () => e[$e][t]), n;
}
function Kn(e) {
	mt() && Gn(Wn(e, zn), Vn)();
}
function qn(e) {
	return Kn(e), Reflect.ownKeys(e);
}
function Yn(e, t, n, r = !1) {
	if (t === `__proto__` || !r && e[t] === n) return;
	let i = e[t], a = e.length;
	n === void 0 ? (delete e[t], e[Bn] && e[Bn][t] && i !== void 0 && e[Bn][t].$()) : (e[t] = n, e[Bn] && e[Bn][t] && i === void 0 && e[Bn][t].$());
	let o = Wn(e, zn), s;
	if ((s = Gn(o, t, i)) && s.$(() => n), Array.isArray(e) && e.length !== a) {
		for (let t = e.length; t < a; t++) (s = o[t]) && s.$();
		(s = Gn(o, `length`, a)) && s.$(e.length);
	}
	(s = o[Vn]) && s.$();
}
function Xn(e, t) {
	let n = Object.keys(t);
	for (let r = 0; r < n.length; r += 1) {
		let i = n[r];
		Zn(i) || Yn(e, i, t[i]);
	}
}
function Zn(e) {
	return e === `__proto__` || e === `constructor` || e === `prototype`;
}
function Qn(e, t) {
	if (typeof t == `function` && (t = t(e)), t = Y(t), Array.isArray(t)) {
		if (e === t) return;
		let n = 0, r = t.length;
		for (; n < r; n++) {
			let r = t[n];
			e[n] !== r && Yn(e, n, r);
		}
		Yn(e, `length`, r);
	} else Xn(e, t);
}
function $n(e, t, n = []) {
	let r, i = e;
	if (t.length > 1) {
		r = t.shift();
		let a = typeof r, o = Array.isArray(e);
		if (a === `string` && (r === `__proto__` || t.length > 1 && Zn(r))) return;
		if (Array.isArray(r)) {
			for (let i = 0; i < r.length; i++) $n(e, [r[i]].concat(t), n);
			return;
		}
		if (o && a === `function`) {
			for (let i = 0; i < e.length; i++) r(e[i], i) && $n(e, [i].concat(t), n);
			return;
		}
		if (o && a === `object`) {
			let { from: i = 0, to: a = e.length - 1, by: o = 1 } = r;
			for (let r = i; r <= a; r += o) $n(e, [r].concat(t), n);
			return;
		}
		if (t.length > 1) {
			$n(e[r], t, [r].concat(n));
			return;
		}
		i = e[r], n = [r].concat(n);
	}
	let a = t[0];
	typeof a == `function` && (a = a(i, n), a === i) || (r !== void 0 || a != null) && (a = Y(a), r === void 0 || Un(i) && Un(a) && !Array.isArray(a) ? Xn(i, a) : Yn(e, r, a));
}
function er(...[e, t]) {
	let n = Y(e || {}), r = Array.isArray(n), i = Hn(n);
	function a(...e) {
		H(() => {
			r && e.length === 1 ? Qn(n, e[0]) : $n(n, e);
		});
	}
	return [i, a];
}
function rr(e) {
	return (t) => {
		if (Un(t)) {
			let n;
			(n = tr.get(t)) || tr.set(t, n = new Proxy(t, nr)), e(n);
		}
		return t;
	};
}
function Os({ actionId: e, actions: t, context: n }) {
	let r = t.find((t) => t.id === e);
	return r ? Ds(r, n) : !1;
}
var M, Qe, $e, N, et, tt, nt, rt, it, P, F, I, L, R, at, _t, vt, xt, Vt, qt, Zt, rn, an, on, sn, ln, un, dn, fn, Ln, Rn, zn, Bn, Vn, Jn, tr, nr, ir, Z, ar, or, sr, cr, lr, ur, dr, fr, pr, mr, hr, gr, _r, vr, yr, br, xr, Sr, Q, Cr, wr, Tr, Er, Dr, Or, kr, Ar, jr, Mr, Nr, Pr, Fr, Ir, Lr, Rr, zr, Br, Vr, Hr, Ur, Wr, Gr, Kr, qr, Jr, Yr, Xr, Zr, Qr, $r, ei, ti, ni, ri, ii, ai, oi, si, ci, li, ui, di, fi, pi, mi, hi, gi, _i, vi, yi, bi, xi, Si, Ci, wi, Ti, Ei, Di, Oi, ki, Ai, ji, Mi, Ni, Pi, Fi, Ii, Li, Ri, zi, Bi, Vi, Hi, Ui, Wi, Gi, Ki, qi, Ji, Yi, Xi, Zi, Qi, $i, ea, ta, na, ra, ia, aa, oa, sa, ca, la, ua, da, fa, pa, ma, ha, ga, _a, va, ya, ba, xa, Sa, Ca, wa, Ta, Ea, Da, Oa, ka, Aa, ja, Ma, Na, Pa, Fa, Ia, La, Ra, za, Ba, Va, Ha, Ua, Wa, Ga, Ka, qa, Ja, Ya, Xa, Za, Qa, $a, eo, to, no, ro, io, ao, oo, so, co, lo, uo, fo, po, mo, ho, go, _o, vo, yo, bo, xo, So, Co, wo, To, Eo, Do, Oo, ko, Ao, jo, Mo, No, Po, Fo, Io, Lo, Ro, zo, Bo, Vo, Ho, Uo, Wo, Go, Ko, qo, Jo, Yo, Xo, Zo, Qo, $o, es, ts, ns, rs, is, as, os, ss, cs, ls, us, ds, fs, ps, ms, hs, gs, _s, vs, ys, bs, xs, Ss, Cs, ws, Ts, Es, Ds, ks, As, $, js, Ms, Ns, Ps, Fs, Is, Ls, Rs, zs, Us, qs, Js, Ys, Xs;
var init_core_DT_3yIQi = __esmMin((() => {
	init_freeze_updates_Cd2QP_tz();
	init_dist();
	M = {
		context: void 0,
		registry: void 0,
		effects: void 0,
		done: !1,
		getContextId() {
			return Xe(this.context.count);
		},
		getNextContextId() {
			return Xe(this.context.count++);
		}
	};
	Qe = (e, t) => e === t;
	$e = Symbol(`solid-proxy`);
	N = typeof Proxy == `function`;
	et = Symbol(`solid-track`);
	tt = { equals: Qe };
	nt = At;
	rt = {
		owned: null,
		cleanups: null,
		context: null,
		owner: null
	};
	it = {};
	P = null;
	F = null;
	I = null;
	L = null;
	R = null;
	at = 0;
	[_t, vt] = /*@__PURE__*/ z(!1);
	Vt = Symbol(`fallback`);
	qt = {
		get(e, t, n) {
			return t === $e ? n : e.get(t);
		},
		has(e, t) {
			return t === $e || e.has(t);
		},
		set: Kt,
		deleteProperty: Kt,
		getOwnPropertyDescriptor(e, t) {
			return {
				configurable: !0,
				enumerable: !0,
				get() {
					return e.get(t);
				},
				set: Kt,
				deleteProperty: Kt
			};
		},
		ownKeys(e) {
			return e.keys();
		}
	};
	Zt = (e) => `Stale read from <${e}>.`;
	rn = /*#__PURE__*/ new Set([
		`className`,
		`value`,
		`readOnly`,
		`noValidate`,
		`formNoValidate`,
		`isMap`,
		`noModule`,
		`playsInline`,
		`adAuctionHeaders`,
		`allowFullscreen`,
		`browsingTopics`,
		`defaultChecked`,
		`defaultMuted`,
		`defaultSelected`,
		`disablePictureInPicture`,
		`disableRemotePlayback`,
		`preservesPitch`,
		`shadowRootClonable`,
		`shadowRootCustomElementRegistry`,
		`shadowRootDelegatesFocus`,
		`shadowRootSerializable`,
		`sharedStorageWritable`,
		.../* @__PURE__ */ `allowfullscreen.async.alpha.autofocus.autoplay.checked.controls.default.disabled.formnovalidate.hidden.indeterminate.inert.ismap.loop.multiple.muted.nomodule.novalidate.open.playsinline.readonly.required.reversed.seamless.selected.adauctionheaders.browsingtopics.credentialless.defaultchecked.defaultmuted.defaultselected.defer.disablepictureinpicture.disableremoteplayback.preservespitch.shadowrootclonable.shadowrootcustomelementregistry.shadowrootdelegatesfocus.shadowrootserializable.sharedstoragewritable`.split(`.`)
	]);
	an = /*#__PURE__*/ new Set([
		`innerHTML`,
		`textContent`,
		`innerText`,
		`children`
	]);
	on = /*#__PURE__*/ Object.assign(Object.create(null), {
		className: `class`,
		htmlFor: `for`
	});
	sn = /*#__PURE__*/ Object.assign(Object.create(null), {
		class: `className`,
		novalidate: {
			$: `noValidate`,
			FORM: 1
		},
		formnovalidate: {
			$: `formNoValidate`,
			BUTTON: 1,
			INPUT: 1
		},
		ismap: {
			$: `isMap`,
			IMG: 1
		},
		nomodule: {
			$: `noModule`,
			SCRIPT: 1
		},
		playsinline: {
			$: `playsInline`,
			VIDEO: 1
		},
		readonly: {
			$: `readOnly`,
			INPUT: 1,
			TEXTAREA: 1
		},
		adauctionheaders: {
			$: `adAuctionHeaders`,
			IFRAME: 1
		},
		allowfullscreen: {
			$: `allowFullscreen`,
			IFRAME: 1
		},
		browsingtopics: {
			$: `browsingTopics`,
			IMG: 1
		},
		defaultchecked: {
			$: `defaultChecked`,
			INPUT: 1
		},
		defaultmuted: {
			$: `defaultMuted`,
			AUDIO: 1,
			VIDEO: 1
		},
		defaultselected: {
			$: `defaultSelected`,
			OPTION: 1
		},
		disablepictureinpicture: {
			$: `disablePictureInPicture`,
			VIDEO: 1
		},
		disableremoteplayback: {
			$: `disableRemotePlayback`,
			AUDIO: 1,
			VIDEO: 1
		},
		preservespitch: {
			$: `preservesPitch`,
			AUDIO: 1,
			VIDEO: 1
		},
		shadowrootclonable: {
			$: `shadowRootClonable`,
			TEMPLATE: 1
		},
		shadowrootdelegatesfocus: {
			$: `shadowRootDelegatesFocus`,
			TEMPLATE: 1
		},
		shadowrootserializable: {
			$: `shadowRootSerializable`,
			TEMPLATE: 1
		},
		sharedstoragewritable: {
			$: `sharedStorageWritable`,
			IFRAME: 1,
			IMG: 1
		}
	});
	ln = /*#__PURE__*/ new Set([
		`beforeinput`,
		`click`,
		`dblclick`,
		`contextmenu`,
		`focusin`,
		`focusout`,
		`input`,
		`keydown`,
		`keyup`,
		`mousedown`,
		`mousemove`,
		`mouseout`,
		`mouseover`,
		`mouseup`,
		`pointerdown`,
		`pointermove`,
		`pointerout`,
		`pointerover`,
		`pointerup`,
		`touchend`,
		`touchmove`,
		`touchstart`
	]);
	un = {
		xlink: `http://www.w3.org/1999/xlink`,
		xml: `http://www.w3.org/XML/1998/namespace`
	};
	dn = (e) => V(() => e());
	fn = `_$DX_DELEGATE`;
	Ln = `/*! tailwindcss v4.3.5 | MIT License | https://tailwindcss.com */
@layer properties{@supports (((-webkit-hyphens:none)) and (not (margin-trim:block))) or ((-moz-orient:inline) and (not (color:rgb(from red r g b)))){*,:before,:after,::backdrop{--tw-rotate-x:initial;--tw-rotate-y:initial;--tw-rotate-z:initial;--tw-skew-x:initial;--tw-skew-y:initial;--tw-border-style:solid;--tw-leading:initial;--tw-font-weight:initial;--tw-ordinal:initial;--tw-slashed-zero:initial;--tw-numeric-figure:initial;--tw-numeric-spacing:initial;--tw-numeric-fraction:initial;--tw-shadow:0 0 #0000;--tw-shadow-color:initial;--tw-shadow-alpha:100%;--tw-inset-shadow:0 0 #0000;--tw-inset-shadow-color:initial;--tw-inset-shadow-alpha:100%;--tw-ring-color:initial;--tw-ring-shadow:0 0 #0000;--tw-inset-ring-color:initial;--tw-inset-ring-shadow:0 0 #0000;--tw-ring-inset:initial;--tw-ring-offset-width:0px;--tw-ring-offset-color:#fff;--tw-ring-offset-shadow:0 0 #0000;--tw-outline-style:solid;--tw-blur:initial;--tw-brightness:initial;--tw-contrast:initial;--tw-grayscale:initial;--tw-hue-rotate:initial;--tw-invert:initial;--tw-opacity:initial;--tw-saturate:initial;--tw-sepia:initial;--tw-drop-shadow:initial;--tw-drop-shadow-color:initial;--tw-drop-shadow-alpha:100%;--tw-drop-shadow-size:initial;--tw-backdrop-blur:initial;--tw-backdrop-brightness:initial;--tw-backdrop-contrast:initial;--tw-backdrop-grayscale:initial;--tw-backdrop-hue-rotate:initial;--tw-backdrop-invert:initial;--tw-backdrop-opacity:initial;--tw-backdrop-saturate:initial;--tw-backdrop-sepia:initial;--tw-duration:initial;--tw-ease:initial;--tw-contain-size:initial;--tw-contain-layout:initial;--tw-contain-paint:initial;--tw-contain-style:initial}}}@layer theme{:root,:host{--font-sans:"Geist", ui-sans-serif, system-ui, sans-serif;--font-serif:ui-serif, Georgia, Cambria, "Times New Roman", Times, serif;--font-mono:ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace;--color-red-500:oklch(63.7% .237 25.331);--spacing:4px;--font-weight-thin:100;--font-weight-extralight:200;--font-weight-light:300;--font-weight-normal:400;--font-weight-medium:500;--font-weight-semibold:600;--font-weight-bold:700;--font-weight-extrabold:800;--font-weight-black:900;--radius-sm:4px;--ease-out:cubic-bezier(0, 0, .2, 1);--default-transition-duration:.15s;--default-transition-timing-function:cubic-bezier(.4, 0, .2, 1);--default-font-family:var(--font-sans);--default-mono-font-family:var(--font-mono);--ease-drawer:cubic-bezier(.32, .72, 0, 1);--ease-spring:cubic-bezier(.34, 1.56, .64, 1)}}@layer base{*,:after,:before,::backdrop{box-sizing:border-box;border:0 solid;margin:0;padding:0}::file-selector-button{box-sizing:border-box;border:0 solid;margin:0;padding:0}html,:host{-webkit-text-size-adjust:100%;tab-size:4;line-height:1.5;font-family:var(--default-font-family,-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", "Noto Sans", Arial, sans-serif, "Apple Color Emoji", "Segoe UI Emoji", "Segoe UI Symbol", "Noto Color Emoji");font-feature-settings:var(--default-font-feature-settings,normal);font-variation-settings:var(--default-font-variation-settings,normal);-webkit-tap-highlight-color:transparent}hr{height:0;color:inherit;border-top-width:1px}abbr:where([title]){-webkit-text-decoration:underline dotted;text-decoration:underline dotted}h1,h2,h3,h4,h5,h6{font-size:inherit;font-weight:inherit}a{color:inherit;-webkit-text-decoration:inherit;-webkit-text-decoration:inherit;-webkit-text-decoration:inherit;text-decoration:inherit}b,strong{font-weight:bolder}code,kbd,samp,pre{font-family:var(--default-mono-font-family,ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace);font-feature-settings:var(--default-mono-font-feature-settings,normal);font-variation-settings:var(--default-mono-font-variation-settings,normal);font-size:1em}small{font-size:80%}sub,sup{vertical-align:baseline;font-size:75%;line-height:0;position:relative}sub{bottom:-.25em}sup{top:-.5em}table{text-indent:0;border-color:inherit;border-collapse:collapse}:-moz-focusring:where(:not(iframe)){outline:auto}progress{vertical-align:baseline}summary{display:list-item}ol,ul,menu{list-style:none}img,svg,video,canvas,audio,iframe,embed,object{vertical-align:middle;display:block}img,video{max-width:100%;height:auto}button,input,select,optgroup,textarea{font:inherit;font-feature-settings:inherit;font-variation-settings:inherit;letter-spacing:inherit;color:inherit;opacity:1;background-color:#0000;border-radius:0}::file-selector-button{font:inherit;font-feature-settings:inherit;font-variation-settings:inherit;letter-spacing:inherit;color:inherit;opacity:1;background-color:#0000;border-radius:0}:where(select:is([multiple],[size])) optgroup{font-weight:bolder}:where(select:is([multiple],[size])) optgroup option{padding-inline-start:20px}::file-selector-button{margin-inline-end:4px}::placeholder{opacity:1}@supports (not ((-webkit-appearance:-apple-pay-button))) or (contain-intrinsic-size:1px){::placeholder{color:currentColor}@supports (color:color-mix(in lab, red, red)){::placeholder{color:color-mix(in oklab, currentcolor 50%, transparent)}}}textarea{resize:vertical}::-webkit-search-decoration{-webkit-appearance:none}::-webkit-date-and-time-value{min-height:1lh;text-align:inherit}::-webkit-datetime-edit{display:inline-flex}::-webkit-datetime-edit-fields-wrapper{padding:0}::-webkit-datetime-edit{padding-block:0}::-webkit-datetime-edit-year-field{padding-block:0}::-webkit-datetime-edit-month-field{padding-block:0}::-webkit-datetime-edit-day-field{padding-block:0}::-webkit-datetime-edit-hour-field{padding-block:0}::-webkit-datetime-edit-minute-field{padding-block:0}::-webkit-datetime-edit-second-field{padding-block:0}::-webkit-datetime-edit-millisecond-field{padding-block:0}::-webkit-datetime-edit-meridiem-field{padding-block:0}::-webkit-calendar-picker-indicator{line-height:1}:-moz-ui-invalid{box-shadow:none}button,input:where([type=button],[type=reset],[type=submit]){appearance:button}::file-selector-button{appearance:button}::-webkit-inner-spin-button{height:auto}::-webkit-outer-spin-button{height:auto}[hidden]:where(:not([hidden=until-found])){display:none!important}}@layer components;@layer utilities{.a11y-hitbox{position:relative}.a11y-hitbox:before{content:"";pointer-events:auto;width:100%;min-width:24px;height:100%;min-height:24px;display:block;position:absolute;top:50%;left:50%;transform:translate(-50%,-50%)}.pointer-events-auto{pointer-events:auto}.pointer-events-none{pointer-events:none}.collapse{visibility:collapse}.invisible{visibility:hidden}.visible{visibility:visible}.absolute{position:absolute}.fixed{position:fixed}.relative{position:relative}.static{position:static}.inset-0{inset:0}.inset-y-0{inset-block:0}.top-0{top:0}.top-1\\/2{top:50%}.top-\\[2px\\]{top:2px}.top-full{top:100%}.right-0{right:0}.right-full{right:100%}.bottom-\\[2px\\]{bottom:2px}.bottom-full{bottom:100%}.left-0{left:0}.left-full{left:100%}.z-1{z-index:1}.z-10{z-index:10}.container{width:100%}@media (min-width:640px){.container{max-width:640px}}@media (min-width:768px){.container{max-width:768px}}@media (min-width:1024px){.container{max-width:1024px}}@media (min-width:1280px){.container{max-width:1280px}}@media (min-width:1536px){.container{max-width:1536px}}.-m-0\\.5{margin:calc(var(--spacing) * -.5)}.-m-4{margin:calc(var(--spacing) * -4)}.m-0{margin:0}.-mx-2{margin-inline:calc(var(--spacing) * -2)}.-my-1\\.5{margin-block:calc(var(--spacing) * -1.5)}.-mt-\\[8px\\]{margin-top:-8px}.mt-2\\.5{margin-top:calc(var(--spacing) * 2.5)}.mr-1{margin-right:var(--spacing)}.mr-1\\.5{margin-right:calc(var(--spacing) * 1.5)}.mr-2\\.5{margin-right:calc(var(--spacing) * 2.5)}.mb-1\\.5{margin-bottom:calc(var(--spacing) * 1.5)}.mb-2\\.5{margin-bottom:calc(var(--spacing) * 2.5)}.ml-1{margin-left:var(--spacing)}.ml-2\\.5{margin-left:calc(var(--spacing) * 2.5)}.ml-4{margin-left:calc(var(--spacing) * 4)}.ml-auto{margin-left:auto}.ml-px{margin-left:1px}.no-scrollbar{scrollbar-width:none;-ms-overflow-style:none}.no-scrollbar::-webkit-scrollbar{width:0;height:0;display:none}.line-clamp-5{-webkit-line-clamp:5;-webkit-box-orient:vertical;display:-webkit-box;overflow:hidden}.block{display:block}.flex{display:flex}.grid{display:grid}.hidden{display:none}.inline{display:inline}.inline-block{display:inline-block}.inline-flex{display:inline-flex}.list-item{display:list-item}.size-0{width:0;height:0}.size-4{width:calc(var(--spacing) * 4);height:calc(var(--spacing) * 4)}.size-\\[12px\\]{width:12px;height:12px}.size-\\[16px\\]{width:16px;height:16px}.h-\\[8px\\]{height:8px}.h-\\[17px\\]{height:17px}.h-\\[20px\\]{height:20px}.h-\\[24px\\]{height:24px}.h-fit{height:fit-content}.max-h-\\[var\\(--rg-edit-list-max-h\\)\\]{max-height:var(--rg-edit-list-max-h)}.min-h-0{min-height:0}.min-h-4{min-height:calc(var(--spacing) * 4)}.min-h-\\[28px\\]{min-height:28px}.w-\\[2px\\]{width:2px}.w-\\[calc\\(100\\%\\+16px\\)\\]{width:calc(100% + 16px)}.w-auto{width:auto}.w-fit{width:fit-content}.w-full{width:100%}.w-px{width:1px}.max-w-\\[280px\\]{max-width:280px}.max-w-full{max-width:100%}.min-w-0{min-width:0}.min-w-\\[36px\\]{min-width:36px}.min-w-\\[100px\\]{min-width:100px}.min-w-\\[150px\\]{min-width:150px}.flex-1{flex:1}.flex-shrink,.shrink{flex-shrink:1}.shrink-0{flex-shrink:0}.flex-grow,.grow{flex-grow:1}.-rotate-90{rotate:-90deg}.rotate-0{rotate:0deg}.rotate-90{rotate:90deg}.rotate-180{rotate:180deg}.interactive-scale{transition:transform .4s cubic-bezier(.34,1.56,.64,1)}@media (hover:hover) and (pointer:fine){.interactive-scale:hover{transform:scale(1.05)}}.interactive-scale:active{transition:transform 60ms cubic-bezier(0,0,.2,1);transform:scale(.96)}.press-scale{transition:transform .4s cubic-bezier(.34,1.56,.64,1)}.press-scale:active{transition:transform 60ms cubic-bezier(0,0,.2,1);transform:scale(.97)}.transform{transform:var(--tw-rotate-x, ) var(--tw-rotate-y, ) var(--tw-rotate-z, ) var(--tw-skew-x, ) var(--tw-skew-y, )}.cursor-grab{cursor:grab}.cursor-grabbing{cursor:grabbing}.cursor-pointer{cursor:pointer}.cursor-text{cursor:text}.resize{resize:both}.resize-none{resize:none}.grid-cols-\\[0fr\\]{grid-template-columns:0fr}.grid-cols-\\[1fr\\]{grid-template-columns:1fr}.grid-rows-\\[0fr\\]{grid-template-rows:0fr}.grid-rows-\\[1fr\\]{grid-template-rows:1fr}.flex-col{flex-direction:column}.flex-wrap{flex-wrap:wrap}.items-center{align-items:center}.items-end{align-items:flex-end}.items-start{align-items:flex-start}.justify-between{justify-content:space-between}.justify-center{justify-content:center}.justify-end{justify-content:flex-end}.gap-0{gap:0}.gap-0\\.5{gap:calc(var(--spacing) * .5)}.gap-1{gap:var(--spacing)}.gap-2{gap:calc(var(--spacing) * 2)}.gap-\\[5px\\]{gap:5px}.self-stretch{align-self:stretch}.truncate{text-overflow:ellipsis;white-space:nowrap;overflow:hidden}.overflow-hidden{overflow:hidden}.overflow-visible{overflow:visible}.overflow-y-auto{overflow-y:auto}.rounded{border-radius:4px}.rounded-\\[1px\\]{border-radius:1px}.rounded-\\[3px\\]{border-radius:3px}.rounded-\\[4px\\]{border-radius:4px}.rounded-\\[6px\\]{border-radius:6px}.rounded-\\[13px\\]{border-radius:13px}.rounded-\\[14px\\]{border-radius:14px}.rounded-full{border-radius:2147483647px}.rounded-sm{border-radius:var(--radius-sm)}.rounded-t{border-top-left-radius:4px;border-top-right-radius:4px}.rounded-t-\\[10px\\]{border-top-left-radius:10px;border-top-right-radius:10px}.rounded-t-none{border-top-left-radius:0;border-top-right-radius:0}.rounded-l{border-top-left-radius:4px;border-bottom-left-radius:4px}.rounded-l-\\[10px\\]{border-top-left-radius:10px;border-bottom-left-radius:10px}.rounded-l-none{border-top-left-radius:0;border-bottom-left-radius:0}.rounded-tl{border-top-left-radius:4px}.rounded-r{border-top-right-radius:4px;border-bottom-right-radius:4px}.rounded-r-\\[10px\\]{border-top-right-radius:10px;border-bottom-right-radius:10px}.rounded-r-none{border-top-right-radius:0;border-bottom-right-radius:0}.rounded-tr{border-top-right-radius:4px}.rounded-b{border-bottom-right-radius:4px;border-bottom-left-radius:4px}.rounded-b-\\[10px\\]{border-bottom-right-radius:10px;border-bottom-left-radius:10px}.rounded-b-none{border-bottom-right-radius:0;border-bottom-left-radius:0}.rounded-br{border-bottom-right-radius:4px}.rounded-bl{border-bottom-left-radius:4px}.border{border-style:var(--tw-border-style);border-width:1px}.border-2{border-style:var(--tw-border-style);border-width:2px}.\\[border-width\\:0\\.5px\\]{border-width:.5px}.border-x{border-inline-style:var(--tw-border-style);border-inline-width:1px}.border-y{border-block-style:var(--tw-border-style);border-block-width:1px}.border-s{border-inline-start-style:var(--tw-border-style);border-inline-start-width:1px}.border-e{border-inline-end-style:var(--tw-border-style);border-inline-end-width:1px}.border-t{border-top-style:var(--tw-border-style);border-top-width:1px}.border-t-4{border-top-style:var(--tw-border-style);border-top-width:4px}.\\[border-top-width\\:0\\.5px\\]{border-top-width:.5px}.border-r{border-right-style:var(--tw-border-style);border-right-width:1px}.border-b{border-bottom-style:var(--tw-border-style);border-bottom-width:1px}.border-l{border-left-style:var(--tw-border-style);border-left-width:1px}.border-dashed{--tw-border-style:dashed;border-style:dashed}.border-dotted{--tw-border-style:dotted;border-style:dotted}.border-none{--tw-border-style:none;border-style:none}.border-solid{--tw-border-style:solid;border-style:solid}.border-\\[var\\(--rg-border-button\\)\\]{border-color:var(--rg-border-button)}.border-t-\\[var\\(--rg-border-subtle\\)\\]{border-top-color:var(--rg-border-subtle)}.bg-\\[\\#f00\\]{background-color:red}.bg-\\[var\\(--rg-error-bg\\)\\]{background-color:var(--rg-error-bg)}.bg-\\[var\\(--rg-panel-bg\\)\\]{background-color:var(--rg-panel-bg)}.bg-\\[var\\(--rg-submit-bg\\)\\]{background-color:var(--rg-submit-bg)}.bg-\\[var\\(--rg-surface-active\\)\\]{background-color:var(--rg-surface-active)}.bg-\\[var\\(--rg-surface-hover\\)\\]{background-color:var(--rg-surface-hover)}.bg-\\[var\\(--rg-text-primary\\)\\]{background-color:var(--rg-text-primary)}.bg-\\[var\\(--rg-text-secondary\\)\\]{background-color:var(--rg-text-secondary)}.bg-red-500{background-color:var(--color-red-500)}.bg-transparent{background-color:#0000}.p-0{padding:0}.p-1{padding:var(--spacing)}.p-6{padding:calc(var(--spacing) * 6)}.px-0{padding-inline:0}.px-0\\.25{padding-inline:calc(var(--spacing) * .25)}.px-1{padding-inline:var(--spacing)}.px-1\\.5{padding-inline:calc(var(--spacing) * 1.5)}.px-2{padding-inline:calc(var(--spacing) * 2)}.px-3{padding-inline:calc(var(--spacing) * 3)}.px-\\[3px\\]{padding-inline:3px}.py-0{padding-block:0}.py-0\\.5{padding-block:calc(var(--spacing) * .5)}.py-0\\.25{padding-block:calc(var(--spacing) * .25)}.py-1{padding-block:var(--spacing)}.py-1\\.5{padding-block:calc(var(--spacing) * 1.5)}.py-2{padding-block:calc(var(--spacing) * 2)}.py-px{padding-block:1px}.pt-1\\.5{padding-top:calc(var(--spacing) * 1.5)}.pt-4{padding-top:calc(var(--spacing) * 4)}.pb-1{padding-bottom:var(--spacing)}.text-center{text-align:center}.text-justify{text-align:justify}.text-left{text-align:left}.text-right{text-align:right}.font-mono{font-family:var(--font-mono)}.font-sans{font-family:var(--font-sans)}.font-serif{font-family:var(--font-serif)}.text-\\[9px\\]{font-size:9px}.text-\\[10px\\]{font-size:10px}.text-\\[11px\\]{font-size:11px}.text-\\[12px\\]{font-size:12px}.text-\\[13px\\]{font-size:13px}.leading-3\\.5{--tw-leading:calc(var(--spacing) * 3.5);line-height:calc(var(--spacing) * 3.5)}.leading-4{--tw-leading:calc(var(--spacing) * 4);line-height:calc(var(--spacing) * 4)}.font-black{--tw-font-weight:var(--font-weight-black);font-weight:var(--font-weight-black)}.font-bold{--tw-font-weight:var(--font-weight-bold);font-weight:var(--font-weight-bold)}.font-extrabold{--tw-font-weight:var(--font-weight-extrabold);font-weight:var(--font-weight-extrabold)}.font-extralight{--tw-font-weight:var(--font-weight-extralight);font-weight:var(--font-weight-extralight)}.font-light{--tw-font-weight:var(--font-weight-light);font-weight:var(--font-weight-light)}.font-medium{--tw-font-weight:var(--font-weight-medium);font-weight:var(--font-weight-medium)}.font-normal{--tw-font-weight:var(--font-weight-normal);font-weight:var(--font-weight-normal)}.font-semibold{--tw-font-weight:var(--font-weight-semibold);font-weight:var(--font-weight-semibold)}.font-thin{--tw-font-weight:var(--font-weight-thin);font-weight:var(--font-weight-thin)}.break-normal{overflow-wrap:normal;word-break:normal}.break-words,.wrap-break-word{overflow-wrap:break-word}.break-all{word-break:break-all}.break-keep{word-break:keep-all}.text-ellipsis{text-overflow:ellipsis}.whitespace-break-spaces{white-space:break-spaces}.whitespace-normal{white-space:normal}.whitespace-nowrap{white-space:nowrap}.whitespace-pre{white-space:pre}.whitespace-pre-line{white-space:pre-line}.whitespace-pre-wrap{white-space:pre-wrap}.text-\\[var\\(--rg-error-text\\)\\]{color:var(--rg-error-text)}.text-\\[var\\(--rg-submit-fg\\)\\]{color:var(--rg-submit-fg)}.text-\\[var\\(--rg-text-primary\\)\\]{color:var(--rg-text-primary)}.text-\\[var\\(--rg-text-primary-85\\)\\]{color:var(--rg-text-primary-85)}.text-\\[var\\(--rg-text-secondary\\)\\]{color:var(--rg-text-secondary)}.capitalize{text-transform:capitalize}.lowercase{text-transform:lowercase}.normal-case{text-transform:none}.uppercase{text-transform:uppercase}.italic{font-style:italic}.not-italic{font-style:normal}.lining-nums{--tw-numeric-figure:lining-nums;font-variant-numeric:var(--tw-ordinal, ) var(--tw-slashed-zero, ) var(--tw-numeric-figure, ) var(--tw-numeric-spacing, ) var(--tw-numeric-fraction, )}.oldstyle-nums{--tw-numeric-figure:oldstyle-nums;font-variant-numeric:var(--tw-ordinal, ) var(--tw-slashed-zero, ) var(--tw-numeric-figure, ) var(--tw-numeric-spacing, ) var(--tw-numeric-fraction, )}.proportional-nums{--tw-numeric-spacing:proportional-nums;font-variant-numeric:var(--tw-ordinal, ) var(--tw-slashed-zero, ) var(--tw-numeric-figure, ) var(--tw-numeric-spacing, ) var(--tw-numeric-fraction, )}.slashed-zero{--tw-slashed-zero:slashed-zero;font-variant-numeric:var(--tw-ordinal, ) var(--tw-slashed-zero, ) var(--tw-numeric-figure, ) var(--tw-numeric-spacing, ) var(--tw-numeric-fraction, )}.tabular-nums{--tw-numeric-spacing:tabular-nums;font-variant-numeric:var(--tw-ordinal, ) var(--tw-slashed-zero, ) var(--tw-numeric-figure, ) var(--tw-numeric-spacing, ) var(--tw-numeric-fraction, )}.line-through{text-decoration-line:line-through}.no-underline{text-decoration-line:none}.underline{text-decoration-line:underline}.antialiased{-webkit-font-smoothing:antialiased;-moz-osx-font-smoothing:grayscale}.opacity-0{opacity:0}.opacity-50{opacity:.5}.opacity-60{opacity:.6}.opacity-100{opacity:1}.shadow{--tw-shadow:0 1px 3px 0 var(--tw-shadow-color,#0000001a), 0 1px 2px -1px var(--tw-shadow-color,#0000001a);box-shadow:var(--tw-inset-shadow), var(--tw-inset-ring-shadow), var(--tw-ring-offset-shadow), var(--tw-ring-shadow), var(--tw-shadow)}.ring{--tw-ring-shadow:var(--tw-ring-inset, ) 0 0 0 calc(1px + var(--tw-ring-offset-width)) var(--tw-ring-color,currentcolor);box-shadow:var(--tw-inset-shadow), var(--tw-inset-ring-shadow), var(--tw-ring-offset-shadow), var(--tw-ring-shadow), var(--tw-shadow)}.\\[box-shadow\\:var\\(--rg-shadow\\)\\]{box-shadow:var(--rg-shadow)}.outline{outline-style:var(--tw-outline-style);outline-width:1px}.blur{--tw-blur:blur(8px);filter:var(--tw-blur, ) var(--tw-brightness, ) var(--tw-contrast, ) var(--tw-grayscale, ) var(--tw-hue-rotate, ) var(--tw-invert, ) var(--tw-saturate, ) var(--tw-sepia, ) var(--tw-drop-shadow, )}.\\[filter\\:var\\(--rg-drop-shadow\\)\\]{filter:var(--rg-drop-shadow)}.filter{filter:var(--tw-blur, ) var(--tw-brightness, ) var(--tw-contrast, ) var(--tw-grayscale, ) var(--tw-hue-rotate, ) var(--tw-invert, ) var(--tw-saturate, ) var(--tw-sepia, ) var(--tw-drop-shadow, )}.backdrop-filter{-webkit-backdrop-filter:var(--tw-backdrop-blur, ) var(--tw-backdrop-brightness, ) var(--tw-backdrop-contrast, ) var(--tw-backdrop-grayscale, ) var(--tw-backdrop-hue-rotate, ) var(--tw-backdrop-invert, ) var(--tw-backdrop-opacity, ) var(--tw-backdrop-saturate, ) var(--tw-backdrop-sepia, );backdrop-filter:var(--tw-backdrop-blur, ) var(--tw-backdrop-brightness, ) var(--tw-backdrop-contrast, ) var(--tw-backdrop-grayscale, ) var(--tw-backdrop-hue-rotate, ) var(--tw-backdrop-invert, ) var(--tw-backdrop-opacity, ) var(--tw-backdrop-saturate, ) var(--tw-backdrop-sepia, )}.transition{transition-property:color,background-color,border-color,outline-color,text-decoration-color,fill,stroke,--tw-gradient-from,--tw-gradient-via,--tw-gradient-to,opacity,box-shadow,transform,translate,scale,rotate,filter,-webkit-backdrop-filter,backdrop-filter,display,content-visibility,overlay,pointer-events;transition-timing-function:var(--tw-ease,var(--default-transition-timing-function));transition-duration:var(--tw-duration,var(--default-transition-duration))}.transition-\\[grid-template-columns\\]{transition-property:grid-template-columns;transition-timing-function:var(--tw-ease,var(--default-transition-timing-function));transition-duration:var(--tw-duration,var(--default-transition-duration))}.transition-\\[grid-template-rows\\]{transition-property:grid-template-rows;transition-timing-function:var(--tw-ease,var(--default-transition-timing-function));transition-duration:var(--tw-duration,var(--default-transition-duration))}.transition-\\[opacity\\,transform\\]{transition-property:opacity,transform;transition-timing-function:var(--tw-ease,var(--default-transition-timing-function));transition-duration:var(--tw-duration,var(--default-transition-duration))}.transition-\\[padding\\,border-radius\\,transform\\]{transition-property:padding,border-radius,transform;transition-timing-function:var(--tw-ease,var(--default-transition-timing-function));transition-duration:var(--tw-duration,var(--default-transition-duration))}.transition-\\[top\\,left\\,width\\,height\\,opacity\\,border-radius\\]{transition-property:top,left,width,height,opacity,border-radius;transition-timing-function:var(--tw-ease,var(--default-transition-timing-function));transition-duration:var(--tw-duration,var(--default-transition-duration))}.transition-\\[transform\\,color\\]{transition-property:transform,color;transition-timing-function:var(--tw-ease,var(--default-transition-timing-function));transition-duration:var(--tw-duration,var(--default-transition-duration))}.transition-\\[transform\\,opacity\\]{transition-property:transform,opacity;transition-timing-function:var(--tw-ease,var(--default-transition-timing-function));transition-duration:var(--tw-duration,var(--default-transition-duration))}.transition-all{transition-property:all;transition-timing-function:var(--tw-ease,var(--default-transition-timing-function));transition-duration:var(--tw-duration,var(--default-transition-duration))}.transition-colors{transition-property:color,background-color,border-color,outline-color,text-decoration-color,fill,stroke,--tw-gradient-from,--tw-gradient-via,--tw-gradient-to;transition-timing-function:var(--tw-ease,var(--default-transition-timing-function));transition-duration:var(--tw-duration,var(--default-transition-duration))}.transition-opacity{transition-property:opacity;transition-timing-function:var(--tw-ease,var(--default-transition-timing-function));transition-duration:var(--tw-duration,var(--default-transition-duration))}.delay-\\[80ms\\]{transition-delay:80ms}.duration-60{--tw-duration:60ms;transition-duration:60ms}.duration-75{--tw-duration:75ms;transition-duration:75ms}.duration-100{--tw-duration:.1s;transition-duration:.1s}.duration-120{--tw-duration:.12s;transition-duration:.12s}.duration-140{--tw-duration:.14s;transition-duration:.14s}.duration-150{--tw-duration:.15s;transition-duration:.15s}.duration-180{--tw-duration:.18s;transition-duration:.18s}.duration-220{--tw-duration:.22s;transition-duration:.22s}.duration-300{--tw-duration:.3s;transition-duration:.3s}.duration-400{--tw-duration:.4s;transition-duration:.4s}.ease-\\[cubic-bezier\\(0\\,0\\,0\\.2\\,1\\)\\]{--tw-ease:cubic-bezier(0,0,.2,1);transition-timing-function:cubic-bezier(0,0,.2,1)}.ease-drawer{--tw-ease:var(--ease-drawer);transition-timing-function:var(--ease-drawer)}.ease-out{--tw-ease:var(--ease-out);transition-timing-function:var(--ease-out)}.ease-spring{--tw-ease:var(--ease-spring);transition-timing-function:var(--ease-spring)}.will-change-\\[opacity\\,transform\\]{will-change:opacity,transform}.will-change-transform{will-change:transform}.contain-layout{--tw-contain-layout:layout;contain:var(--tw-contain-size, ) var(--tw-contain-layout, ) var(--tw-contain-paint, ) var(--tw-contain-style, )}.outline-none{--tw-outline-style:none;outline-style:none}.select-none{-webkit-user-select:none;user-select:none}.\\[corner-shape\\:superellipse\\(1\\.25\\)\\]{corner-shape:superellipse(1.25)}.\\[font-synthesis\\:none\\]{font-synthesis:none}@media (hover:hover){:is(:where(.group):hover .group-hover\\:text-\\[var\\(--rg-text-primary\\)\\]){color:var(--rg-text-primary)}:is(:where(.group):hover .group-hover\\:opacity-100){opacity:1}.hover\\:bg-\\[var\\(--rg-error-bg-hover\\)\\]:hover{background-color:var(--rg-error-bg-hover)}.hover\\:bg-\\[var\\(--rg-surface-active\\)\\]:hover{background-color:var(--rg-surface-active)}.hover\\:bg-\\[var\\(--rg-surface-hover\\)\\]:hover{background-color:var(--rg-surface-hover)}.hover\\:text-\\[var\\(--rg-text-primary\\)\\]:hover{color:var(--rg-text-primary)}}.disabled\\:cursor-default:disabled{cursor:default}.disabled\\:opacity-40:disabled{opacity:.4}}:host{all:initial;direction:ltr}:host,[data-rg-theme=dark]{--rg-panel-bg:#161616;--rg-text-primary:#fff;--rg-text-secondary:#a7a7a7;--rg-surface-hover:#ffffff1a;--rg-surface-active:#ffffff26;--rg-border-subtle:#ffffff1a;--rg-border-button:#fff3;--rg-submit-bg:#fff;--rg-submit-fg:#161616;--rg-error-text:#f87171;--rg-error-bg:#3d1515;--rg-error-bg-hover:#4d1f1f;--rg-shadow:0 2px 8px #00000014;--rg-drop-shadow:drop-shadow(0px 2px 8px #00000014);--rg-shimmer-from:#a1a1aa;--rg-shimmer-to:#d4d4d8;--rg-text-primary-85:#ffffffd9}@media (prefers-color-scheme:dark){:host:not([data-rg-theme]){--rg-panel-bg:#fff;--rg-text-primary:#171717;--rg-text-secondary:#737373;--rg-surface-hover:#0000000d;--rg-surface-active:#00000014;--rg-border-subtle:#00000014;--rg-border-button:#00000026;--rg-submit-bg:#171717;--rg-submit-fg:#fff;--rg-error-text:#dc2626;--rg-error-bg:#fee2e2;--rg-error-bg-hover:#fecaca;--rg-shadow:0 2px 12px #0000001a;--rg-drop-shadow:drop-shadow(0px 2px 12px #0000001a);--rg-shimmer-from:#71717a;--rg-shimmer-to:#a1a1aa;--rg-text-primary-85:#171717d9}}:host([data-rg-theme=light]),[data-rg-theme=light]{--rg-panel-bg:#fff;--rg-text-primary:#171717;--rg-text-secondary:#737373;--rg-surface-hover:#0000000d;--rg-surface-active:#00000014;--rg-border-subtle:#00000014;--rg-border-button:#00000026;--rg-submit-bg:#171717;--rg-submit-fg:#fff;--rg-error-text:#dc2626;--rg-error-bg:#fee2e2;--rg-error-bg-hover:#fecaca;--rg-shadow:0 2px 12px #0000001a;--rg-drop-shadow:drop-shadow(0px 2px 12px #0000001a);--rg-shimmer-from:#71717a;--rg-shimmer-to:#a1a1aa;--rg-text-primary-85:#171717d9}@keyframes rg-slot-enter-up{0%{filter:blur(2px);opacity:0;transform:translateY(.5em)scale(.5)}to{filter:blur();opacity:1;transform:translateY(0)scale(1)}}@keyframes rg-slot-exit-up{0%{filter:blur();opacity:1;transform:translateY(0)scale(1)}to{filter:blur(2px);opacity:0;transform:translateY(-.5em)scale(.5)}}@keyframes rg-slot-enter-down{0%{filter:blur(2px);opacity:0;transform:translateY(-.5em)scale(.5)}to{filter:blur();opacity:1;transform:translateY(0)scale(1)}}@keyframes rg-slot-exit-down{0%{filter:blur();opacity:1;transform:translateY(0)scale(1)}to{filter:blur(2px);opacity:0;transform:translateY(.5em)scale(.5)}}@keyframes rg-slot-enter-fade{0%{filter:blur(2px);opacity:0;transform:scale(.5)}to{filter:blur();opacity:1;transform:scale(1)}}@keyframes rg-slot-exit-fade{0%{filter:blur();opacity:1;transform:scale(1)}to{filter:blur(2px);opacity:0;transform:scale(.5)}}.rg-slot-column{display:inline-block;position:relative}.rg-slot-sizer{visibility:hidden;white-space:pre;display:inline-block}.rg-slot-cell{white-space:pre;will-change:transform, filter, opacity;animation-duration:var(--rg-slot-dur,.35s);animation-timing-function:var(--ease-spring);animation-fill-mode:both;display:inline-block;position:absolute;top:0;left:0}.rg-slot-cell.rg-slot-enter[data-dir="1"]{animation-name:rg-slot-enter-up}.rg-slot-cell.rg-slot-enter[data-dir="-1"]{animation-name:rg-slot-enter-down}.rg-slot-cell.rg-slot-enter[data-dir="0"]{animation-name:rg-slot-enter-fade}.rg-slot-cell.rg-slot-exit[data-dir="1"]{animation-name:rg-slot-exit-up}.rg-slot-cell.rg-slot-exit[data-dir="-1"]{animation-name:rg-slot-exit-down}.rg-slot-cell.rg-slot-exit[data-dir="0"]{animation-name:rg-slot-exit-fade}@media (prefers-reduced-motion:reduce){:host,:host *,:host :before,:host :after{scroll-behavior:auto!important;transition:none!important;animation:none!important}}:host :focus{box-shadow:none!important;outline:none!important}@keyframes shake{0%,to{translate:0}15%,45%{translate:calc(-3px * var(--rg-shake-x,1)) calc(-3px * var(--rg-shake-y,0))}30%,60%{translate:calc(3px * var(--rg-shake-x,1)) calc(3px * var(--rg-shake-y,0))}75%{translate:calc(-2px * var(--rg-shake-x,1)) calc(-2px * var(--rg-shake-y,0))}90%{translate:calc(2px * var(--rg-shake-x,1)) calc(2px * var(--rg-shake-y,0))}}@keyframes icon-loader-spin{0%{opacity:1}50%{opacity:.5}to{opacity:.2}}.icon-loader-bar{animation:.5s linear infinite icon-loader-spin}@keyframes shimmer{0%{background-position:200% 0}to{background-position:-200% 0}}.shimmer-text{background:linear-gradient(90deg, var(--rg-shimmer-from) 0%, var(--rg-shimmer-to) 25%, var(--rg-shimmer-from) 50%, var(--rg-shimmer-to) 75%, var(--rg-shimmer-from) 100%);color:#0000;background-size:200% 100%;-webkit-background-clip:text;background-clip:text;animation:2.5s linear infinite shimmer}.animate-shake,.animate-shake-vertical{will-change:translate;animation:.3s ease-out shake}@keyframes tooltip-fade-in{0%{opacity:0;transform:scale(.97)}to{opacity:1;transform:scale(1)}}.animate-tooltip-fade-in{will-change:transform, opacity;animation:.1s ease-out tooltip-fade-in}.animate-shake-vertical{--rg-shake-x:0;--rg-shake-y:1}@supports (-webkit-touch-callout:none){textarea[data-react-grab-input]{font-size:16px}}@property --tw-rotate-x{syntax:"*";inherits:false}@property --tw-rotate-y{syntax:"*";inherits:false}@property --tw-rotate-z{syntax:"*";inherits:false}@property --tw-skew-x{syntax:"*";inherits:false}@property --tw-skew-y{syntax:"*";inherits:false}@property --tw-border-style{syntax:"*";inherits:false;initial-value:solid}@property --tw-leading{syntax:"*";inherits:false}@property --tw-font-weight{syntax:"*";inherits:false}@property --tw-ordinal{syntax:"*";inherits:false}@property --tw-slashed-zero{syntax:"*";inherits:false}@property --tw-numeric-figure{syntax:"*";inherits:false}@property --tw-numeric-spacing{syntax:"*";inherits:false}@property --tw-numeric-fraction{syntax:"*";inherits:false}@property --tw-shadow{syntax:"*";inherits:false;initial-value:0 0 #0000}@property --tw-shadow-color{syntax:"*";inherits:false}@property --tw-shadow-alpha{syntax:"<percentage>";inherits:false;initial-value:100%}@property --tw-inset-shadow{syntax:"*";inherits:false;initial-value:0 0 #0000}@property --tw-inset-shadow-color{syntax:"*";inherits:false}@property --tw-inset-shadow-alpha{syntax:"<percentage>";inherits:false;initial-value:100%}@property --tw-ring-color{syntax:"*";inherits:false}@property --tw-ring-shadow{syntax:"*";inherits:false;initial-value:0 0 #0000}@property --tw-inset-ring-color{syntax:"*";inherits:false}@property --tw-inset-ring-shadow{syntax:"*";inherits:false;initial-value:0 0 #0000}@property --tw-ring-inset{syntax:"*";inherits:false}@property --tw-ring-offset-width{syntax:"<length>";inherits:false;initial-value:0}@property --tw-ring-offset-color{syntax:"*";inherits:false;initial-value:#fff}@property --tw-ring-offset-shadow{syntax:"*";inherits:false;initial-value:0 0 #0000}@property --tw-outline-style{syntax:"*";inherits:false;initial-value:solid}@property --tw-blur{syntax:"*";inherits:false}@property --tw-brightness{syntax:"*";inherits:false}@property --tw-contrast{syntax:"*";inherits:false}@property --tw-grayscale{syntax:"*";inherits:false}@property --tw-hue-rotate{syntax:"*";inherits:false}@property --tw-invert{syntax:"*";inherits:false}@property --tw-opacity{syntax:"*";inherits:false}@property --tw-saturate{syntax:"*";inherits:false}@property --tw-sepia{syntax:"*";inherits:false}@property --tw-drop-shadow{syntax:"*";inherits:false}@property --tw-drop-shadow-color{syntax:"*";inherits:false}@property --tw-drop-shadow-alpha{syntax:"<percentage>";inherits:false;initial-value:100%}@property --tw-drop-shadow-size{syntax:"*";inherits:false}@property --tw-backdrop-blur{syntax:"*";inherits:false}@property --tw-backdrop-brightness{syntax:"*";inherits:false}@property --tw-backdrop-contrast{syntax:"*";inherits:false}@property --tw-backdrop-grayscale{syntax:"*";inherits:false}@property --tw-backdrop-hue-rotate{syntax:"*";inherits:false}@property --tw-backdrop-invert{syntax:"*";inherits:false}@property --tw-backdrop-opacity{syntax:"*";inherits:false}@property --tw-backdrop-saturate{syntax:"*";inherits:false}@property --tw-backdrop-sepia{syntax:"*";inherits:false}@property --tw-duration{syntax:"*";inherits:false}@property --tw-ease{syntax:"*";inherits:false}@property --tw-contain-size{syntax:"*";inherits:false}@property --tw-contain-layout{syntax:"*";inherits:false}@property --tw-contain-paint{syntax:"*";inherits:false}@property --tw-contain-style{syntax:"*";inherits:false}`;
	Rn = Symbol(`store-raw`);
	zn = Symbol(`store-node`);
	Bn = Symbol(`store-has`);
	Vn = Symbol(`store-self`);
	Jn = {
		get(e, t, n) {
			if (t === Rn) return e;
			if (t === $e) return n;
			if (t === et) return Kn(e), n;
			let r = Wn(e, zn), i = r[t], a = i ? i() : e[t];
			if (t === zn || t === Bn || t === `__proto__`) return a;
			if (!i) {
				let n = Object.getOwnPropertyDescriptor(e, t);
				mt() && (typeof a != `function` || Object.prototype.hasOwnProperty.call(e, t)) && !(n && n.get) && (a = Gn(r, t, a)());
			}
			return Un(a) ? Hn(a) : a;
		},
		has(e, t) {
			return t === Rn || t === $e || t === et || t === zn || t === Bn || t === `__proto__` || (mt() && Gn(Wn(e, Bn), t)(), t in e);
		},
		set() {
			return !0;
		},
		deleteProperty() {
			return !0;
		},
		ownKeys: qn,
		getOwnPropertyDescriptor: X
	};
	tr = /* @__PURE__ */ new WeakMap();
	nr = {
		get(e, t) {
			if (t === Rn) return e;
			let n = e[t];
			if (t === $e || t === et || t === zn || t === Bn || t === `__proto__`) return n;
			let r;
			return Un(n) ? tr.get(n) || (tr.set(n, r = new Proxy(n, nr)), r) : n;
		},
		set(e, t, n) {
			return Yn(e, t, Y(n)), !0;
		},
		deleteProperty(e, t) {
			return Yn(e, t, void 0, !0), !0;
		}
	};
	ir = (e) => ({
		x: e.x + e.width / 2,
		y: e.y + e.height / 2
	});
	Z = (e) => {
		if (!e) return !1;
		let t = S(e);
		return t ? t.isConnected() : !!(e.isConnected || e.ownerDocument?.contains(e));
	};
	ar = (e) => {
		let t = 0, n = e.previousElementSibling;
		for (; n;) t += 1, n = n.previousElementSibling;
		return t;
	};
	or = /* @__PURE__ */ new WeakMap();
	sr = (e) => {
		if (e.anchorElement.isConnected) return e.anchorElement;
		let t = e.anchorElement.tagName, n = ze$1(e.anchorParentFiber);
		if (_$3(n)) {
			let r = n.stateNode;
			if (!v(r) || !r.isConnected) return null;
			let i = r.children[e.anchorIndexInParent];
			return i && i.tagName === t ? i : null;
		}
		for (let e of Ne$1(n)) {
			let n = e.stateNode;
			if (v(n) && n.isConnected && n.tagName === t) return n;
		}
		return null;
	};
	cr = (e, t) => {
		let n = e;
		for (let e of t) {
			let t = (e.isShadowChild ? n.shadowRoot?.children : n.children)?.[e.childIndex];
			if (!t) return null;
			n = t;
		}
		return n;
	};
	lr = (e) => {
		let t = [], n = e, r = Xe$3(n);
		for (; n && !r;) {
			let e = n.parentElement;
			if (e) t.unshift({
				childIndex: ar(n),
				isShadowChild: !1
			}), n = e;
			else {
				let e = n.getRootNode();
				if (!w(e)) break;
				t.unshift({
					childIndex: ar(n),
					isShadowChild: !0
				}), n = e.host;
			}
			r = Xe$3(n);
		}
		let i = r?.return;
		return !n || !i ? null : {
			anchorElement: n,
			anchorParentFiber: i,
			anchorIndexInParent: ar(n),
			domPath: t,
			targetTagName: e.tagName
		};
	};
	ur = (e, t) => e.anchorElement.isConnected && e.anchorIndexInParent === ar(e.anchorElement) && cr(e.anchorElement, e.domPath) === t;
	dr = (e) => {
		if (!e.isConnected) return;
		let t = or.get(e);
		if (t && ur(t, e)) return;
		let n = lr(e);
		n && or.set(e, n);
	};
	fr = (e) => {
		if (e.isConnected) return e;
		let t = or.get(e);
		if (!t) return null;
		let n = sr(t);
		if (!n) return null;
		let r = cr(n, t.domPath);
		return r && r.isConnected && r.tagName === t.targetTagName ? r : null;
	};
	pr = (e) => {
		let t = fr(e) ?? e;
		return dr(t), t;
	};
	mr = (e) => {
		for (let t = 0; t < e.frozenElements.length; t += 1) e.frozenElements[t] = pr(e.frozenElements[t]);
		e.frozenElement = e.frozenElement && pr(e.frozenElement), e.detectedElement = e.detectedElement && pr(e.detectedElement), e.contextMenuElement = e.contextMenuElement && pr(e.contextMenuElement);
		for (let t of e.labelInstances) if (t.element &&= pr(t.element), t.elements) for (let e = 0; e < t.elements.length; e += 1) t.elements[e] = pr(t.elements[e]);
		for (let t of e.grabbedBoxes) t.element &&= pr(t.element);
	};
	hr = (e) => ({
		selectionInteractionLockDepth: 0,
		wasActivatedByToggle: !1,
		pendingCommentMode: !1,
		keyHoldDuration: e.keyHoldDuration,
		dragStart: {
			x: ie,
			y: ie
		},
		copyStart: {
			x: ie,
			y: ie
		},
		copyOffsetFromCenterX: 0,
		detectedElement: null,
		frozenElement: null,
		frozenElements: [],
		frozenDragRect: null,
		lastGrabbedElement: null,
		selectionFilePath: null,
		selectionLineNumber: null,
		inputText: ``,
		grabbedBoxes: [],
		labelInstances: [],
		isTouchMode: !1,
		activationTimestamp: null,
		previouslyFocusedElement: null,
		contextMenuPosition: null,
		contextMenuElement: null,
		contextMenuClickOffset: null
	});
	gr = (e) => {
		let [t, n] = er(hr(e)), [r, i] = z({
			x: ie,
			y: ie
		}), [a, o] = z(0), [s, c] = z({ state: `idle` }), l = (e) => {
			n(rr((t) => {
				e(t), t.frozenElement = t.frozenElements.length > 0 ? t.frozenElements[0] : null, t.frozenDragRect = null;
				for (let e of t.frozenElements) dr(e);
			}));
		}, u = () => {
			l((e) => {
				e.frozenElements = [];
			});
		}, d = (e) => {
			c((t) => t.state === `active` ? {
				...t,
				phase: e
			} : t);
		}, f = {
			startHold: (e) => {
				e !== void 0 && n(`keyHoldDuration`, e), c({
					state: `holding`,
					startedAt: Date.now()
				});
			},
			releaseHold: () => {
				s().state === `holding` && c({ state: `idle` });
			},
			activate: () => {
				H(() => {
					c({
						state: `active`,
						phase: `hovering`,
						isPromptMode: !1,
						isPendingDismiss: !1
					}), n(`activationTimestamp`, Date.now()), n(`previouslyFocusedElement`, document.activeElement);
				});
			},
			deactivate: () => {
				H(() => {
					c({ state: `idle` }), n(rr((e) => {
						e.wasActivatedByToggle = !1, e.pendingCommentMode = !1, e.inputText = ``, e.frozenElement = null, e.frozenElements = [], e.frozenDragRect = null, e.activationTimestamp = null, e.previouslyFocusedElement = null, e.contextMenuPosition = null, e.contextMenuElement = null, e.contextMenuClickOffset = null, e.isTouchMode && (e.detectedElement = null);
					}));
				});
			},
			toggle: () => {
				t.activationTimestamp === null ? (n(`wasActivatedByToggle`, !0), f.activate()) : f.deactivate();
			},
			freeze: () => {
				if (s().state === `active`) {
					let e = t.frozenElement ?? t.detectedElement;
					e && (dr(e), n(`frozenElement`, e)), d(`frozen`);
				}
			},
			unfreeze: () => {
				s().state === `active` && H(() => {
					n(rr((e) => {
						e.frozenElement = null, e.frozenElements = [], e.frozenDragRect = null;
					})), d(`hovering`);
				});
			},
			startDrag: (e, t) => {
				s().state === `active` && H(() => {
					t || u(), n(`dragStart`, {
						x: e.x + window.scrollX,
						y: e.y + window.scrollY
					}), d(`dragging-select`);
				});
			},
			startDragReposition: () => {
				let e = s();
				e.state === `active` && e.phase === `dragging-select` && d(`dragging-reposition`);
			},
			stopDragReposition: () => {
				let e = s();
				e.state === `active` && e.phase === `dragging-reposition` && d(`dragging-select`);
			},
			shiftDragStart: (e) => {
				let t = s();
				t.state === `active` && t.phase === `dragging-reposition` && n(`dragStart`, (t) => ({
					x: t.x + e.x,
					y: t.y + e.y
				}));
			},
			endDrag: () => {
				let e = s();
				e.state === `active` && (e.phase === `dragging-select` || e.phase === `dragging-reposition`) && H(() => {
					n(`dragStart`, {
						x: -1e3,
						y: -1e3
					}), d(`justDragged`);
				});
			},
			cancelDrag: () => {
				let e = s();
				e.state === `active` && (e.phase === `dragging-select` || e.phase === `dragging-reposition`) && H(() => {
					n(`dragStart`, {
						x: -1e3,
						y: -1e3
					}), d(`hovering`);
				});
			},
			finishJustDragged: () => {
				let e = s();
				e.state === `active` && e.phase === `justDragged` && d(`hovering`);
			},
			startCopy: () => {
				let e = s().state === `active`;
				c({
					state: `copying`,
					startedAt: Date.now(),
					wasActive: e
				});
			},
			completeCopy: () => {
				let e = s(), t = e.state === `copying` && e.wasActive;
				c({
					state: `justCopied`,
					copiedAt: Date.now(),
					wasActive: t
				});
			},
			finishJustCopied: () => {
				let e = s();
				e.state === `justCopied` && (e.wasActive && !t.wasActivatedByToggle ? H(() => {
					u(), c({
						state: `active`,
						phase: `hovering`,
						isPromptMode: !1,
						isPendingDismiss: !1
					});
				}) : f.deactivate());
			},
			enterPromptMode: (e, t) => {
				let r = _t$2(t), { x: a } = ir(r);
				dr(t), H(() => {
					n(`copyStart`, e), n(`copyOffsetFromCenterX`, e.x - a), i(e), n(`frozenElement`, t), n(`wasActivatedByToggle`, !0), s().state === `active` ? c((e) => e.state === `active` ? {
						...e,
						isPromptMode: !0,
						phase: `frozen`
					} : e) : (c({
						state: `active`,
						phase: `frozen`,
						isPromptMode: !0,
						isPendingDismiss: !1
					}), n(`activationTimestamp`, Date.now()), n(`previouslyFocusedElement`, document.activeElement));
				});
			},
			exitPromptMode: () => {
				c((e) => e.state === `active` ? {
					...e,
					isPromptMode: !1,
					isPendingDismiss: !1
				} : e);
			},
			setInputText: (e) => {
				n(`inputText`, e);
			},
			clearInputText: () => {
				n(`inputText`, ``);
			},
			setPendingDismiss: (e) => {
				c((t) => t.state === `active` ? {
					...t,
					isPendingDismiss: e
				} : t);
			},
			setPointer: (e) => {
				i((t) => t.x === e.x && t.y === e.y ? t : e);
			},
			setDetectedElement: (e) => {
				e && dr(e), n(`detectedElement`, e);
			},
			setFrozenElement: (e) => {
				l((t) => {
					t.frozenElements = [e];
				});
			},
			setFrozenElements: (e) => {
				l((t) => {
					t.frozenElements = e;
				});
			},
			toggleFrozenElement: (e) => {
				l((t) => {
					let n = t.frozenElements.indexOf(e);
					n >= 0 ? t.frozenElements.splice(n, 1) : t.frozenElements.push(e);
				});
			},
			addFrozenElements: (e) => {
				l((t) => {
					for (let n of e) t.frozenElements.includes(n) || t.frozenElements.push(n);
				});
			},
			setFrozenDragRect: (e) => {
				n(`frozenDragRect`, e);
			},
			relinkLiveElements: () => {
				n(rr(mr));
			},
			setCopyStart: (e, t) => {
				let r = _t$2(t), { x: i } = ir(r);
				n(`copyStart`, e), n(`copyOffsetFromCenterX`, e.x - i);
			},
			setLastGrabbed: (e) => {
				n(`lastGrabbedElement`, e);
			},
			setWasActivatedByToggle: (e) => {
				n(`wasActivatedByToggle`, e);
			},
			setPendingCommentMode: (e) => {
				n(`pendingCommentMode`, e);
			},
			setTouchMode: (e) => {
				n(`isTouchMode`, e);
			},
			incrementSelectionInteractionLockDepth: () => {
				n(`selectionInteractionLockDepth`, (e) => e + 1);
			},
			decrementSelectionInteractionLockDepth: () => {
				n(`selectionInteractionLockDepth`, (e) => Math.max(0, e - 1));
			},
			setSelectionSource: (e, t) => {
				H(() => {
					n(`selectionFilePath`, e), n(`selectionLineNumber`, t);
				});
			},
			incrementViewportVersion: () => {
				o((e) => e + 1);
			},
			addGrabbedBox: (e) => {
				e.element && dr(e.element), n(`grabbedBoxes`, (t) => [...t, e]);
			},
			removeGrabbedBox: (e) => {
				n(`grabbedBoxes`, (t) => t.filter((t) => t.id !== e));
			},
			clearGrabbedBoxes: () => {
				n(`grabbedBoxes`, []);
			},
			addLabelInstance: (e) => {
				e.element && dr(e.element);
				for (let t of e.elements ?? []) dr(t);
				n(`labelInstances`, (t) => [...t, e]);
			},
			updateLabelInstance: (e, r, i) => {
				let a = t.labelInstances.findIndex((t) => t.id === e);
				a !== -1 && H(() => {
					n(`labelInstances`, a, `status`, r), i === void 0 ? r !== `error` && n(`labelInstances`, a, `errorMessage`, void 0) : n(`labelInstances`, a, `errorMessage`, i);
				});
			},
			removeLabelInstance: (e) => {
				n(`labelInstances`, (t) => t.filter((t) => t.id !== e));
			},
			clearLabelInstances: () => {
				n(`labelInstances`, []);
			},
			showContextMenu: (e, t) => {
				let r = _t$2(t), { x: i, y: a } = ir(r);
				dr(t), H(() => {
					n(`contextMenuPosition`, e), n(`contextMenuElement`, t), n(`contextMenuClickOffset`, {
						x: e.x - i,
						y: e.y - a
					});
				});
			},
			hideContextMenu: () => {
				H(() => {
					n(`contextMenuPosition`, null), n(`contextMenuElement`, null), n(`contextMenuClickOffset`, null);
				});
			},
			updateContextMenuPosition: () => {
				let e = t.contextMenuElement, r = t.contextMenuClickOffset;
				if (!e || !r || !Z(e)) return;
				let i = _t$2(e), { x: a, y: o } = ir(i);
				n(`contextMenuPosition`, {
					x: a + r.x,
					y: o + r.y
				});
			}
		};
		return {
			store: t,
			actions: f,
			pointer: r,
			viewportVersion: a,
			current: s
		};
	};
	_r = [
		`input`,
		`textarea`,
		`select`,
		`searchbox`,
		`slider`,
		`spinbutton`,
		`menuitem`,
		`menuitemcheckbox`,
		`menuitemradio`,
		`option`,
		`radio`,
		`textbox`,
		`combobox`
	];
	vr = (e) => {
		if (e.composed) {
			let t = e.composedPath()[0];
			if (t instanceof HTMLElement) return t;
		} else if (e.target instanceof HTMLElement) return e.target;
	};
	yr = (e) => {
		if (document.designMode === `on`) return !0;
		let t = vr(e);
		if (!t) return !1;
		if (t.isContentEditable) return !0;
		let n = D(t);
		return _r.some((e) => e === n || e === t.role);
	};
	br = (e) => {
		let t = e.target;
		if (t instanceof HTMLInputElement || t instanceof HTMLTextAreaElement) {
			let e = t.selectionStart ?? 0;
			return (t.selectionEnd ?? 0) - e > 0;
		}
		return !1;
	};
	xr = () => {
		let e = window.getSelection();
		return e ? e.toString().length > 0 : !1;
	};
	Sr = `data-react-grab`;
	Q = (e) => {
		if (!document.body) return;
		let t = document.querySelectorAll(`[${Sr}]`);
		for (let n of t) n !== e && n.parentNode !== e && (n.shadowRoot || n.remove());
		document.body.appendChild(e);
	};
	Cr = (e) => {
		if (document.body) return Q(e), () => {};
		let t = () => {
			document.removeEventListener(`DOMContentLoaded`, t), Q(e);
		};
		return document.addEventListener(`DOMContentLoaded`, t, { once: !0 }), () => document.removeEventListener(`DOMContentLoaded`, t);
	};
	wr = (e) => {
		let t = window.setTimeout(() => {
			Q(e);
		}, ze), n = new MutationObserver(() => {
			e.parentNode !== document.body && Q(e);
		});
		return n.observe(document.documentElement, { childList: !0 }), () => {
			window.clearTimeout(t), n.disconnect();
		};
	};
	Tr = (e) => {
		let t = document.querySelectorAll(`[${Sr}]`);
		for (let e of t) {
			let t = e.shadowRoot?.querySelector(`[${Sr}]`);
			if (t instanceof HTMLDivElement) return {
				root: t,
				host: e,
				cancelPendingAttachment: wr(e)
			};
			e.remove();
		}
		let n = document.createElement(`div`);
		n.setAttribute(Sr, `true`), Mt$2(n), n.style.zIndex = String(ce), n.style.position = `fixed`, n.style.inset = `0`, n.style.pointerEvents = `none`, n.style.contain = `strict`;
		let r = n.attachShadow({ mode: `open` }), i = document.createElement(`style`), a = At$2();
		a && (i.nonce = a), i.textContent = `@import url("https://fonts.googleapis.com/css2?family=Geist:wght@500&display=swap");\n${e ?? ``}`, r.appendChild(i);
		let o = document.createElement(`div`);
		o.setAttribute(Sr, `true`), r.appendChild(o);
		let s = Cr(n), c = wr(n);
		return {
			root: o,
			host: n,
			cancelPendingAttachment: () => {
				s(), c();
			}
		};
	};
	Er = (e) => {
		let [t, n] = z(void 0), r = 0;
		return B(ft(e, (e) => {
			let t = ++r;
			if (!e) {
				n(void 0);
				return;
			}
			let i = Wr$2(e) ?? void 0;
			n(i), Lr$2(e).then((e) => {
				r === t && n(e ?? i);
			}).catch(() => {
				r === t && n(i);
			});
		})), [t, n];
	};
	Dr = /^rgba?\(\s*([\d.]+)[,\s]+([\d.]+)[,\s]+([\d.]+)(?:\s*[,/]\s*([\d.]+%?))?\s*\)$/;
	Or = (e) => Math.max(0, Math.min(255, Math.round(e))).toString(16).padStart(2, `0`);
	kr = (e, t, n, r = 1) => {
		let i = Or(e), a = Or(t), o = Or(n);
		return r >= 1 ? `#${i}${a}${o}` : `#${i}${a}${o}${Or(r * 255)}`;
	};
	Ar = (e) => {
		if (!e.startsWith(`#`)) return null;
		let t = e.slice(1);
		if (t.length !== 6 && t.length !== 8) return null;
		let n = parseInt(t.slice(0, 2), 16), r = parseInt(t.slice(2, 4), 16), i = parseInt(t.slice(4, 6), 16), a = t.length === 8 ? parseInt(t.slice(6, 8), 16) / 255 : 1;
		return !Number.isFinite(n) || !Number.isFinite(r) || !Number.isFinite(i) ? null : {
			red: n,
			green: r,
			blue: i,
			alpha: a
		};
	};
	jr = (e) => {
		let t = e.match(Dr);
		if (!t) return null;
		let n = 1;
		if (t[4] !== void 0) {
			let e = t[4];
			n = e.endsWith(`%`) ? Number(e.slice(0, -1)) / 100 : Number(e);
		}
		return kr(Number(t[1]), Number(t[2]), Number(t[3]), n);
	};
	Mr = (e) => {
		let t = e.match(Dr);
		return !t || t[4] === void 0 ? !1 : (t[4].endsWith(`%`) ? Number(t[4].slice(0, -1)) / 100 : Number(t[4])) === 0;
	};
	Nr = /^#?([0-9a-fA-F]{3}|[0-9a-fA-F]{4}|[0-9a-fA-F]{6}|[0-9a-fA-F]{8})$/;
	Pr = (e) => {
		let t = e.match(Nr);
		if (!t) return null;
		let n = t[1];
		return n.length === 3 ? `#${n[0]}${n[0]}${n[1]}${n[1]}${n[2]}${n[2]}` : n.length === 4 ? `#${n[0]}${n[0]}${n[1]}${n[1]}${n[2]}${n[2]}${n[3]}${n[3]}` : `#${n}`;
	};
	Fr = `rgba(0, 0, 0, 0)`;
	Ir = /^oklch\(\s*([+-]?\d*\.?\d+%?)\s+([+-]?\d*\.?\d+%?)\s+([+-]?\d*\.?\d+)(deg|grad|rad|turn)?(?:\s*\/\s*([+-]?\d*\.?\d+%?))?\s*\)$/i;
	Lr = 180 / Math.PI;
	Rr = (e, t) => e.endsWith(`%`) ? Number(e.slice(0, -1)) / 100 * t : Number(e);
	zr = (e, t) => {
		let n = Number(e);
		return t === `grad` ? n * .9 : t === `rad` ? n * Lr : t === `turn` ? n * 360 : n;
	};
	Br = (e) => {
		let t = Math.max(0, Math.min(1, e));
		return (t <= .0031308 ? t * 12.92 : 1.055 * t ** (1 / 2.4) - .055) * 255;
	};
	Vr = (e) => {
		let t = e.match(Ir);
		if (!t) return null;
		let n = Rr(t[1], 1), r = Rr(t[2], .4), i = zr(t[3], t[4]) * Math.PI / 180, a = t[5] === void 0 ? 1 : Rr(t[5], 1);
		if (![
			n,
			r,
			i,
			a
		].every(Number.isFinite)) return null;
		let o = r * Math.cos(i), s = r * Math.sin(i), c = n + .3963377774 * o + .2158037573 * s, l = n - .1055613458 * o - .0638541728 * s, u = n - .0894841775 * o - 1.291485548 * s, d = c ** 3, f = l ** 3, p = u ** 3, m = 4.0767416621 * d - 3.3077115913 * f + .2309699292 * p, h = -1.2684380046 * d + 2.6097574011 * f - .3413193965 * p, g = -.0041960863 * d - .7034186147 * f + 1.707614701 * p;
		return kr(Br(m), Br(h), Br(g), Math.max(0, Math.min(1, a)));
	};
	Hr = null;
	Ur = () => Hr || (typeof document > `u` ? null : (Hr = document.createElement(`canvas`).getContext(`2d`, { willReadFrequently: !0 }), Hr));
	Wr = (e, t) => {
		e.clearRect(0, 0, 1, 1), e.fillStyle = t, e.fillRect(0, 0, 1, 1);
		let n = e.getImageData(0, 0, 1, 1).data;
		return n[3] === 0 ? Ue : kr(n[0], n[1], n[2], n[3] / 255);
	};
	Gr = (e) => {
		let t = e.trim();
		if (!t) return null;
		if (t.toLowerCase() === `transparent`) return Ue;
		let n = Pr(t);
		if (n) return n;
		let r = Vr(t);
		if (r) return r;
		let i = Ur();
		if (!i) return null;
		i.fillStyle = Fr, i.fillStyle = t;
		let a = i.fillStyle;
		return typeof a == `string` ? a.toLowerCase() === Fr ? t.toLowerCase().replace(/\s+/g, ``) === `rgba(0,0,0,0)` ? Ue : null : a.startsWith(`#`) && a.length === 7 ? a : a.startsWith(`rgb`) ? jr(a) : Wr(i, t) : null;
	};
	Kr = [
		`data-theme`,
		`data-mode`,
		`data-color-scheme`,
		`data-bs-theme`,
		`data-mui-color-scheme`,
		`data-mantine-color-scheme`
	];
	qr = [{
		attribute: `data-dark`,
		theme: `dark`
	}, {
		attribute: `data-light`,
		theme: `light`
	}];
	Jr = (e, t, n) => {
		let [r, i, a] = [
			e,
			t,
			n
		].map((e) => {
			let t = e / 255;
			return t <= .03928 ? t / 12.92 : ((t + .055) / 1.055) ** 2.4;
		});
		return .2126 * r + .7152 * i + .0722 * a;
	};
	Yr = (e, t = 0) => {
		let n = Gr(e), r = n ? Ar(n) : null;
		return !r || r.alpha <= t ? null : Jr(r.red, r.green, r.blue);
	};
	Xr = (e) => {
		let t = Yr(e);
		return t === null ? null : t < .18 ? `dark` : `light`;
	};
	Zr = (e) => Xr(getComputedStyle(e).backgroundColor);
	Qr = () => {
		let e = Yr(getComputedStyle(document.documentElement).color, Ae);
		return e === null ? null : e > .6 ? `dark` : null;
	};
	$r = () => {
		let e = document.createElement(`div`);
		e.style.cssText = `position:fixed;width:0;height:0;background-color:Canvas!important`, document.documentElement.appendChild(e);
		let t = Xr(getComputedStyle(e).backgroundColor);
		return e.remove(), t;
	};
	ei = (e) => {
		let t = e.toLowerCase();
		return t === `dark` ? `dark` : t === `light` ? `light` : null;
	};
	ti = (e) => {
		let t = e.style.colorScheme || getComputedStyle(e).colorScheme;
		return t ? t.trim().toLowerCase().split(/\s+/) : [];
	};
	ni = (e) => {
		let t = ti(e), n = t.includes(`dark`);
		return n === t.includes(`light`) ? null : n ? `dark` : `light`;
	};
	ri = (e) => {
		if (e.classList.contains(`dark`)) return `dark`;
		if (e.classList.contains(`light`)) return `light`;
		for (let t of Kr) {
			let n = e.getAttribute(t);
			if (!n) continue;
			let r = ei(n);
			if (r) return r;
		}
		for (let { attribute: t, theme: n } of qr) if (e.hasAttribute(t)) return n;
		return null;
	};
	ii = (e, t) => {
		for (let n of e) {
			if (!n) continue;
			let e = t(n);
			if (e) return e;
		}
		return null;
	};
	ai = () => {
		let e = [document.documentElement, document.body], t = [document.body, document.documentElement];
		return ii(e, ri) ?? ii(e, ni) ?? ii(t, Zr) ?? Qr() ?? $r() ?? `light`;
	};
	oi = (e) => e === `dark` ? `light` : `dark`;
	si = () => {
		let e = [];
		for (let t of [document.documentElement, document.body]) {
			if (!t) continue;
			e.push(t.className);
			for (let n of Kr) e.push(t.getAttribute(n) ?? ``);
			for (let { attribute: n } of qr) e.push(t.hasAttribute(n) ? `1` : ``);
			let n = t.style;
			e.push(n.colorScheme, n.backgroundColor, n.color);
			for (let t = 0; t < n.length; t++) {
				let r = n[t];
				r.startsWith(`--`) && e.push(`${r}:${n.getPropertyValue(r)}`);
			}
		}
		return e.join(`|`);
	};
	ci = {
		attributes: !0,
		attributeFilter: [
			`class`,
			`style`,
			...Kr,
			...qr.map(({ attribute: e }) => e)
		]
	};
	li = (e, t) => {
		let n = null, r = null, i = null, a = (i) => {
			if (i === r && n !== null) return getComputedStyle(document.body ?? document.documentElement).backgroundColor, n;
			r = i;
			let a = ai(), o = oi(a);
			return o !== n && (n = o, e.setAttribute(`data-rg-theme`, o), t?.(o)), o;
		}, s = () => a(si()), c = () => {
			i !== null && (ci$2(i), i = null);
		}, l = () => {
			i === null && (i = M$1(() => {
				i = null, s();
			}));
		}, u = s(), d = () => {
			let e = si();
			e === r ? l() : (c(), a(e));
		}, f = new MutationObserver(d);
		f.observe(document.documentElement, ci);
		let p = null;
		document.body ? f.observe(document.body, ci) : (p = new MutationObserver((e, t) => {
			document.body && (t.disconnect(), p = null, f.observe(document.body, ci), d());
		}), p.observe(document.documentElement, { childList: !0 }));
		let m = () => {
			r = null, l();
		}, h = window.matchMedia(`(prefers-color-scheme: dark)`);
		return h.addEventListener(`change`, m), {
			theme: u,
			cleanup: () => {
				p?.disconnect(), f.disconnect(), h.removeEventListener(`change`, m), c();
			}
		};
	};
	ui = () => {};
	di = () => ({
		activate: ui,
		deactivate: ui,
		toggle: ui,
		comment: ui,
		isActive: () => !1,
		isEnabled: () => !1,
		setEnabled: ui,
		getToolbarState: () => null,
		setToolbarState: ui,
		onToolbarStateChange: () => ui,
		reset: ui,
		dispose: ui,
		copyElement: () => Promise.resolve(!1),
		getSource: () => Promise.resolve(null),
		getStackContext: () => Promise.resolve(``),
		getState: () => ({
			isActive: !1,
			isDragging: !1,
			isCopying: !1,
			isPromptMode: !1,
			isSelectionBoxVisible: !1,
			isDragBoxVisible: !1,
			targetElement: null,
			dragBounds: null,
			grabbedBoxes: [],
			labelInstances: [],
			selectionFilePath: null,
			toolbarState: null
		}),
		setOptions: ui,
		registerPlugin: ui,
		unregisterPlugin: ui,
		getPlugins: () => [],
		getDisplayName: () => null
	});
	fi = () => {
		let e = new AbortController();
		return {
			signal: e.signal,
			abort: () => e.abort(),
			addWindowListener: (t, n, r = {}) => {
				window.addEventListener(t, n, {
					...r,
					signal: e.signal
				});
			},
			addDocumentListener: (t, n, r = {}) => {
				document.addEventListener(t, n, {
					...r,
					signal: e.signal
				});
			}
		};
	};
	pi = (e, t = `Unknown error`) => e instanceof Error && e.message ? e.message : t;
	mi = (e) => e instanceof Error ? e : Error(String(e));
	hi = Symbol(`aborted-promise-result`);
	gi = async (e, t) => {
		if (!t) return e;
		if (t.aborted) return hi;
		let n = () => {}, r = new Promise((e) => {
			n = () => e(hi);
		});
		t.addEventListener(`abort`, n, { once: !0 });
		try {
			return await Promise.race([e, r]);
		} finally {
			t.removeEventListener(`abort`, n);
		}
	};
	_i = { status: `cancelled` };
	vi = (e) => ({
		functionName: e.functionName,
		fileName: e.fileName,
		lineNumber: e.lineNumber,
		columnNumber: e.columnNumber,
		isServer: e.isServer,
		isSymbolicated: e.isSymbolicated
	});
	yi = async (e, t) => {
		let n = { maxLines: t }, [r, i, a, o] = await Promise.all([
			ri$2(e, n),
			ti$2(e, n),
			Ur$2(e),
			Ir$2(e)
		]);
		return {
			tagName: D(e),
			componentName: a?.componentName ?? void 0,
			content: `[${r}]`,
			source: a,
			stackContext: i,
			frames: (o ?? []).map(vi)
		};
	};
	bi = async (e, t) => {
		let n = await Promise.all(e.map((e) => yi(e, t))), r = /* @__PURE__ */ new Map();
		for (let e of n) r.has(e.content) || r.set(e.content, e);
		let i = [...r.values()];
		return i.length > 0 ? {
			content: i.map((e) => e.content).join(`
`),
			entries: i
		} : null;
	};
	xi = (e, t, n) => {
		if (e?.entries) return t === e.content ? e.entries : e.entries.length === 1 ? [{
			...e.entries[0],
			content: t,
			commentText: n
		}] : e.entries.map((e) => ({
			...e,
			commentText: n
		}));
	};
	Si = async (e, t, n, r) => {
		if (e.signal?.aborted || await gi(t.onBeforeCopy(n), e.signal) === hi) return _i;
		let i = !1, a = !1, o = ``;
		try {
			let c = e.getContent ? Promise.resolve(e.getContent(n)).then((e) => ({ content: e })) : bi(n, e.maxContextLines), l = await gi(c, e.signal);
			if (l === hi) return _i;
			let u = l?.content;
			if (u?.trim()) {
				let c = await gi(t.transformCopyContent(u, n), e.signal);
				if (c === hi) return _i;
				o = r ? `${r}\n${c}` : c, o.trim() && (a = !0, i = di$2(o, {
					componentName: e.componentName,
					entries: xi(l, o, r)
				}));
			}
		} catch (n) {
			if (!a && e.signal?.aborted) return _i;
			t.onCopyError(mi(n));
		}
		return !a && e.signal?.aborted ? _i : (i && t.onCopySuccess(n, o), t.onAfterCopy(n, i), { status: i ? `succeeded` : `failed` });
	};
	Ci = (e, t, n) => Math.max(t, Math.min(e, n));
	wi = (e) => {
		let t = [], n = e;
		for (; n;) {
			t.unshift(n);
			let e = n.getRootNode();
			if (w(e)) n = e.host;
			else {
				if (!Nt$2(e)) break;
				n = ot$2(e.defaultView);
			}
		}
		return t;
	};
	Ti = (e, t) => {
		if (e === t) return 0;
		let n = wi(e), r = wi(t), i = Math.min(n.length, r.length);
		for (let e = 0; e < i; e += 1) {
			let t = n[e], i = r[e];
			if (t === i) continue;
			let a = t.compareDocumentPosition(i);
			return a & Node.DOCUMENT_POSITION_FOLLOWING ? -1 : a & Node.DOCUMENT_POSITION_PRECEDING ? 1 : 0;
		}
		return n.length - r.length;
	};
	Ei = (e, t) => {
		let n = Math.max(e.left, t.left), r = Math.max(e.top, t.top), i = Math.min(e.right, t.right), a = Math.min(e.bottom, t.bottom);
		return Math.max(0, i - n) * Math.max(0, a - r);
	};
	Di = (e, t) => e.left < t.right && e.right > t.left && e.top < t.bottom && e.bottom > t.top;
	Oi = (e) => e.sort(Ti);
	ki = (e) => {
		if (e.width <= 0 || e.height <= 0) return [];
		let t = window.innerWidth, n = window.innerHeight, r = e.x, i = e.y, a = e.x + e.width, o = e.y + e.height, s = r + e.width / 2, c = i + e.height / 2, l = Ci(Math.ceil(e.width / 32), 3, 20), u = Ci(Math.ceil(e.height / 32), 3, 20), d = l * u, f = d > 100 ? Math.sqrt(100 / d) : 1, p = Ci(Math.floor(l * f), 3, 20), m = Ci(Math.floor(u * f), 3, 20), h = /* @__PURE__ */ new Set(), g = [], _ = (e, r) => {
			let i = Ci(Math.round(e), 0, t - 1), a = Ci(Math.round(r), 0, n - 1), o = `${i}:${a}`;
			h.has(o) || (h.add(o), g.push({
				x: i,
				y: a
			}));
		};
		_(r + 1, i + 1), _(a - 1, i + 1), _(r + 1, o - 1), _(a - 1, o - 1), _(s, i + 1), _(s, o - 1), _(r + 1, c), _(a - 1, c), _(s, c);
		for (let t = 0; t < p; t += 1) {
			let n = r + (t + .5) / p * e.width;
			for (let t = 0; t < m; t += 1) _(n, i + (t + .5) / m * e.height);
		}
		return g;
	};
	Ai = (e, t, n) => {
		let r = {
			left: e.x,
			top: e.y,
			right: e.x + e.width,
			bottom: e.y + e.height
		}, i = /* @__PURE__ */ new Set(), a = ki(e);
		Wi$2();
		try {
			for (let e of a) {
				let t = Ni$2(e.x, e.y);
				for (let e of t) i.add(e);
			}
		} finally {
			Gi$2();
		}
		let o = [];
		for (let e of i) {
			if (b(e) && N$1(e) || !n && bi$2(e) || !Rt$2(e) || !t(e)) continue;
			let i = _t$2(e);
			if (i.width <= 0 || i.height <= 0) continue;
			let a = {
				left: i.x,
				top: i.y,
				right: i.x + i.width,
				bottom: i.y + i.height
			};
			n ? Ei(r, a) / (i.width * i.height) >= .75 && o.push(e) : Di(a, r) && o.push(e);
		}
		return Oi(o);
	};
	ji = (e) => {
		let t = new Set(e), n = [];
		for (let r = e.length - 1; r >= 0; --r) {
			let i = e[r];
			if (!t.has(i)) continue;
			let a = i, o = Pt$2(a), s = !1;
			for (; o;) {
				let e = a.getRootNode();
				if (t.has(o) && w(e) && e.host === o) t.delete(o);
				else if (t.has(o)) {
					s = !0;
					break;
				}
				a = o, o = Pt$2(a);
			}
			s || n.push(i);
		}
		return n.reverse();
	};
	Mi = (e, t, n = !0) => {
		let r = Ai(e, t, n);
		return ji(r);
	};
	Ni = (e, t) => e.width <= 0 ? 0 : Ci((t.x - e.x) / e.width, 0, 1);
	Pi = () => {
		ht$2(), Da$1();
	};
	Fi = (e) => ({
		borderRadius: `0px`,
		height: e.height,
		width: e.width,
		x: e.pageX - window.scrollX,
		y: e.pageY - window.scrollY
	});
	Ii = (e) => ({
		pageX: e.x + window.scrollX,
		pageY: e.y + window.scrollY,
		width: e.width,
		height: e.height
	});
	Li = (e) => ({
		borderRadius: `0px`,
		height: e.height,
		width: e.width,
		x: e.x,
		y: e.y
	});
	Ri = (e, t, n) => {
		let r = (e, t, n, r) => {
			let i = [], a = e, o = 0;
			for (; a && i.length < n && o < 100;) r(a) && i.push(a), a = t(a), o += 1;
			return i;
		}, i = r(Pt$2(e), Pt$2, 6, t);
		i.reverse();
		let a = i.map((e, t) => ({
			element: e,
			depth: t,
			isLast: !0
		})), o = i.length, s = r(e.previousElementSibling, (e) => e.previousElementSibling, 7, n), c = r(e.nextElementSibling, (e) => e.nextElementSibling, 7, n), l = Math.min(s.length, Math.max(3, 7 - c.length)), u = Math.min(c.length, 7 - l), d = [
			...s.slice(0, l).reverse(),
			e,
			...c.slice(0, u)
		], f = r(e.firstElementChild, (e) => e.nextElementSibling, 6, t), p = d.length - 1, m = f.length - 1;
		return d.forEach((t, n) => {
			a.push({
				element: t,
				depth: o,
				isLast: n === p
			}), t === e && f.forEach((e, t) => {
				a.push({
					element: e,
					depth: o + 1,
					isLast: t === m
				});
			});
		}), a;
	};
	zi = (e) => e.namespaceURI === `http://www.w3.org/2000/svg` && e.tagName !== `svg`;
	Bi = (e) => {
		let t = e.getBoundingClientRect();
		return t.width >= 16 || t.height >= 16;
	};
	Vi = (e, t) => t(e) && !zi(e) && Bi(e);
	Hi = /* @__PURE__ */ new Set([
		`c`,
		`C`,
		`с`,
		`С`,
		`ȼ`,
		`Ȼ`,
		`ↄ`,
		`Ↄ`,
		`ᴄ`,
		`ᶜ`,
		`ⱼ`,
		`ⅽ`,
		`Ⅽ`,
		`ç`,
		`Ç`,
		`ć`,
		`Ć`,
		`č`,
		`Č`,
		`ĉ`,
		`Ĉ`,
		`ċ`,
		`Ċ`
	]);
	Ui = (e, t) => t === `KeyC` ? !0 : !e || e.length !== 1 ? !1 : Hi.has(e);
	Wi = null;
	Gi = () => {
		if (typeof navigator > `u` || !(`userAgentData` in navigator)) return null;
		let e = navigator.userAgentData;
		if (typeof e != `object` || !e || !(`platform` in e)) return null;
		let t = e.platform;
		return typeof t == `string` ? t : null;
	};
	Ki = () => {
		if (Wi === null) {
			if (typeof navigator > `u`) return Wi = !1, Wi;
			let e = navigator.platform ?? Gi() ?? navigator.userAgent;
			Wi = /Mac|iPhone|iPad|iPod/i.test(e);
		}
		return Wi;
	};
	qi = (e, t) => {
		if (!e || !t) return !1;
		let n = e.toLowerCase();
		return t === `Space` ? n === `space` || n === ` ` : t.startsWith(`Key`) ? t.slice(3).toLowerCase() === n : t.startsWith(`Digit`) ? t.slice(5) === n : !1;
	};
	Ji = {
		meta: `metaKey`,
		cmd: `metaKey`,
		command: `metaKey`,
		win: `metaKey`,
		windows: `metaKey`,
		ctrl: `ctrlKey`,
		control: `ctrlKey`,
		shift: `shiftKey`,
		alt: `altKey`,
		option: `altKey`,
		opt: `altKey`
	};
	Yi = (e) => {
		let t = e.split(`+`).map((e) => e.trim().toLowerCase()), n = {
			metaKey: !1,
			ctrlKey: !1,
			shiftKey: !1,
			altKey: !1,
			key: null
		};
		for (let e of t) {
			let t = Ji[e];
			t ? n[t] = !0 : n.key = e;
		}
		return n;
	};
	Xi = (e) => {
		if (typeof e == `function`) return (t) => {
			try {
				return e(t);
			} catch {
				return !1;
			}
		};
		let t = Yi(e), n = t.key;
		return (e) => {
			if (n === null) {
				let n = !t.metaKey || e.metaKey || e.key === `Meta`, r = !t.ctrlKey || e.ctrlKey || e.key === `Control`, i = !t.shiftKey || e.shiftKey || e.key === `Shift`, a = !t.altKey || e.altKey || e.key === `Alt`, o = n && r && i && a, s = [
					t.metaKey,
					t.ctrlKey,
					t.shiftKey,
					t.altKey
				].filter(Boolean).length, c = [
					e.metaKey || e.key === `Meta`,
					e.ctrlKey || e.key === `Control`,
					e.shiftKey || e.key === `Shift`,
					e.altKey || e.key === `Alt`
				].filter(Boolean).length;
				return o && c >= s;
			}
			let r = e.key?.toLowerCase() === n || qi(n, e.code), i = t.metaKey || t.ctrlKey || t.shiftKey || t.altKey ? (!t.metaKey || e.metaKey) && (!t.ctrlKey || e.ctrlKey) && (!t.shiftKey || e.shiftKey) && (!t.altKey || e.altKey) : !e.metaKey && !e.ctrlKey && !e.shiftKey && !e.altKey;
			return r && i;
		};
	};
	Zi = (e) => !e || typeof e == `function` ? {
		metaKey: Ki(),
		ctrlKey: !Ki(),
		shiftKey: !1,
		altKey: !1,
		key: null
	} : Yi(e);
	Qi = (e, t) => {
		if (t.activationKey) return Xi(t.activationKey)(e);
		let n = (Ki() ? e.metaKey : e.ctrlKey) && !e.shiftKey && !e.altKey;
		return !!(e.key && n && Ui(e.key, e.code));
	};
	$i = (e, t) => {
		try {
			return e.composedPath().some((e) => y(e) && e.hasAttribute(t));
		} catch {
			return !1;
		}
	};
	ea = (e, t, n) => {
		H$1(e instanceof yt$2 ? e : new yt$2(t, n, e));
	};
	ta = (e, t, n) => {
		try {
			if (n.onOpenFile(e, t)) return;
			ka$1(e, t, n.transformOpenFileUrl).catch((n) => {
				ea(n, e, t);
			});
		} catch (n) {
			ea(n, e, t);
		}
	};
	na = (e) => {
		if (e.length === 0) return {
			x: 0,
			y: 0,
			width: 0,
			height: 0
		};
		if (e.length === 1) return e[0];
		let t = 1 / 0, n = 1 / 0, r = -1 / 0, i = -1 / 0;
		for (let a of e) t = Math.min(t, a.x), n = Math.min(n, a.y), r = Math.max(r, a.x + a.width), i = Math.max(i, a.y + a.height);
		return {
			x: t,
			y: n,
			width: r - t,
			height: i - n
		};
	};
	ra = /* @__PURE__ */ `padding-top.padding-right.padding-bottom.padding-left.margin-top.margin-right.margin-bottom.margin-left.gap.row-gap.column-gap.font-size.line-height.letter-spacing.border-radius.border-top-left-radius.border-top-right-radius.border-bottom-right-radius.border-bottom-left-radius.border-width.border-top-width.border-right-width.border-bottom-width.border-left-width.opacity.width.height.min-width.min-height.max-width.max-height.top.right.bottom.left.z-index`.split(`.`);
	ia = /* @__PURE__ */ new Set([
		`letter-spacing`,
		`gap`,
		`row-gap`,
		`column-gap`
	]);
	aa = /* @__PURE__ */ new Set([`z-index`]);
	oa = [
		{
			key: `padding`,
			label: `padding`,
			longhands: [
				`padding-top`,
				`padding-right`,
				`padding-bottom`,
				`padding-left`
			]
		},
		{
			key: `padding-top,padding-bottom`,
			label: `padding-y`,
			longhands: [`padding-top`, `padding-bottom`]
		},
		{
			key: `padding-left,padding-right`,
			label: `padding-x`,
			longhands: [`padding-left`, `padding-right`]
		},
		{
			key: `padding-top`,
			label: `padding top`,
			longhands: [`padding-top`]
		},
		{
			key: `padding-right`,
			label: `padding right`,
			longhands: [`padding-right`]
		},
		{
			key: `padding-bottom`,
			label: `padding bottom`,
			longhands: [`padding-bottom`]
		},
		{
			key: `padding-left`,
			label: `padding left`,
			longhands: [`padding-left`]
		}
	];
	sa = [
		{
			key: `margin`,
			label: `margin`,
			longhands: [
				`margin-top`,
				`margin-right`,
				`margin-bottom`,
				`margin-left`
			]
		},
		{
			key: `margin-top,margin-bottom`,
			label: `margin-y`,
			longhands: [`margin-top`, `margin-bottom`]
		},
		{
			key: `margin-left,margin-right`,
			label: `margin-x`,
			longhands: [`margin-left`, `margin-right`]
		},
		{
			key: `margin-top`,
			label: `margin top`,
			longhands: [`margin-top`]
		},
		{
			key: `margin-right`,
			label: `margin right`,
			longhands: [`margin-right`]
		},
		{
			key: `margin-bottom`,
			label: `margin bottom`,
			longhands: [`margin-bottom`]
		},
		{
			key: `margin-left`,
			label: `margin left`,
			longhands: [`margin-left`]
		}
	];
	ca = [
		{
			key: `gap`,
			label: `gap`,
			longhands: [`row-gap`, `column-gap`]
		},
		{
			key: `row-gap`,
			label: `row gap`,
			longhands: [`row-gap`]
		},
		{
			key: `column-gap`,
			label: `column gap`,
			longhands: [`column-gap`]
		}
	];
	la = [
		{
			key: `border-radius`,
			label: `border radius`,
			longhands: [
				`border-top-left-radius`,
				`border-top-right-radius`,
				`border-bottom-right-radius`,
				`border-bottom-left-radius`
			]
		},
		{
			key: `border-top-left-radius,border-top-right-radius`,
			label: `top corners`,
			longhands: [`border-top-left-radius`, `border-top-right-radius`]
		},
		{
			key: `border-bottom-left-radius,border-bottom-right-radius`,
			label: `bottom corners`,
			longhands: [`border-bottom-left-radius`, `border-bottom-right-radius`]
		},
		{
			key: `border-top-left-radius,border-bottom-left-radius`,
			label: `left corners`,
			longhands: [`border-top-left-radius`, `border-bottom-left-radius`]
		},
		{
			key: `border-top-right-radius,border-bottom-right-radius`,
			label: `right corners`,
			longhands: [`border-top-right-radius`, `border-bottom-right-radius`]
		},
		{
			key: `border-top-left-radius`,
			label: `top left radius`,
			longhands: [`border-top-left-radius`]
		},
		{
			key: `border-top-right-radius`,
			label: `top right radius`,
			longhands: [`border-top-right-radius`]
		},
		{
			key: `border-bottom-right-radius`,
			label: `bottom right radius`,
			longhands: [`border-bottom-right-radius`]
		},
		{
			key: `border-bottom-left-radius`,
			label: `bottom left radius`,
			longhands: [`border-bottom-left-radius`]
		}
	];
	ua = [
		{
			key: `top,right,bottom,left`,
			label: `inset`,
			longhands: [
				`top`,
				`right`,
				`bottom`,
				`left`
			]
		},
		{
			key: `top,bottom`,
			label: `inset-y`,
			longhands: [`top`, `bottom`]
		},
		{
			key: `left,right`,
			label: `inset-x`,
			longhands: [`left`, `right`]
		},
		{
			key: `top`,
			label: `top`,
			longhands: [`top`]
		},
		{
			key: `right`,
			label: `right`,
			longhands: [`right`]
		},
		{
			key: `bottom`,
			label: `bottom`,
			longhands: [`bottom`]
		},
		{
			key: `left`,
			label: `left`,
			longhands: [`left`]
		}
	];
	da = new Set(ua.map((e) => e.key));
	fa = [
		oa,
		sa,
		ca,
		la,
		[
			{
				key: `border-width`,
				label: `border width`,
				longhands: [
					`border-top-width`,
					`border-right-width`,
					`border-bottom-width`,
					`border-left-width`
				]
			},
			{
				key: `border-top-width`,
				label: `border top width`,
				longhands: [`border-top-width`]
			},
			{
				key: `border-right-width`,
				label: `border right width`,
				longhands: [`border-right-width`]
			},
			{
				key: `border-bottom-width`,
				label: `border bottom width`,
				longhands: [`border-bottom-width`]
			},
			{
				key: `border-left-width`,
				label: `border left width`,
				longhands: [`border-left-width`]
			}
		],
		[
			{
				key: `width,height`,
				label: `size`,
				longhands: [`width`, `height`]
			},
			{
				key: `width`,
				label: `width`,
				longhands: [`width`]
			},
			{
				key: `height`,
				label: `height`,
				longhands: [`height`]
			}
		],
		ua
	];
	pa = [
		{
			key: `font-size`,
			label: `font size`
		},
		{
			key: `line-height`,
			label: `line height`
		},
		{
			key: `letter-spacing`,
			label: `letter spacing`
		},
		{
			key: `min-width`,
			label: `min width`
		},
		{
			key: `min-height`,
			label: `min height`
		},
		{
			key: `max-width`,
			label: `max width`
		},
		{
			key: `max-height`,
			label: `max height`
		},
		{
			key: `z-index`,
			label: `z-index`
		},
		{
			key: `opacity`,
			label: `opacity`
		}
	];
	ma = [
		{
			key: `color`,
			label: `text color`,
			alwaysShow: !0
		},
		{
			key: `background-color`,
			label: `background`,
			alwaysShow: !0
		},
		{
			key: `border-color`,
			label: `border color`
		},
		{
			key: `fill`,
			label: `fill`
		},
		{
			key: `stroke`,
			label: `stroke`
		}
	];
	ha = [
		{
			value: `ui-sans-serif, system-ui, sans-serif, "Apple Color Emoji", "Segoe UI Emoji", "Segoe UI Symbol", "Noto Color Emoji"`,
			label: `sans`
		},
		{
			value: `ui-serif, Georgia, Cambria, "Times New Roman", Times, serif`,
			label: `serif`
		},
		{
			value: `ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace`,
			label: `mono`
		},
		{
			value: `system-ui, sans-serif`,
			label: `system`
		},
		{
			value: `serif`,
			label: `serif`
		},
		{
			value: `monospace`,
			label: `mono`
		}
	];
	ga = (e) => {
		let t = e.trim();
		return ha.some((e) => e.value === t) ? ha : [{
			value: t,
			label: `current`
		}, ...ha];
	};
	_a = {
		key: `font-family`,
		label: `font family`
	};
	va = [
		{
			key: `display`,
			label: `display`,
			options: [
				{
					value: `block`,
					label: `block`
				},
				{
					value: `flex`,
					label: `flex`
				},
				{
					value: `inline-block`,
					label: `inline-block`
				},
				{
					value: `inline`,
					label: `inline`
				},
				{
					value: `grid`,
					label: `grid`
				},
				{
					value: `none`,
					label: `none`
				}
			]
		},
		{
			key: `text-align`,
			label: `text align`,
			options: [
				{
					value: `left`,
					label: `left`
				},
				{
					value: `center`,
					label: `center`
				},
				{
					value: `right`,
					label: `right`
				},
				{
					value: `justify`,
					label: `justify`
				}
			]
		},
		{
			key: `font-style`,
			label: `font style`,
			options: [{
				value: `normal`,
				label: `normal`
			}, {
				value: `italic`,
				label: `italic`
			}]
		},
		{
			key: `text-transform`,
			label: `text transform`,
			options: [
				{
					value: `none`,
					label: `none`
				},
				{
					value: `uppercase`,
					label: `uppercase`
				},
				{
					value: `lowercase`,
					label: `lowercase`
				},
				{
					value: `capitalize`,
					label: `capitalize`
				}
			]
		},
		{
			key: `white-space`,
			label: `white space`,
			options: [
				{
					value: `normal`,
					label: `normal`
				},
				{
					value: `nowrap`,
					label: `nowrap`
				},
				{
					value: `pre`,
					label: `pre`
				},
				{
					value: `pre-wrap`,
					label: `pre-wrap`
				},
				{
					value: `pre-line`,
					label: `pre-line`
				},
				{
					value: `break-spaces`,
					label: `break-spaces`
				}
			]
		},
		{
			key: `word-break`,
			label: `word break`,
			options: [
				{
					value: `normal`,
					label: `normal`
				},
				{
					value: `break-all`,
					label: `break-all`
				},
				{
					value: `keep-all`,
					label: `keep-all`
				}
			]
		},
		{
			key: `overflow-wrap`,
			label: `wrap`,
			options: [
				{
					value: `normal`,
					label: `normal`
				},
				{
					value: `break-word`,
					label: `break-word`
				},
				{
					value: `anywhere`,
					label: `anywhere`
				}
			]
		},
		{
			key: `font-variant-numeric`,
			label: `number style`,
			options: [
				{
					value: `normal`,
					label: `normal`
				},
				{
					value: `tabular-nums`,
					label: `tabular`
				},
				{
					value: `proportional-nums`,
					label: `proportional`
				},
				{
					value: `lining-nums`,
					label: `lining`
				},
				{
					value: `oldstyle-nums`,
					label: `oldstyle`
				},
				{
					value: `slashed-zero`,
					label: `slashed zero`
				}
			]
		},
		{
			key: `text-decoration-line`,
			label: `decoration`,
			options: [
				{
					value: `none`,
					label: `none`
				},
				{
					value: `underline`,
					label: `underline`
				},
				{
					value: `line-through`,
					label: `strike`
				}
			]
		},
		{
			key: `border-style`,
			label: `border style`,
			options: [
				{
					value: `none`,
					label: `none`
				},
				{
					value: `solid`,
					label: `solid`
				},
				{
					value: `dashed`,
					label: `dashed`
				},
				{
					value: `dotted`,
					label: `dotted`
				}
			]
		},
		{
			key: `visibility`,
			label: `visibility`,
			options: [{
				value: `visible`,
				label: `visible`
			}, {
				value: `hidden`,
				label: `hidden`
			}]
		},
		{
			key: `align-items`,
			label: `align items`,
			options: [
				{
					value: `stretch`,
					label: `stretch`
				},
				{
					value: `flex-start`,
					label: `start`
				},
				{
					value: `center`,
					label: `center`
				},
				{
					value: `flex-end`,
					label: `end`
				},
				{
					value: `baseline`,
					label: `baseline`
				}
			]
		},
		{
			key: `justify-content`,
			label: `justify content`,
			options: [
				{
					value: `flex-start`,
					label: `start`
				},
				{
					value: `center`,
					label: `center`
				},
				{
					value: `flex-end`,
					label: `end`
				},
				{
					value: `space-between`,
					label: `between`
				},
				{
					value: `space-around`,
					label: `around`
				},
				{
					value: `space-evenly`,
					label: `evenly`
				}
			]
		},
		{
			key: `font-weight`,
			label: `font weight`,
			options: [
				{
					value: `100`,
					label: `thin`
				},
				{
					value: `200`,
					label: `extra-light`
				},
				{
					value: `300`,
					label: `light`
				},
				{
					value: `400`,
					label: `normal`
				},
				{
					value: `500`,
					label: `medium`
				},
				{
					value: `600`,
					label: `semibold`
				},
				{
					value: `700`,
					label: `bold`
				},
				{
					value: `800`,
					label: `extra-bold`
				},
				{
					value: `900`,
					label: `black`
				}
			]
		}
	];
	ya = [
		...ra,
		`color`,
		`background-color`,
		`border-color`,
		`fill`,
		`stroke`,
		`font-family`,
		`display`,
		`text-align`,
		`font-style`,
		`text-transform`,
		`white-space`,
		`word-break`,
		`overflow-wrap`,
		`font-variant-numeric`,
		`text-decoration-line`,
		`border-style`,
		`visibility`,
		`align-items`,
		`justify-content`,
		`font-weight`
	];
	ba = (e) => {
		let t = getComputedStyle(e), n = {};
		for (let e of ra) n[e] = t.getPropertyValue(e);
		return n;
	};
	xa = (e) => {
		let t = {};
		for (let n of ya) t[n] = e.getPropertyValue(n);
		return t;
	};
	Sa = (e) => {
		try {
			let t = e.parentElement;
			if (!t) return null;
			let n = e.namespaceURI, r = e.tagName.toLowerCase(), i = n && n !== `http://www.w3.org/1999/xhtml` ? e.ownerDocument.createElementNS(n, r) : e.ownerDocument.createElement(r), a = `style` in i ? i.style : null;
			a instanceof CSSStyleDeclaration && a.setProperty(`display`, `none`, `important`), t.appendChild(i);
			try {
				return xa(getComputedStyle(i));
			} finally {
				i.remove();
			}
		} catch {
			return null;
		}
	};
	Ca = /* @__PURE__ */ new Set([
		`display`,
		`width`,
		`height`
	]);
	wa = (e, t, n) => e.kind === `color` ? !1 : e.cssProperties.some((e) => Ca.has(e)) ? Ta(e) : e.cssProperties.every((e) => {
		let r = t[e], i = n[e];
		return r === void 0 || i === void 0 ? !1 : r === i;
	});
	Ta = (e) => {
		if (e.kind === `color`) return !1;
		if (e.kind === `enum`) return e.key === `font-weight` ? e.original === `400` : e.options[0]?.value === e.original;
		let { key: t, original: n } = e;
		return t.startsWith(`padding`) || t.startsWith(`margin`) || t.includes(`gap`) || t === `border-width` || t.endsWith(`-radius`) || t === `border-radius` ? n === 0 : t === `opacity` ? n === 100 : t === `letter-spacing` || t === `z-index` ? n === 0 : !da.has(t) && !!(t === `width` || t === `height` || t === `width,height` || t.startsWith(`max-`) || t.startsWith(`min-`) || t === `line-height`);
	};
	Ea = (e) => Number.isInteger(e) ? e : Math.round(e * 100) / 100;
	Da = (e) => Number.isInteger(e) ? String(e) : e.toFixed(2).replace(/\.?0+$/, ``);
	Oa = (e) => Number.isFinite(e) ? Da(Ea(e)) : String(e);
	ka = (e) => Math.round(e);
	Aa = (e, t) => {
		if (e.kind === `color` || e.kind === `enum`) return typeof t == `string` ? t : e.value;
		let n = typeof t == `number` ? t : e.value;
		return e.key === `opacity` && e.unit === `%` ? Da(Ea(n / 100)) : `${Da(Ea(n))}${e.unit}`;
	};
	ja = (e) => {
		let t = e.trim();
		if (!t || t === `auto` || t === `normal` || t === `none`) return null;
		let n = Number.parseFloat(t);
		return Number.isFinite(n) ? t.endsWith(`%`) ? {
			value: n,
			unit: `%`
		} : t.endsWith(`rem`) || t.endsWith(`em`) ? {
			value: n * 16,
			unit: `px`
		} : t.endsWith(`px`) ? {
			value: n,
			unit: `px`
		} : {
			value: n,
			unit: ``
		} : null;
	};
	Ma = (e, t) => {
		let n = ja(e[t]);
		if (n) return n;
		if (t === `line-height`) {
			let t = ja(e[`font-size`]);
			return t ? {
				value: ka(t.value * et$2),
				unit: t.unit || `px`
			} : null;
		}
		return ia.has(t) ? {
			value: 0,
			unit: `px`
		} : null;
	};
	Na = (e, t) => {
		let n = Ma(e, t[0]);
		if (!n) return null;
		for (let r = 1; r < t.length; r++) {
			let i = Ma(e, t[r]);
			if (!i || i.unit !== n.unit || Math.abs(i.value - n.value) >= .5) return null;
		}
		return n;
	};
	Pa = (e, t) => e === `opacity` ? {
		value: Math.round(t.value * 100),
		unit: `%`
	} : aa.has(e) ? {
		value: ka(t.value),
		unit: ``
	} : {
		value: ka(t.value),
		unit: t.unit || `px`
	};
	Fa = (e) => e !== null;
	Ia = (e, t) => {
		let n = t.map((t) => {
			let n = t.longhands.length === 1 ? Ma(e, t.longhands[0]) : Na(e, t.longhands);
			return n ? {
				definition: t,
				value: n
			} : null;
		}).filter(Fa), r = /* @__PURE__ */ new Map();
		for (let e of n) for (let t of e.definition.longhands) {
			let n = r.get(t);
			(!n || e.definition.longhands.length > n.definition.longhands.length) && r.set(t, e);
		}
		let i = new Set(r.values());
		return n.map((e) => ({
			...e,
			isCanonical: i.has(e)
		}));
	};
	La = (e, t, n) => e === `opacity` ? {
		min: 0,
		max: 100
	} : e === `z-index` ? {
		min: -9999,
		max: 9999
	} : e === `letter-spacing` ? {
		min: -10,
		max: 20
	} : e === `font-size` ? {
		min: 8,
		max: 96
	} : e === `line-height` ? {
		min: 0,
		max: 120
	} : e.includes(`radius`) ? {
		min: 0,
		max: 96
	} : e === `width` || e === `height` || e === `width,height` || e.startsWith(`min-`) || e.startsWith(`max-`) ? {
		min: 0,
		max: Math.max(512, Math.ceil(t * 2))
	} : da.has(e) ? {
		min: -512,
		max: 512
	} : n === `%` ? {
		min: 0,
		max: 100
	} : {
		min: e.startsWith(`margin`) ? -128 : 0,
		max: 384
	};
	Ra = {
		slate: {
			50: `#f8fafc`,
			100: `#f1f5f9`,
			200: `#e2e8f0`,
			300: `#cbd5e1`,
			400: `#94a3b8`,
			500: `#64748b`,
			600: `#475569`,
			700: `#334155`,
			800: `#1e293b`,
			900: `#0f172a`,
			950: `#020617`
		},
		gray: {
			50: `#f9fafb`,
			100: `#f3f4f6`,
			200: `#e5e7eb`,
			300: `#d1d5db`,
			400: `#9ca3af`,
			500: `#6b7280`,
			600: `#4b5563`,
			700: `#374151`,
			800: `#1f2937`,
			900: `#111827`,
			950: `#030712`
		},
		zinc: {
			50: `#fafafa`,
			100: `#f4f4f5`,
			200: `#e4e4e7`,
			300: `#d4d4d8`,
			400: `#a1a1aa`,
			500: `#71717a`,
			600: `#52525b`,
			700: `#3f3f46`,
			800: `#27272a`,
			900: `#18181b`,
			950: `#09090b`
		},
		neutral: {
			50: `#fafafa`,
			100: `#f5f5f5`,
			200: `#e5e5e5`,
			300: `#d4d4d4`,
			400: `#a3a3a3`,
			500: `#737373`,
			600: `#525252`,
			700: `#404040`,
			800: `#262626`,
			900: `#171717`,
			950: `#0a0a0a`
		},
		stone: {
			50: `#fafaf9`,
			100: `#f5f5f4`,
			200: `#e7e5e4`,
			300: `#d6d3d1`,
			400: `#a8a29e`,
			500: `#78716c`,
			600: `#57534e`,
			700: `#44403c`,
			800: `#292524`,
			900: `#1c1917`,
			950: `#0c0a09`
		},
		red: {
			50: `#fef2f2`,
			100: `#fee2e2`,
			200: `#fecaca`,
			300: `#fca5a5`,
			400: `#f87171`,
			500: `#ef4444`,
			600: `#dc2626`,
			700: `#b91c1c`,
			800: `#991b1b`,
			900: `#7f1d1d`,
			950: `#450a0a`
		},
		orange: {
			50: `#fff7ed`,
			100: `#ffedd5`,
			200: `#fed7aa`,
			300: `#fdba74`,
			400: `#fb923c`,
			500: `#f97316`,
			600: `#ea580c`,
			700: `#c2410c`,
			800: `#9a3412`,
			900: `#7c2d12`,
			950: `#431407`
		},
		amber: {
			50: `#fffbeb`,
			100: `#fef3c7`,
			200: `#fde68a`,
			300: `#fcd34d`,
			400: `#fbbf24`,
			500: `#f59e0b`,
			600: `#d97706`,
			700: `#b45309`,
			800: `#92400e`,
			900: `#78350f`,
			950: `#451a03`
		},
		yellow: {
			50: `#fefce8`,
			100: `#fef9c3`,
			200: `#fef08a`,
			300: `#fde047`,
			400: `#facc15`,
			500: `#eab308`,
			600: `#ca8a04`,
			700: `#a16207`,
			800: `#854d0e`,
			900: `#713f12`,
			950: `#422006`
		},
		lime: {
			50: `#f7fee7`,
			100: `#ecfccb`,
			200: `#d9f99d`,
			300: `#bef264`,
			400: `#a3e635`,
			500: `#84cc16`,
			600: `#65a30d`,
			700: `#4d7c0f`,
			800: `#3f6212`,
			900: `#365314`,
			950: `#1a2e05`
		},
		green: {
			50: `#f0fdf4`,
			100: `#dcfce7`,
			200: `#bbf7d0`,
			300: `#86efac`,
			400: `#4ade80`,
			500: `#22c55e`,
			600: `#16a34a`,
			700: `#15803d`,
			800: `#166534`,
			900: `#14532d`,
			950: `#052e16`
		},
		emerald: {
			50: `#ecfdf5`,
			100: `#d1fae5`,
			200: `#a7f3d0`,
			300: `#6ee7b7`,
			400: `#34d399`,
			500: `#10b981`,
			600: `#059669`,
			700: `#047857`,
			800: `#065f46`,
			900: `#064e3b`,
			950: `#022c22`
		},
		teal: {
			50: `#f0fdfa`,
			100: `#ccfbf1`,
			200: `#99f6e4`,
			300: `#5eead4`,
			400: `#2dd4bf`,
			500: `#14b8a6`,
			600: `#0d9488`,
			700: `#0f766e`,
			800: `#115e59`,
			900: `#134e4a`,
			950: `#042f2e`
		},
		cyan: {
			50: `#ecfeff`,
			100: `#cffafe`,
			200: `#a5f3fc`,
			300: `#67e8f9`,
			400: `#22d3ee`,
			500: `#06b6d4`,
			600: `#0891b2`,
			700: `#0e7490`,
			800: `#155e75`,
			900: `#164e63`,
			950: `#083344`
		},
		sky: {
			50: `#f0f9ff`,
			100: `#e0f2fe`,
			200: `#bae6fd`,
			300: `#7dd3fc`,
			400: `#38bdf8`,
			500: `#0ea5e9`,
			600: `#0284c7`,
			700: `#0369a1`,
			800: `#075985`,
			900: `#0c4a6e`,
			950: `#082f49`
		},
		blue: {
			50: `#eff6ff`,
			100: `#dbeafe`,
			200: `#bfdbfe`,
			300: `#93c5fd`,
			400: `#60a5fa`,
			500: `#3b82f6`,
			600: `#2563eb`,
			700: `#1d4ed8`,
			800: `#1e40af`,
			900: `#1e3a8a`,
			950: `#172554`
		},
		indigo: {
			50: `#eef2ff`,
			100: `#e0e7ff`,
			200: `#c7d2fe`,
			300: `#a5b4fc`,
			400: `#818cf8`,
			500: `#6366f1`,
			600: `#4f46e5`,
			700: `#4338ca`,
			800: `#3730a3`,
			900: `#312e81`,
			950: `#1e1b4b`
		},
		violet: {
			50: `#f5f3ff`,
			100: `#ede9fe`,
			200: `#ddd6fe`,
			300: `#c4b5fd`,
			400: `#a78bfa`,
			500: `#8b5cf6`,
			600: `#7c3aed`,
			700: `#6d28d9`,
			800: `#5b21b6`,
			900: `#4c1d95`,
			950: `#2e1065`
		},
		purple: {
			50: `#faf5ff`,
			100: `#f3e8ff`,
			200: `#e9d5ff`,
			300: `#d8b4fe`,
			400: `#c084fc`,
			500: `#a855f7`,
			600: `#9333ea`,
			700: `#7e22ce`,
			800: `#6b21a8`,
			900: `#581c87`,
			950: `#3b0764`
		},
		fuchsia: {
			50: `#fdf4ff`,
			100: `#fae8ff`,
			200: `#f5d0fe`,
			300: `#f0abfc`,
			400: `#e879f9`,
			500: `#d946ef`,
			600: `#c026d3`,
			700: `#a21caf`,
			800: `#86198f`,
			900: `#701a75`,
			950: `#4a044e`
		},
		pink: {
			50: `#fdf2f8`,
			100: `#fce7f3`,
			200: `#fbcfe8`,
			300: `#f9a8d4`,
			400: `#f472b6`,
			500: `#ec4899`,
			600: `#db2777`,
			700: `#be185d`,
			800: `#9d174d`,
			900: `#831843`,
			950: `#500724`
		},
		rose: {
			50: `#fff1f2`,
			100: `#ffe4e6`,
			200: `#fecdd3`,
			300: `#fda4af`,
			400: `#fb7185`,
			500: `#f43f5e`,
			600: `#e11d48`,
			700: `#be123c`,
			800: `#9f1239`,
			900: `#881337`,
			950: `#4c0519`
		}
	};
	za = {
		p: `padding`,
		px: `padding-left,padding-right`,
		py: `padding-top,padding-bottom`,
		pt: `padding-top`,
		pr: `padding-right`,
		pb: `padding-bottom`,
		pl: `padding-left`,
		ps: `padding-left`,
		pe: `padding-right`,
		m: `margin`,
		mx: `margin-left,margin-right`,
		my: `margin-top,margin-bottom`,
		mt: `margin-top`,
		mr: `margin-right`,
		mb: `margin-bottom`,
		ml: `margin-left`,
		ms: `margin-left`,
		me: `margin-right`,
		gap: `gap`,
		"gap-x": `column-gap`,
		"gap-y": `row-gap`,
		text: `font-size`,
		leading: `line-height`,
		tracking: `letter-spacing`,
		rounded: `border-radius`,
		"rounded-t": `border-top-left-radius,border-top-right-radius`,
		"rounded-b": `border-bottom-left-radius,border-bottom-right-radius`,
		"rounded-l": `border-top-left-radius,border-bottom-left-radius`,
		"rounded-r": `border-top-right-radius,border-bottom-right-radius`,
		"rounded-tl": `border-top-left-radius`,
		"rounded-tr": `border-top-right-radius`,
		"rounded-bl": `border-bottom-left-radius`,
		"rounded-br": `border-bottom-right-radius`,
		border: `border-width`,
		"border-t": `border-top-width`,
		"border-r": `border-right-width`,
		"border-b": `border-bottom-width`,
		"border-l": `border-left-width`,
		w: `width`,
		h: `height`,
		size: `width,height`,
		"max-w": `max-width`,
		"max-h": `max-height`,
		"min-w": `min-width`,
		"min-h": `min-height`,
		opacity: `opacity`,
		inset: `top,right,bottom,left`,
		"inset-x": `left,right`,
		"inset-y": `top,bottom`,
		top: `top`,
		right: `right`,
		bottom: `bottom`,
		left: `left`,
		z: `z-index`,
		items: `align-items`,
		justify: `justify-content`,
		font: `font-weight`,
		whitespace: `white-space`
	};
	Ba = {
		"font-size": [
			`font-size`,
			`font size`,
			`text-size`,
			`text size`
		],
		color: [
			`color`,
			`text-color`,
			`text color`,
			`text-colour`,
			`text colour`,
			`foreground`,
			`fg`
		],
		"background-color": [
			`background`,
			`background-color`,
			`background color`,
			`bg`,
			`bg-color`,
			`bg color`
		],
		"border-color": [`border-color`, `border color`],
		fill: [
			`fill`,
			`fill-color`,
			`fill color`
		],
		stroke: [
			`stroke`,
			`stroke-color`,
			`stroke color`
		],
		"font-family": [
			`font`,
			`font-sans`,
			`font-serif`,
			`font-mono`,
			`family`,
			`sans`,
			`serif`,
			`mono`
		],
		"font-style": [`italic`, `not-italic`],
		"text-align": [
			`text-left`,
			`text-center`,
			`text-right`,
			`text-justify`
		],
		"text-transform": [
			`uppercase`,
			`lowercase`,
			`capitalize`,
			`normal-case`
		],
		"text-decoration-line": [
			`underline`,
			`line-through`,
			`no-underline`
		],
		"white-space": [
			`whitespace`,
			`whitespace-normal`,
			`whitespace-nowrap`,
			`whitespace-pre`
		],
		"word-break": [
			`break-normal`,
			`break-all`,
			`break-keep`
		],
		"overflow-wrap": [`break-words`, `break-anywhere`],
		"font-variant-numeric": [
			`tabular-nums`,
			`proportional-nums`,
			`lining-nums`,
			`oldstyle-nums`
		],
		"border-style": [
			`border-solid`,
			`border-dashed`,
			`border-dotted`,
			`border-none`
		]
	};
	Va = (() => {
		let e = {};
		for (let [t, n] of Object.entries(za)) n !== void 0 && (e[n] ??= /* @__PURE__ */ new Set()).add(t);
		let t = {};
		for (let [n, r] of Object.entries(e)) r !== void 0 && (t[n] = Array.from(r));
		for (let [e, n] of Object.entries(Ba)) {
			let r = /* @__PURE__ */ new Set([...t[e] ?? [], ...n]);
			t[e] = Array.from(r);
		}
		return t;
	})();
	Ha = {
		text: `color`,
		bg: `background-color`,
		border: `border-color`,
		"border-x": `border-color`,
		"border-y": `border-color`,
		"border-t": `border-color`,
		"border-r": `border-color`,
		"border-b": `border-color`,
		"border-l": `border-color`,
		"border-s": `border-color`,
		"border-e": `border-color`,
		fill: `fill`,
		stroke: `stroke`
	};
	Ua = (e) => za[e] !== void 0 || Ha[e] !== void 0;
	Wa = (e) => {
		let t = 0, n = -1;
		for (let r = 0; r < e.length; r++) {
			let i = e[r];
			i === `[` ? t++ : i === `]` ? t-- : i === `:` && t === 0 && (n = r);
		}
		let r = n >= 0 ? e.slice(n + 1) : e;
		return r.startsWith(`!`) ? r.slice(1) : r;
	};
	Ga = (e) => {
		let t = Wa(e);
		if (!t) return null;
		let n = t.indexOf(`[`);
		if (n > 0) return t.slice(0, n).replace(/-$/, ``) || null;
		let r = t.split(`-`).filter(Boolean);
		if (r.length === 0) return null;
		for (let e = r.length; e >= 1; e--) {
			let t = r.slice(0, e).join(`-`);
			if (Ua(t)) return t;
		}
		return r[0];
	};
	Ka = /-(red|orange|amber|yellow|lime|green|emerald|teal|cyan|sky|blue|indigo|violet|purple|fuchsia|pink|rose|slate|gray|zinc|neutral|stone|black|white|transparent|current|inherit)\b/;
	qa = (e, t) => e !== `text` && e !== `bg` && e !== `border` && !e.startsWith(`border-`) ? !1 : Ka.test(t);
	Ja = /* @__PURE__ */ new Set(/* @__PURE__ */ `color.red.orange.amber.yellow.lime.green.emerald.teal.cyan.sky.blue.indigo.violet.purple.fuchsia.pink.rose.slate.gray.grey.zinc.neutral.stone.black.white.transparent.current.currentcolor.inherit.primary.secondary.accent.muted.foreground.background.destructive.success.warning.error.info.card.popover.input.ring`.split(`.`));
	Ya = /* @__PURE__ */ new Set([
		`xs`,
		`sm`,
		`base`,
		`lg`,
		`xl`,
		`2xl`,
		`3xl`,
		`4xl`,
		`5xl`,
		`6xl`,
		`7xl`,
		`8xl`,
		`9xl`
	]);
	Xa = /^(#|rgb|rgba|hsl|hsla|hwb|lab|lch|oklab|oklch|color:|color\(|var\(--.*color)/;
	Za = /^(-?\d|length:|size:)/;
	Qa = (e) => e.trim().toLowerCase().replace(/\s+/g, `-`).replace(/([a-z])(\d)/g, `$1-$2`);
	$a = (e) => Xa.test(e) && !Za.test(e);
	eo = (e, t) => {
		if (e === t) return ``;
		let n = `${t}-`;
		return e.startsWith(n) ? e.slice(n.length) : null;
	};
	to = (e) => {
		let t = Wa(Qa(e)), n = Ga(t);
		if (!n) return null;
		let r = Ha[n];
		if (!r) return null;
		let i = t.indexOf(`[`);
		if (i >= 0) {
			let e = t.slice(i + 1).replace(/\]$/, ``);
			return $a(e) ? r : null;
		}
		let a = eo(t, n);
		if (a === null) return null;
		if (n === `bg` && a === `` || (n === `fill` || n === `stroke`) && a === ``) return r;
		let o = a.split(`-`)[0];
		return n === `text` && Ya.has(o) ? null : Ja.has(o) ? r : null;
	};
	no = {
		black: `#000000`,
		white: `#ffffff`,
		transparent: Ue
	};
	ro = (e) => {
		let t = Wa(Qa(e)), n = Ga(t);
		if (!n || Ha[n] === void 0 || t.includes(`[`)) return null;
		let r = eo(t, n);
		if (!r) return null;
		let i = no[r];
		if (i) return i;
		let a = r.lastIndexOf(`-`);
		if (a < 0) return null;
		let o = r.slice(0, a), s = o === `grey` ? `gray` : o, c = Number.parseInt(r.slice(a + 1), 10);
		return Ra[s]?.[c] ?? null;
	};
	io = /^-?\d/;
	ao = (e, t) => {
		if (t === `` || t.startsWith(`[`) || io.test(t)) return !0;
		let n = t.split(`-`)[0];
		return e === `text` ? Ya.has(n) : e === `font` || e === `border` || e.startsWith(`border-`) ? !1 : (e === `rounded` || e.startsWith(`rounded-`), !0);
	};
	oo = {
		"font-sans": {
			property: `font-family`,
			value: `ui-sans-serif, system-ui, sans-serif, "Apple Color Emoji", "Segoe UI Emoji", "Segoe UI Symbol", "Noto Color Emoji"`
		},
		"font-serif": {
			property: `font-family`,
			value: `ui-serif, Georgia, Cambria, "Times New Roman", Times, serif`
		},
		"font-mono": {
			property: `font-family`,
			value: `ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace`
		},
		"font-thin": {
			property: `font-weight`,
			value: `100`
		},
		"font-extralight": {
			property: `font-weight`,
			value: `200`
		},
		"font-light": {
			property: `font-weight`,
			value: `300`
		},
		"font-normal": {
			property: `font-weight`,
			value: `400`
		},
		"font-medium": {
			property: `font-weight`,
			value: `500`
		},
		"font-semibold": {
			property: `font-weight`,
			value: `600`
		},
		"font-bold": {
			property: `font-weight`,
			value: `700`
		},
		"font-extrabold": {
			property: `font-weight`,
			value: `800`
		},
		"font-black": {
			property: `font-weight`,
			value: `900`
		},
		"text-left": {
			property: `text-align`,
			value: `left`
		},
		"text-center": {
			property: `text-align`,
			value: `center`
		},
		"text-right": {
			property: `text-align`,
			value: `right`
		},
		"text-justify": {
			property: `text-align`,
			value: `justify`
		},
		italic: {
			property: `font-style`,
			value: `italic`
		},
		"not-italic": {
			property: `font-style`,
			value: `normal`
		},
		uppercase: {
			property: `text-transform`,
			value: `uppercase`
		},
		lowercase: {
			property: `text-transform`,
			value: `lowercase`
		},
		capitalize: {
			property: `text-transform`,
			value: `capitalize`
		},
		"normal-case": {
			property: `text-transform`,
			value: `none`
		},
		underline: {
			property: `text-decoration-line`,
			value: `underline`
		},
		"line-through": {
			property: `text-decoration-line`,
			value: `line-through`
		},
		"no-underline": {
			property: `text-decoration-line`,
			value: `none`
		},
		"whitespace-normal": {
			property: `white-space`,
			value: `normal`
		},
		"whitespace-nowrap": {
			property: `white-space`,
			value: `nowrap`
		},
		"whitespace-pre": {
			property: `white-space`,
			value: `pre`
		},
		"whitespace-pre-line": {
			property: `white-space`,
			value: `pre-line`
		},
		"whitespace-pre-wrap": {
			property: `white-space`,
			value: `pre-wrap`
		},
		"whitespace-break-spaces": {
			property: `white-space`,
			value: `break-spaces`
		},
		"break-normal": {
			property: `word-break`,
			value: `normal`
		},
		"break-all": {
			property: `word-break`,
			value: `break-all`
		},
		"break-keep": {
			property: `word-break`,
			value: `keep-all`
		},
		"break-words": {
			property: `overflow-wrap`,
			value: `break-word`
		},
		"break-anywhere": {
			property: `overflow-wrap`,
			value: `anywhere`
		},
		"tabular-nums": {
			property: `font-variant-numeric`,
			value: `tabular-nums`
		},
		"proportional-nums": {
			property: `font-variant-numeric`,
			value: `proportional-nums`
		},
		"lining-nums": {
			property: `font-variant-numeric`,
			value: `lining-nums`
		},
		"oldstyle-nums": {
			property: `font-variant-numeric`,
			value: `oldstyle-nums`
		},
		"slashed-zero": {
			property: `font-variant-numeric`,
			value: `slashed-zero`
		},
		"border-solid": {
			property: `border-style`,
			value: `solid`
		},
		"border-dashed": {
			property: `border-style`,
			value: `dashed`
		},
		"border-dotted": {
			property: `border-style`,
			value: `dotted`
		},
		"border-none": {
			property: `border-style`,
			value: `none`
		}
	};
	so = (e) => {
		let t = Qa(e);
		if (!t) return [];
		let n = po(t);
		if (n) return [n.property];
		let r = to(t);
		return r ? [r] : [];
	};
	co = (e) => {
		let t = Qa(e);
		if (!t) return [];
		let n = Ga(t);
		if (!n) return [];
		let r = eo(Wa(t), n);
		if (r === null || !ao(n, r)) return [];
		let i = fo(n);
		return i ? [i] : [];
	};
	lo = (e) => {
		let t = e.getAttribute(`class`);
		if (!t) return /* @__PURE__ */ new Set();
		let n = t.split(/\s+/).filter(Boolean), r = /* @__PURE__ */ new Set();
		for (let e of n) {
			let t = Ga(e);
			if (!t || qa(t, e)) continue;
			let n = po(e);
			if (n) {
				r.add(n.property);
				continue;
			}
			let i = fo(t);
			i && r.add(i);
		}
		return r;
	};
	uo = (e) => Va[e] ?? [];
	fo = (e) => {
		let t = za[e];
		return t === void 0 ? null : t;
	};
	po = (e) => {
		let t = oo[Wa(e)];
		return t === void 0 ? null : t;
	};
	mo = (e, t, n) => {
		let r = Pa(e.key, t), i = r.unit === `px` && Math.abs(r.value) > 1e7, a = La(e.key, i ? 0 : r.value, r.unit), o = r.value < 0 ? a.min : a.max, s = i ? o : r.value;
		return {
			kind: `numeric`,
			key: e.key,
			label: e.label,
			cssProperties: e.longhands,
			min: a.min,
			max: a.max,
			value: s,
			original: s,
			unit: r.unit,
			tailwindAliases: uo(e.key),
			isPrioritized: !1,
			isDefault: !1,
			isCanonical: n
		};
	};
	ho = (e, t, n, r = !1) => {
		let i = Gr(n), a = i ? Ar(i) : null, o = a !== null && a.alpha !== 0;
		if (!o && !r) return null;
		let s = o && i ? i : Ue;
		return {
			kind: `color`,
			key: e,
			label: t,
			cssProperties: [e],
			value: s,
			original: s,
			tailwindAliases: uo(e),
			isPrioritized: r,
			isDefault: !1,
			isCanonical: !0
		};
	};
	go = (e, t, n) => {
		let r = t.trim(), i = n ?? e.options;
		return !i || !i.some((e) => e.value === r) ? null : {
			kind: `enum`,
			key: e.key,
			label: e.label,
			cssProperties: [e.key],
			value: r,
			original: r,
			options: i,
			tailwindAliases: uo(e.key),
			isPrioritized: !1,
			isDefault: !1,
			isCanonical: !0
		};
	};
	_o = new Map((/* @__PURE__ */ `background-color.color.padding.padding-left,padding-right.padding-top,padding-bottom.padding-top.padding-right.padding-bottom.padding-left.margin.margin-left,margin-right.margin-top,margin-bottom.margin-top.margin-right.margin-bottom.margin-left.font-family.font-size.font-weight.font-style.line-height.letter-spacing.text-transform.text-align.text-decoration-line.white-space.word-break.overflow-wrap.font-variant-numeric.border-radius.border-top-left-radius,border-top-right-radius.border-bottom-left-radius,border-bottom-right-radius.border-top-left-radius,border-bottom-left-radius.border-top-right-radius,border-bottom-right-radius.border-top-left-radius.border-top-right-radius.border-bottom-left-radius.border-bottom-right-radius.width,height.width.height.max-width.max-height.min-width.min-height.opacity.gap.row-gap.column-gap.border-width.border-top-width.border-right-width.border-bottom-width.border-left-width.border-color.align-items.justify-content.top,right,bottom,left.left,right.top,bottom.top.right.bottom.left.z-index`.split(`.`)).map((e, t) => [e, t]));
	vo = (e) => _o.get(e.key) ?? 1 / 0;
	yo = (e) => e.map((e, t) => ({
		property: e,
		index: t
	})).sort((e, t) => {
		let n = vo(e.property) - vo(t.property);
		return n === 0 ? e.index - t.index : n;
	}).map((e) => e.property);
	bo = (e) => {
		let t = ba(e), n = getComputedStyle(e), r = xa(n), i = Sa(e), a = [], o = /* @__PURE__ */ new Set(), s = (e) => {
			o.has(e.key) || (a.push(e), o.add(e.key));
		};
		for (let e of fa) for (let n of Ia(t, e)) s(mo(n.definition, n.value, n.isCanonical));
		for (let e of pa) {
			let n = Ma(t, e.key);
			n && s(mo({
				key: e.key,
				label: e.label,
				longhands: [e.key]
			}, n, !0));
		}
		for (let { key: e, label: t, alwaysShow: r } of ma) {
			let i = n.getPropertyValue(e);
			if (!r && (!i || Mr(i))) continue;
			let a = ho(e, t, i, r);
			a && s(a);
		}
		let c = n.getPropertyValue(_a.key);
		if (c) {
			let e = go(_a, c, ga(c));
			e && s(e);
		}
		for (let e of va) {
			let t = n.getPropertyValue(e.key);
			if (!t) continue;
			let r = go(e, t);
			r && s(r);
		}
		return xo(a, lo(e), r, i);
	};
	xo = (e, t, n, r) => {
		let i = [], a = [];
		for (let o of e) if (o.isPrioritized || t.has(o.key)) i.push({
			...o,
			isPrioritized: !0,
			isDefault: !1
		});
		else {
			let e = r ? wa(o, n, r) : Ta(o);
			a.push({
				...o,
				isDefault: e
			});
		}
		return [...yo(i), ...yo(a)];
	};
	So = [
		{
			keyword: `letter`,
			family: `letter-spacing`
		},
		{
			keyword: `tracking`,
			family: `letter-spacing`
		},
		{
			keyword: `leading`,
			family: `line-height`
		},
		{
			keyword: `line-height`,
			family: `line-height`
		},
		{
			keyword: `lineheight`,
			family: `line-height`
		},
		{
			keyword: `radius`,
			family: `radius`
		},
		{
			keyword: `rounded`,
			family: `radius`
		},
		{
			keyword: `corner`,
			family: `radius`
		},
		{
			keyword: `font-size`,
			family: `font-size`
		},
		{
			keyword: `fontsize`,
			family: `font-size`
		},
		{
			keyword: `text`,
			family: `font-size`
		},
		{
			keyword: `stroke`,
			family: `border-width`
		},
		{
			keyword: `border-width`,
			family: `border-width`
		},
		{
			keyword: `borderwidth`,
			family: `border-width`
		},
		{
			keyword: `space`,
			family: `spacing`
		},
		{
			keyword: `gap`,
			family: `spacing`
		},
		{
			keyword: `gutter`,
			family: `spacing`
		},
		{
			keyword: `inset`,
			family: `spacing`
		},
		{
			keyword: `padding`,
			family: `spacing`
		},
		{
			keyword: `margin`,
			family: `spacing`
		},
		{
			keyword: `size`,
			family: `size`
		},
		{
			keyword: `width`,
			family: `size`
		},
		{
			keyword: `height`,
			family: `size`
		}
	];
	Co = (e) => {
		let t = e.toLowerCase();
		for (let { keyword: e, family: n } of So) if (t.includes(e)) return n;
		return null;
	};
	wo = (e) => e.includes(`radius`) ? `radius` : e.endsWith(`-width`) && e.includes(`border`) ? `border-width` : e === `font-size` ? `font-size` : e === `line-height` ? `line-height` : e === `letter-spacing` ? `letter-spacing` : e.startsWith(`padding`) || e.startsWith(`margin`) || e.endsWith(`gap`) || e === `top` || e === `right` || e === `bottom` || e === `left` || e.startsWith(`inset`) ? `spacing` : e === `width` || e === `height` || e.startsWith(`min-`) || e.startsWith(`max-`) ? `size` : null;
	To = (e) => {
		let t = ja(e);
		return !t || t.unit !== `px` ? null : ka(t.value);
	};
	Eo = (e, t) => t === null ? !0 : e.length === t.length ? e < t : e.length < t.length;
	Do = (e, t, n) => {
		if (n === 1) {
			for (let n of e) if (n > t) return n;
			return null;
		}
		for (let n = e.length - 1; n >= 0; n--) if (e[n] < t) return e[n];
		return null;
	};
	Oo = (e, t, n) => {
		let r = (t === 1 ? Math.floor(e / n) + 1 : Math.ceil(e / n) - 1) * n;
		return r < 0 ? null : r;
	};
	ko = null;
	Ao = () => {
		let e = Array.from(document.adoptedStyleSheets ?? []), t = document.styleSheets.length;
		if (ko && ko.documentSheetCount === t && Aa$1(ko.adoptedSheets, e)) return ko.names;
		let n = /* @__PURE__ */ new Set();
		for (let t of [...document.styleSheets, ...e]) {
			let e;
			try {
				e = t.cssRules;
			} catch {
				continue;
			}
			jo(e, n);
		}
		return ko = {
			documentSheetCount: t,
			adoptedSheets: e,
			names: n
		}, n;
	};
	jo = (e, t) => {
		for (let n of Array.from(e)) if (n instanceof CSSStyleRule) {
			let e = n.style;
			for (let n = 0; n < e.length; n++) {
				let r = e.item(n);
				r.startsWith(`--`) && t.add(r);
			}
		} else n instanceof CSSGroupingRule && jo(n.cssRules, t);
	};
	Mo = {
		hasTokens: !1,
		matchColor: () => null,
		matchLength: () => null,
		stepLength: () => null
	};
	No = (e) => {
		if (typeof document > `u`) return Mo;
		let t = Ao();
		if (t.size === 0) return Mo;
		let n = getComputedStyle(e), r = /* @__PURE__ */ new Map(), i = /* @__PURE__ */ new Map(), a = /* @__PURE__ */ new Map(), o = null;
		for (let e of t) {
			let t = n.getPropertyValue(e).trim();
			if (!t) continue;
			let s = To(t);
			if (s !== null) {
				e === `--spacing` && s > 0 && (o = s);
				let t = Co(e);
				if (t !== null) {
					let n = i.get(s);
					n ? n.push({
						name: e,
						family: t
					}) : i.set(s, [{
						name: e,
						family: t
					}]);
					let r = a.get(t);
					r ? r.add(s) : a.set(t, /* @__PURE__ */ new Set([s]));
				}
				continue;
			}
			let c = Gr(t);
			if (c) {
				let t = c.toLowerCase(), n = r.get(t) ?? null;
				Eo(e, n) && r.set(t, e);
			}
		}
		if (r.size === 0 && i.size === 0) return Mo;
		let s = /* @__PURE__ */ new Map();
		for (let [e, t] of a) s.set(e, Array.from(t).sort((e, t) => e - t));
		return {
			hasTokens: !0,
			matchColor: (e) => {
				let t = Gr(e)?.toLowerCase();
				return t ? r.get(t) ?? null : null;
			},
			matchLength: (e, t) => {
				let n = wo(t);
				if (n === null) return null;
				let r = i.get(Math.round(e));
				if (!r) return null;
				let a = null;
				for (let e of r) e.family === n && Eo(e.name, a) && (a = e.name);
				return a;
			},
			stepLength: (e, t, n) => {
				let r = wo(n);
				if (r === null) return null;
				let i = Math.round(e), a = s.get(r);
				return a && a.length >= 2 ? Do(a, i, t) : (r === `spacing` || r === `size`) && o ? Oo(i, t, o) : a ? Do(a, i, t) : null;
			}
		};
	};
	Po = (e) => e.kind === `color` || e.kind === `enum` ? e.value : e.key === `opacity` && e.unit === `%` ? Oa(e.value / 100) : `${Oa(e.value)}${e.unit}`;
	Fo = (e, t, n) => {
		if (!n?.hasTokens) return null;
		let { edit: r } = t;
		return r.kind === `color` ? n.matchColor(r.value) : r.kind === `numeric` && r.unit === `px` ? n.matchLength(r.value, e) : null;
	};
	Io = (e) => {
		let t = /* @__PURE__ */ new Map();
		for (let n of e.edits) {
			let e = Po(n);
			for (let r of n.cssProperties) t.set(r, {
				edit: n,
				cssValue: e
			});
		}
		let n = !1;
		return {
			lines: Array.from(t, ([t, r]) => {
				let i = Fo(t, r, e.designTokens);
				i && (n = !0);
				let a = i ? ` /* var(${i}) */` : ``;
				return `${t}: ${r.cssValue};${a}`;
			}),
			usesToken: n
		};
	};
	Lo = (e) => e.filePath ? `${e.filePath}:${e.lineNumber}` : ``;
	Ro = (e) => [
		"```css",
		...e,
		"```"
	].join(`
`);
	zo = (e) => {
		if (e.length === 0) return ``;
		let t = e.map((e) => {
			let { lines: t, usesToken: n } = Io(e);
			return {
				header: Lo(e),
				cssLines: t,
				usesToken: n
			};
		}), n = [`Apply these style changes canonically (CSS = visual intent; choose the source change that best expresses the underlying layout intent):`];
		if (t.some((e) => e.usesToken) && n.push("Prefer the design token shown in each `/* var(--token) */` comment over the raw value when it matches the project's intent."), t.length === 1) {
			let [e] = t;
			return e.header && n.push(`\n${e.header}`), n.push(Ro(e.cssLines)), n.join(`
`);
		}
		for (let e of t) n.push(e.header ? `\n${e.header}` : ``), n.push(Ro(e.cssLines));
		return n.join(`
`);
	};
	Bo = (e) => `style` in e && e.style instanceof CSSStyleDeclaration;
	Vo = (e) => {
		let t = /* @__PURE__ */ new Map(), n = Bo(e) ? e : null;
		return {
			apply: (e, r) => {
				if (n) for (let i of e) t.has(i) || t.set(i, {
					value: n.style.getPropertyValue(i),
					priority: n.style.getPropertyPriority(i)
				}), n.style.setProperty(i, r);
			},
			restore: () => {
				if (!n) t.clear();
				else {
					for (let [e, { value: r, priority: i }] of t) r ? n.style.setProperty(e, r, i) : n.style.removeProperty(e);
					t.clear();
				}
			},
			hasAppliedStyles: () => t.size > 0
		};
	};
	Ho = (e, t) => {
		for (let n of t) {
			let t = e.edits.findIndex((e) => e.key === n.key);
			t >= 0 ? e.edits[t] = n : e.edits.push(n);
		}
	};
	Uo = (e, t) => ({
		element: e.element,
		preview: e.preview,
		filePath: e.filePath ?? ``,
		lineNumber: e.lineNumber ?? 0,
		edits: t,
		designTokens: e.designTokens
	});
	Wo = (e) => {
		let [t, n] = z(null), [r, i] = z(!1), a = [], o = [], s = (e) => {
			o = e;
		}, c = () => {
			let r = t() !== null;
			n(null), i(!1), a = [], o = [], r && e.onClose?.();
		}, l = () => {
			t()?.preview.restore();
			for (let e = a.length - 1; e >= 0; e--) a[e].preview.restore();
			c();
		}, d = (e) => {
			Lr$2(e).then((t) => {
				t && n((n) => !n || n.element !== e || n.componentName ? n : {
					...n,
					componentName: t
				});
			});
		}, f = (r, i, a = {}) => {
			if (t() !== null) return !1;
			let o = bo(r);
			if (o.length === 0) return !1;
			let s = a.filePath ?? e.store.selectionFilePath ?? void 0, c = a.lineNumber ?? e.store.selectionLineNumber ?? void 0;
			return n({
				element: r,
				position: i,
				properties: o,
				preview: Vo(r),
				filePath: s,
				lineNumber: c,
				componentName: a.componentName,
				tagName: a.tagName ?? D(r),
				initialSearchQuery: a.initialSearchQuery,
				designTokens: No(r)
			}), d(r), e.isActivated() || e.activateRenderer(), e.actions.setPointer(i), e.actions.setFrozenElement(r), e.actions.freeze(), e.onOpen?.(), !0;
		}, p = (r, i) => {
			let s = t();
			if (!s || s.element === r) return !1;
			let c = bo(r);
			return c.length !== 0 && ((o.length > 0 || s.preview.hasAppliedStyles()) && a.push(Uo(s, o)), o = [], n({
				element: r,
				position: i,
				properties: c,
				preview: Vo(r),
				tagName: D(r),
				hasSessionEdits: a.some((e) => e.edits.length > 0),
				designTokens: No(r)
			}), Ur$2(r).then((e) => {
				if (e) {
					for (let t of a) t.element === r && !t.filePath && (t.filePath = e.filePath, t.lineNumber = e.lineNumber ?? 0);
					n((t) => !t || t.element !== r || t.filePath ? t : {
						...t,
						filePath: e.filePath,
						lineNumber: e.lineNumber ?? void 0
					});
				}
			}).catch(() => void 0), d(r), e.actions.setPointer(i), e.actions.setFrozenElement(r), !0);
		}, m = (e, t) => [...a, Uo(e, t)], h = (e) => {
			let t = /* @__PURE__ */ new Map();
			for (let n of e) {
				if (n.edits.length === 0) continue;
				let e = t.get(n.element);
				e ? (!e.filePath && n.filePath && (e.filePath = n.filePath, e.lineNumber = n.lineNumber), Ho(e, n.edits)) : t.set(n.element, {
					filePath: n.filePath,
					lineNumber: n.lineNumber,
					edits: [...n.edits],
					designTokens: n.designTokens
				});
			}
			return Array.from(t.values());
		}, g = (e) => {
			let t = /* @__PURE__ */ new Set();
			for (let n of e) n.edits.length > 0 && t.add(n.element);
			for (let n = e.length - 1; n >= 0; n--) {
				let r = e[n];
				r.edits.length === 0 && !t.has(r.element) && r.preview.restore();
			}
		};
		return {
			state: t,
			trigger: f,
			switchToElement: p,
			setPendingEdits: s,
			dismiss: () => {
				t() !== null && (l(), e.store.wasActivatedByToggle ? e.deactivateRenderer() : e.actions.unfreeze());
			},
			closePreservingRenderer: () => {
				t() !== null && (l(), e.actions.unfreeze());
			},
			submit: (n) => {
				let r = t();
				if (!r) return;
				let i = r.element, a = m(r, n), o = zo(h(a));
				g(a), c(), e.store.wasActivatedByToggle || e.actions.unfreeze(), e.performCopyWithLabel({
					element: i,
					cursorX: r.position.x,
					extraPrompt: o || void 0,
					shouldDeactivateAfter: e.store.wasActivatedByToggle
				});
			},
			resetWithDiscard: l,
			isOpen: () => t() !== null,
			isInteracting: r,
			setInteracting: i
		};
	};
	Go = {
		enabled: !0,
		hue: 0,
		selectionBox: { enabled: !0 },
		dragBox: { enabled: !0 },
		grabbedBoxes: { enabled: !0 },
		elementLabel: { enabled: !0 },
		toolbar: { enabled: !0 }
	};
	Ko = (e, t) => ({
		enabled: t.enabled ?? e.enabled,
		hue: t.hue ?? e.hue,
		selectionBox: { enabled: t.selectionBox?.enabled ?? e.selectionBox.enabled },
		dragBox: { enabled: t.dragBox?.enabled ?? e.dragBox.enabled },
		grabbedBoxes: { enabled: t.grabbedBoxes?.enabled ?? e.grabbedBoxes.enabled },
		elementLabel: { enabled: t.elementLabel?.enabled ?? e.elementLabel.enabled },
		toolbar: { enabled: t.toolbar?.enabled ?? e.toolbar.enabled }
	});
	qo = {
		activationMode: `toggle`,
		keyHoldDuration: 100,
		allowActivationInsideInput: !0,
		activationKey: void 0,
		getContent: void 0,
		maxContextLines: 3,
		freezeReactUpdates: !0
	};
	Jo = (e, t) => {
		if (!e.id) throw Error(`toolbar button requires a non-empty id`);
		if (!e.label) throw Error(`toolbar button "${e.id}" requires a label`);
		if (!e.icon.trimStart().startsWith(`<svg`)) throw Error(`toolbar button "${e.id}" icon must be SVG markup`);
		if (t.has(e.id)) throw Error(`duplicate toolbar button id "${e.id}"`);
		t.add(e.id);
	};
	Yo = (e = {}) => {
		let t = /* @__PURE__ */ new Map(), n = [], r = {}, i = !1, [a, o] = er({
			theme: Go,
			options: {
				...qo,
				...e
			},
			actions: [],
			toolbarButtons: []
		}), s = (t) => {
			let n = Go, r = {
				...qo,
				...e
			}, i = [], a = [], o = /* @__PURE__ */ new Set();
			for (let { config: e } of t) {
				if (e.theme && (n = Ko(n, e.theme)), e.options && (r = {
					...r,
					...e.options
				}), e.actions) for (let t of e.actions) i.push(t);
				if (e.toolbarButtons) for (let t of e.toolbarButtons) Jo(t, o), a.push(t);
			}
			return {
				theme: n,
				options: r,
				actions: i,
				toolbarButtons: a
			};
		}, c = (e) => {
			o(`theme`, e.theme), o(`options`, {
				...e.options,
				...r
			}), o(`actions`, e.actions), o(`toolbarButtons`, e.toolbarButtons);
		}, l = () => {
			c(s(n));
		}, u = (e, t) => {
			r[e] = t, o(`options`, e, () => t);
		}, d = [
			`activationMode`,
			`keyHoldDuration`,
			`allowActivationInsideInput`,
			`activationKey`,
			`getContent`,
			`maxContextLines`,
			`freezeReactUpdates`
		], f = (e) => {
			if (!i) for (let t of d) e[t] !== void 0 && u(t, e[t]);
		}, p = ({ plugin: e, config: t }) => {
			try {
				let n = t.cleanup?.();
				Promise.resolve(n).catch((t) => {
					H$1(new xt$2(e.name, t));
				});
			} catch (t) {
				H$1(new xt$2(e.name, t));
			}
		}, m = (e, t) => {
			let n;
			try {
				n = e.setup?.(t, te) ?? {};
				let r = n.cleanup?.bind(n);
				return {
					...n,
					cleanup: r,
					theme: e.theme ? n.theme ? Ko(Ko(Go, e.theme), n.theme) : e.theme : n.theme,
					options: e.options ? {
						...e.options,
						...n.options
					} : n.options,
					actions: e.actions ? [...e.actions, ...n.actions ?? []] : n.actions,
					toolbarButtons: e.toolbarButtons ? [...e.toolbarButtons, ...n.toolbarButtons ?? []] : n.toolbarButtons,
					hooks: e.hooks ? {
						...e.hooks,
						...n.hooks
					} : n.hooks
				};
			} catch (t) {
				throw n && p({
					plugin: e,
					config: n
				}), new St$2(e.name, t);
			}
		}, h = (e, r) => {
			if (i) return;
			let a = {
				plugin: e,
				config: m(e, r)
			}, o = t.get(e.name), l = new Map(t);
			l.set(e.name, a);
			let u;
			try {
				u = s(l.values());
			} catch (t) {
				throw p(a), new St$2(e.name, t);
			}
			t.set(e.name, a), n = Array.from(t.values()), c(u), o && p(o);
		}, g = (e) => {
			if (i) return;
			let r = t.get(e);
			r && (t.delete(e), n = Array.from(t.values()), p(r), l());
		}, _ = () => {
			if (i) return;
			i = !0;
			let e = n;
			t.clear(), n = [], e.forEach(p), l();
		}, ee = () => n.map(({ plugin: e }) => e.name), v = (e, ...t) => {
			for (let { plugin: r, config: i } of n) {
				let n = i.hooks?.[e];
				if (n) try {
					let i = n(...t);
					i && Promise.resolve(i).catch((t) => {
						H$1(new bt$2(r.name, String(e), t));
					});
				} catch (t) {
					H$1(new bt$2(r.name, String(e), t));
				}
			}
		}, y = (e, ...t) => {
			let r = !1;
			for (let { plugin: i, config: a } of n) {
				let n = a.hooks?.[e];
				if (n) try {
					n(...t) === !0 && (r = !0);
				} catch (t) {
					H$1(new bt$2(i.name, String(e), t));
				}
			}
			return r;
		}, b = async (e, ...t) => {
			for (let { plugin: r, config: i } of n) {
				let n = i.hooks?.[e];
				if (n) try {
					await n(...t);
				} catch (t) {
					H$1(new bt$2(r.name, String(e), t));
				}
			}
		}, x = async (e, t, ...r) => {
			let i = t;
			for (let { plugin: t, config: a } of n) {
				let n = a.hooks?.[e];
				if (n) try {
					i = await n(i, ...r);
				} catch (n) {
					H$1(new bt$2(t.name, String(e), n));
				}
			}
			return i;
		}, S = (e, t, ...r) => {
			let i = t;
			for (let { plugin: t, config: a } of n) {
				let n = a.hooks?.[e];
				if (n) try {
					i = n(i, ...r);
				} catch (n) {
					H$1(new bt$2(t.name, String(e), n));
				}
			}
			return i;
		}, te = {
			onActivate: () => v(`onActivate`),
			onDeactivate: () => v(`onDeactivate`),
			onElementHover: (e) => v(`onElementHover`, e),
			onElementSelect: (e) => {
				let t = !1, r = [];
				for (let { plugin: i, config: a } of n) {
					let n = a.hooks?.onElementSelect;
					if (n) try {
						let a = n(e);
						if (a) {
							if (t = !0, a === !0) continue;
							r.push(a.catch((e) => (H$1(new bt$2(i.name, `onElementSelect`, e)), !1)));
						}
					} catch (e) {
						H$1(new bt$2(i.name, `onElementSelect`, e));
					}
				}
				let i = r.length > 0 ? Promise.all(r).then((e) => e.every(Boolean)) : void 0;
				return {
					wasIntercepted: t,
					pendingResult: i
				};
			},
			onDragStart: (e, t) => v(`onDragStart`, e, t),
			onDragEnd: (e, t) => v(`onDragEnd`, e, t),
			onBeforeCopy: async (e) => b(`onBeforeCopy`, e),
			transformCopyContent: async (e, t) => x(`transformCopyContent`, e, t),
			onAfterCopy: (e, t) => v(`onAfterCopy`, e, t),
			onCopySuccess: (e, t) => v(`onCopySuccess`, e, t),
			onCopyError: (e) => v(`onCopyError`, e),
			onStateChange: (e) => v(`onStateChange`, e),
			onPromptModeChange: (e, t) => v(`onPromptModeChange`, e, t),
			onSelectionBox: (e, t, n) => v(`onSelectionBox`, e, t, n),
			onDragBox: (e, t) => v(`onDragBox`, e, t),
			onGrabbedBox: (e, t) => v(`onGrabbedBox`, e, t),
			onElementLabel: (e, t, n) => v(`onElementLabel`, e, t, n),
			onContextMenu: (e, t) => v(`onContextMenu`, e, t),
			onOpenFile: (e, t) => y(`onOpenFile`, e, t),
			transformHtmlContent: async (e, t) => x(`transformHtmlContent`, e, t),
			transformAgentContext: async (e, t) => x(`transformAgentContext`, e, t),
			transformActionContext: (e) => S(`transformActionContext`, e),
			transformOpenFileUrl: (e, t, n) => S(`transformOpenFileUrl`, e, t, n)
		};
		return ht() && W(_), {
			register: h,
			unregister: g,
			getPluginNames: ee,
			setOptions: f,
			dispose: _,
			store: a,
			hooks: te
		};
	};
	Xo = (e) => `${e}-${Date.now()}-${Math.random().toString(36).slice(2, 9)}`;
	Zo = (e) => {
		let t = e.bounds.x + e.bounds.width / 2, n = e.bounds.width / 2, r = e.mouseX === void 0 ? void 0 : e.mouseX - t;
		return {
			id: Xo(`label`),
			bounds: e.bounds,
			boundsMultiple: e.boundsMultiple,
			tagName: e.tagName,
			componentName: e.componentName,
			status: e.status,
			createdAt: Date.now(),
			element: e.element,
			elements: e.elements,
			mouseX: e.mouseX,
			mouseXOffsetFromCenter: r,
			mouseXOffsetRatio: r !== void 0 && n > 0 ? r / n : void 0,
			hideArrow: e.hideArrow
		};
	};
	Qo = (e, t, n) => {
		let r = /* @__PURE__ */ new Map(), i = (e) => {
			let t = r.get(e);
			t !== void 0 && (window.clearTimeout(t), r.delete(e));
		}, a = () => {
			for (let e of r.values()) window.clearTimeout(e);
			r.clear();
		}, o = () => {
			a(), e.clearLabelInstances(), n?.();
		}, s = (t) => {
			i(t);
			let n = window.setTimeout(() => {
				e.updateLabelInstance(t, `fading`);
				let n = window.setTimeout(() => {
					r.delete(t), e.removeLabelInstance(t);
				}, 125);
				r.set(t, n);
			}, oe);
			r.set(t, n);
		};
		return {
			createInstance: (t, n, r, i, a) => {
				o();
				let s = Zo({
					bounds: t,
					tagName: n,
					componentName: r,
					status: i,
					element: a?.element,
					mouseX: a?.mouseX,
					elements: a?.elements,
					boundsMultiple: a?.boundsMultiple,
					hideArrow: a?.hideArrow
				});
				return e.addLabelInstance(s), s.id;
			},
			createPerElementInstances: (t, n) => {
				o();
				let r = [];
				for (let i of t) {
					let t = Zo({
						bounds: _t$2(i.element),
						tagName: i.tagName,
						componentName: i.componentName,
						status: n,
						element: i.element,
						mouseX: i.mouseX
					});
					e.addLabelInstance(t), r.push(t.id);
				}
				return r;
			},
			updateAfterCopy: (t, n, r) => {
				n ? (e.updateLabelInstance(t, `copied`), s(t)) : (i(t), e.updateLabelInstance(t, `error`, r || `Unknown error`));
			},
			markRetrying: (t) => {
				i(t), e.updateLabelInstance(t, `copying`);
			},
			dismissInstance: (t) => {
				i(t), e.removeLabelInstance(t);
			},
			cancelAllFades: a,
			clearAll: o,
			handleHoverChange: (n, r) => {
				let a = t().find((e) => e.id === n);
				a && (r ? (i(n), a.status === `fading` && e.updateLabelInstance(n, `copied`)) : (a.status === `copied` || a.status === `fading`) && s(n));
			}
		};
	};
	$o = (e) => {
		let t = window.innerWidth, n = window.innerHeight, r = Math.max(0, e.x), i = Math.min(t, e.x + e.width), a = Math.max(0, e.y), o = Math.min(n, e.y + e.height);
		return {
			x: (r + i) / 2,
			y: (a + o) / 2
		};
	};
	es = (e, t, n) => {
		let r = [], i = (t, r) => {
			let i = n(t), a = $o(i), o = Ta$1(a.x, a.y).filter(e), s = o.indexOf(t);
			return s === -1 ? null : o[s + r] ?? null;
		}, a = (e) => {
			let t = i(e, 1);
			return t && (r.push(e), r.length > 50 && (r = r.slice(-50))), t;
		}, o = (e) => {
			if (r.length > 0) {
				let e = r.pop();
				if (Z(e)) return e;
			}
			return i(e, -1);
		}, s = (e, n) => {
			let i = (e) => n ? e.nextElementSibling : e.previousElementSibling, a = i(e);
			for (; a;) {
				if (t(a)) return r = [], a;
				a = i(a);
			}
			return null;
		};
		return {
			findNext: (e, t) => {
				switch (e) {
					case `ArrowUp`: return a(t);
					case `ArrowDown`: return o(t);
					case `ArrowRight`: return s(t, !0);
					case `ArrowLeft`: return s(t, !1);
					default: return null;
				}
			},
			clearHistory: () => {
				r = [];
			}
		};
	};
	ts = () => {
		let e = /* @__PURE__ */ new WeakSet(), t = Object.getOwnPropertyDescriptor(KeyboardEvent.prototype, `key`), n = !1;
		if (t?.get && !t.get.__reactGrabPatched) {
			n = !0;
			let r = t.get, i = function() {
				return e.has(this) ? `` : r.call(this);
			};
			i.__reactGrabPatched = !0, Object.defineProperty(KeyboardEvent.prototype, "key", {
				get: i,
				configurable: !0
			});
		}
		return {
			claimedEvents: e,
			originalKeyDescriptor: t,
			restore: () => {
				n && t && Object.defineProperty(KeyboardEvent.prototype, "key", t);
			}
		};
	};
	ns = (e, t) => ({
		top: t < 25,
		bottom: t > window.innerHeight - 25,
		left: e < 25,
		right: e > window.innerWidth - 25
	});
	rs = (e, t, n) => {
		let r = null, i = () => {
			if (!t()) {
				a();
				return;
			}
			let o = e(), s = ns(o.x, o.y), c = 0, l = 0;
			s.top && (l -= 10), s.bottom && (l += 10), s.left && (c -= 10), s.right && (c += 10);
			let u = !1;
			if (c !== 0 || l !== 0) {
				let e = window.scrollX, t = window.scrollY;
				window.scrollBy({
					behavior: `instant`,
					left: c,
					top: l
				});
				let r = window.scrollX - e, i = window.scrollY - t;
				u = r !== 0 || i !== 0, u && n?.({
					x: r,
					y: i
				});
			}
			r = u ? M$1(i) : null;
		}, a = () => {
			r !== null && (ci$2(r), r = null);
		};
		return {
			start: i,
			stop: a,
			isActive: () => r !== null
		};
	};
	is = () => {
		let e = globalThis;
		return !!(e.chrome?.runtime?.id || e.browser?.runtime?.id);
	};
	as = (e) => {
		try {
			let t = `0.1.50`, n = `data:image/svg+xml;base64,${btoa(`<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 294 294"><path fill="#ff40e0" d="M145 47c25-20 50-27 67-17 16 9 23 30 20 60 0 2-1 5-1 7l-2 13h-1l-12-4c-8-3-17-5-25-6l-17-3c-10-1-20-1-29-1-10 0-20 0-29 1-6 8-11 16-16 24-5 9-9 17-13 26 4 9 8 18 13 26s10 16 16 24l11 14c5 7 11 13 18 19l10 8s-1 0 0 0l-11 9c-17 14-35 22-49 22-6 0-13-2-18-5-16-9-23-30-20-59 1-3 1-5 1-8-30-12-48-29-48-50 0-18 14-35 41-47 2-1 5-2 7-3 0-2 0-5-1-7-3-30 4-51 20-60 18-10 42-3 68 17M71 201c-1 2-1 4-1 5-2 24 3 41 13 47h1c11 7 30 1 51-15-10-9-18-18-26-29-13-1-26-4-38-8m9-38c-3 9-5 17-7 26 8 2 17 4 25 6-3-5-6-10-9-16-3-5-6-10-9-16m-19-53c-2 1-3 1-5 2-21 10-34 23-34 35 0 13 14 27 39 37 3-12 7-25 12-37-5-12-9-24-12-37m37-10c-8 1-17 3-25 6 2 8 4 16 7 25 3-5 6-11 9-16zm-3-61c-4-1-8 0-12 2-10 7-15 24-13 47 0 2 0 3 1 5 12-4 25-6 38-8 8-10 16-20 26-29-15-11-29-17-40-17m111 2c-4-2-8-3-12-2-11 0-25 5-40 17 10 9 19 19 26 29 13 2 26 4 39 8v-5c3-23-2-40-13-47m-61 23c-7 6-13 13-19 19h37c-6-6-12-13-18-19"/><mask id="a"><path fill="#fff" d="m235 85-133 27 28 133 133-27Z"/></mask><path fill="#ff40e0" d="m137 130 76 11c8 1 9 11 3 15l-28 16-4 32c-1 8-10 10-14 4l-41-66c-3-6 1-13 8-12" mask="url(#a)"/></svg>`)}`;
			console.log(`%cReact Grab${` v${t}`}%c\nhttps://react-grab.com`, `background: #330039; color: #ffffff; border: 1px solid #d75fcb; padding: 4px 4px 4px 24px; border-radius: 4px; background-image: url("${n}"); background-size: 16px 16px; background-repeat: no-repeat; background-position: 4px center; display: inline-block; margin-bottom: 4px;`, ``), e && navigator.onLine && !is() && fetch(`https://www.react-grab.com/api/version?source=browser&v=${t}&t=${Date.now()}`, {
				referrerPolicy: `origin`,
				keepalive: !0,
				priority: `low`,
				cache: `no-store`
			}).then((e) => e.text()).then((e) => {
				e && e !== t && console.warn(`[React Grab] v${t} is outdated (latest: v${e}). Run \`npx grab@latest upgrade\` to upgrade.`);
			}).catch(() => null);
		} catch {}
	};
	os = (e) => typeof e == `object` && !!e;
	ss = (e) => {
		if (!os(e)) return null;
		let t = {};
		return typeof e.enabled == `boolean` && (t.enabled = e.enabled), (e.activationMode === `toggle` || e.activationMode === `hold`) && (t.activationMode = e.activationMode), typeof e.keyHoldDuration == `number` && Number.isFinite(e.keyHoldDuration) && (t.keyHoldDuration = e.keyHoldDuration), typeof e.allowActivationInsideInput == `boolean` && (t.allowActivationInsideInput = e.allowActivationInsideInput), typeof e.activationKey == `string` && (t.activationKey = e.activationKey), typeof e.maxContextLines == `number` && Number.isFinite(e.maxContextLines) && (t.maxContextLines = e.maxContextLines), typeof e.freezeReactUpdates == `boolean` && (t.freezeReactUpdates = e.freezeReactUpdates), typeof e.telemetry == `boolean` && (t.telemetry = e.telemetry), Object.keys(t).length === 0 ? null : t;
	};
	cs = () => {
		if (typeof window > `u`) return null;
		try {
			let e = (document.currentScript instanceof HTMLScriptElement ? document.currentScript : null)?.getAttribute(`data-options`);
			return e ? ss(JSON.parse(e)) : null;
		} catch {
			return null;
		}
	};
	ls = (e) => e === `Enter` || e === `NumpadEnter`;
	us = (e, t) => e.x >= t.x && e.x <= t.x + t.width && e.y >= t.y && e.y <= t.y + t.height;
	ds = `react-grab-toolbar-state`;
	fs = () => {
		try {
			let e = localStorage.getItem(ds);
			if (!e) return null;
			let t = JSON.parse(e);
			if (typeof t != `object` || !t) return null;
			let n = t, r = typeof n.collapsed == `boolean` && n.collapsed;
			return {
				edge: n.edge === `top` || n.edge === `bottom` || n.edge === `left` || n.edge === `right` ? n.edge : `bottom`,
				ratio: typeof n.ratio == `number` ? n.ratio : Me,
				collapsed: r,
				enabled: !r,
				defaultAction: typeof n.defaultAction == `string` ? n.defaultAction : Ne
			};
		} catch (e) {
			console.warn(`[react-grab] Failed to load toolbar state from localStorage:`, e);
		}
		return null;
	};
	ps = (e) => {
		try {
			localStorage.setItem(ds, JSON.stringify(e));
		} catch (e) {
			console.warn(`[react-grab] Failed to save toolbar state to localStorage:`, e);
		}
	};
	ms = (e) => {
		let [t, n] = z(!1);
		return pt(() => {
			let t = zt$2((t) => n(e(t))), r = () => n(!1);
			window.addEventListener(`keydown`, t, { capture: !0 }), window.addEventListener(`keyup`, t, { capture: !0 }), window.addEventListener(`blur`, r), W(() => {
				window.removeEventListener(`keydown`, t, { capture: !0 }), window.removeEventListener(`keyup`, t, { capture: !0 }), window.removeEventListener(`blur`, r);
			});
		}), t;
	};
	hs = {
		name: `copy`,
		actions: [{
			id: `copy`,
			label: `Copy`,
			shortcut: `C`,
			showInToolbarMenu: !0,
			onAction: (e) => {
				e.copy?.();
			}
		}]
	};
	gs = {
		name: `comment`,
		actions: [{
			id: `comment`,
			label: `Comment`,
			shortcut: `Enter`,
			showInToolbarMenu: !0,
			onAction: (e) => {
				e.enterPromptMode?.();
			}
		}]
	};
	_s = {
		name: `edit`,
		actions: [{
			id: `edit`,
			label: `Style`,
			shortcut: `S`,
			shortcutModifier: !1,
			showInToolbarMenu: !0,
			enabled: (e) => S(e.element)?.supportsDomEditing !== !1,
			onAction: (e) => {
				e.enterEditMode?.();
			}
		}]
	};
	vs = {
		name: `open`,
		actions: [{
			id: `open`,
			label: `Open`,
			shortcut: `O`,
			enabled: (e) => !!e.filePath,
			onAction: (e) => {
				e.filePath && (ta(e.filePath, e.lineNumber, e.hooks), e.hideContextMenu(), e.cleanup());
			}
		}]
	};
	ys = (e) => {
		let t = e.left + e.width / 2, n = e.top + e.height / 2, r = n, i = window.innerHeight - n, a = t, o = window.innerWidth - t, s = Math.min(r, i, a, o);
		return s === r ? `top` : s === a ? `left` : s === o ? `right` : `bottom`;
	};
	bs = (e) => e.key === `Enter`;
	xs = (e) => e.metaKey || e.ctrlKey;
	Ss = (e, t, n = {}) => {
		if (!t.key) return null;
		if (bs(t)) return e.find((e) => e.shortcut === `Enter`) ?? null;
		if (t.repeat) return null;
		let r = t.key.toLowerCase();
		return xs(t) ? n.includeModifierShortcuts === !0 ? e.find((e) => e.shortcut !== void 0 && e.shortcut !== `Enter` && e.shortcutModifier !== !1 && r === e.shortcut.toLowerCase()) ?? null : null : e.find((e) => e.shortcut !== void 0 && e.shortcutModifier === !1 && r === e.shortcut.toLowerCase()) ?? null;
	};
	Cs = () => {
		let e = !1;
		return {
			arm: () => {
				e = !0;
			},
			clear: () => {
				e = !1;
			},
			consume: () => e ? (e = !1, !0) : !1
		};
	};
	ws = () => {
		let [e, t] = z(!1), [n, r] = z(null), i = Cs(), a = () => {
			let e = n();
			return Z(e) ? e : null;
		}, o = () => {
			r(null), i.clear(), t(!1);
		};
		return {
			selectedElement: a,
			isPendingDismiss: e,
			select: (e, n) => {
				r(e), t(!1), n?.shouldPromptBeforeMouseHandoff ? i.arm() : i.clear();
			},
			clear: o,
			consumeMouseHandoff: () => i.consume() && a() !== null,
			showDismissPrompt: () => a() ? !e() && (t(!0), !0) : (o(), !1),
			takeSelection: (e) => {
				let t = a() ?? (Z(e) ? e : null);
				return o(), t;
			}
		};
	};
	Ts = (e) => e ?? !0;
	Es = (e, t) => {
		if (typeof e.enabled == `function`) {
			if (!t) return !1;
			try {
				return e.enabled(t);
			} catch (t) {
				return H$1(new wt$2(e.id, t)), !1;
			}
		}
		return Ts(e.enabled);
	};
	Ds = (t, n) => {
		if (!Es(t, n)) return !1;
		try {
			let r = t.onAction(n);
			r && r.catch((n) => {
				H$1(new Ct$2(t.id, n));
			});
		} catch (n) {
			H$1(new Ct$2(t.id, n));
		}
		return !0;
	};
	ks = (e) => {
		H$1(new E(`Toolbar state change subscriber failed`, e));
	};
	As = (e, t) => {
		for (let n of e) try {
			Promise.resolve(n(t)).catch(ks);
		} catch (e) {
			ks(e);
		}
	};
	$ = (e, t) => {
		try {
			e();
		} catch (e) {
			t.push(e);
		}
	};
	js = (e) => {
		let t = /* @__PURE__ */ new Map(), n = /* @__PURE__ */ new Map(), r = /* @__PURE__ */ new Map(), a = /* @__PURE__ */ new Map(), o = (e, i) => {
			let o = r.get(e);
			o && (r.delete(e), $(() => o.disconnect(), i));
			for (let t of r.keys()) `host` in t && t.ownerDocument === e && s(t, i);
			for (let t of a.keys()) t.ownerDocument === e && c(t, i);
			let l = n.get(e);
			n.delete(e), l && $(l, i);
			let u = t.get(e);
			t.delete(e), u && $(u, i);
		}, s = (e, t) => {
			let n = r.get(e);
			if (n) {
				r.delete(e), $(() => n.disconnect(), t);
				for (let n of e.querySelectorAll(`*`)) l(n, t);
			}
		}, c = (e, t) => {
			let n = a.get(e);
			n && (a.delete(e), $(() => n.abortController.abort(), t), $(() => e.removeAttribute("data-react-grab-same-origin-frame"), t), n.document && o(n.document, t));
		}, l = (e, t) => {
			b(e) && c(e, t), e.shadowRoot && s(e.shadowRoot, t);
		}, u = (e, t) => {
			for (let n of e.querySelectorAll(`*`)) l(n, t);
			l(e, t);
		}, d = (n) => {
			t.has(n) || (t.set(n, e(n)), f(n), g(n));
		}, f = (e) => {
			let r = e.defaultView?.Element.prototype;
			if (!r) return;
			let i = r.attachShadow, a = new Proxy(i, { apply: (n, r, i) => {
				let a = Reflect.apply(n, r, i);
				return a.mode === `open` && r.isConnected && t.has(e) && g(a), a;
			} });
			r.attachShadow = a, n.set(e, () => {
				r.attachShadow === a && (r.attachShadow = i);
			});
		}, p = (e) => {
			if (a.has(e)) return;
			let t = {
				abortController: new AbortController(),
				document: null
			};
			a.set(e, t);
			let n = () => {
				let n = [], r = N$1(e);
				t.document && t.document !== r && o(t.document, n), t.document = r, r ? (e.setAttribute(De, ``), d(r), q$1(n, `Cleaning up replaced frame document failed`)) : (e.removeAttribute(De), q$1(n, `Cleaning up replaced frame document failed`));
			};
			e.addEventListener(`load`, n, { signal: t.abortController.signal }), n();
		}, m = (e) => {
			Ci$2(e) || (b(e) && p(e), e.shadowRoot && g(e.shadowRoot));
		}, h = (e) => {
			if (m(e), !Ci$2(e)) for (let t of e.querySelectorAll(`*`)) m(t);
		}, g = (e) => {
			if (r.has(e)) return;
			let t = `defaultView` in e ? e.defaultView : e.ownerDocument.defaultView;
			if (!t) return;
			let n = new t.MutationObserver((e) => {
				let t = [];
				for (let n of e) {
					for (let e of n.removedNodes) v(e) && u(e, t);
					for (let e of n.addedNodes) v(e) && h(e);
				}
				q$1(t, `Cleaning up removed frame documents failed`);
			});
			r.set(e, n), n.observe(e, {
				childList: !0,
				subtree: !0
			});
			for (let t of e.querySelectorAll(`*`)) m(t);
		};
		return d(document), () => {
			let e = [];
			for (let t of a.keys()) c(t, e);
			for (let [t, n] of r) r.delete(t), $(() => n.disconnect(), e);
			for (let [t, r] of n) n.delete(t), $(r, e);
			for (let [n, r] of t) t.delete(n), r && $(r, e);
			q$1(e, `Cleaning up same-origin frame documents failed`);
		};
	};
	Ms = (e, t) => {
		window.dispatchEvent(t), t.defaultPrevented && (e.preventDefault(), e.stopImmediatePropagation());
	};
	Ns = (e, t, n) => ({
		bubbles: !0,
		cancelable: !0,
		composed: !0,
		clientX: t,
		clientY: n,
		screenX: e.screenX,
		screenY: e.screenY,
		button: e.button,
		buttons: e.buttons,
		altKey: e.altKey,
		ctrlKey: e.ctrlKey,
		metaKey: e.metaKey,
		shiftKey: e.shiftKey
	});
	Ps = (e) => {
		let t = st$2(e.view, e.clientX, e.clientY), n = new PointerEvent(e.type, {
			...Ns(e, t.x, t.y),
			pointerId: e.pointerId,
			pointerType: e.pointerType,
			isPrimary: e.isPrimary,
			pressure: e.pressure,
			tangentialPressure: e.tangentialPressure,
			tiltX: e.tiltX,
			tiltY: e.tiltY,
			twist: e.twist,
			width: e.width,
			height: e.height
		});
		Ms(e, n);
	};
	Fs = (e) => {
		let t = st$2(e.view, e.clientX, e.clientY), n = new MouseEvent(e.type, Ns(e, t.x, t.y));
		Ms(e, n);
	};
	Is = (e) => {
		let t = new KeyboardEvent(e.type, {
			bubbles: !0,
			cancelable: !0,
			composed: !0,
			key: e.key,
			code: e.code,
			location: e.location,
			repeat: e.repeat,
			isComposing: e.isComposing,
			altKey: e.altKey,
			ctrlKey: e.ctrlKey,
			metaKey: e.metaKey,
			shiftKey: e.shiftKey
		});
		Ms(e, t);
	};
	Ls = (e) => js((t) => {
		if (t === document) return;
		let n = t.defaultView;
		if (!n) return;
		let r = new AbortController(), i = ao$1(t), a = Co$1(t), o = {
			capture: !0,
			signal: r.signal
		}, s = (t) => {
			e.shouldForwardInteraction() && Ps(t);
		}, c = (t) => {
			e.shouldForwardInteraction() && Fs(t);
		}, l = (t) => {
			(e.shouldForwardInteraction() || e.shouldForwardKeyboardEvent(t)) && Is(t);
		}, u = (n) => {
			e.shouldForwardInteraction() && e.shouldForwardViewportEvent(t) && window.dispatchEvent(new Event(n.type));
		};
		return n.addEventListener(`pointermove`, s, o), n.addEventListener(`pointerdown`, s, o), n.addEventListener(`pointerup`, s, o), n.addEventListener(`pointercancel`, s, o), n.addEventListener(`contextmenu`, c, o), n.addEventListener(`click`, c, o), n.addEventListener(`keydown`, l, o), n.addEventListener(`keyup`, l, o), n.addEventListener(`keypress`, l, o), n.addEventListener(`scroll`, u, o), n.addEventListener(`resize`, u, o), () => {
			r.abort();
			let e = [];
			for (let t of [i, a]) try {
				t();
			} catch (t) {
				e.push(t);
			}
			q$1(e, `Cleaning up frame event forwarding failed`);
		};
	});
	Rs = (e, t) => {
		let n = t.ownerDocument;
		for (; n;) {
			if (n === e) return !0;
			n = ot$2(n.defaultView)?.ownerDocument ?? null;
		}
		return !1;
	};
	zs = null;
	Us = (e) => {
		zs === e && (zs = null), typeof window < `u` && window.__REACT_GRAB__ === e && delete window.__REACT_GRAB__;
	};
	qs = [
		hs,
		_s,
		gs,
		vs
	];
	Js = { status: `cancelled` };
	Ys = !1;
	Xs = (e) => {
		if (typeof window > `u`) return di();
		let t = {
			enabled: !0,
			activationMode: `toggle`,
			keyHoldDuration: 100,
			allowActivationInsideInput: !0,
			...cs(),
			...e
		};
		if (t.enabled === !1 || Ys) return di();
		Ys = !0, It$2(t.container ?? null), as(t.telemetry !== !1);
		let { enabled: r, telemetry: i, container: s, ...l } = t;
		return ot((e) => {
			let t = !1, r = new AbortController(), i = null, s, d = /* @__PURE__ */ new Set(), h = Yo(l), { store: g, actions: y$4, pointer: b, viewportVersion: S, current: w } = gr({ keyHoldDuration: h.store.options.keyHoldDuration ?? 100 }), T = V(() => w().state === `holding`), E$4 = V(() => w().state === `active`), ae = V(() => {
				let e = w();
				return e.state === `active` && e.phase === `frozen`;
			}), D$4 = V(() => {
				let e = w();
				return e.state === `active` && (e.phase === `dragging-select` || e.phase === `dragging-reposition`);
			}), oe$3 = V(() => {
				if (!D$4()) return !1;
				let e = Math.abs(b().x + window.scrollX - g.dragStart.x), t = Math.abs(b().y + window.scrollY - g.dragStart.y);
				return e > 2 || t > 2;
			}), se = V(() => {
				let e = w();
				return e.state === `active` && e.phase === `dragging-reposition`;
			}), ce = V(() => {
				let e = w();
				return e.state === `active` && e.phase === `justDragged`;
			}), k = V(() => w().state === `copying`), le = V(() => g.selectionInteractionLockDepth > 0), de = V(() => w().state === `justCopied`), A = V(() => {
				let e = w();
				return e.state === `active` && !!e.isPromptMode;
			}), ge = V(() => g.pendingCommentMode || A()), be = V(() => {
				let e = w();
				return e.state === `active` && !!e.isPromptMode && !!e.isPendingDismiss;
			}), Se = /* @__PURE__ */ new Map(), Ce = (e, t) => {
				Se.has(e) || Se.set(e, document.body.style[e]), document.body.style[e] = t;
			}, Te$2 = (e) => {
				let t = Se.get(e);
				t !== void 0 && (document.body.style[e] = t, Se.delete(e));
			};
			B(ft(E$4, (e, t) => {
				e && !t ? (No$1(b().x, b().y), Ce(`touchAction`, `none`)) : !e && t && (Po$1(), Te$2(`touchAction`));
			}));
			let De = fs(), [Oe, ke] = z(!De || !De.collapsed), [je$2, Fe$1] = z(0), [Le, Re$2] = z(0), [Ue, Ge] = z(De), [Ke, qe] = z(!1), Je = ms((e) => e.shiftKey), [Ye, M] = z(null), [Xe, Ze] = z(null), [Qe, $e] = z(null), N = Wo({
				store: g,
				actions: y$4,
				isActivated: E$4,
				activateRenderer: () => Gn(),
				deactivateRenderer: () => X(),
				performCopyWithLabel: (e) => ln(e),
				onOpen: () => {
					sa(), it?.(), it = oa(aa, Ze);
				},
				onClose: () => {
					it?.(), it = null, Ze(null);
				}
			}), et = V(() => g.contextMenuPosition !== null || N.isOpen()), tt = V(() => et() || Ye() !== null), nt, rt = null, it = null, P = !1, F = /* @__PURE__ */ new WeakMap(), I = ws(), L = () => !Oe() || A() || le() || et() || I.isPendingDismiss(), R = () => {
				St(!1), F = /* @__PURE__ */ new WeakMap();
			}, at = (e) => {
				let t = Ue() ?? fs(), n = {
					edge: t?.edge ?? `bottom`,
					ratio: t?.ratio ?? .5,
					collapsed: t?.collapsed ?? !1,
					enabled: t?.enabled ?? !0,
					defaultAction: t?.defaultAction ?? `copy`,
					...e
				};
				ps(n), Ge(n), As(d, n);
			}, ot = () => {
				G.timerId !== null && (clearTimeout(G.timerId), G.timerId = null);
			}, st = () => {
				G.copyWaiting = !1, G.holdTimerFired = !1, G.startTimestamp = null;
			};
			B(() => {
				w().state === `holding` ? (G.startTimestamp = Date.now(), G.timerId = window.setTimeout(() => {
					G.timerId = null, G.copyWaiting ? G.holdTimerFired = !0 : y$4.activate();
				}, g.keyHoldDuration), W(ot)) : ot();
			}), B(() => {
				let e = w();
				if (e.state !== `active` || e.phase !== `justDragged`) return;
				let t = setTimeout(() => {
					y$4.finishJustDragged();
				}, oe);
				W(() => clearTimeout(t));
			}), B(() => {
				if (w().state !== `justCopied`) return;
				let e = setTimeout(() => {
					y$4.finishJustCopied();
				}, oe);
				W(() => clearTimeout(e));
			}), B(ft(T, (e, t = !1) => {
				t && !e && E$4() && (h.store.options.activationMode !== `hold` && y$4.setWasActivatedByToggle(!0), h.hooks.onActivate());
			}));
			let ct = (e, t, n) => {
				y$4.setCopyStart({
					x: t,
					y: n
				}, e), y$4.clearInputText();
			}, lt = () => {
				let e = g.frozenElement || J();
				e && y$4.enterPromptMode({
					x: b().x,
					y: b().y
				}, e);
			}, ut = {
				lastDetectionTimestamp: 0,
				pendingDetectionScheduledAt: 0,
				latestPointerX: 0,
				latestPointerY: 0
			}, H = null, [pt, mt] = z(null), [ht, gt] = z(0), _t = (e, t) => {
				H !== null && clearTimeout(H), mt(null), H = window.setTimeout(() => {
					mt({
						x: e,
						y: t
					}), H = null;
				}, 32);
			}, vt = () => {
				H !== null && (clearTimeout(H), H = null), mt(null), Nn();
			}, yt = null, G = {
				timerId: null,
				startTimestamp: null,
				copyWaiting: !1,
				holdTimerFired: !1
			}, bt = null, [xt, St] = z(!1), Ct = 0, wt = !1, Tt = null, Et = () => {
				wt = !0, Tt !== null && window.clearTimeout(Tt), Tt = window.setTimeout(() => {
					wt = !1, Tt = null;
				}, oe);
			}, Dt = () => {
				Tt !== null && (window.clearTimeout(Tt), Tt = null), wt = !1;
			}, Ot = 0, K = null, kt = null, [At, jt] = z(!1), [Mt, Nt] = z(null), [Pt, Ft] = z(null), [It, Lt] = Er(Pt), Rt = V(() => N.isOpen() ? Fe : ge() ? Pe : At() ? Mt() : E$4() ? Ne : null), zt = (e) => Vi(e, ki$2), Bt = es(ki$2, zt, _t$2), Vt = rs(b, () => D$4(), (e) => {
				if (se()) {
					if (y$4.shiftDragStart(e), bt) {
						bt = {
							x: bt.x + e.x,
							y: bt.y + e.y
						};
						return;
					}
					let { pageX: t, pageY: n } = En(b().x, b().y);
					bt = {
						x: t,
						y: n
					};
				}
			}), Ht = V(() => E$4() && !k()), Wt = /* @__PURE__ */ new Map(), Kt = (e, t) => {
				let n = Xo(`grabbed`), r = {
					id: n,
					bounds: e,
					createdAt: Date.now(),
					element: t
				};
				y$4.addGrabbedBox(r), h.hooks.onGrabbedBox(e, t);
				let i = window.setTimeout(() => {
					Wt.delete(n), y$4.removeGrabbedBox(n);
				}, oe + 125);
				Wt.set(n, i);
			}, qt = async (e) => {
				let n = await Promise.all(e.map(async (e) => {
					let t = await Ur$2(e), n = t?.componentName ?? null, r = t?.filePath, i = t?.lineNumber ?? void 0, a = t?.columnNumber ?? void 0;
					n ||= Wr$2(e);
					let o = y(e) ? e.innerText?.slice(0, 100) : void 0;
					return {
						tagName: D(e),
						id: e.id || void 0,
						className: e.getAttribute(`class`) || void 0,
						textContent: o,
						componentName: n ?? void 0,
						filePath: r,
						lineNumber: i,
						columnNumber: a
					};
				}));
				t || window.dispatchEvent(new CustomEvent(`react-grab:element-selected`, { detail: { elements: n } }));
			}, Jt = /* @__PURE__ */ new Map(), q = Qo(y$4, () => g.labelInstances, () => Jt.clear()), Yt = (e) => {
				for (let t of e) g.labelInstances.find((e) => e.id === t)?.status === `copying` && (Jt.delete(t), q.dismissInstance(t));
			}, Xt = () => {
				r.abort();
			}, Zt = () => (r.signal.aborted && !t && (r = new AbortController()), r.signal), Qt = () => (Xt(), Zt()), $t = () => {
				Xt(), i = null, Yt(g.labelInstances.filter((e) => e.status === `copying`).map((e) => e.id));
			}, en = async (e, t, n) => {
				let r, i;
				try {
					if (r = await e(n), r.status === `cancelled`) return r;
					r.status === `failed` && (i = `Failed to copy`);
				} catch (e) {
					if (n.aborted) return Js;
					r = { status: `failed` }, i = pi(e, `Action failed`);
				}
				if (t) for (let e of t) q.updateAfterCopy(e, r.status === `succeeded`, i);
				return r;
			}, tn = (e, t, n, r) => {
				if (e) {
					for (let e of n) Jt.delete(e);
					return;
				}
				let i = {
					operation: t,
					siblingIds: new Set(n),
					shouldDeactivateAfter: r
				};
				for (let e of n) Jt.set(e, i);
			}, nn = async (e, t, n, r = Qt()) => {
				if (r.aborted) return !1;
				Dt(), w().state !== `copying` && y$4.startCopy();
				let i = await en(e, t, r);
				if (i.status === `cancelled`) return !1;
				let a = i.status === `succeeded`;
				return t && tn(a, e, t, !!n), w().state === `copying` && (a && y$4.completeCopy(), n ? X() : a ? (y$4.activate(), Et()) : (y$4.activate(), y$4.unfreeze())), !0;
			}, rn = (e) => {
				let t = Jt.get(e);
				if (!t) return;
				let n = [...t.siblingIds];
				for (let e of n) q.markRetrying(e), Jt.delete(e);
				nn(t.operation, n, t.shouldDeactivateAfter).then((e) => {
					e || Yt(n);
				});
			}, an = (e) => {
				Jt.get(e)?.siblingIds.delete(e), Jt.delete(e), q.dismissInstance(e);
			}, on = (e, t, n, r) => {
				let i = e[0], a = r ?? (i ? Wr$2(i) : null), o = i ? D(i) : null, s = a ?? o ?? void 0;
				return Si({
					getContent: h.store.options.getContent,
					componentName: s,
					maxContextLines: h.store.options.maxContextLines,
					signal: t
				}, h.hooks, e, n);
			}, sn = async (e, t, n, r = Zt()) => {
				if (e.length === 0 || r.aborted) return Js;
				let i = [], a = [];
				for (let t of e) {
					let { wasIntercepted: e, pendingResult: n } = h.hooks.onElementSelect(t);
					e || i.push(t), n && a.push(n), h.store.theme.grabbedBoxes.enabled && Kt(_t$2(t), t);
				}
				if (await li$2(), r.aborted) return Js;
				let o;
				if (i.length > 0 && (o = await on(i, r, t, n), o.status === `cancelled`)) return o;
				if (a.length > 0) {
					let e = await gi(Promise.all(a), r);
					if (e === hi) return o?.status === `succeeded` ? o : Js;
					if (!e.every(Boolean)) throw new Ot$2();
				}
				return r.aborted && !o ? Js : (qt(e).catch((e) => {
					H$1(new E(`Element selection notification failed`, e));
				}), o ?? { status: `succeeded` });
			}, cn = (e) => {
				let n = Qt(), r = {}, a = !1;
				i = r, Lr$2(e.primaryElement).then(async (t) => {
					n.aborted || i !== r ? Yt(e.labelInstanceIds) : (i = null, a = !0, await nn((n) => sn(e.targetElements, e.extraPrompt, t ?? void 0, n), e.labelInstanceIds.length > 0 ? e.labelInstanceIds : null, e.shouldDeactivateAfter, n) ? e.onComplete?.() : Yt(e.labelInstanceIds));
				}).catch((o) => {
					if (n.aborted) {
						Yt(e.labelInstanceIds);
						return;
					}
					if (t || !a && i !== r) return;
					H$1(new E(`Copy operation failed`, o));
					let s = pi(o, `Action failed`);
					for (let t of e.labelInstanceIds) q.updateAfterCopy(t, !1, s);
					e.labelInstanceIds.length > 0 && tn(!1, (t) => Lr$2(e.primaryElement).then((n) => t.aborted ? Js : sn(e.targetElements, e.extraPrompt, n ?? void 0, t)), e.labelInstanceIds, !!e.shouldDeactivateAfter), w().state === `copying` && (e.shouldDeactivateAfter ? X() : (y$4.activate(), y$4.unfreeze()));
				}).finally(() => {
					i === r && (i = null);
				});
			}, ln = (e) => {
				let { element: t, cursorX: n, selectedElements: r, extraPrompt: i, shouldDeactivateAfter: a, onComplete: o, dragRect: s } = e, c = r ?? [t], l = s ?? g.frozenDragRect, u = c.length > 1, d = !u && t === yn() ? Tn() : void 0, f = l && u ? Fi(l) : d ?? _t$2(t), p = u ? f.x + f.width / 2 : n, m = D(t);
				Dt(), y$4.startCopy();
				let h = m ? q.createInstance(f, m, void 0, `copying`, {
					element: t,
					mouseX: p,
					elements: r
				}) : null;
				cn({
					primaryElement: t,
					targetElements: c,
					labelInstanceIds: h ? [h] : [],
					extraPrompt: i,
					shouldDeactivateAfter: a,
					onComplete: o
				});
			}, un = (e) => {
				let { elements: t, labelEntries: n, shouldDeactivateAfter: r, onComplete: i } = e, a = t[0];
				Dt(), y$4.startCopy();
				let o = q.createPerElementInstances(n, `copying`);
				cn({
					primaryElement: a,
					targetElements: t,
					labelInstanceIds: o,
					shouldDeactivateAfter: r,
					onComplete: i
				});
			}, J = V(() => {
				if (S(), !Ht() || oe$3() || le() || I.isPendingDismiss()) return null;
				let e = g.detectedElement;
				return Z(e) ? e : null;
			}), fn = V(() => g.frozenElement || (ae() ? null : J())), mn = V(() => !E$4() || A() || xt() || k() || tt() || I.isPendingDismiss() ? null : Je() ? ae() ? fn() : J() ?? g.frozenElement : I.selectedElement()), hn = V(() => {
				let e = mn();
				return e ? Ri(e, ki$2, zt) : [];
			}), gn = V(() => {
				let e = mn();
				return e ? Math.max(0, hn().findIndex((t) => t.element === e)) : 0;
			}), _n = V(() => mn() !== null);
			B(() => {
				if (!g.detectedElement) return;
				let e = 1, t = setInterval(() => {
					if (Z(g.detectedElement) || y$4.relinkLiveElements(), !Z(g.detectedElement)) {
						if (e > 0) {
							--e;
							return;
						}
						Jr();
					}
				}, 100);
				W(() => clearInterval(t));
			}), B(ft(fn, (e) => {
				K !== null && (clearTimeout(K), K = null), e ? K = window.setTimeout(() => {
					K = null, Ft(e);
				}, 100) : Ft(null);
			})), W(() => {
				K !== null && (clearTimeout(K), K = null);
			}), B(() => {
				let e = g.frozenElements;
				W(to$1(e));
			}), B(ft(E$4, (e) => {
				e && h.store.options.freezeReactUpdates && W(hs$1());
			}));
			let vn = () => {
				if (g.isTouchMode && D$4()) {
					let e = g.detectedElement;
					return !Z(e) || bi$2(e) ? void 0 : e;
				}
				let e = fn();
				if (e && !bi$2(e)) return e;
			}, yn = V(() => vn()), bn = () => yn() ? g.isTouchMode && D$4() ? Ht() : Ht() && !oe$3() : !1, xn = Ut(() => g.frozenElements, (e) => V(() => (S(), _t$2(e)))), Sn = V(() => {
				let e = g.frozenElements;
				if (e.length === 0) return [];
				let t = g.frozenDragRect;
				return t && e.length > 1 ? [Fi(t)] : xn().map((e) => e());
			}), Cn = V(() => {
				if (!xt() || g.pendingCommentMode || At()) return null;
				let e = g.detectedElement;
				return !Z(e) || bi$2(e) || g.frozenElements.includes(e) ? null : e;
			}), wn = V(() => {
				S();
				let e = Cn();
				if (e) return _t$2(e);
			}), Tn = V(() => {
				S();
				let e = g.frozenElements;
				if (e.length > 0) {
					let t = Sn();
					if (e.length === 1) {
						let e = t[0];
						if (e) return e;
					}
					let n = g.frozenDragRect;
					return n ? t[0] ?? Fi(n) : Li(na(t));
				}
				let t = yn();
				if (t) return _t$2(t);
			}), En = (e, t) => ({
				pageX: e + window.scrollX,
				pageY: t + window.scrollY
			}), Dn = (e, t) => {
				let { pageX: n, pageY: r } = En(e, t);
				return {
					x: Math.abs(n - g.dragStart.x),
					y: Math.abs(r - g.dragStart.y)
				};
			}, On = (e, t) => {
				let { pageX: n, pageY: r } = En(e, t), i = Math.min(g.dragStart.x, n), a = Math.min(g.dragStart.y, r), o = Math.abs(n - g.dragStart.x), s = Math.abs(r - g.dragStart.y);
				return {
					x: i - window.scrollX,
					y: a - window.scrollY,
					width: o,
					height: s
				};
			}, kn = (e) => e.code === `Space` || e.key === ` `, An = () => {
				if (!D$4()) return;
				y$4.startDragReposition();
				let { pageX: e, pageY: t } = En(b().x, b().y);
				bt = {
					x: e,
					y: t
				};
			}, jn = () => {
				y$4.stopDragReposition(), bt = null;
			}, Mn = V(() => {
				if (S(), !oe$3()) return;
				let e = On(b().x, b().y);
				return Li(e);
			}), Nn = V(() => {
				if (ht(), !oe$3()) return [];
				let e = pt();
				if (!e) return [];
				let t = On(e.x, e.y), n = Mi(t, ki$2);
				return n.length > 0 ? n : Mi(t, ki$2, !1);
			}), Pn = V(() => (S(), Nn().map((e) => _t$2(e)))), Fn = V(() => {
				let e = Pn();
				if (e.length > 0) return e;
				let t = wn();
				return t ? [...Sn(), t] : Sn();
			}), In = Ut(() => g.frozenElements, (e) => {
				let t = D(e) || `element`, n = Wr$2(e) ?? void 0;
				return { read: V(() => {
					if (S(), !Z(e)) return null;
					let r = _t$2(e), i = F.get(e), a = i === void 0 ? void 0 : r.x + r.width * i;
					return {
						tagName: t,
						componentName: n,
						bounds: r,
						mouseX: a
					};
				}) };
			}), Rn = V(() => {
				if (A() || g.frozenElements.length < 2) return [];
				let e = [];
				for (let t of In()) t.read() !== null && e.push(t);
				return e;
			}), zn = V(() => {
				if (A()) return null;
				let e = Cn();
				return e ? (S(), {
					tagName: D(e) || `element`,
					componentName: Wr$2(e) ?? void 0,
					bounds: _t$2(e),
					mouseX: b().x
				}) : null;
			}), Bn = V(() => {
				if (k() || A()) {
					S();
					let e = g.frozenElement || J();
					return e ? {
						x: ir(_t$2(e)).x + g.copyOffsetFromCenterX,
						y: g.copyStart.y
					} : {
						x: g.copyStart.x,
						y: g.copyStart.y
					};
				}
				return {
					x: b().x,
					y: b().y
				};
			}), Vn = V(() => {
				if (!xt() || g.frozenElements.length !== 1) return;
				S();
				let e = g.frozenElements[0];
				if (!Z(e)) return;
				let t = F.get(e);
				if (t === void 0) return;
				let n = _t$2(e);
				return n.x + n.width * t;
			});
			B(ft(() => [J(), g.lastGrabbedElement], ([e, t]) => {
				t && e && t !== e && y$4.setLastGrabbed(null), e && h.hooks.onElementHover(e);
			})), B(ft(() => J(), (e) => {
				let t = ++Ot, n = () => {
					Ot === t && y$4.setSelectionSource(null, null);
				};
				e ? Ur$2(e).then((e) => {
					Ot === t && (e ? y$4.setSelectionSource(e.filePath, e.lineNumber) : n());
				}).catch(() => {
					Ot === t && y$4.setSelectionSource(null, null);
				}) : n();
			}));
			let Hn = V(() => g.grabbedBoxes.map((e) => ({
				id: e.id,
				bounds: e.bounds,
				createdAt: e.createdAt
			}))), Un = V(() => g.labelInstances.map((e) => ({
				id: e.id,
				status: e.status,
				tagName: e.tagName,
				componentName: e.componentName,
				createdAt: e.createdAt
			})));
			B(ft(V(() => {
				let e = E$4(), t = D$4(), n = k(), r = A(), i = J(), a = Mn(), o = h.store.theme.enabled, s = h.store.theme.selectionBox.enabled, c = h.store.theme.dragBox.enabled, l = oe$3(), u = fn(), d = de();
				return {
					isActive: e,
					isDragging: t,
					isCopying: n,
					isPromptMode: r,
					isSelectionBoxVisible: !(!o || !s || !e || n || d || t || u == null),
					isDragBoxVisible: !!(o && c && e && !n && l),
					targetElement: i,
					dragBounds: a ? {
						x: a.x,
						y: a.y,
						width: a.width,
						height: a.height
					} : null,
					grabbedBoxes: [...Hn()],
					labelInstances: [...Un()],
					selectionFilePath: g.selectionFilePath,
					toolbarState: Ue()
				};
			}), (e) => {
				h.hooks.onStateChange(e);
			})), B(ft(() => {
				let e = A();
				return {
					inputMode: e,
					position: e ? b() : U(b),
					target: e ? J() : U(J)
				};
			}, ({ inputMode: e, position: t, target: n }) => {
				h.hooks.onPromptModeChange(e, {
					x: t.x,
					y: t.y,
					targetElement: n
				});
			})), B(ft(() => [
				_i(),
				Tn(),
				J()
			], ([e, t, n]) => {
				h.hooks.onSelectionBox(!!e, t ?? null, n);
			})), B(ft(() => [Oi(), Mn()], ([e, t]) => {
				h.hooks.onDragBox(!!e, t ?? null);
			})), B(ft(() => {
				let e = Ai();
				return [
					e,
					ki(),
					e ? Bn() : U(Bn),
					e ? J() : U(J),
					g.selectionFilePath,
					g.selectionLineNumber
				];
			}, ([e, t, n, r, i, a]) => {
				h.hooks.onElementLabel(e, t, {
					x: n.x,
					y: n.y,
					content: ``,
					element: r ?? void 0,
					tagName: r && D(r) || void 0,
					filePath: i ?? void 0,
					lineNumber: a ?? void 0
				});
			}));
			let Y = null, Wn = (e) => {
				if (e) {
					if (!Y) {
						Y = document.createElement(`style`), Y.setAttribute(`data-react-grab-cursor`, ``);
						let e = At$2();
						e && (Y.nonce = e), Mt$2(Y), document.head.appendChild(Y);
					}
					Y.textContent = `* { cursor: ${e} !important; }`;
				} else Y &&= (Y.remove(), null);
			};
			B(ft(() => [
				E$4(),
				k(),
				A()
			], ([e, t, n]) => {
				Wn(t ? `progress` : e && !n ? `crosshair` : null);
			}));
			let Gn = () => {
				let e = T();
				y$4.activate(), e || h.hooks.onActivate();
			}, X = () => {
				$t();
				let e = D$4(), t = g.previouslyFocusedElement;
				jn(), y$4.deactivate(), N.resetWithDiscard(), sa(), R(), Or(), I.clear(), jt(!1), Nt(null), e && (Te$2(`userSelect`), vt()), yt && window.clearTimeout(yt), Vt.stop(), y(t) && t !== document.body && t !== document.activeElement && Z(t) && t.focus({ preventScroll: !0 }), h.hooks.onDeactivate();
			}, Kn = () => {
				T() && y$4.releaseHold(), E$4() || k() ? X() : $t(), Dt();
			}, qn = () => {
				y$4.setWasActivatedByToggle(!0), Gn();
			}, Jn = () => {
				let e = [...g.frozenElements], t = g.frozenElement || J(), n = A() ? g.inputText.trim() : ``;
				if (!t) {
					X();
					return;
				}
				let r = e.length > 0 ? e : [t], i = r.map((e) => _t$2(e))[0], { x: a, y: o } = ir(i), s = a + g.copyOffsetFromCenterX;
				y$4.setPointer({
					x: a,
					y: o
				}), y$4.exitPromptMode(), y$4.clearInputText(), ln({
					element: t,
					cursorX: s,
					selectedElements: r,
					extraPrompt: n || void 0,
					shouldDeactivateAfter: !0
				});
			}, Yn = () => {
				A() && (be() ? (y$4.clearInputText(), X()) : (y$4.setPendingDismiss(!0), Re$2((e) => e + 1)));
			}, Xn = () => {
				I.isPendingDismiss() ? jr() : (y$4.clearInputText(), X());
			}, Zn = () => {
				y$4.setPendingDismiss(!1);
			}, Qn = () => {
				if (N.isOpen()) {
					N.dismiss();
					return;
				}
				let e = g.frozenElement || J();
				e && N.trigger(e, {
					x: b().x,
					y: b().y
				});
			}, $n = (e, t) => {
				if (!N.isOpen() || g.contextMenuPosition !== null) return !1;
				let n = Ta$1(e, t).find(ki$2);
				if (!n) return !1;
				let r = N.switchToElement(n, {
					x: e,
					y: t
				});
				return r && $a$1([n]), r;
			}, er = () => {
				kt = null, jt(!1), y$4.setPendingCommentMode(!1), Nt(null);
			}, tr = (e) => {
				let t = g.frozenElement || J();
				if (!t) return !1;
				let n = {
					x: b().x,
					y: b().y
				};
				y$4.clearInputText(), y$4.exitPromptMode(), er();
				let r = Lr(t, n);
				return Os({
					actionId: e,
					actions: h.store.actions,
					context: r
				}) || (ua("copy"), or(t, n)), !0;
			}, nr = (e) => {
				if (k()) X();
				else if (E$4()) {
					if (Rt() !== e && A() && tr(e)) return;
					if (Rt() !== e && g.pendingCommentMode) {
						y$4.setPendingCommentMode(!1), kt = e, Nt(e), jt(!0);
						return;
					}
					if (Rt() !== e && At()) {
						kt = e, Nt(e);
						return;
					}
					X();
				} else Oe() && (kt = e, Nt(e), jt(!0), qn());
			}, rr = () => {
				nr(Ue()?.defaultAction ?? `copy`);
			}, ar = (e, t, n) => {
				er(), y$4.clearInputText(), y$4.enterPromptMode({
					x: t,
					y: n
				}, e);
			}, or = (e, t) => {
				R(), ca(), y$4.showContextMenu(t, e), Or(), h.hooks.onContextMenu(e, t);
			}, sr = (e, t) => {
				let n = kt;
				if (kt = null, Nt(null), !n) return;
				let r = Lr(e, t);
				Os({
					actionId: n,
					actions: h.store.actions,
					context: r
				}) || (ua("copy"), or(e, t));
			}, cr = () => {
				Oe() && (E$4() && ge() ? X() : (y$4.setPendingCommentMode(!0), E$4() || qn()));
			}, lr = (e, t, n) => {
				let r = n && xt() && !D$4() && !g.pendingCommentMode && !At();
				if (L() || ae() && !r) return;
				if (y$4.setPointer({
					x: e,
					y: t
				}), ut.latestPointerX = e, ut.latestPointerY = t, r) {
					let n = Ea$1(e, t);
					n !== g.detectedElement && y$4.setDetectedElement(n);
					return;
				}
				let i = performance.now(), a = ut.pendingDetectionScheduledAt > 0 && i - ut.pendingDetectionScheduledAt < 200;
				if (!oe$3() && i - ut.lastDetectionTimestamp >= 32 && !a && (ut.lastDetectionTimestamp = i, ut.pendingDetectionScheduledAt = i, setTimeout(() => {
					if (L() || oe$3()) {
						ut.pendingDetectionScheduledAt = 0;
						return;
					}
					let e = Ea$1(ut.latestPointerX, ut.latestPointerY);
					e !== g.detectedElement && y$4.setDetectedElement(e), ut.pendingDetectionScheduledAt = 0;
				})), D$4()) {
					if (se()) {
						let { pageX: n, pageY: r } = En(e, t);
						bt && y$4.shiftDragStart({
							x: n - bt.x,
							y: r - bt.y
						}), bt = {
							x: n,
							y: r
						};
					}
					_t(e, t);
					let n = ns(e, t), r = n.top || n.bottom || n.left || n.right;
					r && !Vt.isActive() ? Vt.start() : !r && Vt.isActive() && Vt.stop();
				}
			}, ur = (e, t, n) => {
				if (!Ht() || le()) return !1;
				!n && xt() && R();
				let r = I.selectedElement() !== null;
				return y$4.startDrag({
					x: e,
					y: t
				}, n || r), y$4.setPointer({
					x: e,
					y: t
				}), Ce(`userSelect`, `none`), _t(e, t), h.hooks.onDragStart(e + window.scrollX, t + window.scrollY), !0;
			}, dr = (e, t) => {
				let n = g.frozenElements.includes(e), r = g.frozenElements.length === 0;
				if (!n) {
					let n = _t$2(e), i = Ni(n, t);
					if (F.set(e, i), r) {
						let t = Wr$2(e) ?? void 0;
						Lt(t);
					}
				}
				y$4.toggleFrozenElement(e), Da$1();
				let i = g.frozenElements.includes(e);
				i || F.delete(e), g.frozenElements.length === 0 ? (R(), y$4.unfreeze()) : ($a$1(g.frozenElements), St(!0), y$4.setPointer(t), y$4.setLastGrabbed(i ? e : g.frozenElements[g.frozenElements.length - 1]), y$4.freeze(), Or());
			}, fr = () => {
				let e = g.frozenElements.filter(Z), t = e.map((e) => {
					let t = D(e) || `element`, n = Wr$2(e) ?? void 0, r = F.get(e), i = _t$2(e);
					return {
						element: e,
						tagName: t,
						componentName: n,
						mouseX: r === void 0 ? i.x + i.width / 2 : i.x + i.width * r
					};
				});
				R(), e.length === 0 ? y$4.unfreeze() : e.length === 1 ? ln({
					element: e[0],
					cursorX: t[0].mouseX,
					selectedElements: e,
					shouldDeactivateAfter: g.wasActivatedByToggle
				}) : un({
					elements: e,
					labelEntries: t,
					shouldDeactivateAfter: g.wasActivatedByToggle
				});
			}, pr = (e, t, n) => {
				let r = Mi(e, ki$2), i = r.length > 0 ? r : Mi(e, ki$2, !1);
				if (i.length === 0) return;
				let a = n && !g.pendingCommentMode && !At();
				if (a && y$4.addFrozenElements(i), $a$1(a ? g.frozenElements : i), h.hooks.onDragEnd(i, e), a) {
					let e = i[i.length - 1];
					St(!0), Da$1(), y$4.setPointer(ir(_t$2(e))), y$4.setLastGrabbed(e), y$4.freeze(), Or();
					return;
				}
				let o = i[0], s = ir(_t$2(o));
				y$4.setPointer(s), y$4.setFrozenElements(i);
				let c = Ii(e);
				if (y$4.setFrozenDragRect(c), y$4.freeze(), y$4.setLastGrabbed(o), g.pendingCommentMode) {
					ar(o, s.x, s.y);
					return;
				}
				if (At()) {
					jt(!1), kt ? sr(o, s) : or(o, s);
					return;
				}
				let l = g.wasActivatedByToggle && !t;
				ln({
					element: o,
					cursorX: s.x,
					selectedElements: i,
					shouldDeactivateAfter: l,
					dragRect: c
				});
			}, mr = (e) => {
				for (let t of g.frozenElements) if (Z(t) && us(e, _t$2(t))) return t;
				return null;
			}, hr = (e, t, n, r) => {
				let i = Z(g.frozenElement) ? g.frozenElement : null, a = I.selectedElement(), o = () => Ta$1(e, t).find(ki$2) ?? null;
				if (r && !g.pendingCommentMode && !At()) {
					let n = o();
					n !== null && dr(n, {
						x: e,
						y: t
					});
					return;
				}
				let s = o() ?? (Z(g.detectedElement) ? g.detectedElement : null), c = a ?? s ?? i;
				if (!c) return;
				let l, u, d = s === null && i === c, f = a === c;
				if (d) l = b().x, u = b().y;
				else if (f) {
					let e = ir(_t$2(c));
					l = e.x, u = e.y;
				} else l = e, u = t;
				if (g.pendingCommentMode) {
					ar(c, l, u), I.clear();
					return;
				}
				if (At()) {
					jt(!1);
					let { wasIntercepted: e } = h.hooks.onElementSelect(c);
					if (e) return;
					I.clear(), $a$1([c]), y$4.setFrozenElement(c);
					let t = {
						x: l,
						y: u
					};
					y$4.setPointer(t), y$4.freeze(), kt ? sr(c, t) : or(c, t);
					return;
				}
				let p = g.wasActivatedByToggle && !n;
				y$4.setLastGrabbed(c), ln({
					element: c,
					cursorX: l,
					shouldDeactivateAfter: p
				}), I.clear();
			}, _r = () => {
				D$4() && (jn(), vt(), y$4.cancelDrag(), Vt.stop(), Te$2(`userSelect`), Jr());
			}, vr = (e, t, n, r) => {
				if (!D$4()) return;
				vt();
				let i = Dn(e, t), a = i.x > 2 || i.y > 2, o = a ? On(e, t) : null;
				a ? y$4.endDrag() : y$4.cancelDrag(), jn(), Vt.stop(), Te$2(`userSelect`), o ? pr(o, n, r) : hr(e, t, n, r);
			}, Q = fi(), Cr = Ls({
				shouldForwardInteraction: () => E$4() || T() || D$4() || le(),
				shouldForwardKeyboardEvent: (e) => Qi(e, h.store.options),
				shouldForwardViewportEvent: (e) => {
					let t = g.frozenElement ?? J();
					return t && Rs(e, t) ? !0 : g.frozenElements.some((t) => Rs(e, t));
				}
			}), wr = ts(), Dr = (e) => {
				let t;
				try {
					t = wr.originalKeyDescriptor?.get ? wr.originalKeyDescriptor.get.call(e) : e.key;
				} catch {
					return !1;
				}
				let n = t === `Enter` || ls(e.code), r = E$4() || T();
				return n && r && !A() && !I.isPendingDismiss() && !g.wasActivatedByToggle ? !$i(e, `data-react-grab-input`) && (wr.claimedEvents.add(e), e.preventDefault(), e.stopImmediatePropagation(), !0) : !1;
			};
			Q.addDocumentListener(`keydown`, zt$2(Dr), { capture: !0 }), Q.addDocumentListener(`keyup`, zt$2(Dr), { capture: !0 }), Q.addDocumentListener(`keypress`, zt$2(Dr), { capture: !0 });
			let Or = () => {
				Bt.clearHistory(), I.clear();
			}, kr = (e, t = !1) => {
				y$4.setFrozenElement(e), y$4.freeze(), I.select(e, { shouldPromptBeforeMouseHandoff: t });
				let n = ir(_t$2(e));
				y$4.setPointer(n), g.contextMenuPosition !== null && y$4.showContextMenu(n, e);
			}, Ar = () => {
				I.showDismissPrompt() && Re$2((e) => e + 1);
			}, jr = () => {
				I.clear(), y$4.unfreeze(), Or();
			}, Mr = () => {
				let e = I.takeSelection(g.frozenElement);
				if (!e) {
					jr();
					return;
				}
				let t = ir(_t$2(e));
				Or(), y$4.setLastGrabbed(e), ln({
					element: e,
					cursorX: t.x,
					shouldDeactivateAfter: g.wasActivatedByToggle
				});
			}, Nr = (e) => Te.has(e.key) ? e.key : e.key === `Tab` ? e.shiftKey ? `ArrowLeft` : `ArrowRight` : null, Pr = (e, t = {}) => {
				if (!E$4() || A() || xt() || I.isPendingDismiss() && !t.allowPendingKeyboardSelection) return !1;
				let n = Nr(e);
				if (!n || tt()) return !1;
				let r = fn(), i = !r;
				if (!r) {
					let e = Lt$2()?.getBoundingClientRect();
					r = e ? Ea$1(e.left + e.width / 2, e.top + e.height / 2) : Ea$1(window.innerWidth / 2, window.innerHeight / 2);
				}
				if (!r) return !1;
				let a = n === `ArrowUp` || n === `ArrowDown`, o = Bt.findNext(n, r);
				if (!o && !a && !i) return !1;
				let s = o ?? r;
				return e.preventDefault(), e.stopPropagation(), kr(s, !0), !0;
			}, Fr = (e) => e.metaKey || e.ctrlKey || e.altKey || e.repeat || yr(e) || !E$4() || k() || le() || tt() ? null : g.frozenElement || J(), Ir = (e) => {
				let t = Fr(e);
				if (!t) return null;
				let n = Ss(h.store.actions, e);
				return n ? {
					element: t,
					action: n
				} : null;
			}, Lr = (e, t) => {
				let n = _t$2(e);
				return Yi({
					element: e,
					filePath: g.selectionFilePath ?? void 0,
					lineNumber: g.selectionLineNumber ?? void 0,
					tagName: D(e) || void 0,
					componentName: It(),
					position: t,
					shouldDeferHideContextMenu: !1,
					performWithFeedbackOptions: {
						fallbackBounds: n,
						fallbackSelectionBounds: [n],
						position: t
					}
				});
			}, Rr = /^[a-zA-Z0-9-]$/, zr = (e) => {
				if (!e.key || e.key.length !== 1 || !Rr.test(e.key)) return !1;
				let t = Fr(e);
				return !t || !N.trigger(t, {
					x: b().x,
					y: b().y
				}, { initialSearchQuery: e.key }) ? !1 : (er(), e.preventDefault(), e.stopImmediatePropagation(), !0);
			}, Br = (e) => {
				let t = Ir(e);
				if (!t) return !1;
				let { element: n, action: r } = t;
				if (A()) return tr(r.id) ? (e.preventDefault(), e.stopImmediatePropagation(), !0) : !1;
				let i = {
					x: b().x,
					y: b().y
				}, a = Lr(n, i);
				return Ds(r, a) ? (e.preventDefault(), e.stopImmediatePropagation(), !0) : !1;
			}, Vr = () => {
				let e = g.selectionFilePath, t = g.selectionLineNumber;
				e && ta(e, t ?? void 0, h.hooks);
			}, Hr = (e) => e.key?.toLowerCase() !== `o` || !E$4() || !(e.metaKey || e.ctrlKey) || !g.selectionFilePath ? !1 : (e.preventDefault(), e.stopPropagation(), Vr(), !0), Ur = (e) => {
				if (!E$4() || k() || g.contextMenuPosition !== null || N.isOpen()) return !1;
				let t = e.key === `F10` && e.shiftKey, n = e.key === `ContextMenu`;
				if (!t && !n) return !1;
				let r = g.frozenElements, i = r.length > 1, a = (i ? r[0] : null) || g.frozenElement || J();
				if (!a) return !1;
				e.preventDefault(), e.stopPropagation();
				let o = ir(_t$2(a));
				return i ? $a$1(r) : ($a$1([a]), y$4.setFrozenElement(a)), y$4.setPointer(o), y$4.freeze(), or(a, o), !0;
			}, Wr = V(() => hn().map((e) => ({
				tagName: D(e.element) || `element`,
				componentName: Wr$2(e.element) ?? void 0,
				depth: e.depth,
				isLast: e.isLast
			}))), Gr = V(() => ({
				items: Wr(),
				activeIndex: gn()
			})), Kr = (e) => {
				if ((h.store.options.allowActivationInsideInput || !yr(e)) && !(!Qi(e, h.store.options) && ((e.metaKey || e.ctrlKey) && !we.includes(e.key) && !ls(e.code) && (E$4() && !g.wasActivatedByToggle ? X() : T() && (ot(), st(), y$4.releaseHold())), !ls(e.code) || !T()))) {
					if ((E$4() || T()) && !A() && (e.preventDefault(), ls(e.code) && e.stopImmediatePropagation()), E$4()) {
						if (g.wasActivatedByToggle && h.store.options.activationMode !== `hold` || e.repeat) return;
						yt !== null && window.clearTimeout(yt), yt = window.setTimeout(() => {
							X();
						}, 200);
					} else if (T() && e.repeat) {
						if (G.copyWaiting) {
							let e = G.holdTimerFired;
							st(), e && y$4.activate();
						}
					} else if (!(k() || de()) && !T()) {
						let t = h.store.options.keyHoldDuration ?? 100;
						yr(e) ? br(e) ? t += 600 : t += 400 : xr() && (t += 600), st(), y$4.startHold(t);
					}
				}
			}, qr = (e) => {
				if (!I.isPendingDismiss()) return !1;
				if (ls(e.code)) return !0;
				let t = Te.has(e.key), n = Ir(e) !== null;
				return !t && !n ? !1 : t ? Pr(e, { allowPendingKeyboardSelection: !0 }) : Br(e) ? (Or(), !0) : !1;
			};
			Q.addWindowListener(`keydown`, zt$2((e) => {
				if (Nr(e) && yr(e) || qr(e)) return;
				if (Dr(e), !Oe()) {
					Qi(e, h.store.options) && !e.repeat && Fe$1((e) => e + 1);
					return;
				}
				let t = ls(e.code) && T() && !A(), n = $i(e, je);
				if (A() && Qi(e, h.store.options) && !e.repeat && !n) {
					e.preventDefault(), e.stopPropagation(), Yn();
					return;
				}
				if (e.key === `Escape` && k()) {
					X();
					return;
				}
				if (e.key === `Escape` && tt()) {
					Ye() !== null && sa();
					return;
				}
				let r = $i(e, `data-react-grab-ignore-events`) && !t;
				if (A() || r) {
					A() && !n && Br(e) || (e.key === `Escape` && (A() ? Yn() : g.wasActivatedByToggle && X()), r && Te.has(e.key) && Pr(e));
					return;
				}
				if (D$4() && kn(e)) {
					e.repeat || An(), e.preventDefault(), e.stopPropagation();
					return;
				}
				if (e.key === `Escape` && (T() || g.wasActivatedByToggle)) {
					X();
					return;
				}
				E$4() && !we.includes(e.key) && e.preventDefault();
				let i = Date.now() - Ct < 200;
				Pr(e) || Hr(e) || Ur(e) || Br(e) || zr(e) || i || Kr(e);
			}), { capture: !0 }), Q.addWindowListener(`keyup`, zt$2((e) => {
				if (Dr(e)) return;
				if (kn(e) && se() && (jn(), e.preventDefault(), e.stopPropagation()), e.key === `Shift` && xt()) {
					D$4() && _r(), fr();
					return;
				}
				if ($i(e, `data-react-grab-ignore-events`)) return;
				let t = Zi(h.store.options.activationKey), n = t.metaKey || t.ctrlKey ? Ki() ? !e.metaKey : !e.ctrlKey : t.shiftKey && !e.shiftKey || t.altKey && !e.altKey, r = h.store.options.activationKey ? Xi(h.store.options.activationKey)(e) : Ui(e.key, e.code);
				if (de() || wt) {
					(r || n) && (Dt(), X());
					return;
				}
				if (!T() && !E$4() || A()) return;
				let i = !!h.store.options.activationKey, a = h.store.options.activationMode === `hold`, o = D$4();
				if (E$4()) {
					let e = et();
					if (n) {
						if (g.wasActivatedByToggle && h.store.options.activationMode !== `hold` || e) return;
						X();
					} else if (a && r) {
						if (yt !== null && (window.clearTimeout(yt), yt = null), e || o) return;
						X();
					} else !i && r && yt !== null && (window.clearTimeout(yt), yt = null);
				} else if (r || n) {
					if (g.wasActivatedByToggle && h.store.options.activationMode !== `hold`) return;
					if (T() || G.holdTimerFired && n) {
						ot();
						let t = (G.startTimestamp ? Date.now() - G.startTimestamp : 0) >= 200, n = G.holdTimerFired && t && (h.store.options.allowActivationInsideInput || !yr(e));
						st(), n ? y$4.activate() : y$4.releaseHold();
					} else X();
				}
			}), { capture: !0 }), Q.addDocumentListener(`copy`, () => {
				T() && (G.copyWaiting = !0);
			}), Q.addWindowListener(`keypress`, zt$2(Dr), { capture: !0 }), Q.addWindowListener(`pointermove`, zt$2((e) => {
				if (!e.isPrimary) return;
				let t = e.pointerType === `touch`;
				if (y$4.setTouchMode(t), !$i(e, `data-react-grab-ignore-events`) && !L() && (!t || T() || E$4())) {
					if ((t ? T() : E$4()) && !A() && ae() && !e.shiftKey && !xt()) {
						if (I.consumeMouseHandoff()) {
							Ar();
							return;
						}
						y$4.unfreeze(), Or();
					}
					lr(e.clientX, e.clientY, e.shiftKey);
				}
			}), {
				passive: !0,
				capture: !0
			}), Q.addWindowListener(`pointerdown`, zt$2((e) => {
				if (e.button === 0 && e.isPrimary && (y$4.setTouchMode(e.pointerType === `touch`), P = !1, D$4() || !$i(e, `data-react-grab-ignore-events`))) {
					if (et()) $n(e.clientX, e.clientY) && (P = !0, e.preventDefault(), e.stopImmediatePropagation());
					else if (A()) {
						let t = Tn();
						t && e.clientX >= t.x && e.clientX <= t.x + t.width && e.clientY >= t.y && e.clientY <= t.y + t.height ? Jn() : Yn();
					} else if (I.isPendingDismiss()) e.preventDefault(), e.stopImmediatePropagation();
					else if (le()) e.preventDefault(), e.stopImmediatePropagation();
					else if (ur(e.clientX, e.clientY, e.shiftKey)) {
						if (e.pointerId !== void 0) try {
							document.documentElement.setPointerCapture(e.pointerId);
						} catch {}
						e.preventDefault(), e.stopImmediatePropagation();
					}
				}
			}), { capture: !0 }), Q.addWindowListener(`pointerup`, zt$2((e) => {
				if (e.button !== 0 || !e.isPrimary || $i(e, `data-react-grab-ignore-events`) || et()) return;
				let t = Ht() || le() || D$4(), n = e.metaKey || e.ctrlKey;
				vr(e.clientX, e.clientY, n, e.shiftKey), t && (e.preventDefault(), e.stopImmediatePropagation());
			}), { capture: !0 }), Q.addWindowListener(`contextmenu`, zt$2((e) => {
				if (!Ht() || k() || A() || N.isOpen()) return;
				let t = $i(e, `data-react-grab-ignore-events`), n = {
					x: e.clientX,
					y: e.clientY
				}, r = t && g.frozenElements.length > 1 ? mr(n) : null, i = I.isPendingDismiss();
				if (t && !r && !i && !_n()) return;
				if (et()) {
					e.preventDefault();
					return;
				}
				e.preventDefault(), e.stopPropagation();
				let a = r ?? Ea$1(e.clientX, e.clientY);
				if (!a) return;
				let o = g.frozenElements;
				o.length > 1 && o.includes(a) ? $a$1(o) : ($a$1([a]), y$4.setFrozenElement(a)), y$4.setPointer(n), y$4.freeze(), or(a, n);
			}), { capture: !0 }), Q.addWindowListener(`pointercancel`, zt$2((e) => {
				e.isPrimary && _r();
			}), { capture: !0 }), Q.addWindowListener(`click`, zt$2((e) => {
				$i(e, `data-react-grab-ignore-events`) || (P ? (P = !1, e.preventDefault(), e.stopImmediatePropagation()) : et() || (Ht() || ce()) && (e.preventDefault(), e.stopImmediatePropagation(), g.wasActivatedByToggle && !A() && !e.shiftKey && (T() ? y$4.setWasActivatedByToggle(!1) : X())));
			}), { capture: !0 }), Q.addDocumentListener(`visibilitychange`, () => {
				if (document.hidden) {
					y$4.clearGrabbedBoxes();
					let e = g.activationTimestamp;
					E$4() && !A() && e !== null && Date.now() - e > 500 && X();
				}
			}), Q.addWindowListener(`blur`, () => {
				_r(), T() && (ot(), y$4.releaseHold(), st()), R();
			}), Q.addWindowListener(`focus`, () => {
				Ct = Date.now();
			}), Q.addWindowListener(`focusin`, (e) => {
				$i(e, Sr) && e.stopPropagation();
			}, { capture: !0 });
			let Jr = () => {
				if ((!g.isTouchMode || T() || E$4()) && !L() && !ae() && !D$4() && g.frozenElements.length === 0) {
					let e = Ea$1(b().x, b().y);
					y$4.setDetectedElement(e);
				}
			}, Yr = null, Xr = null, Zr = () => {
				Pi(), Jr(), gt((e) => e + 1), y$4.incrementViewportVersion(), y$4.updateContextMenuPosition();
			}, Qr = () => {
				Oi$2(), Zr();
			};
			Q.addWindowListener(`scroll`, Zr, { capture: !0 });
			let $r = window.innerWidth, ei = window.innerHeight;
			Q.addWindowListener(`resize`, () => {
				let e = window.innerWidth, t = window.innerHeight;
				if ($r > 0 && ei > 0) {
					let n = e / $r, r = t / ei, i = Math.abs(n - r) < Re, a = Math.abs(n - 1) > Re;
					i && a && y$4.setPointer({
						x: b().x * n,
						y: b().y * r
					});
				}
				$r = e, ei = t, Qr();
			});
			let ti = window.visualViewport;
			if (ti) {
				let { signal: e } = Q;
				ti.addEventListener(`resize`, Qr, { signal: e }), ti.addEventListener(`scroll`, Zr, { signal: e });
			}
			let ni = () => {
				Xr === null && (Xr = M$1(() => {
					Xr = null, y$4.incrementViewportVersion();
				}));
			};
			B(() => {
				if (h.store.theme.enabled && (E$4() || k() || g.labelInstances.length > 0 || g.grabbedBoxes.length > 0)) {
					if (Yr !== null) return;
					Yr = window.setInterval(() => {
						y$4.relinkLiveElements(), ni();
					}, 100);
				} else Yr !== null && (window.clearInterval(Yr), Yr = null), Xr !== null && (ci$2(Xr), Xr = null);
			}), W(() => {
				Yr !== null && window.clearInterval(Yr), Xr !== null && ci$2(Xr);
			}), Q.addDocumentListener(`copy`, zt$2((e) => {
				A() || $i(e, `data-react-grab-ignore-events`) || Ht() && e.preventDefault();
			}), { capture: !0 }), W(() => {
				let e = [];
				$(Cr, e), $(() => Q.abort(), e), H !== null && window.clearTimeout(H), yt && window.clearTimeout(yt), $(Dt, e), rt && $(rt, e), rt = null, it && $(it, e), it = null, Wt.forEach((e) => window.clearTimeout(e)), Wt.clear(), $(q.cancelAllFades, e), Jt.clear(), $(Vt.stop, e), $(Po$1, e), $(() => Te$2(`userSelect`), e), $(() => Te$2(`touchAction`), e), $(() => Wn(null), e), $(wr.restore, e), $(() => It$2(null), e), q$1(e, `Disposing React Grab failed`);
			});
			let { root: ri, host: ii, cancelPendingAttachment: ai } = Tr(Ln);
			W(ai), W(li(ii).cleanup);
			let oi = V(() => h.store.theme.enabled), si = V(() => h.store.theme.selectionBox.enabled), ci = V(() => h.store.theme.elementLabel.enabled), ui = V(() => h.store.theme.dragBox.enabled), di = V(() => de() || Ke() && !ae() || N.isInteracting()), mi = V(() => Pn().length > 0), _i = V(() => !oi() || !si() || di() ? !1 : mi() ? !0 : bn()), vi = V(() => {
				let e = yn();
				if (e) return D(e) || void 0;
			}), yi = V(() => !oi() || et() || !ci() || di() ? !1 : bn()), bi = /* @__PURE__ */ new Map(), xi = (e) => {
				let t = e.elements?.filter(Z) ?? [], n = e.element, r = null;
				t.length > 0 ? r = t.map(_t$2) : n && Z(n) && (r = [_t$2(n)]);
				let i = e.bounds, a = e.boundsMultiple;
				r && (i = r.length > 1 ? Li(na(r)) : r[0], e.boundsMultiple !== void 0 && (a = e.boundsMultiple.length > 1 && e.boundsMultiple.length === e.elements?.length ? r : [i]));
				let o = bi.get(e.id), s = o?.boundsMultiple, c = s === a || s !== void 0 && a !== void 0 && s.length === a.length && s.every((e, t) => e.x === a[t].x && e.y === a[t].y && e.width === a[t].width && e.height === a[t].height);
				if (o && o.status === e.status && o.errorMessage === e.errorMessage && o.bounds.x === i.x && o.bounds.y === i.y && o.bounds.width === i.width && o.bounds.height === i.height && c) return o;
				let l = i.x + i.width / 2, u = i.width / 2, d;
				d = e.mouseXOffsetRatio !== void 0 && u > 0 ? l + e.mouseXOffsetRatio * u : e.mouseXOffsetFromCenter === void 0 ? e.mouseX ?? l : l + e.mouseXOffsetFromCenter;
				let f = {
					...e,
					bounds: i,
					boundsMultiple: a,
					mouseX: d
				};
				return bi.set(e.id, f), f;
			}, Ci = V(() => {
				if (!oi() || !h.store.theme.grabbedBoxes.enabled) return [];
				S();
				let e = new Set(g.labelInstances.map((e) => e.id));
				for (let t of bi.keys()) e.has(t) || bi.delete(t);
				return g.labelInstances.map(xi);
			}), wi = V(() => new Map(Ci().map((e) => [e.id, e]))), Ti = /* @__PURE__ */ new Map(), Ei = V(() => {
				let e = Ci(), t = new Set(e.map((e) => e.id));
				for (let e of Ti.keys()) t.has(e) || Ti.delete(e);
				return e.map((e) => {
					let t = Ti.get(e.id);
					if (t) return t;
					let n = e.id, r = { read: () => wi().get(n) ?? null };
					return Ti.set(n, r), r;
				});
			}), Di = V(() => !oi() || !h.store.theme.grabbedBoxes.enabled ? [] : (S(), g.grabbedBoxes.map((e) => Z(e.element) ? {
				...e,
				bounds: _t$2(e.element)
			} : e))), Oi = V(() => oi() && ui() && Ht() && oe$3()), ki = V(() => k() ? `processing` : `hover`), Ai = V(() => {
				if (!oi()) return !1;
				let e = ci(), t = A(), n = k(), r = Ht(), i = D$4(), a = !!fn(), o = Ke(), s = ae();
				return !e || t || o && !s ? !1 : n ? !0 : r && !i && a;
			}), ji = V(() => {
				S();
				let e = g.contextMenuElement;
				return e ? _t$2(e) : null;
			}), zi = V(() => (S(), g.contextMenuPosition)), Bi = V(() => {
				let e = g.contextMenuElement;
				if (!e) return;
				let t = g.frozenElements.length;
				return t > 1 ? `${t} elements` : D(e) || void 0;
			}), [Hi] = Er(() => g.frozenElements.length > 1 ? null : g.contextMenuElement), [Wi] = dt(() => g.contextMenuElement, async (e) => e ? Ur$2(e) : null), Gi = async (e) => {
				y$4.incrementSelectionInteractionLockDepth();
				try {
					return await e();
				} finally {
					y$4.decrementSelectionInteractionLockDepth();
				}
			}, qi = (e, t, n, r, i) => async (a) => {
				await Gi(async () => {
					let o = i?.fallbackBounds ?? null, s = i?.fallbackSelectionBounds ?? [], c = i?.position ?? g.contextMenuPosition ?? b(), l = Sn(), u = ji() ?? o, d = t.length > 1, f = d ? Li(na(l)) : u, p = g.wasActivatedByToggle, m;
					if (m = d ? l : u ? [u] : s, y$4.hideContextMenu(), f) {
						let i = d ? f.x + f.width / 2 : c.x, o = q.createInstance(f, n || `element`, r, `copying`, {
							element: e,
							mouseX: i,
							elements: d ? t : void 0,
							boundsMultiple: m
						}), s = !1, l;
						try {
							s = await a(), s || (l = `Failed to copy`);
						} catch (e) {
							l = pi(e, `Action failed`);
						}
						q.updateAfterCopy(o, s, l);
					} else try {
						await a();
					} catch (e) {
						H$1(new E(`Action failed without feedback bounds`, e));
					}
					p ? X() : y$4.unfreeze();
				});
			}, Ji = () => {
				setTimeout(() => {
					y$4.hideContextMenu();
				}, 0);
			}, Yi = (e) => {
				let { element: t, filePath: n, lineNumber: r, tagName: i, componentName: a, position: o, performWithFeedbackOptions: s, shouldDeferHideContextMenu: c, onBeforeCopy: l, onBeforePrompt: u, customEnterPromptMode: d } = e, f = g.frozenElements.length > 0 ? g.frozenElements : [t], p = c ? Ji : y$4.hideContextMenu, m = {
					element: t,
					elements: f,
					filePath: n,
					lineNumber: r,
					componentName: a,
					tagName: i,
					enterPromptMode: d ?? (() => {
						q.clearAll(), er(), u?.(), ct(t, o.x, o.y), y$4.setPointer({
							x: o.x,
							y: o.y
						}), y$4.setFrozenElement(t), lt(), E$4() || Gn(), p();
					}),
					enterEditMode: () => {
						N.trigger(t, o, {
							filePath: n,
							lineNumber: r,
							componentName: a,
							tagName: i
						}) && er(), p();
					},
					copy: () => {
						er(), l?.(), ln({
							element: t,
							cursorX: o.x,
							selectedElements: f.length > 1 ? f : void 0,
							shouldDeactivateAfter: g.wasActivatedByToggle
						}), p();
					},
					hooks: {
						transformHtmlContent: h.hooks.transformHtmlContent,
						onOpenFile: h.hooks.onOpenFile,
						transformOpenFileUrl: h.hooks.transformOpenFileUrl
					},
					performWithFeedback: qi(t, f, i, a, s),
					hideContextMenu: p,
					cleanup: () => {
						g.wasActivatedByToggle ? X() : y$4.unfreeze();
					}
				}, _ = h.hooks.transformActionContext(m);
				return {
					...m,
					..._
				};
			}, ea = V(() => {
				let e = g.contextMenuElement;
				if (!e) return;
				let t = Wi(), n = g.contextMenuPosition ?? b();
				return Yi({
					element: e,
					filePath: t?.filePath,
					lineNumber: t?.lineNumber ?? void 0,
					tagName: Bi(),
					componentName: Hi(),
					position: n,
					shouldDeferHideContextMenu: !0,
					onBeforeCopy: () => {
						I.clear();
					},
					customEnterPromptMode: () => {
						q.clearAll(), er(), y$4.clearInputText(), y$4.enterPromptMode(n, e), Ji();
					}
				});
			}), ra = () => {
				setTimeout(() => {
					y$4.hideContextMenu(), X();
				}, 0);
			}, ia = () => {
				if (!nt) return null;
				let e = nt.getBoundingClientRect(), t = ys(e);
				return t === `left` || t === `right` ? {
					x: t === `left` ? e.right : e.left,
					y: e.top + e.height / 2,
					edge: t
				} : {
					x: e.left + e.width / 2,
					y: t === `top` ? e.bottom : e.top,
					edge: t
				};
			}, aa = () => {
				let e = ia();
				if (e) return e;
				let t = N.state();
				return t ? {
					x: t.position.x,
					y: t.position.y,
					edge: `bottom`
				} : null;
			}, oa = (e, t) => {
				let n = null, r = () => {
					let i = e();
					i && t(i), n = M$1(r);
				};
				return r(), () => {
					n !== null && (ci$2(n), n = null);
				};
			};
			B(() => {
				if (!_n()) return;
				let e = oa(ia, $e);
				W(() => {
					e(), $e(null);
				});
			});
			let sa = () => {
				rt?.(), rt = null, M(null);
			}, ca = () => {
				y$4.hideContextMenu(), sa(), N.dismiss();
			}, la = () => {
				Ye() === null ? (y$4.hideContextMenu(), N.isOpen() && N.closePreservingRenderer(), rt?.(), rt = oa(ia, M)) : sa();
			}, ua = (e) => {
				at({ defaultAction: e });
			}, da = (e) => {
				let t = g.labelInstances.find((t) => t.id === e);
				if (!t?.element || !Z(t.element)) return;
				let n = t.element, r = ir(_t$2(n)), i = {
					x: t.mouseX ?? r.x,
					y: r.y
				}, a = t.elements && t.elements.length > 0 ? t.elements.filter((e) => Z(e)) : [n];
				setTimeout(() => {
					sa(), N.isOpen() && N.closePreservingRenderer(), E$4() || (y$4.setWasActivatedByToggle(!0), Gn()), y$4.setPointer(i), y$4.setFrozenElements(a), a.length > 1 && t.bounds && y$4.setFrozenDragRect(Ii(t.bounds)), y$4.freeze(), y$4.showContextMenu(i, n);
				}, 0);
			};
			B(() => {
				let e = h.store.theme.hue;
				e === 0 ? ri.style.filter = `` : ri.style.filter = `hue-rotate(${e}deg)`;
			}), h.store.theme.enabled && Promise.resolve().then(() => (init_renderer_CyobGvMH(), renderer_CyobGvMH_exports)).then(({ ReactGrabRenderer: e }) => {
				t || (s = pn(() => Gt(e, {
					get selectionVisible() {
						return _i();
					},
					get selectionBounds() {
						return Tn();
					},
					get selectionBoundsMultiple() {
						return Fn();
					},
					get selectionShouldSnap() {
						return g.frozenElements.length > 0 || Pn().length > 0;
					},
					get selectionElementsCount() {
						return g.frozenElements.length;
					},
					get frozenLabelEntryAccessors() {
						return Rn();
					},
					get pendingShiftPreviewEntry() {
						return zn() ?? void 0;
					},
					get selectionFilePath() {
						return g.selectionFilePath ?? void 0;
					},
					get selectionLineNumber() {
						return g.selectionLineNumber ?? void 0;
					},
					get selectionTagName() {
						return vi();
					},
					get selectionComponentName() {
						return It();
					},
					get selectionLabelVisible() {
						return yi();
					},
					selectionLabelStatus: `idle`,
					get hierarchyState() {
						return Gr();
					},
					get hierarchyMenuPosition() {
						return Qe();
					},
					get labelInstances() {
						return Ci();
					},
					get labelInstanceAccessors() {
						return Ei();
					},
					get dragVisible() {
						return Oi();
					},
					get dragBounds() {
						return Mn();
					},
					get grabbedBoxes() {
						return Di();
					},
					get mouseX() {
						return dn(() => g.frozenElements.length > 1)() ? void 0 : Vn() ?? Bn().x;
					},
					get isFrozen() {
						return ae() || E$4() || Ke();
					},
					get inputValue() {
						return g.inputText;
					},
					get isPromptMode() {
						return A();
					},
					onShowContextMenuInstance: da,
					onRetryInstance: rn,
					onAcknowledgeErrorInstance: an,
					get onLabelInstanceHoverChange() {
						return q.handleHoverChange;
					},
					get onInputChange() {
						return y$4.setInputText;
					},
					onInputSubmit: () => void Jn(),
					onToggleExpand: Qn,
					get selectionLabelShakeCount() {
						return Le();
					},
					onConfirmDismiss: Xn,
					onOpenSelectionFile: Vr,
					get discardPrompt() {
						return dn(() => !!I.isPendingDismiss())() ? {
							isKeyboardSelection: !0,
							onConfirm: Xn,
							onCopy: Mr
						} : be() ? {
							onConfirm: Xn,
							onCancel: Zn
						} : void 0;
					},
					get toolbarVisible() {
						return h.store.theme.toolbar.enabled;
					},
					get isActive() {
						return E$4();
					},
					onToggleActive: rr,
					onActivateAction: nr,
					get activeActionId() {
						return Rt();
					},
					get enabled() {
						return Oe();
					},
					get shakeCount() {
						return je$2();
					},
					onToolbarStateChange: (e) => {
						Ge(e), e.enabled !== Oe() && (ke(e.enabled), e.enabled || (Kn(), ca())), As(d, e);
					},
					onSubscribeToToolbarStateChanges: (e) => (d.add(e), () => {
						d.delete(e);
					}),
					onToolbarSelectHoverChange: qe,
					onToolbarRef: (e) => {
						nt = e;
					},
					get contextMenuPosition() {
						return zi();
					},
					get contextMenuBounds() {
						return ji();
					},
					get contextMenuTagName() {
						return Bi();
					},
					get contextMenuComponentName() {
						return Hi();
					},
					get contextMenuHasFilePath() {
						return !!Wi()?.filePath;
					},
					get actions() {
						return h.store.actions;
					},
					get actionContext() {
						return ea();
					},
					onContextMenuDismiss: ra,
					onContextMenuHide: Ji,
					get toolbarMenuPosition() {
						return Ye();
					},
					get toolbarMenuActions() {
						return h.store.actions.filter((e) => e.showInToolbarMenu === !0);
					},
					get toolbarButtons() {
						return h.store.toolbarButtons;
					},
					get defaultActionId() {
						return Ue()?.defaultAction ?? `copy`;
					},
					onSetDefaultAction: ua,
					onToggleToolbarMenu: la,
					onToolbarMenuDismiss: sa,
					get editPanelState() {
						return N.state();
					},
					get editPanelPosition() {
						return Xe();
					},
					get onEditPanelDismiss() {
						return N.dismiss;
					},
					get onEditPanelSubmit() {
						return N.submit;
					},
					get onEditPanelPendingEditsChange() {
						return N.setPendingEdits;
					},
					get onEditPanelInteractingChange() {
						return N.setInteracting;
					}
				}), ri));
			}).catch((e) => {
				console.warn(`[react-grab] Failed to load renderer:`, e);
			});
			let fa = {
				activate: () => {
					y$4.setPendingCommentMode(!1), !E$4() && Oe() && qn();
				},
				deactivate: () => {
					E$4() || k() ? X() : $t();
				},
				toggle: () => {
					E$4() || k() ? X() : Oe() && qn();
				},
				comment: cr,
				isActive: () => E$4(),
				isEnabled: () => Oe(),
				setEnabled: (e) => {
					e !== Oe() && (ke(e), at({
						enabled: e,
						collapsed: !e
					}), e || (Kn(), ca()));
				},
				getToolbarState: () => Ue() ?? fs(),
				setToolbarState: (e) => {
					let t = Ue() ?? fs(), n = e.collapsed ?? t?.collapsed ?? !1, r = {
						edge: e.edge ?? t?.edge ?? `bottom`,
						ratio: e.ratio ?? t?.ratio ?? .5,
						collapsed: n,
						enabled: e.enabled ?? !n,
						defaultAction: e.defaultAction ?? t?.defaultAction ?? `copy`
					};
					ps(r), Ge(r), r.enabled !== Oe() && (ke(r.enabled), r.enabled || (Kn(), ca())), As(d, r);
				},
				onToolbarStateChange: (e) => (d.add(e), () => {
					d.delete(e);
				}),
				reset: () => {
					Kn(), ca(), y$4.clearGrabbedBoxes(), q.clearAll(), y$4.setSelectionSource(null, null);
				},
				dispose: () => {
					if (t) return;
					let n = [];
					t = !0, Ys = !1, $($t, n), $(N.resetWithDiscard, n), $(q.clearAll, n), s && $(s, n), s = void 0, rt && $(rt, n), rt = null, it && $(it, n), it = null, d.clear(), $(e, n), $(() => Us(fa), n), q$1(n, `Disposing React Grab failed`);
				},
				copyElement: async (e) => {
					let t = Array.isArray(e) ? e : [e];
					return t.length !== 0 && (await on(t, Zt())).status === `succeeded`;
				},
				getSource: async (e) => {
					let t = await Ur$2(e);
					return t ? {
						filePath: t.filePath,
						lineNumber: t.lineNumber,
						columnNumber: t.columnNumber ?? null,
						componentName: t.componentName
					} : null;
				},
				getStackContext: (e) => ti$2(e, { maxLines: h.store.options.maxContextLines }),
				getState: () => ({
					isActive: E$4(),
					isDragging: D$4(),
					isCopying: k(),
					isPromptMode: A(),
					isSelectionBoxVisible: !!_i(),
					isDragBoxVisible: !!Oi(),
					targetElement: J(),
					dragBounds: Mn() ?? null,
					grabbedBoxes: [...Hn()],
					labelInstances: [...Un()],
					selectionFilePath: g.selectionFilePath,
					toolbarState: Ue()
				}),
				setOptions: (e) => {
					h.setOptions(e);
				},
				registerPlugin: (e) => {
					h.register(e, fa);
				},
				unregisterPlugin: (e) => {
					h.unregister(e);
				},
				getPlugins: () => h.getPluginNames(),
				getDisplayName: Wr$2
			};
			for (let e of qs) h.register(e, fa);
			return setTimeout(() => {
				Bn$2(!0);
			}, He), fa;
		});
	};
}));
// src/sw/profile.ts
function record(value, path) {
	if (!value || typeof value !== "object" || Array.isArray(value)) throw new Error(`${path}: expected an object`);
	return value;
}
function keys(value, allowed, path) {
	for (const key of Object.keys(value)) if (!allowed.includes(key)) throw new Error(`${path}.${key}: unknown key`);
}
// src/markers.ts
const EDITABLE_SOURCE_ATTRIBUTE = "data-source";
const MARKER = /^(.+):(\d+):(\d+)#(\d+)@([0-9a-f]{12})$/;
function formatEditableSource({ file, line, column, ordinal, hash }) {
	if (!file || /["\\\r\n]/.test(file)) throw new Error(`cannot mark editable source in ${JSON.stringify(file)}: unsupported file name`);
	if (![line, column].every((value) => Number.isSafeInteger(value) && value > 0) || !Number.isSafeInteger(ordinal) || ordinal < 0) throw new Error("invalid editable source position or ordinal");
	if (!/^[0-9a-f]{12}$/.test(hash)) throw new Error(`invalid source hash ${JSON.stringify(hash)}`);
	return `${file}:${line}:${column}#${ordinal}@${hash}`;
}
function parseEditableSource(value) {
	const match = MARKER.exec(value);
	if (!match) throw new Error(`malformed ${EDITABLE_SOURCE_ATTRIBUTE} ${JSON.stringify(value)}; expected <file>:<line>:<column>#<ordinal>@<hash>`);
	const marker = {
		file: match[1],
		line: Number(match[2]),
		column: Number(match[3]),
		ordinal: Number(match[4]),
		hash: match[5]
	};
	formatEditableSource(marker);
	return marker;
}
// src/host/agent-source.ts
function isPageSource(source) {
	if (/^(?:file|data|blob|javascript):/i.test(source)) return false;
	if (source.includes("#")) try {
		parseEditableSource(source);
	} catch {
		return false;
	}
	else if (!/^.+:[1-9]\d*:[1-9]\d*$/.test(source)) return false;
	return !/(?:^|\/)(?:node_modules|npm|link|entry|__runtime)\/|(?:^|\/)packages\/(?:ui|react-grab|facet)\//.test(source);
}
function validateDocumentSource(url) {
	const parsed = new URL(url);
	if (![
		"http:",
		"https:",
		"file:"
	].includes(parsed.protocol) || parsed.href !== url || parsed.username || parsed.password) throw new Error("external-document source requires a canonical credential-free HTTP(S) or file document URL");
}
// src/host/message-values.ts
function text$1(value, max = 4096, empty = false) {
	if (typeof value !== "string" || !empty && !value || value.length > max) throw new Error("agent: invalid text");
}
function number(value, min = -Infinity) {
	if (typeof value !== "number" || !Number.isFinite(value) || value < min) throw new Error("agent: invalid number");
}
function list(value) {
	if (!Array.isArray(value) || value.length > 1e4) throw new Error("agent: invalid list");
	return value;
}
function rect(value) {
	const r = record(value, "agent.rect");
	keys(r, [
		"left",
		"top",
		"width",
		"height"
	], "agent.rect");
	number(r.left);
	number(r.top);
	number(r.width, 0);
	number(r.height, 0);
}
function locator(value) {
	const l = record(value, "agent.locator");
	if (l.kind === "css") {
		keys(l, ["kind", "selector"], "locator");
		text$1(l.selector, 1024);
		return;
	}
	text$1(l.tag, 100);
	if (l.kind === "role") {
		keys(l, [
			"kind",
			"tag",
			"role",
			"name"
		], "locator");
		text$1(l.role, 100);
		text$1(l.name, 200);
		return;
	}
	text$1(l.text, 200, true);
	if (l.kind === "text") {
		keys(l, [
			"kind",
			"tag",
			"text",
			"heading",
			"scope"
		], "locator");
		if (l.scope !== void 0) text$1(l.scope, 1024);
		if (l.heading !== void 0) {
			const h = record(l.heading, "heading");
			keys(h, ["selector", "text"], "heading");
			text$1(h.selector, 1024);
			text$1(h.text, 200);
		}
	} else if (l.kind === "markdown") {
		keys(l, [
			"kind",
			"tag",
			"text",
			"path",
			"line"
		], "locator");
		list(l.path).forEach((part) => text$1(part, 200));
		if (l.line !== void 0) {
			number(l.line, 1);
			if (!Number.isInteger(l.line)) throw new Error("locator: invalid line");
		}
	} else throw new Error("agent: invalid locator kind");
}
function validateAgentSelection(value) {
	const s = record(value, "selection");
	keys(s, [
		"anchor",
		"label",
		"element",
		"rect"
	], "selection");
	text$1(s.anchor);
	text$1(s.label);
	rect(s.rect);
	const e = record(s.element, "element");
	keys(e, [
		"source",
		"sourceContext",
		"componentName",
		"html",
		"locators"
	], "element");
	text$1(e.source);
	text$1(e.componentName);
	text$1(e.html, 4200, true);
	if (e.sourceContext !== void 0) {
		const context = record(e.sourceContext, "sourceContext");
		keys(context, ["mode", "url"], "sourceContext");
		if (context.mode !== "external-document" || context.url !== e.source) throw new Error("selection: external source does not match document context");
		validateDocumentSource(e.source);
	} else if (!isPageSource(e.source)) throw new Error("selection: expected page source");
	list(e.locators).forEach(locator);
}
function validateJson(value, depth = 0) {
	if (depth > 50) throw new Error("data: maximum nesting exceeded");
	if (value === null || typeof value === "string" || typeof value === "boolean" || typeof value === "number" && Number.isFinite(value)) return;
	if (Array.isArray(value)) {
		value.forEach((v) => validateJson(v, depth + 1));
		return;
	}
	for (const item of Object.values(record(value, "data"))) validateJson(item, depth + 1);
}
// src/host/design-validation.ts
function text(value, empty = false) {
	if (typeof value !== "string" || !empty && !value || value.length > 1e4) throw new Error("design: invalid text");
}
function position(value) {
	if (typeof value !== "number" || !Number.isSafeInteger(value) || value < 1) throw new Error("design: invalid position");
}
function source(value) {
	text(value);
	parseEditableSource(value);
	if (!isPageSource(value)) throw new Error("design: expected editable page source");
}
function prop(name) {
	if (!/^[a-zA-Z][\w-]*$/.test(name) || /^(?:on[A-Z]|data-source$|__|ref$|key$|dangerouslySetInnerHTML$)/.test(name)) throw new Error(`design: invalid prop ${name}`);
}
function props(value) {
	for (const [name, item] of Object.entries(record(value, "design.props"))) {
		prop(name);
		validateJson(item);
	}
}
function edit(value) {
	const e = record(value, "design.edit");
	switch (e.kind) {
		case "delete":
			keys(e, ["kind"], "edit");
			break;
		case "style":
			keys(e, [
				"kind",
				"prop",
				"value"
			], "edit");
			if (e.prop !== "variant" && e.prop !== "role") throw new Error("design: invalid style prop");
			text(e.value, true);
			break;
		case "text":
			keys(e, [
				"kind",
				"prop",
				"from",
				"to"
			], "edit");
			text(e.prop);
			prop(e.prop);
			text(e.from, true);
			text(e.to, true);
			break;
		case "attributes":
			keys(e, [
				"kind",
				"values",
				"textFrom"
			], "edit");
			if (e.textFrom !== void 0) text(e.textFrom, true);
			if (!Object.keys(record(e.values, "attributes")).length) throw new Error("design: attributes are empty");
			for (const [name, item] of Object.entries(record(e.values, "attributes"))) {
				prop(name);
				if (item !== null && typeof item !== "string" && !(typeof item === "number" && Number.isFinite(item))) throw new Error("design: invalid attribute value");
			}
			break;
		default: throw new Error("design: invalid edit kind");
	}
}
function result(value) {
	const r = record(value, "design.result");
	switch (r.status) {
		case "applied":
			keys(r, [
				"status",
				"file",
				"path",
				"line",
				"hash"
			], "result");
			text(r.file);
			text(r.path);
			position(r.line);
			text(r.hash);
			if (!/^[0-9a-f]{12}$/.test(r.hash)) throw new Error("design: invalid result hash");
			break;
		case "stale":
			keys(r, [
				"status",
				"error",
				"file"
			], "result");
			text(r.error);
			text(r.file);
			break;
		case "invalid":
			keys(r, ["status", "error"], "result");
			text(r.error);
			break;
		case "refused":
			keys(r, [
				"status",
				"error",
				"reason",
				"file",
				"line",
				"column",
				"attribute"
			], "result");
			text(r.error);
			text(r.reason);
			text(r.file);
			position(r.line);
			position(r.column);
			if (r.attribute !== void 0) text(r.attribute);
			break;
		case "ambiguous":
			keys(r, [
				"status",
				"error",
				"candidates"
			], "result");
			text(r.error);
			if (!Array.isArray(r.candidates) || !r.candidates.length || r.candidates.length > 1e4) throw new Error("design: invalid candidates");
			for (const value of r.candidates) {
				const c = record(value, "candidate");
				keys(c, [
					"id",
					"file",
					"line",
					"column",
					"context"
				], "candidate");
				text(c.id);
				text(c.file);
				position(c.line);
				position(c.column);
				text(c.context, true);
			}
			break;
		default: throw new Error("design: invalid result status");
	}
}
function validateDesignMessage(m, direction) {
	if (m.type === "design-selection" && direction === "frame") {
		keys(m, [
			"type",
			"selection",
			"source",
			"props"
		], "design");
		validateAgentSelection(m.selection);
		source(m.source);
		props(m.props);
		const element = record(record(m.selection, "selection").element, "element");
		if (element.source !== m.source || element.sourceContext !== void 0) throw new Error("design: selection must match the editable page source");
		return m;
	}
	text(m.requestId);
	if (m.type === "design-write-result") {
		keys(m, [
			"type",
			"requestId",
			"result"
		], "design");
		result(m.result);
	} else if (m.type === "design-override" && direction === "host") {
		keys(m, [
			"type",
			"requestId",
			"source",
			"props"
		], "design");
		source(m.source);
		props(m.props);
		if (!Object.keys(m.props).length) throw new Error("design: override is empty");
	} else if (m.type === "design-override-result" && direction === "frame") {
		keys(m, m.ok === true ? [
			"type",
			"requestId",
			"source",
			"ok"
		] : [
			"type",
			"requestId",
			"source",
			"ok",
			"error"
		], "design");
		source(m.source);
		if (m.ok === false) text(m.error);
		else if (m.ok !== true) throw new Error("design: invalid override result");
	} else if (m.type === "design-write" && direction === "host") {
		keys(m, [
			"type",
			"requestId",
			"page",
			"source",
			"edit",
			"choose"
		], "design");
		text(m.page);
		source(m.source);
		edit(m.edit);
		if (m.choose !== void 0) text(m.choose);
	} else if (m.type === "design-write-request" && direction === "frame") {
		keys(m, [
			"type",
			"requestId",
			"request"
		], "design");
		const r = record(m.request, "request");
		keys(r, [
			"page",
			"source",
			"edit",
			"choose"
		], "request");
		text(r.page);
		source(r.source);
		edit(r.edit);
		if (r.choose !== void 0) text(r.choose);
	} else throw new Error("design: unsupported message direction");
	return m;
}
// src/host/agent-validation.ts
function validateAgentHostMessage(value) {
	const m = record(value, "agent");
	switch (m.type) {
		case "design-override":
		case "design-write":
		case "design-write-result": return validateDesignMessage(m, "host");
		case "mode":
			keys(m, ["type", "mode"], "agent");
			if (m.mode !== "off" && m.mode !== "comment" && m.mode !== "design") throw new Error("invalid agent mode");
			break;
		case "pick":
		case "page":
		case "capture-vocabulary":
			keys(m, ["type"], "agent");
			break;
		case "pins":
			keys(m, [
				"type",
				"anchors",
				"focus"
			], "agent");
			if (m.focus !== void 0 && m.focus !== null) text$1(m.focus);
			list(m.anchors).forEach((value) => {
				const p = record(value, "pin");
				keys(p, [
					"anchor",
					"count",
					"locators"
				], "pin");
				text$1(p.anchor);
				number(p.count, 1);
				if (!Number.isInteger(p.count)) throw new Error("invalid pin count");
				list(p.locators).forEach(locator);
			});
			break;
		case "highlight":
			keys(m, ["type", "anchor"], "agent");
			if (m.anchor !== null) text$1(m.anchor);
			break;
		case "scroll":
			keys(m, [
				"type",
				"anchor",
				"behavior",
				"block",
				"offset"
			], "agent");
			text$1(m.anchor);
			if (m.behavior !== void 0 && ![
				"auto",
				"smooth",
				"instant"
			].includes(String(m.behavior))) throw new Error("invalid scroll behavior");
			if (m.block !== void 0 && m.block !== "start" && m.block !== "center") throw new Error("invalid scroll block");
			if (m.offset !== void 0) number(m.offset);
			break;
		default: throw new Error("unsupported agent host message");
	}
	return m;
}
// src/frame/agent/owner.ts
init_source();
init_dist();
function isPageElement(options) {
	if (!options.element.isConnected || options.element.ownerDocument !== options.root.ownerDocument) return false;
	for (let element = options.element; element;) {
		if (element.closest("[data-page-agent],[data-react-grab],[data-react-grab-ignore]")) return false;
		const tree = element.getRootNode();
		element = tree instanceof ShadowRoot ? tree.host : null;
	}
	if (options.root.contains(options.element)) return true;
	const visited = /* @__PURE__ */ new Set();
	for (let fiber = options.fiber === void 0 ? Xe$3(options.element) : options.fiber; fiber && !visited.has(fiber); fiber = fiber.return) {
		visited.add(fiber);
		const node = fiber.stateNode;
		if (node instanceof Node && options.root.contains(node)) return true;
		if (node && "containerInfo" in node && node.containerInfo === options.root) return true;
	}
	return false;
}
function resolvePageOwner(options) {
	if (!isPageElement(options)) throw new Error("picker selected outside the page root");
	if (options.sourceContext) {
		validateDocumentSource(options.sourceContext.url);
		return {
			element: options.element,
			source: options.sourceContext.url,
			componentName: options.componentName ?? options.element.tagName.toLowerCase()
		};
	}
	for (let element = options.element; element && isPageElement({
		element,
		root: options.root
	}); element = element.parentElement) {
		const source = element.getAttribute("data-source");
		if (source && isPageSource(source)) return {
			element,
			source,
			componentName: options.componentName ?? element.tagName.toLowerCase()
		};
	}
	const pending = [options.fiber];
	const visited = /* @__PURE__ */ new Set();
	while (pending.length) {
		const fiber = pending.shift();
		if (!fiber || visited.has(fiber)) continue;
		visited.add(fiber);
		const source = fiber._debugSource;
		const location = source ? `${source.fileName}:${source.lineNumber}:${source.columnNumber}` : "";
		if (isPageSource(location)) return {
			element: options.element,
			source: location,
			componentName: options.componentName ?? options.element.tagName.toLowerCase()
		};
		if (fiber._debugStack?.stack) {
			const frame = m(fiber._debugStack.stack).find((frame) => isPageSource(`${frame.fileName}:${frame.lineNumber}:${frame.columnNumber}`));
			if (frame) return {
				element: options.element,
				source: `${frame.fileName}:${frame.lineNumber}:${frame.columnNumber}`,
				componentName: options.componentName ?? options.element.tagName.toLowerCase()
			};
		}
		pending.push(fiber._debugOwner ?? null, fiber.return ?? null);
	}
	throw new Error("picker could not resolve a page source location for the selected element");
}
// src/frame/agent/overlay.ts
function elementRect(element) {
	const { left, top, width, height } = element.getBoundingClientRect();
	return {
		left,
		top,
		width,
		height
	};
}
function mountAgentOverlay(options) {
	const host = document.createElement("div");
	host.setAttribute("data-page-agent", "");
	host.setAttribute("data-react-grab-ignore", "");
	const shadow = host.attachShadow({ mode: "open" });
	const style = document.createElement("style");
	style.textContent = `:host{all:initial;position:fixed;inset:0;pointer-events:none;z-index:2147483646}
    button{position:absolute;pointer-events:auto;min-width:22px;height:22px;padding:0 4px;border:1px solid white;border-radius:12px;color:white;background:#2563eb;font:600 12px system-ui;cursor:pointer;transform:translate(-50%,-50%)}
    .highlight{position:absolute;border:2px solid #2563eb;background:#2563eb18;box-sizing:border-box;pointer-events:none}`;
	const pins = document.createElement("div");
	const highlight = document.createElement("div");
	highlight.className = "highlight";
	highlight.hidden = true;
	shadow.append(style, pins, highlight);
	document.documentElement.append(host);
	let saved = [];
	let selected = null;
	let highlighted = null;
	let pending = 0;
	let last = "";
	function resolve(pin) {
		const registered = options.registered.get(pin.anchor);
		const match = registered?.isConnected ? {
			element: registered,
			matchedBy: "anchor",
			conflict: false
		} : resolveComment(options.root, pin.anchor, pin.locators);
		return match && isPageElement({
			element: match.element,
			root: options.root
		}) ? match : null;
	}
	function target(anchor) {
		if (anchor === "__document__") return options.root;
		return resolve(saved.find((pin) => pin.anchor === anchor) ?? {
			anchor,
			count: 1,
			locators: []
		})?.element ?? null;
	}
	function drawPin(pin, index) {
		const match = resolve(pin);
		if (!match) return {
			anchor: pin.anchor,
			state: "orphaned",
			label: pin.anchor,
			rect: null
		};
		const box = elementRect(match.element);
		const button = document.createElement("button");
		button.type = "button";
		button.textContent = String(index + 1);
		button.title = `${describeElement(match.element)} (${pin.count})`;
		button.setAttribute("aria-label", `Open ${pin.count} comments on ${describeElement(match.element)}`);
		button.setAttribute("aria-pressed", String(selected === pin.anchor));
		button.style.left = `${box.left + box.width}px`;
		button.style.top = `${box.top}px`;
		button.addEventListener("click", () => options.send({
			type: "focus",
			anchor: pin.anchor
		}));
		pins.append(button);
		return {
			anchor: pin.anchor,
			state: match.matchedBy === "anchor" && !match.conflict ? "found" : "moved",
			label: describeElement(match.element),
			rect: box
		};
	}
	function recompute() {
		pending = 0;
		pins.replaceChildren();
		const states = saved.map(drawPin);
		const element = highlighted ? target(highlighted) : null;
		highlight.hidden = !element;
		if (element) for (const [key, value] of Object.entries(elementRect(element))) highlight.style.setProperty(key, `${value}px`);
		const geometry = {
			type: "geometry",
			scrollX: window.scrollX,
			scrollY: window.scrollY,
			width: window.innerWidth,
			height: window.innerHeight
		};
		const state = JSON.stringify([states, geometry]);
		if (state !== last) {
			last = state;
			options.send({
				type: "anchors",
				anchors: states
			});
			options.send(geometry);
		}
	}
	function schedule() {
		if (!pending) pending = requestAnimationFrame(recompute);
	}
	const observer = new MutationObserver(schedule);
	observer.observe(options.root, {
		subtree: true,
		childList: true,
		characterData: true,
		attributes: true
	});
	observer.observe(options.root.ownerDocument.body, {
		subtree: true,
		childList: true,
		characterData: true,
		attributes: true
	});
	const resize = new ResizeObserver(schedule);
	resize.observe(options.root);
	window.addEventListener("scroll", schedule, true);
	window.addEventListener("resize", schedule);
	schedule();
	return {
		host,
		target,
		schedule,
		containedAnchors(node) {
			return [...new Set(saved.filter((pin) => {
				const element = target(pin.anchor);
				return element !== null && node.contains(element);
			}).map((pin) => pin.anchor))];
		},
		update(anchors, focus) {
			saved = anchors;
			if (focus !== void 0) selected = focus;
			last = "";
			schedule();
		},
		highlight(anchor) {
			highlighted = anchor;
			schedule();
		},
		dispose() {
			observer.disconnect();
			resize.disconnect();
			cancelAnimationFrame(pending);
			window.removeEventListener("scroll", schedule, true);
			window.removeEventListener("resize", schedule);
			host.remove();
		}
	};
}
// src/frame/agent/react-override.ts
function renderer() {
	const hook = window.__REACT_DEVTOOLS_GLOBAL_HOOK__;
	if (!hook?.renderers?.size) throw new Error("Design mode requires the DevTools hook before react-dom loads");
	const renderers = [...hook.renderers.values()].filter((value) => typeof value.overrideProps === "function");
	if (renderers.length !== 1) throw new Error("Design mode requires exactly one React development renderer with overrideProps");
	return renderers[0];
}
function currentFiber(fiber) {
	if (!fiber.alternate) return fiber;
	let top = fiber;
	while (top.return) top = top.return;
	if (top.tag !== 3) throw new Error("fiber is detached from its React root");
	const stack = [top.stateNode.current];
	for (let next = stack.pop(); next; next = stack.pop()) {
		if (next === fiber || next === fiber.alternate) return next;
		if (next.sibling) stack.push(next.sibling);
		if (next.child) stack.push(next.child);
	}
	throw new Error("fiber is not in the committed React tree");
}
function fiberOf(node) {
	const key = Object.keys(node).find((name) => name.startsWith("__reactFiber$"));
	if (!key) throw new Error(`<${node.localName}> was not rendered by React`);
	return node[key];
}
function callSiteFiber(options) {
	let match = null;
	for (let fiber = fiberOf(options.node); fiber; fiber = fiber.return) if (fiber.memoizedProps?.["data-source"] === options.source) match = fiber;
	if (!match) throw new Error(`no React call site for ${options.source} above <${options.node.localName}>`);
	return currentFiber(match);
}
function readCallSiteProps(options) {
	renderer();
	return { ...callSiteFiber(options).memoizedProps };
}
function overrideCallSite(options) {
	const active = renderer();
	const fibers = new Set(options.nodes.map((node) => callSiteFiber({
		node,
		source: options.source
	})));
	for (const fiber of fibers) active.overrideProps(fiber, [], {
		...fiber.memoizedProps,
		...options.props
	});
}
// src/editing.ts
const scope = globalThis;
const { removed, listeners } = scope.__pageEditing ??= {
	removed: /* @__PURE__ */ new Set(),
	listeners: /* @__PURE__ */ new Set()
};
createContext("");
function setRemoved(identity, value) {
	if (value) removed.add(identity);
	else removed.delete(identity);
	for (const listener of listeners) listener();
}
// src/frame/agent/design.ts
function resolveDesignTarget(options) {
	for (let element = options.element; element && options.root.contains(element); element = element.parentElement) {
		const source = element.getAttribute("data-source");
		if (source?.includes("#") && isPageSource(source)) {
			parseEditableSource(source);
			return {
				element,
				source
			};
		}
	}
	for (let fiber = fiberOf(options.element); fiber; fiber = fiber.return) {
		const source = fiber.memoizedProps?.["data-source"];
		if (typeof source === "string" && source.includes("#") && isPageSource(source)) {
			parseEditableSource(source);
			return {
				element: options.element,
				source
			};
		}
	}
	throw new Error("Design selection has no editable page call site");
}
function panelProps(props) {
	return Object.fromEntries(Object.entries(props).filter(([name, value]) => name !== "data-source" && !name.startsWith("on") && (value === null || typeof value === "string" || typeof value === "boolean" || typeof value === "number" && Number.isFinite(value))));
}
function createDesignAgent(options) {
	let selected;
	const pending = /* @__PURE__ */ new Map();
	function nodes(source) {
		if (!selected || selected.source !== source) throw new Error("Design request does not match the live selection");
		const matches = [.../* @__PURE__ */ new Set([...options.root.contains(selected.node) ? [selected.node] : [], ...[...options.root.querySelectorAll("[data-source]")].filter((node) => node.getAttribute("data-source") === source)])];
		if (!matches.length) throw new Error("Design target is no longer rendered");
		return matches;
	}
	return {
		select(target, selection) {
			const props = panelProps(readCallSiteProps({
				node: target.element,
				source: target.source
			}));
			selected = {
				node: target.element,
				source: target.source
			};
			options.send({
				type: "design-selection",
				selection,
				source: target.source,
				props
			});
		},
		receive(message) {
			if (message.type === "design-write-result") {
				const attempt = pending.get(message.requestId);
				if (!attempt) throw new Error(`unexpected Design write result ${message.requestId}`);
				pending.delete(message.requestId);
				if (attempt.deletion && message.result.status !== "applied") setRemoved(`source:${attempt.source}`, false);
				options.send(message);
				return;
			}
			if (message.type === "design-override") {
				try {
					overrideCallSite({
						nodes: nodes(message.source),
						source: message.source,
						props: Object.fromEntries(Object.entries(message.props).map(([name, value]) => [name, value === null ? void 0 : value]))
					});
					options.send({
						type: "design-override-result",
						requestId: message.requestId,
						source: message.source,
						ok: true
					});
				} catch (error) {
					options.send({
						type: "design-override-result",
						requestId: message.requestId,
						source: message.source,
						ok: false,
						error: error instanceof Error ? error.message : String(error)
					});
				}
				return;
			}
			nodes(message.source);
			if (pending.has(message.requestId)) throw new Error(`duplicate Design write ${message.requestId}`);
			pending.set(message.requestId, {
				source: message.source,
				deletion: message.edit.kind === "delete"
			});
			if (message.edit.kind === "delete") setRemoved(`source:${message.source}`, true);
			options.send({
				type: "design-write-request",
				requestId: message.requestId,
				request: {
					page: message.page,
					source: message.source,
					edit: message.edit,
					...message.choose === void 0 ? {} : { choose: message.choose }
				}
			});
		},
		dispose() {
			selected = void 0;
			pending.clear();
		}
	};
}
// src/host/review/validation.ts
const REVIEW_ID = /^[A-Za-z][A-Za-z0-9:_-]*$/;
const REVIEW_ACTIONS = [
	"mark",
	"approved",
	"rejected",
	"revision_requested",
	"delete",
	"export"
];
function id(value) {
	text$1(value);
	if (!REVIEW_ID.test(value)) throw new Error("Invalid review id");
}
function bool(value) {
	if (typeof value !== "boolean") throw new Error("Invalid review visibility");
}
function action(value) {
	if (!REVIEW_ACTIONS.some((action) => action === value)) throw new Error("Invalid review action");
}
function validateReviewHostMessage(value) {
	const m = record(value, "review");
	text$1(m.requestId);
	switch (m.type) {
		case "review-remove":
			keys(m, [
				"type",
				"requestId",
				"id",
				"removed"
			], "review");
			id(m.id);
			bool(m.removed);
			break;
		case "review-list":
			keys(m, ["type", "requestId"], "review");
			break;
		case "review-select":
			keys(m, [
				"type",
				"requestId",
				"id"
			], "review");
			if (m.id !== null) id(m.id);
			break;
		case "review-visibility":
			keys(m, [
				"type",
				"requestId",
				"visible",
				"showRejected"
			], "review");
			bool(m.visible);
			bool(m.showRejected);
			break;
		case "review-action":
			keys(m, [
				"type",
				"requestId",
				"id",
				"action"
			], "review");
			id(m.id);
			action(m.action);
			break;
		default: throw new Error("Unsupported review host message");
	}
	return m;
}
// src/frame/agent/review.ts
init_dist();
function createReviewAgent(options) {
	let visible = false;
	let showRejected = false;
	let selected = null;
	let scheduled;
	let disposed = false;
	const previousRejected = /* @__PURE__ */ new Map();
	function applyVisibility() {
		for (const node of [options.root, ...options.root.querySelectorAll("[data-review-content]")]) {
			if (!previousRejected.has(node)) previousRejected.set(node, node.getAttribute("data-show-rejected"));
			if (node.getAttribute("data-show-rejected") !== String(showRejected)) node.setAttribute("data-show-rejected", String(showRejected));
		}
	}
	function nodes() {
		return [...options.root.querySelectorAll("[data-review-id]")].filter((node) => isPageElement({
			element: node,
			root: options.root
		}));
	}
	function describe(node, id) {
		const owner = resolvePageOwner({
			element: node,
			root: options.root,
			componentName: options.componentName(node),
			fiber: Xe$3(node)
		});
		const html = node.outerHTML;
		return {
			id,
			...node.id ? { domId: node.id } : {},
			selection: {
				anchor: `[data-review-id="${id}"]`,
				label: describeElement(node),
				rect: elementRect(node),
				element: {
					source: owner.source,
					componentName: owner.componentName,
					locators: captureLocators(node, options.root),
					html: html.length <= COMMENT_ELEMENT_HTML_LIMIT ? html : `${html.slice(0, COMMENT_ELEMENT_HTML_LIMIT)}${COMMENT_ELEMENT_HTML_TRUNCATION}`
				}
			}
		};
	}
	function enumerate() {
		const seen = /* @__PURE__ */ new Set();
		return nodes().flatMap((node) => {
			const id = node.getAttribute("data-review-id");
			if (!REVIEW_ID.test(id) || seen.has(id)) throw new Error(`Review id ${JSON.stringify(id)} must be valid and unique on this page`);
			seen.add(id);
			return node.getClientRects().length ? [describe(node, id)] : [];
		});
	}
	function state(requestId) {
		applyVisibility();
		const elements = visible ? enumerate() : [];
		if (!elements.some((element) => element.id === selected)) selected = null;
		options.send({
			type: "review-state",
			...requestId ? { requestId } : {},
			elements,
			selected,
			visible,
			showRejected
		});
	}
	function schedule() {
		if (!visible || disposed || scheduled !== void 0) return;
		scheduled = requestAnimationFrame(() => {
			scheduled = void 0;
			try {
				state();
			} catch (error) {
				options.error(error instanceof Error ? error : new Error(String(error)));
			}
		});
	}
	function target(id) {
		const found = nodes().filter((node) => node.getAttribute("data-review-id") === id);
		if (found.length !== 1) throw new Error(`Review element ${JSON.stringify(id)} must exist exactly once`);
		return found[0];
	}
	function action(optionsForAction) {
		if (!visible) throw new Error("Review actions require visible review controls");
		const { node, id, action, requestId } = optionsForAction;
		if (action === "delete") {
			options.send({
				type: "review-action-request",
				requestId,
				action,
				element: describe(node, id),
				containedAnchors: options.containedAnchors(node)
			});
			return;
		}
		let content;
		if (action === "export") {
			const clone = ([...node.querySelectorAll("[data-review-content]")].find((region) => region.parentElement?.closest("[data-review-id]") === node) ?? node).cloneNode(true);
			for (const ignored of clone.querySelectorAll("script,[data-page-agent],[data-react-grab],[data-react-grab-ignore]")) ignored.remove();
			content = {
				html: clone.outerHTML,
				...clone.matches("svg") ? { svg: clone.outerHTML } : {}
			};
		}
		options.send({
			type: "review-action-request",
			requestId,
			action,
			element: describe(node, id),
			...content ? { content } : {}
		});
	}
	const mutation = new MutationObserver(schedule);
	mutation.observe(options.root, {
		childList: true,
		subtree: true,
		attributes: true,
		characterData: true
	});
	const resize = new ResizeObserver(schedule);
	resize.observe(options.root);
	window.addEventListener("scroll", schedule, true);
	window.addEventListener("resize", schedule);
	return {
		receive(value) {
			if (disposed) throw new Error("Review agent is disposed");
			const m = validateReviewHostMessage(value);
			try {
				if (m.type === "review-remove") {
					setRemoved(`review:${m.id}`, m.removed);
					selected = null;
					state(m.requestId);
					return;
				}
				if (m.type === "review-action") action({
					node: target(m.id),
					id: m.id,
					action: m.action,
					requestId: m.requestId
				});
				else {
					if (m.type === "review-visibility") {
						visible = m.visible;
						showRejected = m.showRejected;
					}
					if (m.type === "review-select") {
						if (m.id !== null) {
							if (!visible) throw new Error("Review selection requires visible review controls");
							target(m.id).scrollIntoView({ block: "center" });
						}
						selected = m.id;
					}
					state(m.requestId);
				}
			} catch (error) {
				options.send({
					type: "review-error",
					requestId: m.requestId,
					error: error instanceof Error ? error.message : String(error)
				});
			}
		},
		action(node, kind) {
			if (!isPageElement({
				element: node,
				root: options.root
			})) throw new Error("Review target must belong to the page");
			const reviewable = node.closest("[data-review-id]");
			if (kind !== "mark" && (!reviewable || !options.root.contains(reviewable))) throw new Error("Selected element is not reviewable");
			const id = kind === "mark" ? `new:${randomUUID()}` : reviewable.getAttribute("data-review-id");
			action({
				node: kind === "mark" ? node : reviewable,
				id,
				action: kind,
				requestId: randomUUID()
			});
		},
		enabled: () => visible,
		dispose() {
			disposed = true;
			mutation.disconnect();
			resize.disconnect();
			window.removeEventListener("scroll", schedule, true);
			window.removeEventListener("resize", schedule);
			if (scheduled !== void 0) cancelAnimationFrame(scheduled);
			for (const [node, value] of previousRejected) if (value === null) node.removeAttribute("data-show-rejected");
			else node.setAttribute("data-show-rejected", value);
			previousRejected.clear();
		}
	};
}
// src/host/validation.ts
function validateSourcePath(path) {
	if (!path || path.startsWith("/") || path.includes("\\") || path.split("/").some((segment) => segment === ".." || segment === "." || segment === "") || /[?#]/.test(path) || Array.from(path).some((character) => character.charCodeAt(0) < 32)) throw new Error(`invalid source path ${path}`);
}
// src/link/page-sources.ts
function path(value, empty = false) {
	if (empty && value === "") return;
	validateSourcePath(value);
	if (/%|:/.test(value)) throw new Error(`Noncanonical page source path ${value}`);
}
function validatePageSources(value) {
	if (!Array.isArray(value) || value.length > 1e3) throw new Error("Invalid page source roots");
	const roots = [];
	for (const item of value) {
		if (!item || typeof item !== "object" || Object.keys(item).some((key) => ![
			"link",
			"root",
			"prefix",
			"editable"
		].includes(key))) throw new Error("Invalid page source root");
		const root = item;
		if (typeof root.link !== "string" || !/^[A-Za-z0-9_-]+$/.test(root.link) || typeof root.root !== "string" || typeof root.prefix !== "string" || typeof root.editable !== "boolean") throw new Error("Invalid page source declaration");
		path(root.root);
		path(root.prefix, true);
		if (!isPageSource(`${root.root}/page.tsx:1:1`) || !isPageSource(`${root.prefix ? `${root.prefix}/` : ""}page.tsx:1:1`)) throw new Error("Page source root or destination is a library or runtime path");
		for (const previous of roots) if (previous.link === root.link && overlaps(previous.root, root.root) || previous.prefix === root.prefix) throw new Error("Page source roots overlap");
		roots.push(root);
	}
}
function overlaps(a, b) {
	return !a || !b || a === b || a.startsWith(`${b}/`) || b.startsWith(`${a}/`);
}
// src/frame/agent/vocabulary.ts
async function captureVocabulary(options) {
	try {
		if (!options.ready) throw new Error("page is not ready for vocabulary capture");
		return {
			type: "vocabulary-previews",
			images: await captureVocabularyPreviews(document)
		};
	} catch (error) {
		return {
			type: "error",
			error: {
				phase: "render",
				message: error instanceof Error ? error.message : String(error),
				file: options.page
			}
		};
	}
}
// src/frame/agent/index.ts
init_core_DT_3yIQi();
init_dist();
function startPageAgent(options) {
	if (!options.root.isConnected || !options.page) throw new Error("page agent requires a mounted root and page path");
	if (options.sourceContext) validateDocumentSource(options.sourceContext.url);
	if (options.pageSources) validatePageSources(options.pageSources);
	const design = createDesignAgent(options);
	const api = options.picker ?? Xs({
		enabled: true,
		telemetry: false,
		freezeReactUpdates: false
	});
	const registered = /* @__PURE__ */ new Map();
	const overlay = mountAgentOverlay({
		root: options.root,
		send: options.send,
		registered
	});
	const review = createReviewAgent({
		...options,
		componentName: (element) => api.getDisplayName(element),
		containedAnchors: overlay.containedAnchors
	});
	const plugin = `page-agent-${randomUUID()}`;
	const previousAction = api.getToolbarState()?.defaultAction;
	const previousEnabled = api.isEnabled();
	let picking = false;
	let disposed = false;
	let version = 0;
	let mode = "off";
	async function select(element) {
		if (element.closest("[data-react-grab-ignore]")) return;
		picking = false;
		const revision = ++version;
		const componentName = api.getDisplayName(element);
		const owner = mode === "design" ? {
			...resolveDesignTarget({
				element,
				root: options.root
			}),
			componentName: componentName ?? element.tagName.toLowerCase()
		} : resolvePageOwner({
			element,
			root: options.root,
			componentName,
			fiber: Xe$3(element),
			...options.sourceContext ? { sourceContext: options.sourceContext } : {}
		});
		const anchor = commentAnchor(owner.element, options.root);
		if (!anchor) throw new Error("selected page element cannot be anchored");
		const html = owner.element.outerHTML;
		const context = {
			source: owner.source,
			componentName: owner.componentName,
			...options.sourceContext ? { sourceContext: options.sourceContext } : {},
			html: html.length <= COMMENT_ELEMENT_HTML_LIMIT ? html : `${html.slice(0, COMMENT_ELEMENT_HTML_LIMIT)}${COMMENT_ELEMENT_HTML_TRUNCATION}`,
			locators: captureLocators(owner.element, options.root)
		};
		api.deactivate();
		if (disposed || version !== revision) return;
		overlay.highlight(mode === "design" ? anchor : null);
		const selection = {
			anchor,
			label: describeElement(owner.element),
			element: context,
			rect: elementRect(owner.element)
		};
		if (mode === "design") design.select(owner, selection);
		else options.send({
			type: "selection",
			selection
		});
	}
	api.registerPlugin({
		name: plugin,
		...options.toolbar === false ? { theme: { toolbar: { enabled: false } } } : {},
		actions: REVIEW_ACTIONS.map((action) => ({
			id: `review-${action}`,
			label: action === "mark" ? "Make reviewable" : action === "approved" ? "Approve" : action === "rejected" ? "Reject with comment" : action === "revision_requested" ? "Request revision" : action === "delete" ? "Delete source and feedback" : "Export review content",
			enabled: ({ element }) => review.enabled() && isPageElement({
				element,
				root: options.root
			}) && (action === "mark" ? !element.closest("[data-review-id]") : Boolean(element.closest("[data-review-id]"))),
			onAction: ({ element, hideContextMenu, cleanup }) => {
				hideContextMenu();
				cleanup();
				review.action(element, action);
			}
		})),
		hooks: {
			onElementSelect(element) {
				if (!picking) return;
				select(element).catch((error) => options.error(error instanceof Error ? error : new Error(String(error))));
				return true;
			},
			onDeactivate() {
				picking = false;
			},
			transformActionContext(context) {
				return {
					...context,
					enterPromptMode: () => {
						select(context.element).catch((error) => options.error(error instanceof Error ? error : new Error(String(error))));
					}
				};
			}
		}
	});
	function pick() {
		picking = true;
		api.setEnabled(true);
		api.setToolbarState({ defaultAction: "comment" });
		api.activate();
	}
	function receive(value) {
		if (disposed) throw new Error("page agent is disposed");
		if (value.type === "review-remove" || value.type === "review-list" || value.type === "review-select" || value.type === "review-visibility" || value.type === "review-action") {
			review.receive(value);
			return;
		}
		const m = validateAgentHostMessage(value);
		switch (m.type) {
			case "mode":
				if (m.mode === "design" && options.sourceContext) throw new Error("Design mode requires runtime page sources");
				mode = m.mode;
				if (mode !== "off") pick();
				else {
					version++;
					picking = false;
					api.deactivate();
				}
				break;
			case "design-override":
			case "design-write":
			case "design-write-result":
				if (mode !== "design") throw new Error("Design request requires Design mode");
				design.receive(m);
				break;
			case "pick":
				pick();
				break;
			case "page": {
				version++;
				picking = false;
				api.deactivate();
				const html = options.root.outerHTML;
				options.send({
					type: "selection",
					selection: {
						anchor: "__document__",
						label: "Whole page",
						rect: elementRect(options.root),
						element: {
							source: options.sourceContext?.url ?? `${options.page}:1:1`,
							...options.sourceContext ? { sourceContext: options.sourceContext } : {},
							componentName: "Page",
							locators: [],
							html: html.length <= COMMENT_ELEMENT_HTML_LIMIT ? html : `${html.slice(0, COMMENT_ELEMENT_HTML_LIMIT)}${COMMENT_ELEMENT_HTML_TRUNCATION}`
						}
					}
				});
				break;
			}
			case "pins":
				overlay.update(m.anchors, m.focus);
				break;
			case "highlight":
				overlay.highlight(m.anchor);
				break;
			case "scroll": {
				const element = overlay.target(m.anchor);
				if (!element) throw new Error(`cannot scroll to orphaned anchor ${m.anchor}`);
				element.scrollIntoView({
					behavior: m.behavior ?? "smooth",
					block: m.block ?? "center"
				});
				if (m.offset !== void 0) window.scrollBy({
					top: -m.offset,
					behavior: "instant"
				});
				break;
			}
			default: throw new Error(`page agent message ${m.type} requires a dedicated handler`);
		}
	}
	function keys(event) {
		const toggle = event.key.toLowerCase() === "c" && event.altKey && !event.ctrlKey && !event.metaKey;
		if (event.key !== "Escape" && !toggle) return;
		if (event.isComposing || toggle && event.composedPath().some((node) => node instanceof Element && node.matches("input,textarea,select,[contenteditable=\"true\"]"))) return;
		event.preventDefault();
		if (event.key === "Escape") {
			version++;
			picking = false;
			api.deactivate();
		}
		options.send({
			type: "keys",
			key: event.key === "Escape" ? "Escape" : "c",
			ctrlKey: event.ctrlKey,
			metaKey: event.metaKey,
			altKey: event.altKey,
			shiftKey: event.shiftKey
		});
	}
	window.addEventListener("keydown", keys, true);
	return {
		receive,
		captureVocabulary(request) {
			if (disposed) throw new Error("page agent is disposed");
			return captureVocabulary({
				ready: request.ready,
				page: options.page
			});
		},
		registerAnchor(anchor, element) {
			if (!anchor || element && !isPageElement({
				element,
				root: options.root
			})) throw new Error("registered anchor must belong to the page root");
			if (element) {
				const existing = registered.get(anchor);
				if (existing && existing !== element && existing.isConnected) throw new Error(`duplicate registered anchor ${anchor}`);
				registered.set(anchor, element);
			} else registered.delete(anchor);
			overlay.schedule();
		},
		focusAnchor(anchor) {
			options.send({
				type: "focus",
				anchor
			});
		},
		dispose() {
			disposed = true;
			version++;
			window.removeEventListener("keydown", keys, true);
			api.deactivate();
			api.unregisterPlugin(plugin);
			if (options.picker) {
				api.setToolbarState({ defaultAction: previousAction ?? "copy" });
				api.setEnabled(previousEnabled);
			} else api.dispose();
			overlay.dispose();
			registered.clear();
			design.dispose();
			review.dispose();
		}
	};
}
// src/frame/agent/provider.tsx
var _jsxFileName = "src/frame/agent/provider.tsx";
function PageAgentProvider({ agent, children }) {
	return /* @__PURE__ */ jsxDEV(CommentProvider, {
		comments: [],
		config: { statuses: [] },
		resolveAnchor: strictAnchorResolver,
		onScrollToAnchor: (anchor, options) => {
			agent.receive({
				type: "scroll",
				anchor,
				...options
			});
			return true;
		},
		children: /* @__PURE__ */ jsxDEV(CommentAnchorBridge, {
			registerAnchor: agent.registerAnchor,
			focusAnchor: agent.focusAnchor,
			children
		}, void 0, false, {
			fileName: _jsxFileName,
			lineNumber: 9,
			columnNumber: 7
		}, this)
	}, void 0, false, {
		fileName: _jsxFileName,
		lineNumber: 7,
		columnNumber: 5
	}, this);
}
export { PageAgentProvider, startPageAgent };

//# sourceMappingURL=page-agent.js.map