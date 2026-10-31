// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/clsx/2.1.1/2b14a624e52246853e8086215d43767b3f05a70db70d9393672d28f32b6c54a6/node_modules/clsx/dist/clsx.mjs
function e(t) {
	var n, r, i = "";
	if (typeof t == "string" || typeof t == "number") i += t;
	else if (typeof t == "object") {
		if (Array.isArray(t)) {
			var a = t.length;
			for (n = 0; n < a; n++) t[n] && (r = e(t[n])) && (i && (i += " "), i += r);
		} else for (r in t) t[r] && (i && (i += " "), i += r);
	}
	return i;
}
function t() {
	for (var t, n, r = 0, i = "", a = arguments.length; r < a; r++) (t = arguments[r]) && (n = e(t)) && (i && (i += " "), i += n);
	return i;
}
export { t };
