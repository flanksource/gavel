// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/zwitch/2.0.4/ba2b41af332e04cd54d4405c6fa9d719437fc9b4066df42228ac8d79d844ef5b/node_modules/zwitch/index.js
var e = {}.hasOwnProperty;
function t(t, n) {
	let r = n || {};
	function i(n, ...r) {
		let a = i.invalid, o = i.handlers;
		if (n && e.call(n, t)) {
			let r = String(n[t]);
			a = e.call(o, r) ? o[r] : i.unknown;
		}
		if (a) return a.call(this, n, ...r);
	}
	return i.handlers = r.handlers || {}, i.invalid = r.invalid, i.unknown = r.unknown, i;
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/ccount/2.0.1/a9b1838d5e863942aa9dd2258a84a046030ae3abfc784374a38c04c5753bde52/node_modules/ccount/index.js
function n(e, t) {
	let n = String(e);
	if (typeof t != "string") throw TypeError("Expected character");
	let r = 0, i = n.indexOf(t);
	for (; i !== -1;) r++, i = n.indexOf(t, i + t.length);
	return r;
}
export { t as n, n as t };
