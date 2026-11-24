import { a as e } from "./rolldown-runtime-DUcC5eAW.js";
import { n as t } from "./jsx-runtime-B4NxTvNR.js";
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
var s = "(prefers-color-scheme: dark)", c = (0, n.createContext)(null);
function l() {
	return r(s);
}
function u() {
	return (0, n.useContext)(c);
}
function d(e) {
	let t = u(), [r, a] = (0, n.useState)(() => l() ? "dark" : "light");
	return (0, n.useEffect)(() => i(s, (e) => a(e ? "dark" : "light")), []), e && e !== "system" ? e : e === "system" ? r : t ? t.resolvedTheme : r;
}
export { r as n, a as r, d as t };
