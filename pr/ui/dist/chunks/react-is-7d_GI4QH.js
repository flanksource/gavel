import { t as e } from "./rolldown-runtime-QGRDbDUS.js";
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/dequal/2.0.3/9b5133530171b3f5834fc6daef4f201fdb414fc35bda1e6865305a17b415b237/node_modules/dequal/dist/index.mjs
var t = Object.prototype.hasOwnProperty;
function n(e, t, n) {
	for (n of e.keys()) if (r(n, t)) return n;
}
function r(e, i) {
	var a, o, s;
	if (e === i) return !0;
	if (e && i && (a = e.constructor) === i.constructor) {
		if (a === Date) return e.getTime() === i.getTime();
		if (a === RegExp) return e.toString() === i.toString();
		if (a === Array) {
			if ((o = e.length) === i.length) for (; o-- && r(e[o], i[o]););
			return o === -1;
		}
		if (a === Set) {
			if (e.size !== i.size) return !1;
			for (o of e) if (s = o, s && typeof s == "object" && (s = n(i, s), !s) || !i.has(s)) return !1;
			return !0;
		}
		if (a === Map) {
			if (e.size !== i.size) return !1;
			for (o of e) if (s = o[0], s && typeof s == "object" && (s = n(i, s), !s) || !r(o[1], i.get(s))) return !1;
			return !0;
		}
		if (a === ArrayBuffer) e = new Uint8Array(e), i = new Uint8Array(i);
		else if (a === DataView) {
			if ((o = e.byteLength) === i.byteLength) for (; o-- && e.getInt8(o) === i.getInt8(o););
			return o === -1;
		}
		if (ArrayBuffer.isView(e)) {
			if ((o = e.byteLength) === i.byteLength) for (; o-- && e[o] === i[o];);
			return o === -1;
		}
		if (!a || typeof e == "object") {
			for (a in o = 0, e) if (t.call(e, a) && ++o && !t.call(i, a) || !(a in i) || !r(e[a], i[a])) return !1;
			return Object.keys(i).length === o;
		}
	}
	return e !== e && i !== i;
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/react-is/17.0.2/e53048f60b2073eee40d849a125f3baf7933a0a3de91026ec02385c06de3e410/node_modules/react-is/cjs/react-is.production.min.js
var i = /* @__PURE__ */ e(((e) => {
	if (typeof Symbol == "function" && Symbol.for) {
		var t = Symbol.for;
		t("react.element"), t("react.portal"), t("react.fragment"), t("react.strict_mode"), t("react.profiler"), t("react.provider"), t("react.context"), t("react.forward_ref"), t("react.suspense"), t("react.suspense_list"), t("react.memo"), t("react.lazy"), t("react.block"), t("react.server.block"), t("react.fundamental"), t("react.debug_trace_mode"), t("react.legacy_hidden");
	}
})), a = /* @__PURE__ */ e(((e, t) => {
	t.exports = i();
}));
export { r as n, a as t };
