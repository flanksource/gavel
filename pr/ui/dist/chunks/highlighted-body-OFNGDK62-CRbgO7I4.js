import { r as e } from "./rolldown-runtime-QGRDbDUS.js";
import { n as t, t as n } from "./react-DHvKIBht.js";
import { n as r, r as i, t as a } from "./dist-C-bR18hx.js";
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/streamdown/2.5.0/3bc14f2eb408a90c753497ac7cef03eb04d3869230b04a3264b6e3a53d1cc113/node_modules/streamdown/dist/highlighted-body-OFNGDK62.js
var o = /* @__PURE__ */ e(n(), 1), s = /* @__PURE__ */ e(t(), 1), c = ({ code: e, language: t, raw: n, className: c, startLine: l, lineNumbers: u, ...d }) => {
	let { shikiTheme: f } = (0, o.useContext)(i), p = r(), [m, h] = (0, o.useState)(n);
	return (0, o.useEffect)(() => {
		if (!p) {
			h(n);
			return;
		}
		let r = p.highlight({
			code: e,
			language: t,
			themes: f
		}, (e) => {
			h(e);
		});
		r && h(r);
	}, [
		e,
		t,
		f,
		p,
		n
	]), (0, s.jsx)(a, {
		className: c,
		language: t,
		lineNumbers: u,
		result: m,
		startLine: l,
		...d
	});
};
export { c as HighlightedCodeBlockBody };
