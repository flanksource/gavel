import { r as e } from "./rolldown-runtime-QGRDbDUS.js";
import { t } from "./react-DHvKIBht.js";
// ../../../clicky-ui/packages/ui/src/hooks/use-media-query.ts
var n = /* @__PURE__ */ e(t(), 1);
function r(e) {
	let t = o(e);
	return t ? t.matches : !1;
}
function i(e, t) {
	let n = o(e);
	if (!n) return () => {};
	let r = () => t(n.matches);
	return n.addEventListener ? (n.addEventListener("change", r), () => n.removeEventListener("change", r)) : (n.addListener?.(r), () => n.removeListener?.(r));
}
function a(e) {
	let [t, a] = (0, n.useState)(() => r(e));
	return (0, n.useEffect)(() => (a(r(e)), i(e, a)), [e]), t;
}
function o(e) {
	return typeof window > "u" || typeof window.matchMedia != "function" ? null : window.matchMedia(e);
}
// ../../../clicky-ui/packages/ui/src/hooks/use-theme.ts
var s = "(prefers-color-scheme: dark)", c = "clicky-ui-theme", l = "data-theme", u = (0, n.createContext)(null);
function d() {
	return r(s);
}
function f() {
	if (typeof window > "u") return "system";
	let e = window.localStorage.getItem(c);
	return e === "light" || e === "dark" || e === "system" ? e : "system";
}
function p(e) {
	return e === "system" ? d() ? "dark" : "light" : e;
}
function m(e) {
	typeof document < "u" && document.documentElement.setAttribute(l, e);
}
function h() {
	let e = (0, n.useContext)(u);
	if (!e) throw Error("useTheme must be used inside <ThemeProvider>");
	return e;
}
function g() {
	return (0, n.useContext)(u);
}
function _(e) {
	let t = g(), [r, a] = (0, n.useState)(() => d() ? "dark" : "light");
	return (0, n.useEffect)(() => i(s, (e) => a(e ? "dark" : "light")), []), e && e !== "system" ? e : e === "system" ? r : t ? t.resolvedTheme : r;
}
export { f as a, h as c, a as d, m as i, r as l, c as n, p as o, u as r, _ as s, s as t, i as u };
