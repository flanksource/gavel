import { a as e } from "./rolldown-runtime-DUcC5eAW.js";
import { n as t, t as n } from "./jsx-runtime-B4NxTvNR.js";
import { n as r, r as i, t as a } from "./dist-Jk7JYRzo.js";
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/streamdown/2.7.1/cd86399a25cc60a263bb5bc5019500d5ee112acb0fe8c90f08576a062b1a2e36/node_modules/streamdown/dist/highlighted-body-JLMTFACH.js
var o = /* @__PURE__ */ e(t(), 1), s = n(), c = ({ code: e, isIncomplete: t = !1, language: n, maxHeight: c, raw: l, className: u, startLine: d, lineNumbers: f, ...p }) => {
	let { shikiTheme: m } = (0, o.useContext)(r), h = i(), [g, _] = (0, o.useState)(l);
	return (0, o.useEffect)(() => {
		if (!h) {
			_(l);
			return;
		}
		let r = !0, i = h.highlight({
			code: e,
			isIncomplete: t,
			language: n,
			themes: m
		}, (e) => {
			r && _(e);
		});
		return i && _(i), () => {
			r = !1;
		};
	}, [
		e,
		t,
		n,
		m,
		h,
		l
	]), (0, s.jsx)(a, {
		className: u,
		language: n,
		lineNumbers: f,
		maxHeight: c,
		result: g,
		startLine: d,
		...p
	});
};
export { c as HighlightedCodeBlockBody };
