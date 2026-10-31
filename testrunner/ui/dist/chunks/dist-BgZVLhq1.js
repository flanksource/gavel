import { r as e, t } from "./rolldown-runtime-QGRDbDUS.js";
import { n, t as r } from "./jsx-runtime-lVm_Myqd.js";
import { t as i } from "./react-dom-BBUmKE87.js";
import { n as a, r as o } from "./bundle-mjs--8zjrAdI.js";
import { a as s, c, i as l, l as u, n as d, o as f, r as p, s as m, t as h, u as g } from "./lib-BeutSJOr.js";
import { n as _, t as v } from "./ccount-ClnF6NWF.js";
import { A as y, H as b, I as x, M as S, N as ee, O as te, P as ne, R as re, U as ie, V as C, W as ae, _ as oe, a as se, c as ce, d as le, f as ue, g as de, j as fe, k as pe, l as me, m as he, n as ge, s as w, t as _e, v as ve, y as ye, z as T } from "./lib-CPWjiPzk.js";
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/property-information/7.1.0/4cced2366fa5df45a5566b4fa132b0b9878e1c1ed65413aeaf33cb36c564a08f/node_modules/property-information/lib/hast-to-react.js
var be = {
	classId: "classID",
	dataType: "datatype",
	itemId: "itemID",
	strokeDashArray: "strokeDasharray",
	strokeDashOffset: "strokeDashoffset",
	strokeLineCap: "strokeLinecap",
	strokeLineJoin: "strokeLinejoin",
	strokeMiterLimit: "strokeMiterlimit",
	typeOf: "typeof",
	xLinkActuate: "xlinkActuate",
	xLinkArcRole: "xlinkArcrole",
	xLinkHref: "xlinkHref",
	xLinkRole: "xlinkRole",
	xLinkShow: "xlinkShow",
	xLinkTitle: "xlinkTitle",
	xLinkType: "xlinkType",
	xmlnsXLink: "xmlnsXlink"
}, E = /* @__PURE__ */ e(n(), 1), xe = {
	indicator: "indicator",
	textOnly: "text-only",
	remove: "remove"
};
function Se({ defaultOrigin: e = "", allowedLinkPrefixes: t = [], allowedImagePrefixes: n = [], allowDataImages: r = !1, allowedProtocols: i = [], blockedImageClass: a = "inline-block bg-gray-200 dark:bg-gray-700 text-gray-600 dark:text-gray-400 px-3 py-1 rounded text-sm", blockedLinkClass: o = "text-gray-500", linkBlockPolicy: s = xe.indicator, imageBlockPolicy: c = xe.indicator }) {
	let l = t.length && !t.every((e) => e === "*"), u = n.length && !n.every((e) => e === "*");
	if (!e && (l || u)) throw Error("defaultOrigin is required when allowedLinkPrefixes or allowedImagePrefixes are provided");
	return (l) => {
		let u = Me(e, t, n, r, i, a, o, s, c);
		Oe(l), C(l, u);
	};
}
function Ce(e, t) {
	if (typeof e != "string") return null;
	try {
		return new URL(e);
	} catch {
		if (t) try {
			return new URL(e, t);
		} catch {
			return null;
		}
		if (e.startsWith("/") || e.startsWith("./") || e.startsWith("../")) try {
			return new URL(e, "http://example.com");
		} catch {
			return null;
		}
		return null;
	}
}
function we(e) {
	return typeof e == "string" ? e.startsWith("/") || e.startsWith("./") || e.startsWith("../") : !1;
}
var Te = /* @__PURE__ */ new Set([
	"https:",
	"http:",
	"irc:",
	"ircs:",
	"mailto:",
	"xmpp:",
	"blob:"
]), Ee = /* @__PURE__ */ new Set([
	"javascript:",
	"data:",
	"file:",
	"vbscript:"
]);
function De(e, t, n, r = !1, i = !1, a = []) {
	if (!e) return null;
	if (typeof e == "string" && e.startsWith("#") && !i) try {
		if (new URL(e, "http://example.com").hash === e) return e;
	} catch {}
	if (typeof e == "string" && e.startsWith("data:")) return i && r && e.startsWith("data:image/") ? e : null;
	if (typeof e == "string" && e.startsWith("blob:")) {
		try {
			if (new URL(e).protocol === "blob:" && e.length > 5) {
				let t = e.substring(5);
				if (t && t.length > 0 && t !== "invalid") return e;
			}
		} catch {
			return null;
		}
		return null;
	}
	let o = Ce(e, n);
	if (!o || Ee.has(o.protocol) || !(Te.has(o.protocol) || a.includes(o.protocol) || a.includes("*"))) return null;
	if (o.protocol === "mailto:" || !o.protocol.match(/^https?:$/)) return o.href;
	let s = we(e);
	return o && t.some((e) => {
		let t = Ce(e, n);
		return !t || t.origin !== o.origin ? !1 : o.href.startsWith(t.href);
	}) ? s ? o.pathname + o.search + o.hash : o.href : t.includes("*") ? o.protocol !== "https:" && o.protocol !== "http:" ? null : s ? o.pathname + o.search + o.hash : o.href : null;
}
function Oe(e) {
	if ("children" in e && Array.isArray(e.children)) {
		e.children = e.children.filter((e) => e != null);
		for (let t of e.children) Oe(t);
	}
}
var ke = Symbol("node-seen");
function Ae(e, t, n) {
	return t === xe.remove ? { type: "remove" } : t === xe.textOnly ? {
		type: "replace",
		element: {
			type: "element",
			tagName: "span",
			properties: {},
			children: [...e.children]
		}
	} : {
		type: "replace",
		element: {
			type: "element",
			tagName: "span",
			properties: {
				title: "Blocked URL: " + String(e.properties.href),
				class: n
			},
			children: [...e.children, {
				type: "text",
				value: " [blocked]"
			}]
		}
	};
}
function je(e, t, n) {
	if (t === xe.remove) return { type: "remove" };
	if (t === xe.textOnly) {
		let t = String(e.properties.alt || "");
		return t ? {
			type: "replace",
			element: {
				type: "element",
				tagName: "span",
				properties: {},
				children: [{
					type: "text",
					value: t
				}]
			}
		} : { type: "remove" };
	}
	return {
		type: "replace",
		element: {
			type: "element",
			tagName: "span",
			properties: { class: n },
			children: [{
				type: "text",
				value: "[Image blocked: " + String(e.properties.alt || "No description") + "]"
			}]
		}
	};
}
var Me = (e, t, n, r, i, a, o, s, c) => {
	let l = (u, d, f) => {
		if (u.type !== "element" || u[ke]) return !0;
		if (u.tagName === "a") {
			let n = De(u.properties.href, t, e, !1, !1, i);
			if (n === null) {
				if (u[ke] = !0, C(u, l), f && typeof d == "number") {
					let e = Ae(u, s, o);
					if (e.type === "remove") return f.children.splice(d, 1), [b, d];
					f.children[d] = e.element;
				}
				return b;
			}
			return u.properties.href = n, u.properties.target = "_blank", u.properties.rel = "noopener noreferrer", !0;
		}
		if (u.tagName === "img") {
			let t = De(u.properties.src, n, e, r, !0, i);
			if (t === null) {
				if (u[ke] = !0, C(u, l), f && typeof d == "number") {
					let e = je(u, c, a);
					if (e.type === "remove") return f.children.splice(d, 1), [b, d];
					f.children[d] = e.element;
				}
				return b;
			}
			return u.properties.src = t, !0;
		}
		return !0;
	};
	return l;
}, Ne = typeof self == "object" ? self : globalThis, Pe = (e, t) => {
	let n = (t, n) => (e.set(n, t), t), r = (i) => {
		if (e.has(i)) return e.get(i);
		let [a, o] = t[i];
		switch (a) {
			case 0:
			case -1: return n(o, i);
			case 1: {
				let e = n([], i);
				for (let t of o) e.push(r(t));
				return e;
			}
			case 2: {
				let e = n({}, i);
				for (let [t, n] of o) e[r(t)] = r(n);
				return e;
			}
			case 3: return n(new Date(o), i);
			case 4: {
				let { source: e, flags: t } = o;
				return n(new RegExp(e, t), i);
			}
			case 5: {
				let e = n(/* @__PURE__ */ new Map(), i);
				for (let [t, n] of o) e.set(r(t), r(n));
				return e;
			}
			case 6: {
				let e = n(/* @__PURE__ */ new Set(), i);
				for (let t of o) e.add(r(t));
				return e;
			}
			case 7: {
				let { name: e, message: t } = o;
				return n(new Ne[e](t), i);
			}
			case 8: return n(BigInt(o), i);
			case "BigInt": return n(Object(BigInt(o)), i);
			case "ArrayBuffer": return n(new Uint8Array(o).buffer, o);
			case "DataView": {
				let { buffer: e } = new Uint8Array(o);
				return n(new DataView(e), o);
			}
		}
		return n(new Ne[a](o), i);
	};
	return r;
}, Fe = (e) => Pe(/* @__PURE__ */ new Map(), e)(0), Ie = "", { toString: Le } = {}, { keys: Re } = Object, ze = (e) => {
	let t = typeof e;
	if (t !== "object" || !e) return [0, t];
	let n = Le.call(e).slice(8, -1);
	switch (n) {
		case "Array": return [1, Ie];
		case "Object": return [2, Ie];
		case "Date": return [3, Ie];
		case "RegExp": return [4, Ie];
		case "Map": return [5, Ie];
		case "Set": return [6, Ie];
		case "DataView": return [1, n];
	}
	return n.includes("Array") ? [1, n] : n.includes("Error") ? [7, n] : [2, n];
}, Be = ([e, t]) => e === 0 && (t === "function" || t === "symbol"), Ve = (e, t, n, r) => {
	let i = (e, t) => {
		let i = r.push(e) - 1;
		return n.set(t, i), i;
	}, a = (r) => {
		if (n.has(r)) return n.get(r);
		let [o, s] = ze(r);
		switch (o) {
			case 0: {
				let t = r;
				switch (s) {
					case "bigint":
						o = 8, t = r.toString();
						break;
					case "function":
					case "symbol":
						if (e) throw TypeError("unable to serialize " + s);
						t = null;
						break;
					case "undefined": return i([-1], r);
				}
				return i([o, t], r);
			}
			case 1: {
				if (s) {
					let e = r;
					return s === "DataView" ? e = new Uint8Array(r.buffer) : s === "ArrayBuffer" && (e = new Uint8Array(r)), i([s, [...e]], r);
				}
				let e = [], t = i([o, e], r);
				for (let t of r) e.push(a(t));
				return t;
			}
			case 2: {
				if (s) switch (s) {
					case "BigInt": return i([s, r.toString()], r);
					case "Boolean":
					case "Number":
					case "String": return i([s, r.valueOf()], r);
				}
				if (t && "toJSON" in r) return a(r.toJSON());
				let n = [], c = i([o, n], r);
				for (let t of Re(r)) (e || !Be(ze(r[t]))) && n.push([a(t), a(r[t])]);
				return c;
			}
			case 3: return i([o, r.toISOString()], r);
			case 4: {
				let { source: e, flags: t } = r;
				return i([o, {
					source: e,
					flags: t
				}], r);
			}
			case 5: {
				let t = [], n = i([o, t], r);
				for (let [n, i] of r) (e || !(Be(ze(n)) || Be(ze(i)))) && t.push([a(n), a(i)]);
				return n;
			}
			case 6: {
				let t = [], n = i([o, t], r);
				for (let n of r) (e || !Be(ze(n))) && t.push(a(n));
				return n;
			}
		}
		let { message: c } = r;
		return i([o, {
			name: s,
			message: c
		}], r);
	};
	return a;
}, He = (e, { json: t, lossy: n } = {}) => {
	let r = [];
	return Ve(!(t || n), !!t, /* @__PURE__ */ new Map(), r)(e), r;
}, Ue = typeof structuredClone == "function" ? 
/* c8 ignore start */
(e, t) => t && ("json" in t || "lossy" in t) ? Fe(He(e, t)) : structuredClone(e) : (e, t) => Fe(He(e, t)), We = /[#.]/g;
function Ge(e, t) {
	let n = e || "", r = {}, i = 0, a, o;
	for (; i < n.length;) {
		We.lastIndex = i;
		let e = We.exec(n), t = n.slice(i, e ? e.index : n.length);
		t && (a ? a === "#" ? r.id = t : Array.isArray(r.className) ? r.className.push(t) : r.className = [t] : o = t, i += t.length), e && (a = e[0], i++);
	}
	return {
		type: "element",
		tagName: o || t || "div",
		properties: r,
		children: []
	};
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/hastscript/9.0.1/f1041d8b2b794b38426a856edcb87717b47c8c35fd8ecb8e4bd6f5194e9b0ec1/node_modules/hastscript/lib/create-h.js
function Ke(e, t, n) {
	let r = n ? Qe(n) : void 0;
	function i(n, i, ...a) {
		let o;
		if (n == null) {
			o = {
				type: "root",
				children: []
			};
			let e = i;
			a.unshift(e);
		} else {
			o = Ge(n, t);
			let s = o.tagName.toLowerCase(), c = r ? r.get(s) : void 0;
			if (o.tagName = c || s, qe(i)) a.unshift(i);
			else for (let [t, n] of Object.entries(i)) Je(e, o.properties, t, n);
		}
		for (let e of a) Ye(o.children, e);
		return o.type === "element" && o.tagName === "template" && (o.content = {
			type: "root",
			children: o.children
		}, o.children = []), o;
	}
	return i;
}
function qe(e) {
	if (typeof e != "object" || !e || Array.isArray(e)) return !0;
	if (typeof e.type != "string") return !1;
	let t = e, n = Object.keys(e);
	for (let e of n) {
		let n = t[e];
		if (n && typeof n == "object") {
			if (!Array.isArray(n)) return !0;
			let e = n;
			for (let t of e) if (typeof t != "number" && typeof t != "string") return !0;
		}
	}
	return !!("children" in e && Array.isArray(e.children));
}
function Je(e, t, n, r) {
	let i = c(e, n), a;
	if (r != null) {
		if (typeof r == "number") {
			if (Number.isNaN(r)) return;
			a = r;
		} else a = typeof r == "boolean" ? r : typeof r == "string" ? i.spaceSeparated ? d(r) : i.commaSeparated ? l(r) : i.commaOrSpaceSeparated ? d(l(r).join(" ")) : Xe(i, i.property, r) : Array.isArray(r) ? [...r] : i.property === "style" ? Ze(r) : String(r);
		if (Array.isArray(a)) {
			let e = [];
			for (let t of a) e.push(Xe(i, i.property, t));
			a = e;
		}
		i.property === "className" && Array.isArray(t.className) && (a = t.className.concat(a)), t[i.property] = a;
	}
}
function Ye(e, t) {
	if (t != null) {
		if (typeof t == "number" || typeof t == "string") e.push({
			type: "text",
			value: String(t)
		});
		else if (Array.isArray(t)) for (let n of t) Ye(e, n);
		else if (typeof t == "object" && "type" in t) t.type === "root" ? Ye(e, t.children) : e.push(t);
		else throw Error("Expected node, nodes, or string, got `" + t + "`");
	}
}
function Xe(e, t, n) {
	if (typeof n == "string") {
		if (e.number && n && !Number.isNaN(Number(n))) return Number(n);
		if ((e.boolean || e.overloadedBoolean) && (n === "" || u(n) === u(t))) return !0;
	}
	return n;
}
function Ze(e) {
	let t = [];
	for (let [n, r] of Object.entries(e)) t.push([n, r].join(": "));
	return t.join("; ");
}
function Qe(e) {
	let t = /* @__PURE__ */ new Map();
	for (let n of e) t.set(n.toLowerCase(), n);
	return t;
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/hastscript/9.0.1/f1041d8b2b794b38426a856edcb87717b47c8c35fd8ecb8e4bd6f5194e9b0ec1/node_modules/hastscript/lib/svg-case-sensitive-tag-names.js
var $e = /* @__PURE__ */ "altGlyph.altGlyphDef.altGlyphItem.animateColor.animateMotion.animateTransform.clipPath.feBlend.feColorMatrix.feComponentTransfer.feComposite.feConvolveMatrix.feDiffuseLighting.feDisplacementMap.feDistantLight.feDropShadow.feFlood.feFuncA.feFuncB.feFuncG.feFuncR.feGaussianBlur.feImage.feMerge.feMergeNode.feMorphology.feOffset.fePointLight.feSpecularLighting.feSpotLight.feTile.feTurbulence.foreignObject.glyphRef.linearGradient.radialGradient.solidColor.textArea.textPath".split("."), et = Ke(f, "div"), tt = Ke(m, "g", $e);
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/vfile-location/5.0.3/3cde0bbab7d096571f618bad4a88ce0b0bb691db81f61327b8f914cdf2762f15/node_modules/vfile-location/lib/index.js
function nt(e) {
	let t = String(e), n = [];
	return {
		toOffset: i,
		toPoint: r
	};
	function r(e) {
		if (typeof e == "number" && e > -1 && e <= t.length) {
			let r = 0;
			for (;;) {
				let i = n[r];
				if (i === void 0) {
					let e = rt(t, n[r - 1]);
					i = e === -1 ? t.length + 1 : e + 1, n[r] = i;
				}
				if (i > e) return {
					line: r + 1,
					column: e - (r > 0 ? n[r - 1] : 0) + 1,
					offset: e
				};
				r++;
			}
		}
	}
	function i(e) {
		if (e && typeof e.line == "number" && typeof e.column == "number" && !Number.isNaN(e.line) && !Number.isNaN(e.column)) {
			for (; n.length < e.line;) {
				let e = n[n.length - 1], r = rt(t, e), i = r === -1 ? t.length + 1 : r + 1;
				if (e === i) break;
				n.push(i);
			}
			let r = (e.line > 1 ? n[e.line - 2] : 0) + e.column - 1;
			if (r < n[e.line - 1]) return r;
		}
	}
}
function rt(e, t) {
	let n = e.indexOf("\r", t), r = e.indexOf("\n", t);
	return r === -1 ? n : n === -1 || n + 1 === r ? r : n < r ? n : r;
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/web-namespaces/2.0.1/1e7c04fc430e08cefe5b44c2483610eceb3a425439cbb1ed916eb11aca6d9599/node_modules/web-namespaces/index.js
var it = {
	html: "http://www.w3.org/1999/xhtml",
	mathml: "http://www.w3.org/1998/Math/MathML",
	svg: "http://www.w3.org/2000/svg",
	xlink: "http://www.w3.org/1999/xlink",
	xml: "http://www.w3.org/XML/1998/namespace",
	xmlns: "http://www.w3.org/2000/xmlns/"
}, at = {}.hasOwnProperty, ot = Object.prototype;
function st(e, t) {
	let n = t || {};
	return ct({
		file: n.file || void 0,
		location: !1,
		schema: n.space === "svg" ? m : f,
		verbose: n.verbose || !1
	}, e);
}
function ct(e, t) {
	let n;
	switch (t.nodeName) {
		case "#comment": {
			let r = t;
			return n = {
				type: "comment",
				value: r.data
			}, dt(e, r, n), n;
		}
		case "#document":
		case "#document-fragment": {
			let r = t, i = "mode" in r ? r.mode === "quirks" || r.mode === "limited-quirks" : !1;
			if (n = {
				type: "root",
				children: lt(e, t.childNodes),
				data: { quirksMode: i }
			}, e.file && e.location) {
				let t = String(e.file), r = nt(t), i = r.toPoint(0), a = r.toPoint(t.length);
				n.position = {
					start: i,
					end: a
				};
			}
			return n;
		}
		case "#documentType": {
			let r = t;
			return n = { type: "doctype" }, dt(e, r, n), n;
		}
		case "#text": {
			let r = t;
			return n = {
				type: "text",
				value: r.value
			}, dt(e, r, n), n;
		}
		default: return n = ut(e, t), n;
	}
}
function lt(e, t) {
	let n = -1, r = [];
	for (; ++n < t.length;) {
		let i = ct(e, t[n]);
		r.push(i);
	}
	return r;
}
function ut(e, t) {
	let n = e.schema;
	e.schema = t.namespaceURI === it.svg ? m : f;
	let r = -1, i = {};
	for (; ++r < t.attrs.length;) {
		let e = t.attrs[r], n = (e.prefix ? e.prefix + ":" : "") + e.name;
		at.call(ot, n) || (i[n] = e.value);
	}
	let a = (e.schema.space === "svg" ? tt : et)(t.tagName, i, lt(e, t.childNodes));
	if (dt(e, t, a), a.tagName === "template") {
		let n = t, r = n.sourceCodeLocation, i = r && r.startTag && pt(r.startTag), o = r && r.endTag && pt(r.endTag), s = ct(e, n.content);
		i && o && e.file && (s.position = {
			start: i.end,
			end: o.start
		}), a.content = s;
	}
	return e.schema = n, a;
}
function dt(e, t, n) {
	if ("sourceCodeLocation" in t && t.sourceCodeLocation && e.file) {
		let r = ft(e, n, t.sourceCodeLocation);
		r && (e.location = !0, n.position = r);
	}
}
function ft(e, t, n) {
	let r = pt(n);
	if (t.type === "element") {
		let i = t.children[t.children.length - 1];
		if (r && !n.endTag && i && i.position && i.position.end && (r.end = Object.assign({}, i.position.end)), e.verbose) {
			let r = {}, i;
			if (n.attrs) for (i in n.attrs) at.call(n.attrs, i) && (r[c(e.schema, i).property] = pt(n.attrs[i]));
			n.startTag;
			let a = pt(n.startTag), o = n.endTag ? pt(n.endTag) : void 0, s = { opening: a };
			o && (s.closing = o), s.properties = r, t.data = { position: s };
		}
	}
	return r;
}
function pt(e) {
	let t = mt({
		line: e.startLine,
		column: e.startCol,
		offset: e.startOffset
	}), n = mt({
		line: e.endLine,
		column: e.endCol,
		offset: e.endOffset
	});
	return t || n ? {
		start: t,
		end: n
	} : void 0;
}
function mt(e) {
	return e.line && e.column ? e : void 0;
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/hast-util-to-parse5/8.0.1/e6c3cee0d5fab529e1b82c3f7dfd465e7942c580e2946bfc513d9dbdc1b718b0/node_modules/hast-util-to-parse5/lib/index.js
var ht = {}, gt = {}.hasOwnProperty, _t = _("type", { handlers: {
	root: yt,
	element: wt,
	text: St,
	comment: Ct,
	doctype: xt
} });
function vt(e, t) {
	let n = (t || ht).space;
	return _t(e, n === "svg" ? m : f);
}
function yt(e, t) {
	let n = {
		nodeName: "#document",
		mode: (e.data || {}).quirksMode ? "quirks" : "no-quirks",
		childNodes: []
	};
	return n.childNodes = Et(e.children, n, t), Dt(e, n), n;
}
function bt(e, t) {
	let n = {
		nodeName: "#document-fragment",
		childNodes: []
	};
	return n.childNodes = Et(e.children, n, t), Dt(e, n), n;
}
function xt(e) {
	let t = {
		nodeName: "#documentType",
		name: "html",
		publicId: "",
		systemId: "",
		parentNode: null
	};
	return Dt(e, t), t;
}
function St(e) {
	let t = {
		nodeName: "#text",
		value: e.value,
		parentNode: null
	};
	return Dt(e, t), t;
}
function Ct(e) {
	let t = {
		nodeName: "#comment",
		data: e.value,
		parentNode: null
	};
	return Dt(e, t), t;
}
function wt(e, t) {
	let n = t, r = n;
	e.type === "element" && e.tagName.toLowerCase() === "svg" && n.space === "html" && (r = m);
	let i = [], a;
	if (e.properties) {
		for (a in e.properties) if (a !== "children" && gt.call(e.properties, a)) {
			let t = Tt(r, a, e.properties[a]);
			t && i.push(t);
		}
	}
	let o = r.space, s = {
		nodeName: e.tagName,
		tagName: e.tagName,
		attrs: i,
		namespaceURI: it[o],
		childNodes: [],
		parentNode: null
	};
	return s.childNodes = Et(e.children, s, r), Dt(e, s), e.tagName === "template" && e.content && (s.content = bt(e.content, r)), s;
}
function Tt(e, t, n) {
	let r = c(e, t);
	if (n === !1 || n == null || typeof n == "number" && Number.isNaN(n) || !n && r.boolean) return;
	Array.isArray(n) && (n = r.commaSeparated ? s(n) : p(n));
	let i = {
		name: r.attribute,
		value: n === !0 ? "" : String(n)
	};
	if (r.space && r.space !== "html" && r.space !== "svg") {
		let e = i.name.indexOf(":");
		e < 0 ? i.prefix = "" : (i.name = i.name.slice(e + 1), i.prefix = r.attribute.slice(0, e)), i.namespace = it[r.space];
	}
	return i;
}
function Et(e, t, n) {
	let r = -1, i = [];
	if (e) for (; ++r < e.length;) {
		let a = _t(e[r], n);
		a.parentNode = t, i.push(a);
	}
	return i;
}
function Dt(e, t) {
	let n = e.position;
	n && n.start && n.end && (n.start.offset, n.end.offset, t.sourceCodeLocation = {
		startLine: n.start.line,
		startCol: n.start.column,
		startOffset: n.start.offset,
		endLine: n.end.line,
		endCol: n.end.column,
		endOffset: n.end.offset
	});
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/parse5/7.3.0/4acbced08702cc2b068edf7ec6fbf0576d666aa2a4d5c21d9bdc1830acf8a8e2/node_modules/parse5/dist/common/unicode.js
var Ot = /* @__PURE__ */ new Set([
	65534,
	65535,
	131070,
	131071,
	196606,
	196607,
	262142,
	262143,
	327678,
	327679,
	393214,
	393215,
	458750,
	458751,
	524286,
	524287,
	589822,
	589823,
	655358,
	655359,
	720894,
	720895,
	786430,
	786431,
	851966,
	851967,
	917502,
	917503,
	983038,
	983039,
	1048574,
	1048575,
	1114110,
	1114111
]), D;
(function(e) {
	e[e.EOF = -1] = "EOF", e[e.NULL = 0] = "NULL", e[e.TABULATION = 9] = "TABULATION", e[e.CARRIAGE_RETURN = 13] = "CARRIAGE_RETURN", e[e.LINE_FEED = 10] = "LINE_FEED", e[e.FORM_FEED = 12] = "FORM_FEED", e[e.SPACE = 32] = "SPACE", e[e.EXCLAMATION_MARK = 33] = "EXCLAMATION_MARK", e[e.QUOTATION_MARK = 34] = "QUOTATION_MARK", e[e.AMPERSAND = 38] = "AMPERSAND", e[e.APOSTROPHE = 39] = "APOSTROPHE", e[e.HYPHEN_MINUS = 45] = "HYPHEN_MINUS", e[e.SOLIDUS = 47] = "SOLIDUS", e[e.DIGIT_0 = 48] = "DIGIT_0", e[e.DIGIT_9 = 57] = "DIGIT_9", e[e.SEMICOLON = 59] = "SEMICOLON", e[e.LESS_THAN_SIGN = 60] = "LESS_THAN_SIGN", e[e.EQUALS_SIGN = 61] = "EQUALS_SIGN", e[e.GREATER_THAN_SIGN = 62] = "GREATER_THAN_SIGN", e[e.QUESTION_MARK = 63] = "QUESTION_MARK", e[e.LATIN_CAPITAL_A = 65] = "LATIN_CAPITAL_A", e[e.LATIN_CAPITAL_Z = 90] = "LATIN_CAPITAL_Z", e[e.RIGHT_SQUARE_BRACKET = 93] = "RIGHT_SQUARE_BRACKET", e[e.GRAVE_ACCENT = 96] = "GRAVE_ACCENT", e[e.LATIN_SMALL_A = 97] = "LATIN_SMALL_A", e[e.LATIN_SMALL_Z = 122] = "LATIN_SMALL_Z";
})(D ||= {});
var O = {
	DASH_DASH: "--",
	CDATA_START: "[CDATA[",
	DOCTYPE: "doctype",
	SCRIPT: "script",
	PUBLIC: "public",
	SYSTEM: "system"
};
function kt(e) {
	return e >= 55296 && e <= 57343;
}
function At(e) {
	return e >= 56320 && e <= 57343;
}
function jt(e, t) {
	return (e - 55296) * 1024 + 9216 + t;
}
function Mt(e) {
	return e !== 32 && e !== 10 && e !== 13 && e !== 9 && e !== 12 && e >= 1 && e <= 31 || e >= 127 && e <= 159;
}
function Nt(e) {
	return e >= 64976 && e <= 65007 || Ot.has(e);
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/parse5/7.3.0/4acbced08702cc2b068edf7ec6fbf0576d666aa2a4d5c21d9bdc1830acf8a8e2/node_modules/parse5/dist/common/error-codes.js
var k;
(function(e) {
	e.controlCharacterInInputStream = "control-character-in-input-stream", e.noncharacterInInputStream = "noncharacter-in-input-stream", e.surrogateInInputStream = "surrogate-in-input-stream", e.nonVoidHtmlElementStartTagWithTrailingSolidus = "non-void-html-element-start-tag-with-trailing-solidus", e.endTagWithAttributes = "end-tag-with-attributes", e.endTagWithTrailingSolidus = "end-tag-with-trailing-solidus", e.unexpectedSolidusInTag = "unexpected-solidus-in-tag", e.unexpectedNullCharacter = "unexpected-null-character", e.unexpectedQuestionMarkInsteadOfTagName = "unexpected-question-mark-instead-of-tag-name", e.invalidFirstCharacterOfTagName = "invalid-first-character-of-tag-name", e.unexpectedEqualsSignBeforeAttributeName = "unexpected-equals-sign-before-attribute-name", e.missingEndTagName = "missing-end-tag-name", e.unexpectedCharacterInAttributeName = "unexpected-character-in-attribute-name", e.unknownNamedCharacterReference = "unknown-named-character-reference", e.missingSemicolonAfterCharacterReference = "missing-semicolon-after-character-reference", e.unexpectedCharacterAfterDoctypeSystemIdentifier = "unexpected-character-after-doctype-system-identifier", e.unexpectedCharacterInUnquotedAttributeValue = "unexpected-character-in-unquoted-attribute-value", e.eofBeforeTagName = "eof-before-tag-name", e.eofInTag = "eof-in-tag", e.missingAttributeValue = "missing-attribute-value", e.missingWhitespaceBetweenAttributes = "missing-whitespace-between-attributes", e.missingWhitespaceAfterDoctypePublicKeyword = "missing-whitespace-after-doctype-public-keyword", e.missingWhitespaceBetweenDoctypePublicAndSystemIdentifiers = "missing-whitespace-between-doctype-public-and-system-identifiers", e.missingWhitespaceAfterDoctypeSystemKeyword = "missing-whitespace-after-doctype-system-keyword", e.missingQuoteBeforeDoctypePublicIdentifier = "missing-quote-before-doctype-public-identifier", e.missingQuoteBeforeDoctypeSystemIdentifier = "missing-quote-before-doctype-system-identifier", e.missingDoctypePublicIdentifier = "missing-doctype-public-identifier", e.missingDoctypeSystemIdentifier = "missing-doctype-system-identifier", e.abruptDoctypePublicIdentifier = "abrupt-doctype-public-identifier", e.abruptDoctypeSystemIdentifier = "abrupt-doctype-system-identifier", e.cdataInHtmlContent = "cdata-in-html-content", e.incorrectlyOpenedComment = "incorrectly-opened-comment", e.eofInScriptHtmlCommentLikeText = "eof-in-script-html-comment-like-text", e.eofInDoctype = "eof-in-doctype", e.nestedComment = "nested-comment", e.abruptClosingOfEmptyComment = "abrupt-closing-of-empty-comment", e.eofInComment = "eof-in-comment", e.incorrectlyClosedComment = "incorrectly-closed-comment", e.eofInCdata = "eof-in-cdata", e.absenceOfDigitsInNumericCharacterReference = "absence-of-digits-in-numeric-character-reference", e.nullCharacterReference = "null-character-reference", e.surrogateCharacterReference = "surrogate-character-reference", e.characterReferenceOutsideUnicodeRange = "character-reference-outside-unicode-range", e.controlCharacterReference = "control-character-reference", e.noncharacterCharacterReference = "noncharacter-character-reference", e.missingWhitespaceBeforeDoctypeName = "missing-whitespace-before-doctype-name", e.missingDoctypeName = "missing-doctype-name", e.invalidCharacterSequenceAfterDoctypeName = "invalid-character-sequence-after-doctype-name", e.duplicateAttribute = "duplicate-attribute", e.nonConformingDoctype = "non-conforming-doctype", e.missingDoctype = "missing-doctype", e.misplacedDoctype = "misplaced-doctype", e.endTagWithoutMatchingOpenElement = "end-tag-without-matching-open-element", e.closingOfElementWithOpenChildElements = "closing-of-element-with-open-child-elements", e.disallowedContentInNoscriptInHead = "disallowed-content-in-noscript-in-head", e.openElementsLeftAfterEof = "open-elements-left-after-eof", e.abandonedHeadElementChild = "abandoned-head-element-child", e.misplacedStartTagForHeadElement = "misplaced-start-tag-for-head-element", e.nestedNoscriptInHead = "nested-noscript-in-head", e.eofInElementThatCanContainOnlyText = "eof-in-element-that-can-contain-only-text";
})(k ||= {});
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/parse5/7.3.0/4acbced08702cc2b068edf7ec6fbf0576d666aa2a4d5c21d9bdc1830acf8a8e2/node_modules/parse5/dist/tokenizer/preprocessor.js
var Pt = 65536, Ft = class {
	constructor(e) {
		this.handler = e, this.html = "", this.pos = -1, this.lastGapPos = -2, this.gapStack = [], this.skipNextNewLine = !1, this.lastChunkWritten = !1, this.endOfChunkHit = !1, this.bufferWaterline = Pt, this.isEol = !1, this.lineStartPos = 0, this.droppedBufferSize = 0, this.line = 1, this.lastErrOffset = -1;
	}
	get col() {
		return this.pos - this.lineStartPos + Number(this.lastGapPos !== this.pos);
	}
	get offset() {
		return this.droppedBufferSize + this.pos;
	}
	getError(e, t) {
		let { line: n, col: r, offset: i } = this, a = r + t, o = i + t;
		return {
			code: e,
			startLine: n,
			endLine: n,
			startCol: a,
			endCol: a,
			startOffset: o,
			endOffset: o
		};
	}
	_err(e) {
		this.handler.onParseError && this.lastErrOffset !== this.offset && (this.lastErrOffset = this.offset, this.handler.onParseError(this.getError(e, 0)));
	}
	_addGap() {
		this.gapStack.push(this.lastGapPos), this.lastGapPos = this.pos;
	}
	_processSurrogate(e) {
		if (this.pos !== this.html.length - 1) {
			let t = this.html.charCodeAt(this.pos + 1);
			if (At(t)) return this.pos++, this._addGap(), jt(e, t);
		} else if (!this.lastChunkWritten) return this.endOfChunkHit = !0, D.EOF;
		return this._err(k.surrogateInInputStream), e;
	}
	willDropParsedChunk() {
		return this.pos > this.bufferWaterline;
	}
	dropParsedChunk() {
		this.willDropParsedChunk() && (this.html = this.html.substring(this.pos), this.lineStartPos -= this.pos, this.droppedBufferSize += this.pos, this.pos = 0, this.lastGapPos = -2, this.gapStack.length = 0);
	}
	write(e, t) {
		this.html.length > 0 ? this.html += e : this.html = e, this.endOfChunkHit = !1, this.lastChunkWritten = t;
	}
	insertHtmlAtCurrentPos(e) {
		this.html = this.html.substring(0, this.pos + 1) + e + this.html.substring(this.pos + 1), this.endOfChunkHit = !1;
	}
	startsWith(e, t) {
		if (this.pos + e.length > this.html.length) return this.endOfChunkHit = !this.lastChunkWritten, !1;
		if (t) return this.html.startsWith(e, this.pos);
		for (let t = 0; t < e.length; t++) if ((this.html.charCodeAt(this.pos + t) | 32) !== e.charCodeAt(t)) return !1;
		return !0;
	}
	peek(e) {
		let t = this.pos + e;
		if (t >= this.html.length) return this.endOfChunkHit = !this.lastChunkWritten, D.EOF;
		let n = this.html.charCodeAt(t);
		return n === D.CARRIAGE_RETURN ? D.LINE_FEED : n;
	}
	advance() {
		if (this.pos++, this.isEol && (this.isEol = !1, this.line++, this.lineStartPos = this.pos), this.pos >= this.html.length) return this.endOfChunkHit = !this.lastChunkWritten, D.EOF;
		let e = this.html.charCodeAt(this.pos);
		return e === D.CARRIAGE_RETURN ? (this.isEol = !0, this.skipNextNewLine = !0, D.LINE_FEED) : e === D.LINE_FEED && (this.isEol = !0, this.skipNextNewLine) ? (this.line--, this.skipNextNewLine = !1, this._addGap(), this.advance()) : (this.skipNextNewLine = !1, kt(e) && (e = this._processSurrogate(e)), this.handler.onParseError === null || e > 31 && e < 127 || e === D.LINE_FEED || e === D.CARRIAGE_RETURN || e > 159 && e < 64976 || this._checkForProblematicCharacters(e), e);
	}
	_checkForProblematicCharacters(e) {
		Mt(e) ? this._err(k.controlCharacterInInputStream) : Nt(e) && this._err(k.noncharacterInInputStream);
	}
	retreat(e) {
		for (this.pos -= e; this.pos < this.lastGapPos;) this.lastGapPos = this.gapStack.pop(), this.pos--;
		this.isEol = !1;
	}
}, A;
(function(e) {
	e[e.CHARACTER = 0] = "CHARACTER", e[e.NULL_CHARACTER = 1] = "NULL_CHARACTER", e[e.WHITESPACE_CHARACTER = 2] = "WHITESPACE_CHARACTER", e[e.START_TAG = 3] = "START_TAG", e[e.END_TAG = 4] = "END_TAG", e[e.COMMENT = 5] = "COMMENT", e[e.DOCTYPE = 6] = "DOCTYPE", e[e.EOF = 7] = "EOF", e[e.HIBERNATION = 8] = "HIBERNATION";
})(A ||= {});
function It(e, t) {
	for (let n = e.attrs.length - 1; n >= 0; n--) if (e.attrs[n].name === t) return e.attrs[n].value;
	return null;
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/entities/6.0.1/b31f3809f7f0ff09a74ef817ab8d031a53f2e09782aac6b2d83fa34f895449ba/node_modules/entities/dist/esm/generated/decode-data-html.js
var Lt = /* #__PURE__ */ new Uint16Array(/* #__PURE__ */ "ᵁ<Õıʊҝջאٵ۞ޢߖࠏ੊ઑඡ๭༉༦჊ረዡᐕᒝᓃᓟᔥ\0\0\0\0\0\0ᕫᛍᦍᰒᷝ὾⁠↰⊍⏀⏻⑂⠤⤒ⴈ⹈⿎〖㊺㘹㞬㣾㨨㩱㫠㬮ࠀEMabcfglmnoprstu\\bfms¦³¹ÈÏlig耻Æ䃆P耻&䀦cute耻Á䃁reve;䄂Āiyx}rc耻Â䃂;䐐r;쀀𝔄rave耻À䃀pha;䎑acr;䄀d;橓Āgp¡on;䄄f;쀀𝔸plyFunction;恡ing耻Å䃅Ācs¾Ãr;쀀𝒜ign;扔ilde耻Ã䃃ml耻Ä䃄ЀaceforsuåûþėĜĢħĪĀcrêòkslash;或Ŷöø;櫧ed;挆y;䐑ƀcrtąċĔause;戵noullis;愬a;䎒r;쀀𝔅pf;쀀𝔹eve;䋘còēmpeq;扎܀HOacdefhilorsuōőŖƀƞƢƵƷƺǜȕɳɸɾcy;䐧PY耻©䂩ƀcpyŝŢźute;䄆Ā;iŧŨ拒talDifferentialD;慅leys;愭ȀaeioƉƎƔƘron;䄌dil耻Ç䃇rc;䄈nint;戰ot;䄊ĀdnƧƭilla;䂸terDot;䂷òſi;䎧rcleȀDMPTǇǋǑǖot;抙inus;抖lus;投imes;抗oĀcsǢǸkwiseContourIntegral;戲eCurlyĀDQȃȏoubleQuote;思uote;怙ȀlnpuȞȨɇɕonĀ;eȥȦ户;橴ƀgitȯȶȺruent;扡nt;戯ourIntegral;戮ĀfrɌɎ;愂oduct;成nterClockwiseContourIntegral;戳oss;樯cr;쀀𝒞pĀ;Cʄʅ拓ap;才րDJSZacefiosʠʬʰʴʸˋ˗ˡ˦̳ҍĀ;oŹʥtrahd;椑cy;䐂cy;䐅cy;䐏ƀgrsʿ˄ˇger;怡r;憡hv;櫤Āayː˕ron;䄎;䐔lĀ;t˝˞戇a;䎔r;쀀𝔇Āaf˫̧Ācm˰̢riticalȀADGT̖̜̀̆cute;䂴oŴ̋̍;䋙bleAcute;䋝rave;䁠ilde;䋜ond;拄ferentialD;慆Ѱ̽\0\0\0͔͂\0Ѕf;쀀𝔻ƀ;DE͈͉͍䂨ot;惜qual;扐blèCDLRUVͣͲ΂ϏϢϸontourIntegraìȹoɴ͹\0\0ͻ»͉nArrow;懓Āeo·ΤftƀARTΐΖΡrrow;懐ightArrow;懔eåˊngĀLRΫτeftĀARγιrrow;柸ightArrow;柺ightArrow;柹ightĀATϘϞrrow;懒ee;抨pɁϩ\0\0ϯrrow;懑ownArrow;懕erticalBar;戥ǹABLRTaВЪаўѿͼrrowƀ;BUНОТ憓ar;椓pArrow;懵reve;䌑eft˒к\0ц\0ѐightVector;楐eeVector;楞ectorĀ;Bљњ憽ar;楖ightǔѧ\0ѱeeVector;楟ectorĀ;BѺѻ懁ar;楗eeĀ;A҆҇护rrow;憧ĀctҒҗr;쀀𝒟rok;䄐ࠀNTacdfglmopqstuxҽӀӄӋӞӢӧӮӵԡԯԶՒ՝ՠեG;䅊H耻Ð䃐cute耻É䃉ƀaiyӒӗӜron;䄚rc耻Ê䃊;䐭ot;䄖r;쀀𝔈rave耻È䃈ement;戈ĀapӺӾcr;䄒tyɓԆ\0\0ԒmallSquare;旻erySmallSquare;斫ĀgpԦԪon;䄘f;쀀𝔼silon;䎕uĀaiԼՉlĀ;TՂՃ橵ilde;扂librium;懌Āci՗՚r;愰m;橳a;䎗ml耻Ë䃋Āipժկsts;戃onentialE;慇ʀcfiosօֈ֍ֲ׌y;䐤r;쀀𝔉lledɓ֗\0\0֣mallSquare;旼erySmallSquare;斪Ͱֺ\0ֿ\0\0ׄf;쀀𝔽All;戀riertrf;愱cò׋؀JTabcdfgorstר׬ׯ׺؀ؒؖ؛؝أ٬ٲcy;䐃耻>䀾mmaĀ;d׷׸䎓;䏜reve;䄞ƀeiy؇،ؐdil;䄢rc;䄜;䐓ot;䄠r;쀀𝔊;拙pf;쀀𝔾eater̀EFGLSTصلَٖٛ٦qualĀ;Lؾؿ扥ess;招ullEqual;执reater;檢ess;扷lantEqual;橾ilde;扳cr;쀀𝒢;扫ЀAacfiosuڅڋږڛڞڪھۊRDcy;䐪Āctڐڔek;䋇;䁞irc;䄤r;愌lbertSpace;愋ǰگ\0ڲf;愍izontalLine;攀Āctۃۅòکrok;䄦mpńېۘownHumðįqual;扏܀EJOacdfgmnostuۺ۾܃܇܎ܚܞܡܨ݄ݸދޏޕcy;䐕lig;䄲cy;䐁cute耻Í䃍Āiyܓܘrc耻Î䃎;䐘ot;䄰r;愑rave耻Ì䃌ƀ;apܠܯܿĀcgܴܷr;䄪inaryI;慈lieóϝǴ݉\0ݢĀ;eݍݎ戬Āgrݓݘral;戫section;拂isibleĀCTݬݲomma;恣imes;恢ƀgptݿރވon;䄮f;쀀𝕀a;䎙cr;愐ilde;䄨ǫޚ\0ޞcy;䐆l耻Ï䃏ʀcfosuެ޷޼߂ߐĀiyޱ޵rc;䄴;䐙r;쀀𝔍pf;쀀𝕁ǣ߇\0ߌr;쀀𝒥rcy;䐈kcy;䐄΀HJacfosߤߨ߽߬߱ࠂࠈcy;䐥cy;䐌ppa;䎚Āey߶߻dil;䄶;䐚r;쀀𝔎pf;쀀𝕂cr;쀀𝒦րJTaceflmostࠥࠩࠬࡐࡣ঳সে্਷ੇcy;䐉耻<䀼ʀcmnpr࠷࠼ࡁࡄࡍute;䄹bda;䎛g;柪lacetrf;愒r;憞ƀaeyࡗ࡜ࡡron;䄽dil;䄻;䐛Āfsࡨ॰tԀACDFRTUVarࡾࢩࢱࣦ࣠ࣼयज़ΐ४Ānrࢃ࢏gleBracket;柨rowƀ;BR࢙࢚࢞憐ar;懤ightArrow;懆eiling;挈oǵࢷ\0ࣃbleBracket;柦nǔࣈ\0࣒eeVector;楡ectorĀ;Bࣛࣜ懃ar;楙loor;挊ightĀAV࣯ࣵrrow;憔ector;楎Āerँगeƀ;AVउऊऐ抣rrow;憤ector;楚iangleƀ;BEतथऩ抲ar;槏qual;抴pƀDTVषूौownVector;楑eeVector;楠ectorĀ;Bॖॗ憿ar;楘ectorĀ;B॥०憼ar;楒ightáΜs̀EFGLSTॾঋকঝঢভqualGreater;拚ullEqual;扦reater;扶ess;檡lantEqual;橽ilde;扲r;쀀𝔏Ā;eঽা拘ftarrow;懚idot;䄿ƀnpw৔ਖਛgȀLRlr৞৷ਂਐeftĀAR০৬rrow;柵ightArrow;柷ightArrow;柶eftĀarγਊightáοightáϊf;쀀𝕃erĀLRਢਬeftArrow;憙ightArrow;憘ƀchtਾੀੂòࡌ;憰rok;䅁;扪Ѐacefiosuਗ਼੝੠੷੼અઋ઎p;椅y;䐜Ādl੥੯iumSpace;恟lintrf;愳r;쀀𝔐nusPlus;戓pf;쀀𝕄cò੶;䎜ҀJacefostuણધભીଔଙඑ඗ඞcy;䐊cute;䅃ƀaey઴હાron;䅇dil;䅅;䐝ƀgswે૰଎ativeƀMTV૓૟૨ediumSpace;怋hiĀcn૦૘ë૙eryThiî૙tedĀGL૸ଆreaterGreateòٳessLesóੈLine;䀊r;쀀𝔑ȀBnptଢନଷ଺reak;恠BreakingSpace;䂠f;愕ڀ;CDEGHLNPRSTV୕ୖ୪୼஡௫ఄ౞಄ದ೘ൡඅ櫬Āou୛୤ngruent;扢pCap;扭oubleVerticalBar;戦ƀlqxஃஊ஛ement;戉ualĀ;Tஒஓ扠ilde;쀀≂̸ists;戄reater΀;EFGLSTஶஷ஽௉௓௘௥扯qual;扱ullEqual;쀀≧̸reater;쀀≫̸ess;批lantEqual;쀀⩾̸ilde;扵umpń௲௽ownHump;쀀≎̸qual;쀀≏̸eĀfsఊధtTriangleƀ;BEచఛడ拪ar;쀀⧏̸qual;括s̀;EGLSTవశ఼ౄోౘ扮qual;扰reater;扸ess;쀀≪̸lantEqual;쀀⩽̸ilde;扴estedĀGL౨౹reaterGreater;쀀⪢̸essLess;쀀⪡̸recedesƀ;ESಒಓಛ技qual;쀀⪯̸lantEqual;拠ĀeiಫಹverseElement;戌ghtTriangleƀ;BEೋೌ೒拫ar;쀀⧐̸qual;拭ĀquೝഌuareSuĀbp೨೹setĀ;E೰ೳ쀀⊏̸qual;拢ersetĀ;Eഃആ쀀⊐̸qual;拣ƀbcpഓതൎsetĀ;Eഛഞ쀀⊂⃒qual;抈ceedsȀ;ESTലള഻െ抁qual;쀀⪰̸lantEqual;拡ilde;쀀≿̸ersetĀ;E൘൛쀀⊃⃒qual;抉ildeȀ;EFT൮൯൵ൿ扁qual;扄ullEqual;扇ilde;扉erticalBar;戤cr;쀀𝒩ilde耻Ñ䃑;䎝܀Eacdfgmoprstuvලෂ෉෕ෛ෠෧෼ขภยา฿ไlig;䅒cute耻Ó䃓Āiy෎ීrc耻Ô䃔;䐞blac;䅐r;쀀𝔒rave耻Ò䃒ƀaei෮ෲ෶cr;䅌ga;䎩cron;䎟pf;쀀𝕆enCurlyĀDQฎบoubleQuote;怜uote;怘;橔Āclวฬr;쀀𝒪ash耻Ø䃘iŬื฼de耻Õ䃕es;樷ml耻Ö䃖erĀBP๋๠Āar๐๓r;怾acĀek๚๜;揞et;掴arenthesis;揜Ҁacfhilors๿ງຊຏຒດຝະ໼rtialD;戂y;䐟r;쀀𝔓i;䎦;䎠usMinus;䂱Āipຢອncareplanåڝf;愙Ȁ;eio຺ູ໠໤檻cedesȀ;EST່້໏໚扺qual;檯lantEqual;扼ilde;找me;怳Ādp໩໮uct;戏ortionĀ;aȥ໹l;戝Āci༁༆r;쀀𝒫;䎨ȀUfos༑༖༛༟OT耻\"䀢r;쀀𝔔pf;愚cr;쀀𝒬؀BEacefhiorsu༾གྷཇའཱིྦྷྪྭ႖ႩႴႾarr;椐G耻®䂮ƀcnrཎནབute;䅔g;柫rĀ;tཛྷཝ憠l;椖ƀaeyཧཬཱron;䅘dil;䅖;䐠Ā;vླྀཹ愜erseĀEUྂྙĀlq྇ྎement;戋uilibrium;懋pEquilibrium;楯r»ཹo;䎡ghtЀACDFTUVa࿁࿫࿳ဢဨၛႇϘĀnr࿆࿒gleBracket;柩rowƀ;BL࿜࿝࿡憒ar;懥eftArrow;懄eiling;按oǵ࿹\0စbleBracket;柧nǔည\0နeeVector;楝ectorĀ;Bဝသ懂ar;楕loor;挋Āerိ၃eƀ;AVဵံြ抢rrow;憦ector;楛iangleƀ;BEၐၑၕ抳ar;槐qual;抵pƀDTVၣၮၸownVector;楏eeVector;楜ectorĀ;Bႂႃ憾ar;楔ectorĀ;B႑႒懀ar;楓Āpuႛ႞f;愝ndImplies;楰ightarrow;懛ĀchႹႼr;愛;憱leDelayed;槴ڀHOacfhimoqstuფჱჷჽᄙᄞᅑᅖᅡᅧᆵᆻᆿĀCcჩხHcy;䐩y;䐨FTcy;䐬cute;䅚ʀ;aeiyᄈᄉᄎᄓᄗ檼ron;䅠dil;䅞rc;䅜;䐡r;쀀𝔖ortȀDLRUᄪᄴᄾᅉownArrow»ОeftArrow»࢚ightArrow»࿝pArrow;憑gma;䎣allCircle;战pf;쀀𝕊ɲᅭ\0\0ᅰt;戚areȀ;ISUᅻᅼᆉᆯ斡ntersection;抓uĀbpᆏᆞsetĀ;Eᆗᆘ抏qual;抑ersetĀ;Eᆨᆩ抐qual;抒nion;抔cr;쀀𝒮ar;拆ȀbcmpᇈᇛሉላĀ;sᇍᇎ拐etĀ;Eᇍᇕqual;抆ĀchᇠህeedsȀ;ESTᇭᇮᇴᇿ扻qual;檰lantEqual;扽ilde;承Tháྌ;我ƀ;esሒሓሣ拑rsetĀ;Eሜም抃qual;抇et»ሓրHRSacfhiorsሾቄ቉ቕ቞ቱቶኟዂወዑORN耻Þ䃞ADE;愢ĀHc቎ቒcy;䐋y;䐦Ābuቚቜ;䀉;䎤ƀaeyብቪቯron;䅤dil;䅢;䐢r;쀀𝔗Āeiቻ኉ǲኀ\0ኇefore;戴a;䎘Ācn኎ኘkSpace;쀀  Space;怉ldeȀ;EFTካኬኲኼ戼qual;扃ullEqual;扅ilde;扈pf;쀀𝕋ipleDot;惛Āctዖዛr;쀀𝒯rok;䅦ૡዷጎጚጦ\0ጬጱ\0\0\0\0\0ጸጽ፷ᎅ\0᏿ᐄᐊᐐĀcrዻጁute耻Ú䃚rĀ;oጇገ憟cir;楉rǣጓ\0጖y;䐎ve;䅬Āiyጞጣrc耻Û䃛;䐣blac;䅰r;쀀𝔘rave耻Ù䃙acr;䅪Ādiፁ፩erĀBPፈ፝Āarፍፐr;䁟acĀekፗፙ;揟et;掵arenthesis;揝onĀ;P፰፱拃lus;抎Āgp፻፿on;䅲f;쀀𝕌ЀADETadps᎕ᎮᎸᏄϨᏒᏗᏳrrowƀ;BDᅐᎠᎤar;椒ownArrow;懅ownArrow;憕quilibrium;楮eeĀ;AᏋᏌ报rrow;憥ownáϳerĀLRᏞᏨeftArrow;憖ightArrow;憗iĀ;lᏹᏺ䏒on;䎥ing;䅮cr;쀀𝒰ilde;䅨ml耻Ü䃜ҀDbcdefosvᐧᐬᐰᐳᐾᒅᒊᒐᒖash;披ar;櫫y;䐒ashĀ;lᐻᐼ抩;櫦Āerᑃᑅ;拁ƀbtyᑌᑐᑺar;怖Ā;iᑏᑕcalȀBLSTᑡᑥᑪᑴar;戣ine;䁼eparator;杘ilde;所ThinSpace;怊r;쀀𝔙pf;쀀𝕍cr;쀀𝒱dash;抪ʀcefosᒧᒬᒱᒶᒼirc;䅴dge;拀r;쀀𝔚pf;쀀𝕎cr;쀀𝒲Ȁfiosᓋᓐᓒᓘr;쀀𝔛;䎞pf;쀀𝕏cr;쀀𝒳ҀAIUacfosuᓱᓵᓹᓽᔄᔏᔔᔚᔠcy;䐯cy;䐇cy;䐮cute耻Ý䃝Āiyᔉᔍrc;䅶;䐫r;쀀𝔜pf;쀀𝕐cr;쀀𝒴ml;䅸ЀHacdefosᔵᔹᔿᕋᕏᕝᕠᕤcy;䐖cute;䅹Āayᕄᕉron;䅽;䐗ot;䅻ǲᕔ\0ᕛoWidtè૙a;䎖r;愨pf;愤cr;쀀𝒵௡ᖃᖊᖐ\0ᖰᖶᖿ\0\0\0\0ᗆᗛᗫᙟ᙭\0ᚕ᚛ᚲᚹ\0ᚾcute耻á䃡reve;䄃̀;Ediuyᖜᖝᖡᖣᖨᖭ戾;쀀∾̳;房rc耻â䃢te肻´̆;䐰lig耻æ䃦Ā;r²ᖺ;쀀𝔞rave耻à䃠ĀepᗊᗖĀfpᗏᗔsym;愵èᗓha;䎱ĀapᗟcĀclᗤᗧr;䄁g;樿ɤᗰ\0\0ᘊʀ;adsvᗺᗻᗿᘁᘇ戧nd;橕;橜lope;橘;橚΀;elmrszᘘᘙᘛᘞᘿᙏᙙ戠;榤e»ᘙsdĀ;aᘥᘦ戡ѡᘰᘲᘴᘶᘸᘺᘼᘾ;榨;榩;榪;榫;榬;榭;榮;榯tĀ;vᙅᙆ戟bĀ;dᙌᙍ抾;榝Āptᙔᙗh;戢»¹arr;捼Āgpᙣᙧon;䄅f;쀀𝕒΀;Eaeiop዁ᙻᙽᚂᚄᚇᚊ;橰cir;橯;扊d;手s;䀧roxĀ;e዁ᚒñᚃing耻å䃥ƀctyᚡᚦᚨr;쀀𝒶;䀪mpĀ;e዁ᚯñʈilde耻ã䃣ml耻ä䃤Āciᛂᛈoninôɲnt;樑ࠀNabcdefiklnoprsu᛭ᛱᜰ᜼ᝃᝈ᝸᝽០៦ᠹᡐᜍ᤽᥈ᥰot;櫭Ācrᛶ᜞kȀcepsᜀᜅᜍᜓong;扌psilon;䏶rime;怵imĀ;e᜚᜛戽q;拍Ŷᜢᜦee;抽edĀ;gᜬᜭ挅e»ᜭrkĀ;t፜᜷brk;掶Āoyᜁᝁ;䐱quo;怞ʀcmprtᝓ᝛ᝡᝤᝨausĀ;eĊĉptyv;榰séᜌnoõēƀahwᝯ᝱ᝳ;䎲;愶een;扬r;쀀𝔟g΀costuvwឍឝឳេ៕៛៞ƀaiuបពរðݠrc;旯p»፱ƀdptឤឨឭot;樀lus;樁imes;樂ɱឹ\0\0ើcup;樆ar;昅riangleĀdu៍្own;施p;斳plus;樄eåᑄåᒭarow;植ƀako៭ᠦᠵĀcn៲ᠣkƀlst៺֫᠂ozenge;槫riangleȀ;dlr᠒᠓᠘᠝斴own;斾eft;旂ight;斸k;搣Ʊᠫ\0ᠳƲᠯ\0ᠱ;斒;斑4;斓ck;斈ĀeoᠾᡍĀ;qᡃᡆ쀀=⃥uiv;쀀≡⃥t;挐Ȁptwxᡙᡞᡧᡬf;쀀𝕓Ā;tᏋᡣom»Ꮜtie;拈؀DHUVbdhmptuvᢅᢖᢪᢻᣗᣛᣬ᣿ᤅᤊᤐᤡȀLRlrᢎᢐᢒᢔ;敗;敔;敖;敓ʀ;DUduᢡᢢᢤᢦᢨ敐;敦;敩;敤;敧ȀLRlrᢳᢵᢷᢹ;敝;敚;敜;教΀;HLRhlrᣊᣋᣍᣏᣑᣓᣕ救;敬;散;敠;敫;敢;敟ox;槉ȀLRlrᣤᣦᣨᣪ;敕;敒;攐;攌ʀ;DUduڽ᣷᣹᣻᣽;敥;敨;攬;攴inus;抟lus;択imes;抠ȀLRlrᤙᤛᤝ᤟;敛;敘;攘;攔΀;HLRhlrᤰᤱᤳᤵᤷ᤻᤹攂;敪;敡;敞;攼;攤;攜Āevģ᥂bar耻¦䂦Ȁceioᥑᥖᥚᥠr;쀀𝒷mi;恏mĀ;e᜚᜜lƀ;bhᥨᥩᥫ䁜;槅sub;柈Ŭᥴ᥾lĀ;e᥹᥺怢t»᥺pƀ;Eeįᦅᦇ;檮Ā;qۜۛೡᦧ\0᧨ᨑᨕᨲ\0ᨷᩐ\0\0᪴\0\0᫁\0\0ᬡᬮ᭍᭒\0᯽\0ᰌƀcpr᦭ᦲ᧝ute;䄇̀;abcdsᦿᧀᧄ᧊᧕᧙戩nd;橄rcup;橉Āau᧏᧒p;橋p;橇ot;橀;쀀∩︀Āeo᧢᧥t;恁îړȀaeiu᧰᧻ᨁᨅǰ᧵\0᧸s;橍on;䄍dil耻ç䃧rc;䄉psĀ;sᨌᨍ橌m;橐ot;䄋ƀdmnᨛᨠᨦil肻¸ƭptyv;榲t脀¢;eᨭᨮ䂢räƲr;쀀𝔠ƀceiᨽᩀᩍy;䑇ckĀ;mᩇᩈ朓ark»ᩈ;䏇r΀;Ecefms᩟᩠ᩢᩫ᪤᪪᪮旋;槃ƀ;elᩩᩪᩭ䋆q;扗eɡᩴ\0\0᪈rrowĀlr᩼᪁eft;憺ight;憻ʀRSacd᪒᪔᪖᪚᪟»ཇ;擈st;抛irc;抚ash;抝nint;樐id;櫯cir;槂ubsĀ;u᪻᪼晣it»᪼ˬ᫇᫔᫺\0ᬊonĀ;eᫍᫎ䀺Ā;qÇÆɭ᫙\0\0᫢aĀ;t᫞᫟䀬;䁀ƀ;fl᫨᫩᫫戁îᅠeĀmx᫱᫶ent»᫩eóɍǧ᫾\0ᬇĀ;dኻᬂot;橭nôɆƀfryᬐᬔᬗ;쀀𝕔oäɔ脀©;sŕᬝr;愗Āaoᬥᬩrr;憵ss;朗Ācuᬲᬷr;쀀𝒸Ābpᬼ᭄Ā;eᭁᭂ櫏;櫑Ā;eᭉᭊ櫐;櫒dot;拯΀delprvw᭠᭬᭷ᮂᮬᯔ᯹arrĀlr᭨᭪;椸;椵ɰ᭲\0\0᭵r;拞c;拟arrĀ;p᭿ᮀ憶;椽̀;bcdosᮏᮐᮖᮡᮥᮨ截rcap;橈Āauᮛᮞp;橆p;橊ot;抍r;橅;쀀∪︀Ȁalrv᮵ᮿᯞᯣrrĀ;mᮼᮽ憷;椼yƀevwᯇᯔᯘqɰᯎ\0\0ᯒreã᭳uã᭵ee;拎edge;拏en耻¤䂤earrowĀlrᯮ᯳eft»ᮀight»ᮽeäᯝĀciᰁᰇoninôǷnt;戱lcty;挭ঀAHabcdefhijlorstuwz᰸᰻᰿ᱝᱩᱵᲊᲞᲬᲷ᳻᳿ᴍᵻᶑᶫᶻ᷆᷍rò΁ar;楥Ȁglrs᱈ᱍ᱒᱔ger;怠eth;愸òᄳhĀ;vᱚᱛ怐»ऊūᱡᱧarow;椏aã̕Āayᱮᱳron;䄏;䐴ƀ;ao̲ᱼᲄĀgrʿᲁr;懊tseq;橷ƀglmᲑᲔᲘ耻°䂰ta;䎴ptyv;榱ĀirᲣᲨsht;楿;쀀𝔡arĀlrᲳᲵ»ࣜ»သʀaegsv᳂͸᳖᳜᳠mƀ;oș᳊᳔ndĀ;ș᳑uit;晦amma;䏝in;拲ƀ;io᳧᳨᳸䃷de脀÷;o᳧ᳰntimes;拇nø᳷cy;䑒cɯᴆ\0\0ᴊrn;挞op;挍ʀlptuwᴘᴝᴢᵉᵕlar;䀤f;쀀𝕕ʀ;emps̋ᴭᴷᴽᵂqĀ;d͒ᴳot;扑inus;戸lus;戔quare;抡blebarwedgåúnƀadhᄮᵝᵧownarrowóᲃarpoonĀlrᵲᵶefôᲴighôᲶŢᵿᶅkaro÷གɯᶊ\0\0ᶎrn;挟op;挌ƀcotᶘᶣᶦĀryᶝᶡ;쀀𝒹;䑕l;槶rok;䄑Ādrᶰᶴot;拱iĀ;fᶺ᠖斿Āah᷀᷃ròЩaòྦangle;榦Āci᷒ᷕy;䑟grarr;柿ऀDacdefglmnopqrstuxḁḉḙḸոḼṉṡṾấắẽỡἪἷὄ὎὚ĀDoḆᴴoôᲉĀcsḎḔute耻é䃩ter;橮ȀaioyḢḧḱḶron;䄛rĀ;cḭḮ扖耻ê䃪lon;払;䑍ot;䄗ĀDrṁṅot;扒;쀀𝔢ƀ;rsṐṑṗ檚ave耻è䃨Ā;dṜṝ檖ot;檘Ȁ;ilsṪṫṲṴ檙nters;揧;愓Ā;dṹṺ檕ot;檗ƀapsẅẉẗcr;䄓tyƀ;svẒẓẕ戅et»ẓpĀ1;ẝẤĳạả;怄;怅怃ĀgsẪẬ;䅋p;怂ĀgpẴẸon;䄙f;쀀𝕖ƀalsỄỎỒrĀ;sỊị拕l;槣us;橱iƀ;lvỚớở䎵on»ớ;䏵ȀcsuvỪỳἋἣĀioữḱrc»Ḯɩỹ\0\0ỻíՈantĀglἂἆtr»ṝess»Ṻƀaeiἒ἖Ἒls;䀽st;扟vĀ;DȵἠD;橸parsl;槥ĀDaἯἳot;打rr;楱ƀcdiἾὁỸr;愯oô͒ĀahὉὋ;䎷耻ð䃰Āmrὓὗl耻ë䃫o;悬ƀcipὡὤὧl;䀡sôծĀeoὬὴctatioîՙnentialåչৡᾒ\0ᾞ\0ᾡᾧ\0\0ῆῌ\0ΐ\0ῦῪ \0 ⁚llingdotseñṄy;䑄male;晀ƀilrᾭᾳ῁lig;耀ﬃɩᾹ\0\0᾽g;耀ﬀig;耀ﬄ;쀀𝔣lig;耀ﬁlig;쀀fjƀaltῙ῜ῡt;晭ig;耀ﬂns;斱of;䆒ǰ΅\0ῳf;쀀𝕗ĀakֿῷĀ;vῼ´拔;櫙artint;樍Āao‌⁕Ācs‑⁒α‚‰‸⁅⁈\0⁐β•‥‧‪‬\0‮耻½䂽;慓耻¼䂼;慕;慙;慛Ƴ‴\0‶;慔;慖ʴ‾⁁\0\0⁃耻¾䂾;慗;慜5;慘ƶ⁌\0⁎;慚;慝8;慞l;恄wn;挢cr;쀀𝒻ࢀEabcdefgijlnorstv₂₉₟₥₰₴⃰⃵⃺⃿℃ℒℸ̗ℾ⅒↞Ā;lٍ₇;檌ƀcmpₐₕ₝ute;䇵maĀ;dₜ᳚䎳;檆reve;䄟Āiy₪₮rc;䄝;䐳ot;䄡Ȁ;lqsؾق₽⃉ƀ;qsؾٌ⃄lanô٥Ȁ;cdl٥⃒⃥⃕c;檩otĀ;o⃜⃝檀Ā;l⃢⃣檂;檄Ā;e⃪⃭쀀⋛︀s;檔r;쀀𝔤Ā;gٳ؛mel;愷cy;䑓Ȁ;Eajٚℌℎℐ;檒;檥;檤ȀEaesℛℝ℩ℴ;扩pĀ;p℣ℤ檊rox»ℤĀ;q℮ℯ檈Ā;q℮ℛim;拧pf;쀀𝕘Āci⅃ⅆr;愊mƀ;el٫ⅎ⅐;檎;檐茀>;cdlqr׮ⅠⅪⅮⅳⅹĀciⅥⅧ;檧r;橺ot;拗Par;榕uest;橼ʀadelsↄⅪ←ٖ↛ǰ↉\0↎proø₞r;楸qĀlqؿ↖lesó₈ií٫Āen↣↭rtneqq;쀀≩︀Å↪ԀAabcefkosy⇄⇇⇱⇵⇺∘∝∯≨≽ròΠȀilmr⇐⇔⇗⇛rsðᒄf»․ilôکĀdr⇠⇤cy;䑊ƀ;cwࣴ⇫⇯ir;楈;憭ar;意irc;䄥ƀalr∁∎∓rtsĀ;u∉∊晥it»∊lip;怦con;抹r;쀀𝔥sĀew∣∩arow;椥arow;椦ʀamopr∺∾≃≞≣rr;懿tht;戻kĀlr≉≓eftarrow;憩ightarrow;憪f;쀀𝕙bar;怕ƀclt≯≴≸r;쀀𝒽asè⇴rok;䄧Ābp⊂⊇ull;恃hen»ᱛૡ⊣\0⊪\0⊸⋅⋎\0⋕⋳\0\0⋸⌢⍧⍢⍿\0⎆⎪⎴cute耻í䃭ƀ;iyݱ⊰⊵rc耻î䃮;䐸Ācx⊼⊿y;䐵cl耻¡䂡ĀfrΟ⋉;쀀𝔦rave耻ì䃬Ȁ;inoܾ⋝⋩⋮Āin⋢⋦nt;樌t;戭fin;槜ta;愩lig;䄳ƀaop⋾⌚⌝ƀcgt⌅⌈⌗r;䄫ƀelpܟ⌏⌓inåގarôܠh;䄱f;抷ed;䆵ʀ;cfotӴ⌬⌱⌽⍁are;愅inĀ;t⌸⌹戞ie;槝doô⌙ʀ;celpݗ⍌⍐⍛⍡al;抺Āgr⍕⍙eróᕣã⍍arhk;樗rod;樼Ȁcgpt⍯⍲⍶⍻y;䑑on;䄯f;쀀𝕚a;䎹uest耻¿䂿Āci⎊⎏r;쀀𝒾nʀ;EdsvӴ⎛⎝⎡ӳ;拹ot;拵Ā;v⎦⎧拴;拳Ā;iݷ⎮lde;䄩ǫ⎸\0⎼cy;䑖l耻ï䃯̀cfmosu⏌⏗⏜⏡⏧⏵Āiy⏑⏕rc;䄵;䐹r;쀀𝔧ath;䈷pf;쀀𝕛ǣ⏬\0⏱r;쀀𝒿rcy;䑘kcy;䑔Ѐacfghjos␋␖␢␧␭␱␵␻ppaĀ;v␓␔䎺;䏰Āey␛␠dil;䄷;䐺r;쀀𝔨reen;䄸cy;䑅cy;䑜pf;쀀𝕜cr;쀀𝓀஀ABEHabcdefghjlmnoprstuv⑰⒁⒆⒍⒑┎┽╚▀♎♞♥♹♽⚚⚲⛘❝❨➋⟀⠁⠒ƀart⑷⑺⑼rò৆òΕail;椛arr;椎Ā;gঔ⒋;檋ar;楢ॣ⒥\0⒪\0⒱\0\0\0\0\0⒵Ⓔ\0ⓆⓈⓍ\0⓹ute;䄺mptyv;榴raîࡌbda;䎻gƀ;dlࢎⓁⓃ;榑åࢎ;檅uo耻«䂫rЀ;bfhlpst࢙ⓞⓦⓩ⓫⓮⓱⓵Ā;f࢝ⓣs;椟s;椝ë≒p;憫l;椹im;楳l;憢ƀ;ae⓿─┄檫il;椙Ā;s┉┊檭;쀀⪭︀ƀabr┕┙┝rr;椌rk;杲Āak┢┬cĀek┨┪;䁻;䁛Āes┱┳;榋lĀdu┹┻;榏;榍Ȁaeuy╆╋╖╘ron;䄾Ādi═╔il;䄼ìࢰâ┩;䐻Ȁcqrs╣╦╭╽a;椶uoĀ;rนᝆĀdu╲╷har;楧shar;楋h;憲ʀ;fgqs▋▌উ◳◿扤tʀahlrt▘▤▷◂◨rrowĀ;t࢙□aé⓶arpoonĀdu▯▴own»њp»०eftarrows;懇ightƀahs◍◖◞rrowĀ;sࣴࢧarpoonó྘quigarro÷⇰hreetimes;拋ƀ;qs▋ও◺lanôবʀ;cdgsব☊☍☝☨c;檨otĀ;o☔☕橿Ā;r☚☛檁;檃Ā;e☢☥쀀⋚︀s;檓ʀadegs☳☹☽♉♋pproøⓆot;拖qĀgq♃♅ôউgtò⒌ôছiíলƀilr♕࣡♚sht;楼;쀀𝔩Ā;Eজ♣;檑š♩♶rĀdu▲♮Ā;l॥♳;楪lk;斄cy;䑙ʀ;achtੈ⚈⚋⚑⚖rò◁orneòᴈard;楫ri;旺Āio⚟⚤dot;䅀ustĀ;a⚬⚭掰che»⚭ȀEaes⚻⚽⛉⛔;扨pĀ;p⛃⛄檉rox»⛄Ā;q⛎⛏檇Ā;q⛎⚻im;拦Ѐabnoptwz⛩⛴⛷✚✯❁❇❐Ānr⛮⛱g;柬r;懽rëࣁgƀlmr⛿✍✔eftĀar০✇ightá৲apsto;柼ightá৽parrowĀlr✥✩efô⓭ight;憬ƀafl✶✹✽r;榅;쀀𝕝us;樭imes;樴š❋❏st;戗áፎƀ;ef❗❘᠀旊nge»❘arĀ;l❤❥䀨t;榓ʀachmt❳❶❼➅➇ròࢨorneòᶌarĀ;d྘➃;業;怎ri;抿̀achiqt➘➝ੀ➢➮➻quo;怹r;쀀𝓁mƀ;egল➪➬;檍;檏Ābu┪➳oĀ;rฟ➹;怚rok;䅂萀<;cdhilqrࠫ⟒☹⟜⟠⟥⟪⟰Āci⟗⟙;檦r;橹reå◲mes;拉arr;楶uest;橻ĀPi⟵⟹ar;榖ƀ;ef⠀भ᠛旃rĀdu⠇⠍shar;楊har;楦Āen⠗⠡rtneqq;쀀≨︀Å⠞܀Dacdefhilnopsu⡀⡅⢂⢎⢓⢠⢥⢨⣚⣢⣤ઃ⣳⤂Dot;戺Ȁclpr⡎⡒⡣⡽r耻¯䂯Āet⡗⡙;時Ā;e⡞⡟朠se»⡟Ā;sျ⡨toȀ;dluျ⡳⡷⡻owîҌefôएðᏑker;斮Āoy⢇⢌mma;権;䐼ash;怔asuredangle»ᘦr;쀀𝔪o;愧ƀcdn⢯⢴⣉ro耻µ䂵Ȁ;acdᑤ⢽⣀⣄sôᚧir;櫰ot肻·Ƶusƀ;bd⣒ᤃ⣓戒Ā;uᴼ⣘;横ţ⣞⣡p;櫛ò−ðઁĀdp⣩⣮els;抧f;쀀𝕞Āct⣸⣽r;쀀𝓂pos»ᖝƀ;lm⤉⤊⤍䎼timap;抸ఀGLRVabcdefghijlmoprstuvw⥂⥓⥾⦉⦘⧚⧩⨕⨚⩘⩝⪃⪕⪤⪨⬄⬇⭄⭿⮮ⰴⱧⱼ⳩Āgt⥇⥋;쀀⋙̸Ā;v⥐௏쀀≫⃒ƀelt⥚⥲⥶ftĀar⥡⥧rrow;懍ightarrow;懎;쀀⋘̸Ā;v⥻ే쀀≪⃒ightarrow;懏ĀDd⦎⦓ash;抯ash;抮ʀbcnpt⦣⦧⦬⦱⧌la»˞ute;䅄g;쀀∠⃒ʀ;Eiop඄⦼⧀⧅⧈;쀀⩰̸d;쀀≋̸s;䅉roø඄urĀ;a⧓⧔普lĀ;s⧓ସǳ⧟\0⧣p肻\xA0ଷmpĀ;e௹ఀʀaeouy⧴⧾⨃⨐⨓ǰ⧹\0⧻;橃on;䅈dil;䅆ngĀ;dൾ⨊ot;쀀⩭̸p;橂;䐽ash;怓΀;Aadqsxஒ⨩⨭⨻⩁⩅⩐rr;懗rĀhr⨳⨶k;椤Ā;oᏲᏰot;쀀≐̸uiöୣĀei⩊⩎ar;椨í஘istĀ;s஠டr;쀀𝔫ȀEest௅⩦⩹⩼ƀ;qs஼⩭௡ƀ;qs஼௅⩴lanô௢ií௪Ā;rஶ⪁»ஷƀAap⪊⪍⪑rò⥱rr;憮ar;櫲ƀ;svྍ⪜ྌĀ;d⪡⪢拼;拺cy;䑚΀AEadest⪷⪺⪾⫂⫅⫶⫹rò⥦;쀀≦̸rr;憚r;急Ȁ;fqs఻⫎⫣⫯tĀar⫔⫙rro÷⫁ightarro÷⪐ƀ;qs఻⪺⫪lanôౕĀ;sౕ⫴»శiíౝĀ;rవ⫾iĀ;eచథiäඐĀpt⬌⬑f;쀀𝕟膀¬;in⬙⬚⬶䂬nȀ;Edvஉ⬤⬨⬮;쀀⋹̸ot;쀀⋵̸ǡஉ⬳⬵;拷;拶iĀ;vಸ⬼ǡಸ⭁⭃;拾;拽ƀaor⭋⭣⭩rȀ;ast୻⭕⭚⭟lleì୻l;쀀⫽⃥;쀀∂̸lint;樔ƀ;ceಒ⭰⭳uåಥĀ;cಘ⭸Ā;eಒ⭽ñಘȀAait⮈⮋⮝⮧rò⦈rrƀ;cw⮔⮕⮙憛;쀀⤳̸;쀀↝̸ghtarrow»⮕riĀ;eೋೖ΀chimpqu⮽⯍⯙⬄୸⯤⯯Ȁ;cerല⯆ഷ⯉uå൅;쀀𝓃ortɭ⬅\0\0⯖ará⭖mĀ;e൮⯟Ā;q൴൳suĀbp⯫⯭å೸åഋƀbcp⯶ⰑⰙȀ;Ees⯿ⰀഢⰄ抄;쀀⫅̸etĀ;eഛⰋqĀ;qണⰀcĀ;eലⰗñസȀ;EesⰢⰣൟⰧ抅;쀀⫆̸etĀ;e൘ⰮqĀ;qൠⰣȀgilrⰽⰿⱅⱇìௗlde耻ñ䃱çృiangleĀlrⱒⱜeftĀ;eచⱚñదightĀ;eೋⱥñ೗Ā;mⱬⱭ䎽ƀ;esⱴⱵⱹ䀣ro;愖p;怇ҀDHadgilrsⲏⲔⲙⲞⲣⲰⲶⳓⳣash;抭arr;椄p;쀀≍⃒ash;抬ĀetⲨⲬ;쀀≥⃒;쀀>⃒nfin;槞ƀAetⲽⳁⳅrr;椂;쀀≤⃒Ā;rⳊⳍ쀀<⃒ie;쀀⊴⃒ĀAtⳘⳜrr;椃rie;쀀⊵⃒im;쀀∼⃒ƀAan⳰⳴ⴂrr;懖rĀhr⳺⳽k;椣Ā;oᏧᏥear;椧ቓ᪕\0\0\0\0\0\0\0\0\0\0\0\0\0ⴭ\0ⴸⵈⵠⵥ⵲ⶄᬇ\0\0ⶍⶫ\0ⷈⷎ\0ⷜ⸙⸫⸾⹃Ācsⴱ᪗ute耻ó䃳ĀiyⴼⵅrĀ;c᪞ⵂ耻ô䃴;䐾ʀabios᪠ⵒⵗǈⵚlac;䅑v;樸old;榼lig;䅓Ācr⵩⵭ir;榿;쀀𝔬ͯ⵹\0\0⵼\0ⶂn;䋛ave耻ò䃲;槁Ābmⶈ෴ar;榵Ȁacitⶕ⶘ⶥⶨrò᪀Āir⶝ⶠr;榾oss;榻nå๒;槀ƀaeiⶱⶵⶹcr;䅍ga;䏉ƀcdnⷀⷅǍron;䎿;榶pf;쀀𝕠ƀaelⷔ⷗ǒr;榷rp;榹΀;adiosvⷪⷫⷮ⸈⸍⸐⸖戨rò᪆Ȁ;efmⷷⷸ⸂⸅橝rĀ;oⷾⷿ愴f»ⷿ耻ª䂪耻º䂺gof;抶r;橖lope;橗;橛ƀclo⸟⸡⸧ò⸁ash耻ø䃸l;折iŬⸯ⸴de耻õ䃵esĀ;aǛ⸺s;樶ml耻ö䃶bar;挽ૡ⹞\0⹽\0⺀⺝\0⺢⺹\0\0⻋ຜ\0⼓\0\0⼫⾼\0⿈rȀ;astЃ⹧⹲຅脀¶;l⹭⹮䂶leìЃɩ⹸\0\0⹻m;櫳;櫽y;䐿rʀcimpt⺋⺏⺓ᡥ⺗nt;䀥od;䀮il;怰enk;怱r;쀀𝔭ƀimo⺨⺰⺴Ā;v⺭⺮䏆;䏕maô੶ne;明ƀ;tv⺿⻀⻈䏀chfork»´;䏖Āau⻏⻟nĀck⻕⻝kĀ;h⇴⻛;愎ö⇴sҀ;abcdemst⻳⻴ᤈ⻹⻽⼄⼆⼊⼎䀫cir;樣ir;樢Āouᵀ⼂;樥;橲n肻±ຝim;樦wo;樧ƀipu⼙⼠⼥ntint;樕f;쀀𝕡nd耻£䂣Ԁ;Eaceinosu່⼿⽁⽄⽇⾁⾉⾒⽾⾶;檳p;檷uå໙Ā;c໎⽌̀;acens່⽙⽟⽦⽨⽾pproø⽃urlyeñ໙ñ໎ƀaes⽯⽶⽺pprox;檹qq;檵im;拨iíໟmeĀ;s⾈ຮ怲ƀEas⽸⾐⽺ð⽵ƀdfp໬⾙⾯ƀals⾠⾥⾪lar;挮ine;挒urf;挓Ā;t໻⾴ï໻rel;抰Āci⿀⿅r;쀀𝓅;䏈ncsp;怈̀fiopsu⿚⋢⿟⿥⿫⿱r;쀀𝔮pf;쀀𝕢rime;恗cr;쀀𝓆ƀaeo⿸〉〓tĀei⿾々rnionóڰnt;樖stĀ;e【】䀿ñἙô༔઀ABHabcdefhilmnoprstux぀けさすムㄎㄫㅇㅢㅲㆎ㈆㈕㈤㈩㉘㉮㉲㊐㊰㊷ƀartぇおがròႳòϝail;検aròᱥar;楤΀cdenqrtとふへみわゔヌĀeuねぱ;쀀∽̱te;䅕iãᅮmptyv;榳gȀ;del࿑らるろ;榒;榥å࿑uo耻»䂻rր;abcfhlpstw࿜ガクシスゼゾダッデナp;極Ā;f࿠ゴs;椠;椳s;椞ë≝ð✮l;楅im;楴l;憣;憝Āaiパフil;椚oĀ;nホボ戶aló༞ƀabrョリヮrò៥rk;杳ĀakンヽcĀekヹ・;䁽;䁝Āes㄂㄄;榌lĀduㄊㄌ;榎;榐Ȁaeuyㄗㄜㄧㄩron;䅙Ādiㄡㄥil;䅗ì࿲âヺ;䑀Ȁclqsㄴㄷㄽㅄa;椷dhar;楩uoĀ;rȎȍh;憳ƀacgㅎㅟངlȀ;ipsླྀㅘㅛႜnåႻarôྩt;断ƀilrㅩဣㅮsht;楽;쀀𝔯ĀaoㅷㆆrĀduㅽㅿ»ѻĀ;l႑ㆄ;楬Ā;vㆋㆌ䏁;䏱ƀgns㆕ㇹㇼht̀ahlrstㆤㆰ㇂㇘㇤㇮rrowĀ;t࿜ㆭaéトarpoonĀduㆻㆿowîㅾp»႒eftĀah㇊㇐rrowó࿪arpoonóՑightarrows;應quigarro÷ニhreetimes;拌g;䋚ingdotseñἲƀahm㈍㈐㈓rò࿪aòՑ;怏oustĀ;a㈞㈟掱che»㈟mid;櫮Ȁabpt㈲㈽㉀㉒Ānr㈷㈺g;柭r;懾rëဃƀafl㉇㉊㉎r;榆;쀀𝕣us;樮imes;樵Āap㉝㉧rĀ;g㉣㉤䀩t;榔olint;樒arò㇣Ȁachq㉻㊀Ⴜ㊅quo;怺r;쀀𝓇Ābu・㊊oĀ;rȔȓƀhir㊗㊛㊠reåㇸmes;拊iȀ;efl㊪ၙᠡ㊫方tri;槎luhar;楨;愞ൡ㋕㋛㋟㌬㌸㍱\0㍺㎤\0\0㏬㏰\0㐨㑈㑚㒭㒱㓊㓱\0㘖\0\0㘳cute;䅛quï➺Ԁ;Eaceinpsyᇭ㋳㋵㋿㌂㌋㌏㌟㌦㌩;檴ǰ㋺\0㋼;檸on;䅡uåᇾĀ;dᇳ㌇il;䅟rc;䅝ƀEas㌖㌘㌛;檶p;檺im;择olint;樓iíሄ;䑁otƀ;be㌴ᵇ㌵担;橦΀Aacmstx㍆㍊㍗㍛㍞㍣㍭rr;懘rĀhr㍐㍒ë∨Ā;oਸ਼਴t耻§䂧i;䀻war;椩mĀin㍩ðnuóñt;朶rĀ;o㍶⁕쀀𝔰Ȁacoy㎂㎆㎑㎠rp;景Āhy㎋㎏cy;䑉;䑈rtɭ㎙\0\0㎜iäᑤaraì⹯耻­䂭Āgm㎨㎴maƀ;fv㎱㎲㎲䏃;䏂Ѐ;deglnprካ㏅㏉㏎㏖㏞㏡㏦ot;橪Ā;q኱ኰĀ;E㏓㏔檞;檠Ā;E㏛㏜檝;檟e;扆lus;樤arr;楲aròᄽȀaeit㏸㐈㐏㐗Āls㏽㐄lsetmé㍪hp;樳parsl;槤Ādlᑣ㐔e;挣Ā;e㐜㐝檪Ā;s㐢㐣檬;쀀⪬︀ƀflp㐮㐳㑂tcy;䑌Ā;b㐸㐹䀯Ā;a㐾㐿槄r;挿f;쀀𝕤aĀdr㑍ЂesĀ;u㑔㑕晠it»㑕ƀcsu㑠㑹㒟Āau㑥㑯pĀ;sᆈ㑫;쀀⊓︀pĀ;sᆴ㑵;쀀⊔︀uĀbp㑿㒏ƀ;esᆗᆜ㒆etĀ;eᆗ㒍ñᆝƀ;esᆨᆭ㒖etĀ;eᆨ㒝ñᆮƀ;afᅻ㒦ְrť㒫ֱ»ᅼaròᅈȀcemt㒹㒾㓂㓅r;쀀𝓈tmîñiì㐕aræᆾĀar㓎㓕rĀ;f㓔ឿ昆Āan㓚㓭ightĀep㓣㓪psiloîỠhé⺯s»⡒ʀbcmnp㓻㕞ሉ㖋㖎Ҁ;Edemnprs㔎㔏㔑㔕㔞㔣㔬㔱㔶抂;櫅ot;檽Ā;dᇚ㔚ot;櫃ult;櫁ĀEe㔨㔪;櫋;把lus;檿arr;楹ƀeiu㔽㕒㕕tƀ;en㔎㕅㕋qĀ;qᇚ㔏eqĀ;q㔫㔨m;櫇Ābp㕚㕜;櫕;櫓c̀;acensᇭ㕬㕲㕹㕻㌦pproø㋺urlyeñᇾñᇳƀaes㖂㖈㌛pproø㌚qñ㌗g;晪ڀ123;Edehlmnps㖩㖬㖯ሜ㖲㖴㗀㗉㗕㗚㗟㗨㗭耻¹䂹耻²䂲耻³䂳;櫆Āos㖹㖼t;檾ub;櫘Ā;dሢ㗅ot;櫄sĀou㗏㗒l;柉b;櫗arr;楻ult;櫂ĀEe㗤㗦;櫌;抋lus;櫀ƀeiu㗴㘉㘌tƀ;enሜ㗼㘂qĀ;qሢ㖲eqĀ;q㗧㗤m;櫈Ābp㘑㘓;櫔;櫖ƀAan㘜㘠㘭rr;懙rĀhr㘦㘨ë∮Ā;oਫ਩war;椪lig耻ß䃟௡㙑㙝㙠ዎ㙳㙹\0㙾㛂\0\0\0\0\0㛛㜃\0㜉㝬\0\0\0㞇ɲ㙖\0\0㙛get;挖;䏄rë๟ƀaey㙦㙫㙰ron;䅥dil;䅣;䑂lrec;挕r;쀀𝔱Ȁeiko㚆㚝㚵㚼ǲ㚋\0㚑eĀ4fኄኁaƀ;sv㚘㚙㚛䎸ym;䏑Ācn㚢㚲kĀas㚨㚮pproø዁im»ኬsðኞĀas㚺㚮ð዁rn耻þ䃾Ǭ̟㛆⋧es膀×;bd㛏㛐㛘䃗Ā;aᤏ㛕r;樱;樰ƀeps㛡㛣㜀á⩍Ȁ;bcf҆㛬㛰㛴ot;挶ir;櫱Ā;o㛹㛼쀀𝕥rk;櫚á㍢rime;怴ƀaip㜏㜒㝤dåቈ΀adempst㜡㝍㝀㝑㝗㝜㝟ngleʀ;dlqr㜰㜱㜶㝀㝂斵own»ᶻeftĀ;e⠀㜾ñम;扜ightĀ;e㊪㝋ñၚot;旬inus;樺lus;樹b;槍ime;樻ezium;揢ƀcht㝲㝽㞁Āry㝷㝻;쀀𝓉;䑆cy;䑛rok;䅧Āio㞋㞎xô᝷headĀlr㞗㞠eftarro÷ࡏightarrow»ཝऀAHabcdfghlmoprstuw㟐㟓㟗㟤㟰㟼㠎㠜㠣㠴㡑㡝㡫㢩㣌㣒㣪㣶ròϭar;楣Ācr㟜㟢ute耻ú䃺òᅐrǣ㟪\0㟭y;䑞ve;䅭Āiy㟵㟺rc耻û䃻;䑃ƀabh㠃㠆㠋ròᎭlac;䅱aòᏃĀir㠓㠘sht;楾;쀀𝔲rave耻ù䃹š㠧㠱rĀlr㠬㠮»ॗ»ႃlk;斀Āct㠹㡍ɯ㠿\0\0㡊rnĀ;e㡅㡆挜r»㡆op;挏ri;旸Āal㡖㡚cr;䅫肻¨͉Āgp㡢㡦on;䅳f;쀀𝕦̀adhlsuᅋ㡸㡽፲㢑㢠ownáᎳarpoonĀlr㢈㢌efô㠭ighô㠯iƀ;hl㢙㢚㢜䏅»ᏺon»㢚parrows;懈ƀcit㢰㣄㣈ɯ㢶\0\0㣁rnĀ;e㢼㢽挝r»㢽op;挎ng;䅯ri;旹cr;쀀𝓊ƀdir㣙㣝㣢ot;拰lde;䅩iĀ;f㜰㣨»᠓Āam㣯㣲rò㢨l耻ü䃼angle;榧ހABDacdeflnoprsz㤜㤟㤩㤭㦵㦸㦽㧟㧤㧨㧳㧹㧽㨁㨠ròϷarĀ;v㤦㤧櫨;櫩asèϡĀnr㤲㤷grt;榜΀eknprst㓣㥆㥋㥒㥝㥤㦖appá␕othinçẖƀhir㓫⻈㥙opô⾵Ā;hᎷ㥢ïㆍĀiu㥩㥭gmá㎳Ābp㥲㦄setneqĀ;q㥽㦀쀀⊊︀;쀀⫋︀setneqĀ;q㦏㦒쀀⊋︀;쀀⫌︀Āhr㦛㦟etá㚜iangleĀlr㦪㦯eft»थight»ၑy;䐲ash»ံƀelr㧄㧒㧗ƀ;beⷪ㧋㧏ar;抻q;扚lip;拮Ābt㧜ᑨaòᑩr;쀀𝔳tré㦮suĀbp㧯㧱»ജ»൙pf;쀀𝕧roð໻tré㦴Ācu㨆㨋r;쀀𝓋Ābp㨐㨘nĀEe㦀㨖»㥾nĀEe㦒㨞»㦐igzag;榚΀cefoprs㨶㨻㩖㩛㩔㩡㩪irc;䅵Ādi㩀㩑Ābg㩅㩉ar;機eĀ;qᗺ㩏;扙erp;愘r;쀀𝔴pf;쀀𝕨Ā;eᑹ㩦atèᑹcr;쀀𝓌ૣណ㪇\0㪋\0㪐㪛\0\0㪝㪨㪫㪯\0\0㫃㫎\0㫘ៜ៟tré៑r;쀀𝔵ĀAa㪔㪗ròσrò৶;䎾ĀAa㪡㪤ròθrò৫að✓is;拻ƀdptឤ㪵㪾Āfl㪺ឩ;쀀𝕩imåឲĀAa㫇㫊ròώròਁĀcq㫒ីr;쀀𝓍Āpt៖㫜ré។Ѐacefiosu㫰㫽㬈㬌㬑㬕㬛㬡cĀuy㫶㫻te耻ý䃽;䑏Āiy㬂㬆rc;䅷;䑋n耻¥䂥r;쀀𝔶cy;䑗pf;쀀𝕪cr;쀀𝓎Ācm㬦㬩y;䑎l耻ÿ䃿Ԁacdefhiosw㭂㭈㭔㭘㭤㭩㭭㭴㭺㮀cute;䅺Āay㭍㭒ron;䅾;䐷ot;䅼Āet㭝㭡træᕟa;䎶r;쀀𝔷cy;䐶grarr;懝pf;쀀𝕫cr;쀀𝓏Ājn㮅㮇;怍j;怌".split("").map((e) => e.charCodeAt(0))), Rt = /* @__PURE__ */ new Map([
	[0, 65533],
	[128, 8364],
	[130, 8218],
	[131, 402],
	[132, 8222],
	[133, 8230],
	[134, 8224],
	[135, 8225],
	[136, 710],
	[137, 8240],
	[138, 352],
	[139, 8249],
	[140, 338],
	[142, 381],
	[145, 8216],
	[146, 8217],
	[147, 8220],
	[148, 8221],
	[149, 8226],
	[150, 8211],
	[151, 8212],
	[152, 732],
	[153, 8482],
	[154, 353],
	[155, 8250],
	[156, 339],
	[158, 382],
	[159, 376]
]);
String.fromCodePoint;
function zt(e) {
	return e >= 55296 && e <= 57343 || e > 1114111 ? 65533 : Rt.get(e) ?? e;
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/entities/6.0.1/b31f3809f7f0ff09a74ef817ab8d031a53f2e09782aac6b2d83fa34f895449ba/node_modules/entities/dist/esm/decode.js
var j;
(function(e) {
	e[e.NUM = 35] = "NUM", e[e.SEMI = 59] = "SEMI", e[e.EQUALS = 61] = "EQUALS", e[e.ZERO = 48] = "ZERO", e[e.NINE = 57] = "NINE", e[e.LOWER_A = 97] = "LOWER_A", e[e.LOWER_F = 102] = "LOWER_F", e[e.LOWER_X = 120] = "LOWER_X", e[e.LOWER_Z = 122] = "LOWER_Z", e[e.UPPER_A = 65] = "UPPER_A", e[e.UPPER_F = 70] = "UPPER_F", e[e.UPPER_Z = 90] = "UPPER_Z";
})(j ||= {});
var Bt = 32, Vt;
(function(e) {
	e[e.VALUE_LENGTH = 49152] = "VALUE_LENGTH", e[e.BRANCH_LENGTH = 16256] = "BRANCH_LENGTH", e[e.JUMP_TABLE = 127] = "JUMP_TABLE";
})(Vt ||= {});
function Ht(e) {
	return e >= j.ZERO && e <= j.NINE;
}
function Ut(e) {
	return e >= j.UPPER_A && e <= j.UPPER_F || e >= j.LOWER_A && e <= j.LOWER_F;
}
function Wt(e) {
	return e >= j.UPPER_A && e <= j.UPPER_Z || e >= j.LOWER_A && e <= j.LOWER_Z || Ht(e);
}
function Gt(e) {
	return e === j.EQUALS || Wt(e);
}
var M;
(function(e) {
	e[e.EntityStart = 0] = "EntityStart", e[e.NumericStart = 1] = "NumericStart", e[e.NumericDecimal = 2] = "NumericDecimal", e[e.NumericHex = 3] = "NumericHex", e[e.NamedEntity = 4] = "NamedEntity";
})(M ||= {});
var Kt;
(function(e) {
	e[e.Legacy = 0] = "Legacy", e[e.Strict = 1] = "Strict", e[e.Attribute = 2] = "Attribute";
})(Kt ||= {});
var qt = class {
	constructor(e, t, n) {
		this.decodeTree = e, this.emitCodePoint = t, this.errors = n, this.state = M.EntityStart, this.consumed = 1, this.result = 0, this.treeIndex = 0, this.excess = 1, this.decodeMode = Kt.Strict;
	}
	startEntity(e) {
		this.decodeMode = e, this.state = M.EntityStart, this.result = 0, this.treeIndex = 0, this.excess = 1, this.consumed = 1;
	}
	write(e, t) {
		switch (this.state) {
			case M.EntityStart: return e.charCodeAt(t) === j.NUM ? (this.state = M.NumericStart, this.consumed += 1, this.stateNumericStart(e, t + 1)) : (this.state = M.NamedEntity, this.stateNamedEntity(e, t));
			case M.NumericStart: return this.stateNumericStart(e, t);
			case M.NumericDecimal: return this.stateNumericDecimal(e, t);
			case M.NumericHex: return this.stateNumericHex(e, t);
			case M.NamedEntity: return this.stateNamedEntity(e, t);
		}
	}
	stateNumericStart(e, t) {
		return t >= e.length ? -1 : (e.charCodeAt(t) | Bt) === j.LOWER_X ? (this.state = M.NumericHex, this.consumed += 1, this.stateNumericHex(e, t + 1)) : (this.state = M.NumericDecimal, this.stateNumericDecimal(e, t));
	}
	addToNumericResult(e, t, n, r) {
		if (t !== n) {
			let i = n - t;
			this.result = this.result * r ** +i + Number.parseInt(e.substr(t, i), r), this.consumed += i;
		}
	}
	stateNumericHex(e, t) {
		let n = t;
		for (; t < e.length;) {
			let r = e.charCodeAt(t);
			if (Ht(r) || Ut(r)) t += 1;
			else return this.addToNumericResult(e, n, t, 16), this.emitNumericEntity(r, 3);
		}
		return this.addToNumericResult(e, n, t, 16), -1;
	}
	stateNumericDecimal(e, t) {
		let n = t;
		for (; t < e.length;) {
			let r = e.charCodeAt(t);
			if (Ht(r)) t += 1;
			else return this.addToNumericResult(e, n, t, 10), this.emitNumericEntity(r, 2);
		}
		return this.addToNumericResult(e, n, t, 10), -1;
	}
	emitNumericEntity(e, t) {
		var n;
		if (this.consumed <= t) return (n = this.errors) == null || n.absenceOfDigitsInNumericCharacterReference(this.consumed), 0;
		if (e === j.SEMI) this.consumed += 1;
		else if (this.decodeMode === Kt.Strict) return 0;
		return this.emitCodePoint(zt(this.result), this.consumed), this.errors && (e !== j.SEMI && this.errors.missingSemicolonAfterCharacterReference(), this.errors.validateNumericCharacterReference(this.result)), this.consumed;
	}
	stateNamedEntity(e, t) {
		let { decodeTree: n } = this, r = n[this.treeIndex], i = (r & Vt.VALUE_LENGTH) >> 14;
		for (; t < e.length; t++, this.excess++) {
			let a = e.charCodeAt(t);
			if (this.treeIndex = Jt(n, r, this.treeIndex + Math.max(1, i), a), this.treeIndex < 0) return this.result === 0 || this.decodeMode === Kt.Attribute && (i === 0 || Gt(a)) ? 0 : this.emitNotTerminatedNamedEntity();
			if (r = n[this.treeIndex], i = (r & Vt.VALUE_LENGTH) >> 14, i !== 0) {
				if (a === j.SEMI) return this.emitNamedEntityData(this.treeIndex, i, this.consumed + this.excess);
				this.decodeMode !== Kt.Strict && (this.result = this.treeIndex, this.consumed += this.excess, this.excess = 0);
			}
		}
		return -1;
	}
	emitNotTerminatedNamedEntity() {
		var e;
		let { result: t, decodeTree: n } = this, r = (n[t] & Vt.VALUE_LENGTH) >> 14;
		return this.emitNamedEntityData(t, r, this.consumed), (e = this.errors) == null || e.missingSemicolonAfterCharacterReference(), this.consumed;
	}
	emitNamedEntityData(e, t, n) {
		let { decodeTree: r } = this;
		return this.emitCodePoint(t === 1 ? r[e] & ~Vt.VALUE_LENGTH : r[e + 1], n), t === 3 && this.emitCodePoint(r[e + 2], n), n;
	}
	end() {
		var e;
		switch (this.state) {
			case M.NamedEntity: return this.result !== 0 && (this.decodeMode !== Kt.Attribute || this.result === this.treeIndex) ? this.emitNotTerminatedNamedEntity() : 0;
			case M.NumericDecimal: return this.emitNumericEntity(0, 2);
			case M.NumericHex: return this.emitNumericEntity(0, 3);
			case M.NumericStart: return (e = this.errors) == null || e.absenceOfDigitsInNumericCharacterReference(this.consumed), 0;
			case M.EntityStart: return 0;
		}
	}
};
function Jt(e, t, n, r) {
	let i = (t & Vt.BRANCH_LENGTH) >> 7, a = t & Vt.JUMP_TABLE;
	if (i === 0) return a !== 0 && r === a ? n : -1;
	if (a) {
		let t = r - a;
		return t < 0 || t >= i ? -1 : e[n + t] - 1;
	}
	let o = n, s = o + i - 1;
	for (; o <= s;) {
		let t = o + s >>> 1, n = e[t];
		if (n < r) o = t + 1;
		else if (n > r) s = t - 1;
		else return e[t + i];
	}
	return -1;
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/parse5/7.3.0/4acbced08702cc2b068edf7ec6fbf0576d666aa2a4d5c21d9bdc1830acf8a8e2/node_modules/parse5/dist/common/html.js
var N;
(function(e) {
	e.HTML = "http://www.w3.org/1999/xhtml", e.MATHML = "http://www.w3.org/1998/Math/MathML", e.SVG = "http://www.w3.org/2000/svg", e.XLINK = "http://www.w3.org/1999/xlink", e.XML = "http://www.w3.org/XML/1998/namespace", e.XMLNS = "http://www.w3.org/2000/xmlns/";
})(N ||= {});
var Yt;
(function(e) {
	e.TYPE = "type", e.ACTION = "action", e.ENCODING = "encoding", e.PROMPT = "prompt", e.NAME = "name", e.COLOR = "color", e.FACE = "face", e.SIZE = "size";
})(Yt ||= {});
var Xt;
(function(e) {
	e.NO_QUIRKS = "no-quirks", e.QUIRKS = "quirks", e.LIMITED_QUIRKS = "limited-quirks";
})(Xt ||= {});
var P;
(function(e) {
	e.A = "a", e.ADDRESS = "address", e.ANNOTATION_XML = "annotation-xml", e.APPLET = "applet", e.AREA = "area", e.ARTICLE = "article", e.ASIDE = "aside", e.B = "b", e.BASE = "base", e.BASEFONT = "basefont", e.BGSOUND = "bgsound", e.BIG = "big", e.BLOCKQUOTE = "blockquote", e.BODY = "body", e.BR = "br", e.BUTTON = "button", e.CAPTION = "caption", e.CENTER = "center", e.CODE = "code", e.COL = "col", e.COLGROUP = "colgroup", e.DD = "dd", e.DESC = "desc", e.DETAILS = "details", e.DIALOG = "dialog", e.DIR = "dir", e.DIV = "div", e.DL = "dl", e.DT = "dt", e.EM = "em", e.EMBED = "embed", e.FIELDSET = "fieldset", e.FIGCAPTION = "figcaption", e.FIGURE = "figure", e.FONT = "font", e.FOOTER = "footer", e.FOREIGN_OBJECT = "foreignObject", e.FORM = "form", e.FRAME = "frame", e.FRAMESET = "frameset", e.H1 = "h1", e.H2 = "h2", e.H3 = "h3", e.H4 = "h4", e.H5 = "h5", e.H6 = "h6", e.HEAD = "head", e.HEADER = "header", e.HGROUP = "hgroup", e.HR = "hr", e.HTML = "html", e.I = "i", e.IMG = "img", e.IMAGE = "image", e.INPUT = "input", e.IFRAME = "iframe", e.KEYGEN = "keygen", e.LABEL = "label", e.LI = "li", e.LINK = "link", e.LISTING = "listing", e.MAIN = "main", e.MALIGNMARK = "malignmark", e.MARQUEE = "marquee", e.MATH = "math", e.MENU = "menu", e.META = "meta", e.MGLYPH = "mglyph", e.MI = "mi", e.MO = "mo", e.MN = "mn", e.MS = "ms", e.MTEXT = "mtext", e.NAV = "nav", e.NOBR = "nobr", e.NOFRAMES = "noframes", e.NOEMBED = "noembed", e.NOSCRIPT = "noscript", e.OBJECT = "object", e.OL = "ol", e.OPTGROUP = "optgroup", e.OPTION = "option", e.P = "p", e.PARAM = "param", e.PLAINTEXT = "plaintext", e.PRE = "pre", e.RB = "rb", e.RP = "rp", e.RT = "rt", e.RTC = "rtc", e.RUBY = "ruby", e.S = "s", e.SCRIPT = "script", e.SEARCH = "search", e.SECTION = "section", e.SELECT = "select", e.SOURCE = "source", e.SMALL = "small", e.SPAN = "span", e.STRIKE = "strike", e.STRONG = "strong", e.STYLE = "style", e.SUB = "sub", e.SUMMARY = "summary", e.SUP = "sup", e.TABLE = "table", e.TBODY = "tbody", e.TEMPLATE = "template", e.TEXTAREA = "textarea", e.TFOOT = "tfoot", e.TD = "td", e.TH = "th", e.THEAD = "thead", e.TITLE = "title", e.TR = "tr", e.TRACK = "track", e.TT = "tt", e.U = "u", e.UL = "ul", e.SVG = "svg", e.VAR = "var", e.WBR = "wbr", e.XMP = "xmp";
})(P ||= {});
var F;
(function(e) {
	e[e.UNKNOWN = 0] = "UNKNOWN", e[e.A = 1] = "A", e[e.ADDRESS = 2] = "ADDRESS", e[e.ANNOTATION_XML = 3] = "ANNOTATION_XML", e[e.APPLET = 4] = "APPLET", e[e.AREA = 5] = "AREA", e[e.ARTICLE = 6] = "ARTICLE", e[e.ASIDE = 7] = "ASIDE", e[e.B = 8] = "B", e[e.BASE = 9] = "BASE", e[e.BASEFONT = 10] = "BASEFONT", e[e.BGSOUND = 11] = "BGSOUND", e[e.BIG = 12] = "BIG", e[e.BLOCKQUOTE = 13] = "BLOCKQUOTE", e[e.BODY = 14] = "BODY", e[e.BR = 15] = "BR", e[e.BUTTON = 16] = "BUTTON", e[e.CAPTION = 17] = "CAPTION", e[e.CENTER = 18] = "CENTER", e[e.CODE = 19] = "CODE", e[e.COL = 20] = "COL", e[e.COLGROUP = 21] = "COLGROUP", e[e.DD = 22] = "DD", e[e.DESC = 23] = "DESC", e[e.DETAILS = 24] = "DETAILS", e[e.DIALOG = 25] = "DIALOG", e[e.DIR = 26] = "DIR", e[e.DIV = 27] = "DIV", e[e.DL = 28] = "DL", e[e.DT = 29] = "DT", e[e.EM = 30] = "EM", e[e.EMBED = 31] = "EMBED", e[e.FIELDSET = 32] = "FIELDSET", e[e.FIGCAPTION = 33] = "FIGCAPTION", e[e.FIGURE = 34] = "FIGURE", e[e.FONT = 35] = "FONT", e[e.FOOTER = 36] = "FOOTER", e[e.FOREIGN_OBJECT = 37] = "FOREIGN_OBJECT", e[e.FORM = 38] = "FORM", e[e.FRAME = 39] = "FRAME", e[e.FRAMESET = 40] = "FRAMESET", e[e.H1 = 41] = "H1", e[e.H2 = 42] = "H2", e[e.H3 = 43] = "H3", e[e.H4 = 44] = "H4", e[e.H5 = 45] = "H5", e[e.H6 = 46] = "H6", e[e.HEAD = 47] = "HEAD", e[e.HEADER = 48] = "HEADER", e[e.HGROUP = 49] = "HGROUP", e[e.HR = 50] = "HR", e[e.HTML = 51] = "HTML", e[e.I = 52] = "I", e[e.IMG = 53] = "IMG", e[e.IMAGE = 54] = "IMAGE", e[e.INPUT = 55] = "INPUT", e[e.IFRAME = 56] = "IFRAME", e[e.KEYGEN = 57] = "KEYGEN", e[e.LABEL = 58] = "LABEL", e[e.LI = 59] = "LI", e[e.LINK = 60] = "LINK", e[e.LISTING = 61] = "LISTING", e[e.MAIN = 62] = "MAIN", e[e.MALIGNMARK = 63] = "MALIGNMARK", e[e.MARQUEE = 64] = "MARQUEE", e[e.MATH = 65] = "MATH", e[e.MENU = 66] = "MENU", e[e.META = 67] = "META", e[e.MGLYPH = 68] = "MGLYPH", e[e.MI = 69] = "MI", e[e.MO = 70] = "MO", e[e.MN = 71] = "MN", e[e.MS = 72] = "MS", e[e.MTEXT = 73] = "MTEXT", e[e.NAV = 74] = "NAV", e[e.NOBR = 75] = "NOBR", e[e.NOFRAMES = 76] = "NOFRAMES", e[e.NOEMBED = 77] = "NOEMBED", e[e.NOSCRIPT = 78] = "NOSCRIPT", e[e.OBJECT = 79] = "OBJECT", e[e.OL = 80] = "OL", e[e.OPTGROUP = 81] = "OPTGROUP", e[e.OPTION = 82] = "OPTION", e[e.P = 83] = "P", e[e.PARAM = 84] = "PARAM", e[e.PLAINTEXT = 85] = "PLAINTEXT", e[e.PRE = 86] = "PRE", e[e.RB = 87] = "RB", e[e.RP = 88] = "RP", e[e.RT = 89] = "RT", e[e.RTC = 90] = "RTC", e[e.RUBY = 91] = "RUBY", e[e.S = 92] = "S", e[e.SCRIPT = 93] = "SCRIPT", e[e.SEARCH = 94] = "SEARCH", e[e.SECTION = 95] = "SECTION", e[e.SELECT = 96] = "SELECT", e[e.SOURCE = 97] = "SOURCE", e[e.SMALL = 98] = "SMALL", e[e.SPAN = 99] = "SPAN", e[e.STRIKE = 100] = "STRIKE", e[e.STRONG = 101] = "STRONG", e[e.STYLE = 102] = "STYLE", e[e.SUB = 103] = "SUB", e[e.SUMMARY = 104] = "SUMMARY", e[e.SUP = 105] = "SUP", e[e.TABLE = 106] = "TABLE", e[e.TBODY = 107] = "TBODY", e[e.TEMPLATE = 108] = "TEMPLATE", e[e.TEXTAREA = 109] = "TEXTAREA", e[e.TFOOT = 110] = "TFOOT", e[e.TD = 111] = "TD", e[e.TH = 112] = "TH", e[e.THEAD = 113] = "THEAD", e[e.TITLE = 114] = "TITLE", e[e.TR = 115] = "TR", e[e.TRACK = 116] = "TRACK", e[e.TT = 117] = "TT", e[e.U = 118] = "U", e[e.UL = 119] = "UL", e[e.SVG = 120] = "SVG", e[e.VAR = 121] = "VAR", e[e.WBR = 122] = "WBR", e[e.XMP = 123] = "XMP";
})(F ||= {});
var Zt = /* @__PURE__ */ new Map([
	[P.A, F.A],
	[P.ADDRESS, F.ADDRESS],
	[P.ANNOTATION_XML, F.ANNOTATION_XML],
	[P.APPLET, F.APPLET],
	[P.AREA, F.AREA],
	[P.ARTICLE, F.ARTICLE],
	[P.ASIDE, F.ASIDE],
	[P.B, F.B],
	[P.BASE, F.BASE],
	[P.BASEFONT, F.BASEFONT],
	[P.BGSOUND, F.BGSOUND],
	[P.BIG, F.BIG],
	[P.BLOCKQUOTE, F.BLOCKQUOTE],
	[P.BODY, F.BODY],
	[P.BR, F.BR],
	[P.BUTTON, F.BUTTON],
	[P.CAPTION, F.CAPTION],
	[P.CENTER, F.CENTER],
	[P.CODE, F.CODE],
	[P.COL, F.COL],
	[P.COLGROUP, F.COLGROUP],
	[P.DD, F.DD],
	[P.DESC, F.DESC],
	[P.DETAILS, F.DETAILS],
	[P.DIALOG, F.DIALOG],
	[P.DIR, F.DIR],
	[P.DIV, F.DIV],
	[P.DL, F.DL],
	[P.DT, F.DT],
	[P.EM, F.EM],
	[P.EMBED, F.EMBED],
	[P.FIELDSET, F.FIELDSET],
	[P.FIGCAPTION, F.FIGCAPTION],
	[P.FIGURE, F.FIGURE],
	[P.FONT, F.FONT],
	[P.FOOTER, F.FOOTER],
	[P.FOREIGN_OBJECT, F.FOREIGN_OBJECT],
	[P.FORM, F.FORM],
	[P.FRAME, F.FRAME],
	[P.FRAMESET, F.FRAMESET],
	[P.H1, F.H1],
	[P.H2, F.H2],
	[P.H3, F.H3],
	[P.H4, F.H4],
	[P.H5, F.H5],
	[P.H6, F.H6],
	[P.HEAD, F.HEAD],
	[P.HEADER, F.HEADER],
	[P.HGROUP, F.HGROUP],
	[P.HR, F.HR],
	[P.HTML, F.HTML],
	[P.I, F.I],
	[P.IMG, F.IMG],
	[P.IMAGE, F.IMAGE],
	[P.INPUT, F.INPUT],
	[P.IFRAME, F.IFRAME],
	[P.KEYGEN, F.KEYGEN],
	[P.LABEL, F.LABEL],
	[P.LI, F.LI],
	[P.LINK, F.LINK],
	[P.LISTING, F.LISTING],
	[P.MAIN, F.MAIN],
	[P.MALIGNMARK, F.MALIGNMARK],
	[P.MARQUEE, F.MARQUEE],
	[P.MATH, F.MATH],
	[P.MENU, F.MENU],
	[P.META, F.META],
	[P.MGLYPH, F.MGLYPH],
	[P.MI, F.MI],
	[P.MO, F.MO],
	[P.MN, F.MN],
	[P.MS, F.MS],
	[P.MTEXT, F.MTEXT],
	[P.NAV, F.NAV],
	[P.NOBR, F.NOBR],
	[P.NOFRAMES, F.NOFRAMES],
	[P.NOEMBED, F.NOEMBED],
	[P.NOSCRIPT, F.NOSCRIPT],
	[P.OBJECT, F.OBJECT],
	[P.OL, F.OL],
	[P.OPTGROUP, F.OPTGROUP],
	[P.OPTION, F.OPTION],
	[P.P, F.P],
	[P.PARAM, F.PARAM],
	[P.PLAINTEXT, F.PLAINTEXT],
	[P.PRE, F.PRE],
	[P.RB, F.RB],
	[P.RP, F.RP],
	[P.RT, F.RT],
	[P.RTC, F.RTC],
	[P.RUBY, F.RUBY],
	[P.S, F.S],
	[P.SCRIPT, F.SCRIPT],
	[P.SEARCH, F.SEARCH],
	[P.SECTION, F.SECTION],
	[P.SELECT, F.SELECT],
	[P.SOURCE, F.SOURCE],
	[P.SMALL, F.SMALL],
	[P.SPAN, F.SPAN],
	[P.STRIKE, F.STRIKE],
	[P.STRONG, F.STRONG],
	[P.STYLE, F.STYLE],
	[P.SUB, F.SUB],
	[P.SUMMARY, F.SUMMARY],
	[P.SUP, F.SUP],
	[P.TABLE, F.TABLE],
	[P.TBODY, F.TBODY],
	[P.TEMPLATE, F.TEMPLATE],
	[P.TEXTAREA, F.TEXTAREA],
	[P.TFOOT, F.TFOOT],
	[P.TD, F.TD],
	[P.TH, F.TH],
	[P.THEAD, F.THEAD],
	[P.TITLE, F.TITLE],
	[P.TR, F.TR],
	[P.TRACK, F.TRACK],
	[P.TT, F.TT],
	[P.U, F.U],
	[P.UL, F.UL],
	[P.SVG, F.SVG],
	[P.VAR, F.VAR],
	[P.WBR, F.WBR],
	[P.XMP, F.XMP]
]);
function Qt(e) {
	return Zt.get(e) ?? F.UNKNOWN;
}
var I = F, $t = {
	[N.HTML]: /* @__PURE__ */ new Set([
		I.ADDRESS,
		I.APPLET,
		I.AREA,
		I.ARTICLE,
		I.ASIDE,
		I.BASE,
		I.BASEFONT,
		I.BGSOUND,
		I.BLOCKQUOTE,
		I.BODY,
		I.BR,
		I.BUTTON,
		I.CAPTION,
		I.CENTER,
		I.COL,
		I.COLGROUP,
		I.DD,
		I.DETAILS,
		I.DIR,
		I.DIV,
		I.DL,
		I.DT,
		I.EMBED,
		I.FIELDSET,
		I.FIGCAPTION,
		I.FIGURE,
		I.FOOTER,
		I.FORM,
		I.FRAME,
		I.FRAMESET,
		I.H1,
		I.H2,
		I.H3,
		I.H4,
		I.H5,
		I.H6,
		I.HEAD,
		I.HEADER,
		I.HGROUP,
		I.HR,
		I.HTML,
		I.IFRAME,
		I.IMG,
		I.INPUT,
		I.LI,
		I.LINK,
		I.LISTING,
		I.MAIN,
		I.MARQUEE,
		I.MENU,
		I.META,
		I.NAV,
		I.NOEMBED,
		I.NOFRAMES,
		I.NOSCRIPT,
		I.OBJECT,
		I.OL,
		I.P,
		I.PARAM,
		I.PLAINTEXT,
		I.PRE,
		I.SCRIPT,
		I.SECTION,
		I.SELECT,
		I.SOURCE,
		I.STYLE,
		I.SUMMARY,
		I.TABLE,
		I.TBODY,
		I.TD,
		I.TEMPLATE,
		I.TEXTAREA,
		I.TFOOT,
		I.TH,
		I.THEAD,
		I.TITLE,
		I.TR,
		I.TRACK,
		I.UL,
		I.WBR,
		I.XMP
	]),
	[N.MATHML]: /* @__PURE__ */ new Set([
		I.MI,
		I.MO,
		I.MN,
		I.MS,
		I.MTEXT,
		I.ANNOTATION_XML
	]),
	[N.SVG]: /* @__PURE__ */ new Set([
		I.TITLE,
		I.FOREIGN_OBJECT,
		I.DESC
	]),
	[N.XLINK]: /* @__PURE__ */ new Set(),
	[N.XML]: /* @__PURE__ */ new Set(),
	[N.XMLNS]: /* @__PURE__ */ new Set()
}, en = /* @__PURE__ */ new Set([
	I.H1,
	I.H2,
	I.H3,
	I.H4,
	I.H5,
	I.H6
]);
P.STYLE, P.SCRIPT, P.XMP, P.IFRAME, P.NOEMBED, P.NOFRAMES, P.PLAINTEXT;
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/parse5/7.3.0/4acbced08702cc2b068edf7ec6fbf0576d666aa2a4d5c21d9bdc1830acf8a8e2/node_modules/parse5/dist/tokenizer/index.js
var L;
(function(e) {
	e[e.DATA = 0] = "DATA", e[e.RCDATA = 1] = "RCDATA", e[e.RAWTEXT = 2] = "RAWTEXT", e[e.SCRIPT_DATA = 3] = "SCRIPT_DATA", e[e.PLAINTEXT = 4] = "PLAINTEXT", e[e.TAG_OPEN = 5] = "TAG_OPEN", e[e.END_TAG_OPEN = 6] = "END_TAG_OPEN", e[e.TAG_NAME = 7] = "TAG_NAME", e[e.RCDATA_LESS_THAN_SIGN = 8] = "RCDATA_LESS_THAN_SIGN", e[e.RCDATA_END_TAG_OPEN = 9] = "RCDATA_END_TAG_OPEN", e[e.RCDATA_END_TAG_NAME = 10] = "RCDATA_END_TAG_NAME", e[e.RAWTEXT_LESS_THAN_SIGN = 11] = "RAWTEXT_LESS_THAN_SIGN", e[e.RAWTEXT_END_TAG_OPEN = 12] = "RAWTEXT_END_TAG_OPEN", e[e.RAWTEXT_END_TAG_NAME = 13] = "RAWTEXT_END_TAG_NAME", e[e.SCRIPT_DATA_LESS_THAN_SIGN = 14] = "SCRIPT_DATA_LESS_THAN_SIGN", e[e.SCRIPT_DATA_END_TAG_OPEN = 15] = "SCRIPT_DATA_END_TAG_OPEN", e[e.SCRIPT_DATA_END_TAG_NAME = 16] = "SCRIPT_DATA_END_TAG_NAME", e[e.SCRIPT_DATA_ESCAPE_START = 17] = "SCRIPT_DATA_ESCAPE_START", e[e.SCRIPT_DATA_ESCAPE_START_DASH = 18] = "SCRIPT_DATA_ESCAPE_START_DASH", e[e.SCRIPT_DATA_ESCAPED = 19] = "SCRIPT_DATA_ESCAPED", e[e.SCRIPT_DATA_ESCAPED_DASH = 20] = "SCRIPT_DATA_ESCAPED_DASH", e[e.SCRIPT_DATA_ESCAPED_DASH_DASH = 21] = "SCRIPT_DATA_ESCAPED_DASH_DASH", e[e.SCRIPT_DATA_ESCAPED_LESS_THAN_SIGN = 22] = "SCRIPT_DATA_ESCAPED_LESS_THAN_SIGN", e[e.SCRIPT_DATA_ESCAPED_END_TAG_OPEN = 23] = "SCRIPT_DATA_ESCAPED_END_TAG_OPEN", e[e.SCRIPT_DATA_ESCAPED_END_TAG_NAME = 24] = "SCRIPT_DATA_ESCAPED_END_TAG_NAME", e[e.SCRIPT_DATA_DOUBLE_ESCAPE_START = 25] = "SCRIPT_DATA_DOUBLE_ESCAPE_START", e[e.SCRIPT_DATA_DOUBLE_ESCAPED = 26] = "SCRIPT_DATA_DOUBLE_ESCAPED", e[e.SCRIPT_DATA_DOUBLE_ESCAPED_DASH = 27] = "SCRIPT_DATA_DOUBLE_ESCAPED_DASH", e[e.SCRIPT_DATA_DOUBLE_ESCAPED_DASH_DASH = 28] = "SCRIPT_DATA_DOUBLE_ESCAPED_DASH_DASH", e[e.SCRIPT_DATA_DOUBLE_ESCAPED_LESS_THAN_SIGN = 29] = "SCRIPT_DATA_DOUBLE_ESCAPED_LESS_THAN_SIGN", e[e.SCRIPT_DATA_DOUBLE_ESCAPE_END = 30] = "SCRIPT_DATA_DOUBLE_ESCAPE_END", e[e.BEFORE_ATTRIBUTE_NAME = 31] = "BEFORE_ATTRIBUTE_NAME", e[e.ATTRIBUTE_NAME = 32] = "ATTRIBUTE_NAME", e[e.AFTER_ATTRIBUTE_NAME = 33] = "AFTER_ATTRIBUTE_NAME", e[e.BEFORE_ATTRIBUTE_VALUE = 34] = "BEFORE_ATTRIBUTE_VALUE", e[e.ATTRIBUTE_VALUE_DOUBLE_QUOTED = 35] = "ATTRIBUTE_VALUE_DOUBLE_QUOTED", e[e.ATTRIBUTE_VALUE_SINGLE_QUOTED = 36] = "ATTRIBUTE_VALUE_SINGLE_QUOTED", e[e.ATTRIBUTE_VALUE_UNQUOTED = 37] = "ATTRIBUTE_VALUE_UNQUOTED", e[e.AFTER_ATTRIBUTE_VALUE_QUOTED = 38] = "AFTER_ATTRIBUTE_VALUE_QUOTED", e[e.SELF_CLOSING_START_TAG = 39] = "SELF_CLOSING_START_TAG", e[e.BOGUS_COMMENT = 40] = "BOGUS_COMMENT", e[e.MARKUP_DECLARATION_OPEN = 41] = "MARKUP_DECLARATION_OPEN", e[e.COMMENT_START = 42] = "COMMENT_START", e[e.COMMENT_START_DASH = 43] = "COMMENT_START_DASH", e[e.COMMENT = 44] = "COMMENT", e[e.COMMENT_LESS_THAN_SIGN = 45] = "COMMENT_LESS_THAN_SIGN", e[e.COMMENT_LESS_THAN_SIGN_BANG = 46] = "COMMENT_LESS_THAN_SIGN_BANG", e[e.COMMENT_LESS_THAN_SIGN_BANG_DASH = 47] = "COMMENT_LESS_THAN_SIGN_BANG_DASH", e[e.COMMENT_LESS_THAN_SIGN_BANG_DASH_DASH = 48] = "COMMENT_LESS_THAN_SIGN_BANG_DASH_DASH", e[e.COMMENT_END_DASH = 49] = "COMMENT_END_DASH", e[e.COMMENT_END = 50] = "COMMENT_END", e[e.COMMENT_END_BANG = 51] = "COMMENT_END_BANG", e[e.DOCTYPE = 52] = "DOCTYPE", e[e.BEFORE_DOCTYPE_NAME = 53] = "BEFORE_DOCTYPE_NAME", e[e.DOCTYPE_NAME = 54] = "DOCTYPE_NAME", e[e.AFTER_DOCTYPE_NAME = 55] = "AFTER_DOCTYPE_NAME", e[e.AFTER_DOCTYPE_PUBLIC_KEYWORD = 56] = "AFTER_DOCTYPE_PUBLIC_KEYWORD", e[e.BEFORE_DOCTYPE_PUBLIC_IDENTIFIER = 57] = "BEFORE_DOCTYPE_PUBLIC_IDENTIFIER", e[e.DOCTYPE_PUBLIC_IDENTIFIER_DOUBLE_QUOTED = 58] = "DOCTYPE_PUBLIC_IDENTIFIER_DOUBLE_QUOTED", e[e.DOCTYPE_PUBLIC_IDENTIFIER_SINGLE_QUOTED = 59] = "DOCTYPE_PUBLIC_IDENTIFIER_SINGLE_QUOTED", e[e.AFTER_DOCTYPE_PUBLIC_IDENTIFIER = 60] = "AFTER_DOCTYPE_PUBLIC_IDENTIFIER", e[e.BETWEEN_DOCTYPE_PUBLIC_AND_SYSTEM_IDENTIFIERS = 61] = "BETWEEN_DOCTYPE_PUBLIC_AND_SYSTEM_IDENTIFIERS", e[e.AFTER_DOCTYPE_SYSTEM_KEYWORD = 62] = "AFTER_DOCTYPE_SYSTEM_KEYWORD", e[e.BEFORE_DOCTYPE_SYSTEM_IDENTIFIER = 63] = "BEFORE_DOCTYPE_SYSTEM_IDENTIFIER", e[e.DOCTYPE_SYSTEM_IDENTIFIER_DOUBLE_QUOTED = 64] = "DOCTYPE_SYSTEM_IDENTIFIER_DOUBLE_QUOTED", e[e.DOCTYPE_SYSTEM_IDENTIFIER_SINGLE_QUOTED = 65] = "DOCTYPE_SYSTEM_IDENTIFIER_SINGLE_QUOTED", e[e.AFTER_DOCTYPE_SYSTEM_IDENTIFIER = 66] = "AFTER_DOCTYPE_SYSTEM_IDENTIFIER", e[e.BOGUS_DOCTYPE = 67] = "BOGUS_DOCTYPE", e[e.CDATA_SECTION = 68] = "CDATA_SECTION", e[e.CDATA_SECTION_BRACKET = 69] = "CDATA_SECTION_BRACKET", e[e.CDATA_SECTION_END = 70] = "CDATA_SECTION_END", e[e.CHARACTER_REFERENCE = 71] = "CHARACTER_REFERENCE", e[e.AMBIGUOUS_AMPERSAND = 72] = "AMBIGUOUS_AMPERSAND";
})(L ||= {});
var R = {
	DATA: L.DATA,
	RCDATA: L.RCDATA,
	RAWTEXT: L.RAWTEXT,
	SCRIPT_DATA: L.SCRIPT_DATA,
	PLAINTEXT: L.PLAINTEXT,
	CDATA_SECTION: L.CDATA_SECTION
};
function tn(e) {
	return e >= D.DIGIT_0 && e <= D.DIGIT_9;
}
function nn(e) {
	return e >= D.LATIN_CAPITAL_A && e <= D.LATIN_CAPITAL_Z;
}
function rn(e) {
	return e >= D.LATIN_SMALL_A && e <= D.LATIN_SMALL_Z;
}
function an(e) {
	return rn(e) || nn(e);
}
function on(e) {
	return an(e) || tn(e);
}
function sn(e) {
	return e + 32;
}
function cn(e) {
	return e === D.SPACE || e === D.LINE_FEED || e === D.TABULATION || e === D.FORM_FEED;
}
function ln(e) {
	return cn(e) || e === D.SOLIDUS || e === D.GREATER_THAN_SIGN;
}
function un(e) {
	return e === D.NULL ? k.nullCharacterReference : e > 1114111 ? k.characterReferenceOutsideUnicodeRange : kt(e) ? k.surrogateCharacterReference : Nt(e) ? k.noncharacterCharacterReference : Mt(e) || e === D.CARRIAGE_RETURN ? k.controlCharacterReference : null;
}
var dn = class {
	constructor(e, t) {
		this.options = e, this.handler = t, this.paused = !1, this.inLoop = !1, this.inForeignNode = !1, this.lastStartTagName = "", this.active = !1, this.state = L.DATA, this.returnState = L.DATA, this.entityStartPos = 0, this.consumedAfterSnapshot = -1, this.currentCharacterToken = null, this.currentToken = null, this.currentAttr = {
			name: "",
			value: ""
		}, this.preprocessor = new Ft(t), this.currentLocation = this.getCurrentLocation(-1), this.entityDecoder = new qt(Lt, (e, t) => {
			this.preprocessor.pos = this.entityStartPos + t - 1, this._flushCodePointConsumedAsCharacterReference(e);
		}, t.onParseError ? {
			missingSemicolonAfterCharacterReference: () => {
				this._err(k.missingSemicolonAfterCharacterReference, 1);
			},
			absenceOfDigitsInNumericCharacterReference: (e) => {
				this._err(k.absenceOfDigitsInNumericCharacterReference, this.entityStartPos - this.preprocessor.pos + e);
			},
			validateNumericCharacterReference: (e) => {
				let t = un(e);
				t && this._err(t, 1);
			}
		} : void 0);
	}
	_err(e, t = 0) {
		var n, r;
		(r = (n = this.handler).onParseError) == null || r.call(n, this.preprocessor.getError(e, t));
	}
	getCurrentLocation(e) {
		return this.options.sourceCodeLocationInfo ? {
			startLine: this.preprocessor.line,
			startCol: this.preprocessor.col - e,
			startOffset: this.preprocessor.offset - e,
			endLine: -1,
			endCol: -1,
			endOffset: -1
		} : null;
	}
	_runParsingLoop() {
		if (!this.inLoop) {
			for (this.inLoop = !0; this.active && !this.paused;) {
				this.consumedAfterSnapshot = 0;
				let e = this._consume();
				this._ensureHibernation() || this._callState(e);
			}
			this.inLoop = !1;
		}
	}
	pause() {
		this.paused = !0;
	}
	resume(e) {
		if (!this.paused) throw Error("Parser was already resumed");
		this.paused = !1, !this.inLoop && (this._runParsingLoop(), this.paused || e?.());
	}
	write(e, t, n) {
		this.active = !0, this.preprocessor.write(e, t), this._runParsingLoop(), this.paused || n?.();
	}
	insertHtmlAtCurrentPos(e) {
		this.active = !0, this.preprocessor.insertHtmlAtCurrentPos(e), this._runParsingLoop();
	}
	_ensureHibernation() {
		return this.preprocessor.endOfChunkHit ? (this.preprocessor.retreat(this.consumedAfterSnapshot), this.consumedAfterSnapshot = 0, this.active = !1, !0) : !1;
	}
	_consume() {
		return this.consumedAfterSnapshot++, this.preprocessor.advance();
	}
	_advanceBy(e) {
		this.consumedAfterSnapshot += e;
		for (let t = 0; t < e; t++) this.preprocessor.advance();
	}
	_consumeSequenceIfMatch(e, t) {
		return this.preprocessor.startsWith(e, t) ? (this._advanceBy(e.length - 1), !0) : !1;
	}
	_createStartTagToken() {
		this.currentToken = {
			type: A.START_TAG,
			tagName: "",
			tagID: F.UNKNOWN,
			selfClosing: !1,
			ackSelfClosing: !1,
			attrs: [],
			location: this.getCurrentLocation(1)
		};
	}
	_createEndTagToken() {
		this.currentToken = {
			type: A.END_TAG,
			tagName: "",
			tagID: F.UNKNOWN,
			selfClosing: !1,
			ackSelfClosing: !1,
			attrs: [],
			location: this.getCurrentLocation(2)
		};
	}
	_createCommentToken(e) {
		this.currentToken = {
			type: A.COMMENT,
			data: "",
			location: this.getCurrentLocation(e)
		};
	}
	_createDoctypeToken(e) {
		this.currentToken = {
			type: A.DOCTYPE,
			name: e,
			forceQuirks: !1,
			publicId: null,
			systemId: null,
			location: this.currentLocation
		};
	}
	_createCharacterToken(e, t) {
		this.currentCharacterToken = {
			type: e,
			chars: t,
			location: this.currentLocation
		};
	}
	_createAttr(e) {
		this.currentAttr = {
			name: e,
			value: ""
		}, this.currentLocation = this.getCurrentLocation(0);
	}
	_leaveAttrName() {
		var e;
		let t = this.currentToken;
		if (It(t, this.currentAttr.name) === null) {
			if (t.attrs.push(this.currentAttr), t.location && this.currentLocation) {
				let n = (e = t.location).attrs ?? (e.attrs = Object.create(null));
				n[this.currentAttr.name] = this.currentLocation, this._leaveAttrValue();
			}
		} else this._err(k.duplicateAttribute);
	}
	_leaveAttrValue() {
		this.currentLocation && (this.currentLocation.endLine = this.preprocessor.line, this.currentLocation.endCol = this.preprocessor.col, this.currentLocation.endOffset = this.preprocessor.offset);
	}
	prepareToken(e) {
		this._emitCurrentCharacterToken(e.location), this.currentToken = null, e.location && (e.location.endLine = this.preprocessor.line, e.location.endCol = this.preprocessor.col + 1, e.location.endOffset = this.preprocessor.offset + 1), this.currentLocation = this.getCurrentLocation(-1);
	}
	emitCurrentTagToken() {
		let e = this.currentToken;
		this.prepareToken(e), e.tagID = Qt(e.tagName), e.type === A.START_TAG ? (this.lastStartTagName = e.tagName, this.handler.onStartTag(e)) : (e.attrs.length > 0 && this._err(k.endTagWithAttributes), e.selfClosing && this._err(k.endTagWithTrailingSolidus), this.handler.onEndTag(e)), this.preprocessor.dropParsedChunk();
	}
	emitCurrentComment(e) {
		this.prepareToken(e), this.handler.onComment(e), this.preprocessor.dropParsedChunk();
	}
	emitCurrentDoctype(e) {
		this.prepareToken(e), this.handler.onDoctype(e), this.preprocessor.dropParsedChunk();
	}
	_emitCurrentCharacterToken(e) {
		if (this.currentCharacterToken) {
			switch (e && this.currentCharacterToken.location && (this.currentCharacterToken.location.endLine = e.startLine, this.currentCharacterToken.location.endCol = e.startCol, this.currentCharacterToken.location.endOffset = e.startOffset), this.currentCharacterToken.type) {
				case A.CHARACTER:
					this.handler.onCharacter(this.currentCharacterToken);
					break;
				case A.NULL_CHARACTER:
					this.handler.onNullCharacter(this.currentCharacterToken);
					break;
				case A.WHITESPACE_CHARACTER: this.handler.onWhitespaceCharacter(this.currentCharacterToken);
			}
			this.currentCharacterToken = null;
		}
	}
	_emitEOFToken() {
		let e = this.getCurrentLocation(0);
		e && (e.endLine = e.startLine, e.endCol = e.startCol, e.endOffset = e.startOffset), this._emitCurrentCharacterToken(e), this.handler.onEof({
			type: A.EOF,
			location: e
		}), this.active = !1;
	}
	_appendCharToCurrentCharacterToken(e, t) {
		if (this.currentCharacterToken) {
			if (this.currentCharacterToken.type === e) {
				this.currentCharacterToken.chars += t;
				return;
			}
			this.currentLocation = this.getCurrentLocation(0), this._emitCurrentCharacterToken(this.currentLocation), this.preprocessor.dropParsedChunk();
		}
		this._createCharacterToken(e, t);
	}
	_emitCodePoint(e) {
		let t = cn(e) ? A.WHITESPACE_CHARACTER : e === D.NULL ? A.NULL_CHARACTER : A.CHARACTER;
		this._appendCharToCurrentCharacterToken(t, String.fromCodePoint(e));
	}
	_emitChars(e) {
		this._appendCharToCurrentCharacterToken(A.CHARACTER, e);
	}
	_startCharacterReference() {
		this.returnState = this.state, this.state = L.CHARACTER_REFERENCE, this.entityStartPos = this.preprocessor.pos, this.entityDecoder.startEntity(this._isCharacterReferenceInAttribute() ? Kt.Attribute : Kt.Legacy);
	}
	_isCharacterReferenceInAttribute() {
		return this.returnState === L.ATTRIBUTE_VALUE_DOUBLE_QUOTED || this.returnState === L.ATTRIBUTE_VALUE_SINGLE_QUOTED || this.returnState === L.ATTRIBUTE_VALUE_UNQUOTED;
	}
	_flushCodePointConsumedAsCharacterReference(e) {
		this._isCharacterReferenceInAttribute() ? this.currentAttr.value += String.fromCodePoint(e) : this._emitCodePoint(e);
	}
	_callState(e) {
		switch (this.state) {
			case L.DATA:
				this._stateData(e);
				break;
			case L.RCDATA:
				this._stateRcdata(e);
				break;
			case L.RAWTEXT:
				this._stateRawtext(e);
				break;
			case L.SCRIPT_DATA:
				this._stateScriptData(e);
				break;
			case L.PLAINTEXT:
				this._statePlaintext(e);
				break;
			case L.TAG_OPEN:
				this._stateTagOpen(e);
				break;
			case L.END_TAG_OPEN:
				this._stateEndTagOpen(e);
				break;
			case L.TAG_NAME:
				this._stateTagName(e);
				break;
			case L.RCDATA_LESS_THAN_SIGN:
				this._stateRcdataLessThanSign(e);
				break;
			case L.RCDATA_END_TAG_OPEN:
				this._stateRcdataEndTagOpen(e);
				break;
			case L.RCDATA_END_TAG_NAME:
				this._stateRcdataEndTagName(e);
				break;
			case L.RAWTEXT_LESS_THAN_SIGN:
				this._stateRawtextLessThanSign(e);
				break;
			case L.RAWTEXT_END_TAG_OPEN:
				this._stateRawtextEndTagOpen(e);
				break;
			case L.RAWTEXT_END_TAG_NAME:
				this._stateRawtextEndTagName(e);
				break;
			case L.SCRIPT_DATA_LESS_THAN_SIGN:
				this._stateScriptDataLessThanSign(e);
				break;
			case L.SCRIPT_DATA_END_TAG_OPEN:
				this._stateScriptDataEndTagOpen(e);
				break;
			case L.SCRIPT_DATA_END_TAG_NAME:
				this._stateScriptDataEndTagName(e);
				break;
			case L.SCRIPT_DATA_ESCAPE_START:
				this._stateScriptDataEscapeStart(e);
				break;
			case L.SCRIPT_DATA_ESCAPE_START_DASH:
				this._stateScriptDataEscapeStartDash(e);
				break;
			case L.SCRIPT_DATA_ESCAPED:
				this._stateScriptDataEscaped(e);
				break;
			case L.SCRIPT_DATA_ESCAPED_DASH:
				this._stateScriptDataEscapedDash(e);
				break;
			case L.SCRIPT_DATA_ESCAPED_DASH_DASH:
				this._stateScriptDataEscapedDashDash(e);
				break;
			case L.SCRIPT_DATA_ESCAPED_LESS_THAN_SIGN:
				this._stateScriptDataEscapedLessThanSign(e);
				break;
			case L.SCRIPT_DATA_ESCAPED_END_TAG_OPEN:
				this._stateScriptDataEscapedEndTagOpen(e);
				break;
			case L.SCRIPT_DATA_ESCAPED_END_TAG_NAME:
				this._stateScriptDataEscapedEndTagName(e);
				break;
			case L.SCRIPT_DATA_DOUBLE_ESCAPE_START:
				this._stateScriptDataDoubleEscapeStart(e);
				break;
			case L.SCRIPT_DATA_DOUBLE_ESCAPED:
				this._stateScriptDataDoubleEscaped(e);
				break;
			case L.SCRIPT_DATA_DOUBLE_ESCAPED_DASH:
				this._stateScriptDataDoubleEscapedDash(e);
				break;
			case L.SCRIPT_DATA_DOUBLE_ESCAPED_DASH_DASH:
				this._stateScriptDataDoubleEscapedDashDash(e);
				break;
			case L.SCRIPT_DATA_DOUBLE_ESCAPED_LESS_THAN_SIGN:
				this._stateScriptDataDoubleEscapedLessThanSign(e);
				break;
			case L.SCRIPT_DATA_DOUBLE_ESCAPE_END:
				this._stateScriptDataDoubleEscapeEnd(e);
				break;
			case L.BEFORE_ATTRIBUTE_NAME:
				this._stateBeforeAttributeName(e);
				break;
			case L.ATTRIBUTE_NAME:
				this._stateAttributeName(e);
				break;
			case L.AFTER_ATTRIBUTE_NAME:
				this._stateAfterAttributeName(e);
				break;
			case L.BEFORE_ATTRIBUTE_VALUE:
				this._stateBeforeAttributeValue(e);
				break;
			case L.ATTRIBUTE_VALUE_DOUBLE_QUOTED:
				this._stateAttributeValueDoubleQuoted(e);
				break;
			case L.ATTRIBUTE_VALUE_SINGLE_QUOTED:
				this._stateAttributeValueSingleQuoted(e);
				break;
			case L.ATTRIBUTE_VALUE_UNQUOTED:
				this._stateAttributeValueUnquoted(e);
				break;
			case L.AFTER_ATTRIBUTE_VALUE_QUOTED:
				this._stateAfterAttributeValueQuoted(e);
				break;
			case L.SELF_CLOSING_START_TAG:
				this._stateSelfClosingStartTag(e);
				break;
			case L.BOGUS_COMMENT:
				this._stateBogusComment(e);
				break;
			case L.MARKUP_DECLARATION_OPEN:
				this._stateMarkupDeclarationOpen(e);
				break;
			case L.COMMENT_START:
				this._stateCommentStart(e);
				break;
			case L.COMMENT_START_DASH:
				this._stateCommentStartDash(e);
				break;
			case L.COMMENT:
				this._stateComment(e);
				break;
			case L.COMMENT_LESS_THAN_SIGN:
				this._stateCommentLessThanSign(e);
				break;
			case L.COMMENT_LESS_THAN_SIGN_BANG:
				this._stateCommentLessThanSignBang(e);
				break;
			case L.COMMENT_LESS_THAN_SIGN_BANG_DASH:
				this._stateCommentLessThanSignBangDash(e);
				break;
			case L.COMMENT_LESS_THAN_SIGN_BANG_DASH_DASH:
				this._stateCommentLessThanSignBangDashDash(e);
				break;
			case L.COMMENT_END_DASH:
				this._stateCommentEndDash(e);
				break;
			case L.COMMENT_END:
				this._stateCommentEnd(e);
				break;
			case L.COMMENT_END_BANG:
				this._stateCommentEndBang(e);
				break;
			case L.DOCTYPE:
				this._stateDoctype(e);
				break;
			case L.BEFORE_DOCTYPE_NAME:
				this._stateBeforeDoctypeName(e);
				break;
			case L.DOCTYPE_NAME:
				this._stateDoctypeName(e);
				break;
			case L.AFTER_DOCTYPE_NAME:
				this._stateAfterDoctypeName(e);
				break;
			case L.AFTER_DOCTYPE_PUBLIC_KEYWORD:
				this._stateAfterDoctypePublicKeyword(e);
				break;
			case L.BEFORE_DOCTYPE_PUBLIC_IDENTIFIER:
				this._stateBeforeDoctypePublicIdentifier(e);
				break;
			case L.DOCTYPE_PUBLIC_IDENTIFIER_DOUBLE_QUOTED:
				this._stateDoctypePublicIdentifierDoubleQuoted(e);
				break;
			case L.DOCTYPE_PUBLIC_IDENTIFIER_SINGLE_QUOTED:
				this._stateDoctypePublicIdentifierSingleQuoted(e);
				break;
			case L.AFTER_DOCTYPE_PUBLIC_IDENTIFIER:
				this._stateAfterDoctypePublicIdentifier(e);
				break;
			case L.BETWEEN_DOCTYPE_PUBLIC_AND_SYSTEM_IDENTIFIERS:
				this._stateBetweenDoctypePublicAndSystemIdentifiers(e);
				break;
			case L.AFTER_DOCTYPE_SYSTEM_KEYWORD:
				this._stateAfterDoctypeSystemKeyword(e);
				break;
			case L.BEFORE_DOCTYPE_SYSTEM_IDENTIFIER:
				this._stateBeforeDoctypeSystemIdentifier(e);
				break;
			case L.DOCTYPE_SYSTEM_IDENTIFIER_DOUBLE_QUOTED:
				this._stateDoctypeSystemIdentifierDoubleQuoted(e);
				break;
			case L.DOCTYPE_SYSTEM_IDENTIFIER_SINGLE_QUOTED:
				this._stateDoctypeSystemIdentifierSingleQuoted(e);
				break;
			case L.AFTER_DOCTYPE_SYSTEM_IDENTIFIER:
				this._stateAfterDoctypeSystemIdentifier(e);
				break;
			case L.BOGUS_DOCTYPE:
				this._stateBogusDoctype(e);
				break;
			case L.CDATA_SECTION:
				this._stateCdataSection(e);
				break;
			case L.CDATA_SECTION_BRACKET:
				this._stateCdataSectionBracket(e);
				break;
			case L.CDATA_SECTION_END:
				this._stateCdataSectionEnd(e);
				break;
			case L.CHARACTER_REFERENCE:
				this._stateCharacterReference();
				break;
			case L.AMBIGUOUS_AMPERSAND:
				this._stateAmbiguousAmpersand(e);
				break;
			default: throw Error("Unknown state");
		}
	}
	_stateData(e) {
		switch (e) {
			case D.LESS_THAN_SIGN:
				this.state = L.TAG_OPEN;
				break;
			case D.AMPERSAND:
				this._startCharacterReference();
				break;
			case D.NULL:
				this._err(k.unexpectedNullCharacter), this._emitCodePoint(e);
				break;
			case D.EOF:
				this._emitEOFToken();
				break;
			default: this._emitCodePoint(e);
		}
	}
	_stateRcdata(e) {
		switch (e) {
			case D.AMPERSAND:
				this._startCharacterReference();
				break;
			case D.LESS_THAN_SIGN:
				this.state = L.RCDATA_LESS_THAN_SIGN;
				break;
			case D.NULL:
				this._err(k.unexpectedNullCharacter), this._emitChars("�");
				break;
			case D.EOF:
				this._emitEOFToken();
				break;
			default: this._emitCodePoint(e);
		}
	}
	_stateRawtext(e) {
		switch (e) {
			case D.LESS_THAN_SIGN:
				this.state = L.RAWTEXT_LESS_THAN_SIGN;
				break;
			case D.NULL:
				this._err(k.unexpectedNullCharacter), this._emitChars("�");
				break;
			case D.EOF:
				this._emitEOFToken();
				break;
			default: this._emitCodePoint(e);
		}
	}
	_stateScriptData(e) {
		switch (e) {
			case D.LESS_THAN_SIGN:
				this.state = L.SCRIPT_DATA_LESS_THAN_SIGN;
				break;
			case D.NULL:
				this._err(k.unexpectedNullCharacter), this._emitChars("�");
				break;
			case D.EOF:
				this._emitEOFToken();
				break;
			default: this._emitCodePoint(e);
		}
	}
	_statePlaintext(e) {
		switch (e) {
			case D.NULL:
				this._err(k.unexpectedNullCharacter), this._emitChars("�");
				break;
			case D.EOF:
				this._emitEOFToken();
				break;
			default: this._emitCodePoint(e);
		}
	}
	_stateTagOpen(e) {
		if (an(e)) this._createStartTagToken(), this.state = L.TAG_NAME, this._stateTagName(e);
		else switch (e) {
			case D.EXCLAMATION_MARK:
				this.state = L.MARKUP_DECLARATION_OPEN;
				break;
			case D.SOLIDUS:
				this.state = L.END_TAG_OPEN;
				break;
			case D.QUESTION_MARK:
				this._err(k.unexpectedQuestionMarkInsteadOfTagName), this._createCommentToken(1), this.state = L.BOGUS_COMMENT, this._stateBogusComment(e);
				break;
			case D.EOF:
				this._err(k.eofBeforeTagName), this._emitChars("<"), this._emitEOFToken();
				break;
			default: this._err(k.invalidFirstCharacterOfTagName), this._emitChars("<"), this.state = L.DATA, this._stateData(e);
		}
	}
	_stateEndTagOpen(e) {
		if (an(e)) this._createEndTagToken(), this.state = L.TAG_NAME, this._stateTagName(e);
		else switch (e) {
			case D.GREATER_THAN_SIGN:
				this._err(k.missingEndTagName), this.state = L.DATA;
				break;
			case D.EOF:
				this._err(k.eofBeforeTagName), this._emitChars("</"), this._emitEOFToken();
				break;
			default: this._err(k.invalidFirstCharacterOfTagName), this._createCommentToken(2), this.state = L.BOGUS_COMMENT, this._stateBogusComment(e);
		}
	}
	_stateTagName(e) {
		let t = this.currentToken;
		switch (e) {
			case D.SPACE:
			case D.LINE_FEED:
			case D.TABULATION:
			case D.FORM_FEED:
				this.state = L.BEFORE_ATTRIBUTE_NAME;
				break;
			case D.SOLIDUS:
				this.state = L.SELF_CLOSING_START_TAG;
				break;
			case D.GREATER_THAN_SIGN:
				this.state = L.DATA, this.emitCurrentTagToken();
				break;
			case D.NULL:
				this._err(k.unexpectedNullCharacter), t.tagName += "�";
				break;
			case D.EOF:
				this._err(k.eofInTag), this._emitEOFToken();
				break;
			default: t.tagName += String.fromCodePoint(nn(e) ? sn(e) : e);
		}
	}
	_stateRcdataLessThanSign(e) {
		e === D.SOLIDUS ? this.state = L.RCDATA_END_TAG_OPEN : (this._emitChars("<"), this.state = L.RCDATA, this._stateRcdata(e));
	}
	_stateRcdataEndTagOpen(e) {
		an(e) ? (this.state = L.RCDATA_END_TAG_NAME, this._stateRcdataEndTagName(e)) : (this._emitChars("</"), this.state = L.RCDATA, this._stateRcdata(e));
	}
	handleSpecialEndTag(e) {
		if (!this.preprocessor.startsWith(this.lastStartTagName, !1)) return !this._ensureHibernation();
		this._createEndTagToken();
		let t = this.currentToken;
		switch (t.tagName = this.lastStartTagName, this.preprocessor.peek(this.lastStartTagName.length)) {
			case D.SPACE:
			case D.LINE_FEED:
			case D.TABULATION:
			case D.FORM_FEED: return this._advanceBy(this.lastStartTagName.length), this.state = L.BEFORE_ATTRIBUTE_NAME, !1;
			case D.SOLIDUS: return this._advanceBy(this.lastStartTagName.length), this.state = L.SELF_CLOSING_START_TAG, !1;
			case D.GREATER_THAN_SIGN: return this._advanceBy(this.lastStartTagName.length), this.emitCurrentTagToken(), this.state = L.DATA, !1;
			default: return !this._ensureHibernation();
		}
	}
	_stateRcdataEndTagName(e) {
		this.handleSpecialEndTag(e) && (this._emitChars("</"), this.state = L.RCDATA, this._stateRcdata(e));
	}
	_stateRawtextLessThanSign(e) {
		e === D.SOLIDUS ? this.state = L.RAWTEXT_END_TAG_OPEN : (this._emitChars("<"), this.state = L.RAWTEXT, this._stateRawtext(e));
	}
	_stateRawtextEndTagOpen(e) {
		an(e) ? (this.state = L.RAWTEXT_END_TAG_NAME, this._stateRawtextEndTagName(e)) : (this._emitChars("</"), this.state = L.RAWTEXT, this._stateRawtext(e));
	}
	_stateRawtextEndTagName(e) {
		this.handleSpecialEndTag(e) && (this._emitChars("</"), this.state = L.RAWTEXT, this._stateRawtext(e));
	}
	_stateScriptDataLessThanSign(e) {
		switch (e) {
			case D.SOLIDUS:
				this.state = L.SCRIPT_DATA_END_TAG_OPEN;
				break;
			case D.EXCLAMATION_MARK:
				this.state = L.SCRIPT_DATA_ESCAPE_START, this._emitChars("<!");
				break;
			default: this._emitChars("<"), this.state = L.SCRIPT_DATA, this._stateScriptData(e);
		}
	}
	_stateScriptDataEndTagOpen(e) {
		an(e) ? (this.state = L.SCRIPT_DATA_END_TAG_NAME, this._stateScriptDataEndTagName(e)) : (this._emitChars("</"), this.state = L.SCRIPT_DATA, this._stateScriptData(e));
	}
	_stateScriptDataEndTagName(e) {
		this.handleSpecialEndTag(e) && (this._emitChars("</"), this.state = L.SCRIPT_DATA, this._stateScriptData(e));
	}
	_stateScriptDataEscapeStart(e) {
		e === D.HYPHEN_MINUS ? (this.state = L.SCRIPT_DATA_ESCAPE_START_DASH, this._emitChars("-")) : (this.state = L.SCRIPT_DATA, this._stateScriptData(e));
	}
	_stateScriptDataEscapeStartDash(e) {
		e === D.HYPHEN_MINUS ? (this.state = L.SCRIPT_DATA_ESCAPED_DASH_DASH, this._emitChars("-")) : (this.state = L.SCRIPT_DATA, this._stateScriptData(e));
	}
	_stateScriptDataEscaped(e) {
		switch (e) {
			case D.HYPHEN_MINUS:
				this.state = L.SCRIPT_DATA_ESCAPED_DASH, this._emitChars("-");
				break;
			case D.LESS_THAN_SIGN:
				this.state = L.SCRIPT_DATA_ESCAPED_LESS_THAN_SIGN;
				break;
			case D.NULL:
				this._err(k.unexpectedNullCharacter), this._emitChars("�");
				break;
			case D.EOF:
				this._err(k.eofInScriptHtmlCommentLikeText), this._emitEOFToken();
				break;
			default: this._emitCodePoint(e);
		}
	}
	_stateScriptDataEscapedDash(e) {
		switch (e) {
			case D.HYPHEN_MINUS:
				this.state = L.SCRIPT_DATA_ESCAPED_DASH_DASH, this._emitChars("-");
				break;
			case D.LESS_THAN_SIGN:
				this.state = L.SCRIPT_DATA_ESCAPED_LESS_THAN_SIGN;
				break;
			case D.NULL:
				this._err(k.unexpectedNullCharacter), this.state = L.SCRIPT_DATA_ESCAPED, this._emitChars("�");
				break;
			case D.EOF:
				this._err(k.eofInScriptHtmlCommentLikeText), this._emitEOFToken();
				break;
			default: this.state = L.SCRIPT_DATA_ESCAPED, this._emitCodePoint(e);
		}
	}
	_stateScriptDataEscapedDashDash(e) {
		switch (e) {
			case D.HYPHEN_MINUS:
				this._emitChars("-");
				break;
			case D.LESS_THAN_SIGN:
				this.state = L.SCRIPT_DATA_ESCAPED_LESS_THAN_SIGN;
				break;
			case D.GREATER_THAN_SIGN:
				this.state = L.SCRIPT_DATA, this._emitChars(">");
				break;
			case D.NULL:
				this._err(k.unexpectedNullCharacter), this.state = L.SCRIPT_DATA_ESCAPED, this._emitChars("�");
				break;
			case D.EOF:
				this._err(k.eofInScriptHtmlCommentLikeText), this._emitEOFToken();
				break;
			default: this.state = L.SCRIPT_DATA_ESCAPED, this._emitCodePoint(e);
		}
	}
	_stateScriptDataEscapedLessThanSign(e) {
		e === D.SOLIDUS ? this.state = L.SCRIPT_DATA_ESCAPED_END_TAG_OPEN : an(e) ? (this._emitChars("<"), this.state = L.SCRIPT_DATA_DOUBLE_ESCAPE_START, this._stateScriptDataDoubleEscapeStart(e)) : (this._emitChars("<"), this.state = L.SCRIPT_DATA_ESCAPED, this._stateScriptDataEscaped(e));
	}
	_stateScriptDataEscapedEndTagOpen(e) {
		an(e) ? (this.state = L.SCRIPT_DATA_ESCAPED_END_TAG_NAME, this._stateScriptDataEscapedEndTagName(e)) : (this._emitChars("</"), this.state = L.SCRIPT_DATA_ESCAPED, this._stateScriptDataEscaped(e));
	}
	_stateScriptDataEscapedEndTagName(e) {
		this.handleSpecialEndTag(e) && (this._emitChars("</"), this.state = L.SCRIPT_DATA_ESCAPED, this._stateScriptDataEscaped(e));
	}
	_stateScriptDataDoubleEscapeStart(e) {
		if (this.preprocessor.startsWith(O.SCRIPT, !1) && ln(this.preprocessor.peek(O.SCRIPT.length))) {
			this._emitCodePoint(e);
			for (let e = 0; e < O.SCRIPT.length; e++) this._emitCodePoint(this._consume());
			this.state = L.SCRIPT_DATA_DOUBLE_ESCAPED;
		} else this._ensureHibernation() || (this.state = L.SCRIPT_DATA_ESCAPED, this._stateScriptDataEscaped(e));
	}
	_stateScriptDataDoubleEscaped(e) {
		switch (e) {
			case D.HYPHEN_MINUS:
				this.state = L.SCRIPT_DATA_DOUBLE_ESCAPED_DASH, this._emitChars("-");
				break;
			case D.LESS_THAN_SIGN:
				this.state = L.SCRIPT_DATA_DOUBLE_ESCAPED_LESS_THAN_SIGN, this._emitChars("<");
				break;
			case D.NULL:
				this._err(k.unexpectedNullCharacter), this._emitChars("�");
				break;
			case D.EOF:
				this._err(k.eofInScriptHtmlCommentLikeText), this._emitEOFToken();
				break;
			default: this._emitCodePoint(e);
		}
	}
	_stateScriptDataDoubleEscapedDash(e) {
		switch (e) {
			case D.HYPHEN_MINUS:
				this.state = L.SCRIPT_DATA_DOUBLE_ESCAPED_DASH_DASH, this._emitChars("-");
				break;
			case D.LESS_THAN_SIGN:
				this.state = L.SCRIPT_DATA_DOUBLE_ESCAPED_LESS_THAN_SIGN, this._emitChars("<");
				break;
			case D.NULL:
				this._err(k.unexpectedNullCharacter), this.state = L.SCRIPT_DATA_DOUBLE_ESCAPED, this._emitChars("�");
				break;
			case D.EOF:
				this._err(k.eofInScriptHtmlCommentLikeText), this._emitEOFToken();
				break;
			default: this.state = L.SCRIPT_DATA_DOUBLE_ESCAPED, this._emitCodePoint(e);
		}
	}
	_stateScriptDataDoubleEscapedDashDash(e) {
		switch (e) {
			case D.HYPHEN_MINUS:
				this._emitChars("-");
				break;
			case D.LESS_THAN_SIGN:
				this.state = L.SCRIPT_DATA_DOUBLE_ESCAPED_LESS_THAN_SIGN, this._emitChars("<");
				break;
			case D.GREATER_THAN_SIGN:
				this.state = L.SCRIPT_DATA, this._emitChars(">");
				break;
			case D.NULL:
				this._err(k.unexpectedNullCharacter), this.state = L.SCRIPT_DATA_DOUBLE_ESCAPED, this._emitChars("�");
				break;
			case D.EOF:
				this._err(k.eofInScriptHtmlCommentLikeText), this._emitEOFToken();
				break;
			default: this.state = L.SCRIPT_DATA_DOUBLE_ESCAPED, this._emitCodePoint(e);
		}
	}
	_stateScriptDataDoubleEscapedLessThanSign(e) {
		e === D.SOLIDUS ? (this.state = L.SCRIPT_DATA_DOUBLE_ESCAPE_END, this._emitChars("/")) : (this.state = L.SCRIPT_DATA_DOUBLE_ESCAPED, this._stateScriptDataDoubleEscaped(e));
	}
	_stateScriptDataDoubleEscapeEnd(e) {
		if (this.preprocessor.startsWith(O.SCRIPT, !1) && ln(this.preprocessor.peek(O.SCRIPT.length))) {
			this._emitCodePoint(e);
			for (let e = 0; e < O.SCRIPT.length; e++) this._emitCodePoint(this._consume());
			this.state = L.SCRIPT_DATA_ESCAPED;
		} else this._ensureHibernation() || (this.state = L.SCRIPT_DATA_DOUBLE_ESCAPED, this._stateScriptDataDoubleEscaped(e));
	}
	_stateBeforeAttributeName(e) {
		switch (e) {
			case D.SPACE:
			case D.LINE_FEED:
			case D.TABULATION:
			case D.FORM_FEED: break;
			case D.SOLIDUS:
			case D.GREATER_THAN_SIGN:
			case D.EOF:
				this.state = L.AFTER_ATTRIBUTE_NAME, this._stateAfterAttributeName(e);
				break;
			case D.EQUALS_SIGN:
				this._err(k.unexpectedEqualsSignBeforeAttributeName), this._createAttr("="), this.state = L.ATTRIBUTE_NAME;
				break;
			default: this._createAttr(""), this.state = L.ATTRIBUTE_NAME, this._stateAttributeName(e);
		}
	}
	_stateAttributeName(e) {
		switch (e) {
			case D.SPACE:
			case D.LINE_FEED:
			case D.TABULATION:
			case D.FORM_FEED:
			case D.SOLIDUS:
			case D.GREATER_THAN_SIGN:
			case D.EOF:
				this._leaveAttrName(), this.state = L.AFTER_ATTRIBUTE_NAME, this._stateAfterAttributeName(e);
				break;
			case D.EQUALS_SIGN:
				this._leaveAttrName(), this.state = L.BEFORE_ATTRIBUTE_VALUE;
				break;
			case D.QUOTATION_MARK:
			case D.APOSTROPHE:
			case D.LESS_THAN_SIGN:
				this._err(k.unexpectedCharacterInAttributeName), this.currentAttr.name += String.fromCodePoint(e);
				break;
			case D.NULL:
				this._err(k.unexpectedNullCharacter), this.currentAttr.name += "�";
				break;
			default: this.currentAttr.name += String.fromCodePoint(nn(e) ? sn(e) : e);
		}
	}
	_stateAfterAttributeName(e) {
		switch (e) {
			case D.SPACE:
			case D.LINE_FEED:
			case D.TABULATION:
			case D.FORM_FEED: break;
			case D.SOLIDUS:
				this.state = L.SELF_CLOSING_START_TAG;
				break;
			case D.EQUALS_SIGN:
				this.state = L.BEFORE_ATTRIBUTE_VALUE;
				break;
			case D.GREATER_THAN_SIGN:
				this.state = L.DATA, this.emitCurrentTagToken();
				break;
			case D.EOF:
				this._err(k.eofInTag), this._emitEOFToken();
				break;
			default: this._createAttr(""), this.state = L.ATTRIBUTE_NAME, this._stateAttributeName(e);
		}
	}
	_stateBeforeAttributeValue(e) {
		switch (e) {
			case D.SPACE:
			case D.LINE_FEED:
			case D.TABULATION:
			case D.FORM_FEED: break;
			case D.QUOTATION_MARK:
				this.state = L.ATTRIBUTE_VALUE_DOUBLE_QUOTED;
				break;
			case D.APOSTROPHE:
				this.state = L.ATTRIBUTE_VALUE_SINGLE_QUOTED;
				break;
			case D.GREATER_THAN_SIGN:
				this._err(k.missingAttributeValue), this.state = L.DATA, this.emitCurrentTagToken();
				break;
			default: this.state = L.ATTRIBUTE_VALUE_UNQUOTED, this._stateAttributeValueUnquoted(e);
		}
	}
	_stateAttributeValueDoubleQuoted(e) {
		switch (e) {
			case D.QUOTATION_MARK:
				this.state = L.AFTER_ATTRIBUTE_VALUE_QUOTED;
				break;
			case D.AMPERSAND:
				this._startCharacterReference();
				break;
			case D.NULL:
				this._err(k.unexpectedNullCharacter), this.currentAttr.value += "�";
				break;
			case D.EOF:
				this._err(k.eofInTag), this._emitEOFToken();
				break;
			default: this.currentAttr.value += String.fromCodePoint(e);
		}
	}
	_stateAttributeValueSingleQuoted(e) {
		switch (e) {
			case D.APOSTROPHE:
				this.state = L.AFTER_ATTRIBUTE_VALUE_QUOTED;
				break;
			case D.AMPERSAND:
				this._startCharacterReference();
				break;
			case D.NULL:
				this._err(k.unexpectedNullCharacter), this.currentAttr.value += "�";
				break;
			case D.EOF:
				this._err(k.eofInTag), this._emitEOFToken();
				break;
			default: this.currentAttr.value += String.fromCodePoint(e);
		}
	}
	_stateAttributeValueUnquoted(e) {
		switch (e) {
			case D.SPACE:
			case D.LINE_FEED:
			case D.TABULATION:
			case D.FORM_FEED:
				this._leaveAttrValue(), this.state = L.BEFORE_ATTRIBUTE_NAME;
				break;
			case D.AMPERSAND:
				this._startCharacterReference();
				break;
			case D.GREATER_THAN_SIGN:
				this._leaveAttrValue(), this.state = L.DATA, this.emitCurrentTagToken();
				break;
			case D.NULL:
				this._err(k.unexpectedNullCharacter), this.currentAttr.value += "�";
				break;
			case D.QUOTATION_MARK:
			case D.APOSTROPHE:
			case D.LESS_THAN_SIGN:
			case D.EQUALS_SIGN:
			case D.GRAVE_ACCENT:
				this._err(k.unexpectedCharacterInUnquotedAttributeValue), this.currentAttr.value += String.fromCodePoint(e);
				break;
			case D.EOF:
				this._err(k.eofInTag), this._emitEOFToken();
				break;
			default: this.currentAttr.value += String.fromCodePoint(e);
		}
	}
	_stateAfterAttributeValueQuoted(e) {
		switch (e) {
			case D.SPACE:
			case D.LINE_FEED:
			case D.TABULATION:
			case D.FORM_FEED:
				this._leaveAttrValue(), this.state = L.BEFORE_ATTRIBUTE_NAME;
				break;
			case D.SOLIDUS:
				this._leaveAttrValue(), this.state = L.SELF_CLOSING_START_TAG;
				break;
			case D.GREATER_THAN_SIGN:
				this._leaveAttrValue(), this.state = L.DATA, this.emitCurrentTagToken();
				break;
			case D.EOF:
				this._err(k.eofInTag), this._emitEOFToken();
				break;
			default: this._err(k.missingWhitespaceBetweenAttributes), this.state = L.BEFORE_ATTRIBUTE_NAME, this._stateBeforeAttributeName(e);
		}
	}
	_stateSelfClosingStartTag(e) {
		switch (e) {
			case D.GREATER_THAN_SIGN: {
				let e = this.currentToken;
				e.selfClosing = !0, this.state = L.DATA, this.emitCurrentTagToken();
				break;
			}
			case D.EOF:
				this._err(k.eofInTag), this._emitEOFToken();
				break;
			default: this._err(k.unexpectedSolidusInTag), this.state = L.BEFORE_ATTRIBUTE_NAME, this._stateBeforeAttributeName(e);
		}
	}
	_stateBogusComment(e) {
		let t = this.currentToken;
		switch (e) {
			case D.GREATER_THAN_SIGN:
				this.state = L.DATA, this.emitCurrentComment(t);
				break;
			case D.EOF:
				this.emitCurrentComment(t), this._emitEOFToken();
				break;
			case D.NULL:
				this._err(k.unexpectedNullCharacter), t.data += "�";
				break;
			default: t.data += String.fromCodePoint(e);
		}
	}
	_stateMarkupDeclarationOpen(e) {
		this._consumeSequenceIfMatch(O.DASH_DASH, !0) ? (this._createCommentToken(O.DASH_DASH.length + 1), this.state = L.COMMENT_START) : this._consumeSequenceIfMatch(O.DOCTYPE, !1) ? (this.currentLocation = this.getCurrentLocation(O.DOCTYPE.length + 1), this.state = L.DOCTYPE) : this._consumeSequenceIfMatch(O.CDATA_START, !0) ? this.inForeignNode ? this.state = L.CDATA_SECTION : (this._err(k.cdataInHtmlContent), this._createCommentToken(O.CDATA_START.length + 1), this.currentToken.data = "[CDATA[", this.state = L.BOGUS_COMMENT) : this._ensureHibernation() || (this._err(k.incorrectlyOpenedComment), this._createCommentToken(2), this.state = L.BOGUS_COMMENT, this._stateBogusComment(e));
	}
	_stateCommentStart(e) {
		switch (e) {
			case D.HYPHEN_MINUS:
				this.state = L.COMMENT_START_DASH;
				break;
			case D.GREATER_THAN_SIGN: {
				this._err(k.abruptClosingOfEmptyComment), this.state = L.DATA;
				let e = this.currentToken;
				this.emitCurrentComment(e);
				break;
			}
			default: this.state = L.COMMENT, this._stateComment(e);
		}
	}
	_stateCommentStartDash(e) {
		let t = this.currentToken;
		switch (e) {
			case D.HYPHEN_MINUS:
				this.state = L.COMMENT_END;
				break;
			case D.GREATER_THAN_SIGN:
				this._err(k.abruptClosingOfEmptyComment), this.state = L.DATA, this.emitCurrentComment(t);
				break;
			case D.EOF:
				this._err(k.eofInComment), this.emitCurrentComment(t), this._emitEOFToken();
				break;
			default: t.data += "-", this.state = L.COMMENT, this._stateComment(e);
		}
	}
	_stateComment(e) {
		let t = this.currentToken;
		switch (e) {
			case D.HYPHEN_MINUS:
				this.state = L.COMMENT_END_DASH;
				break;
			case D.LESS_THAN_SIGN:
				t.data += "<", this.state = L.COMMENT_LESS_THAN_SIGN;
				break;
			case D.NULL:
				this._err(k.unexpectedNullCharacter), t.data += "�";
				break;
			case D.EOF:
				this._err(k.eofInComment), this.emitCurrentComment(t), this._emitEOFToken();
				break;
			default: t.data += String.fromCodePoint(e);
		}
	}
	_stateCommentLessThanSign(e) {
		let t = this.currentToken;
		switch (e) {
			case D.EXCLAMATION_MARK:
				t.data += "!", this.state = L.COMMENT_LESS_THAN_SIGN_BANG;
				break;
			case D.LESS_THAN_SIGN:
				t.data += "<";
				break;
			default: this.state = L.COMMENT, this._stateComment(e);
		}
	}
	_stateCommentLessThanSignBang(e) {
		e === D.HYPHEN_MINUS ? this.state = L.COMMENT_LESS_THAN_SIGN_BANG_DASH : (this.state = L.COMMENT, this._stateComment(e));
	}
	_stateCommentLessThanSignBangDash(e) {
		e === D.HYPHEN_MINUS ? this.state = L.COMMENT_LESS_THAN_SIGN_BANG_DASH_DASH : (this.state = L.COMMENT_END_DASH, this._stateCommentEndDash(e));
	}
	_stateCommentLessThanSignBangDashDash(e) {
		e !== D.GREATER_THAN_SIGN && e !== D.EOF && this._err(k.nestedComment), this.state = L.COMMENT_END, this._stateCommentEnd(e);
	}
	_stateCommentEndDash(e) {
		let t = this.currentToken;
		switch (e) {
			case D.HYPHEN_MINUS:
				this.state = L.COMMENT_END;
				break;
			case D.EOF:
				this._err(k.eofInComment), this.emitCurrentComment(t), this._emitEOFToken();
				break;
			default: t.data += "-", this.state = L.COMMENT, this._stateComment(e);
		}
	}
	_stateCommentEnd(e) {
		let t = this.currentToken;
		switch (e) {
			case D.GREATER_THAN_SIGN:
				this.state = L.DATA, this.emitCurrentComment(t);
				break;
			case D.EXCLAMATION_MARK:
				this.state = L.COMMENT_END_BANG;
				break;
			case D.HYPHEN_MINUS:
				t.data += "-";
				break;
			case D.EOF:
				this._err(k.eofInComment), this.emitCurrentComment(t), this._emitEOFToken();
				break;
			default: t.data += "--", this.state = L.COMMENT, this._stateComment(e);
		}
	}
	_stateCommentEndBang(e) {
		let t = this.currentToken;
		switch (e) {
			case D.HYPHEN_MINUS:
				t.data += "--!", this.state = L.COMMENT_END_DASH;
				break;
			case D.GREATER_THAN_SIGN:
				this._err(k.incorrectlyClosedComment), this.state = L.DATA, this.emitCurrentComment(t);
				break;
			case D.EOF:
				this._err(k.eofInComment), this.emitCurrentComment(t), this._emitEOFToken();
				break;
			default: t.data += "--!", this.state = L.COMMENT, this._stateComment(e);
		}
	}
	_stateDoctype(e) {
		switch (e) {
			case D.SPACE:
			case D.LINE_FEED:
			case D.TABULATION:
			case D.FORM_FEED:
				this.state = L.BEFORE_DOCTYPE_NAME;
				break;
			case D.GREATER_THAN_SIGN:
				this.state = L.BEFORE_DOCTYPE_NAME, this._stateBeforeDoctypeName(e);
				break;
			case D.EOF: {
				this._err(k.eofInDoctype), this._createDoctypeToken(null);
				let e = this.currentToken;
				e.forceQuirks = !0, this.emitCurrentDoctype(e), this._emitEOFToken();
				break;
			}
			default: this._err(k.missingWhitespaceBeforeDoctypeName), this.state = L.BEFORE_DOCTYPE_NAME, this._stateBeforeDoctypeName(e);
		}
	}
	_stateBeforeDoctypeName(e) {
		if (nn(e)) this._createDoctypeToken(String.fromCharCode(sn(e))), this.state = L.DOCTYPE_NAME;
		else switch (e) {
			case D.SPACE:
			case D.LINE_FEED:
			case D.TABULATION:
			case D.FORM_FEED: break;
			case D.NULL:
				this._err(k.unexpectedNullCharacter), this._createDoctypeToken("�"), this.state = L.DOCTYPE_NAME;
				break;
			case D.GREATER_THAN_SIGN: {
				this._err(k.missingDoctypeName), this._createDoctypeToken(null);
				let e = this.currentToken;
				e.forceQuirks = !0, this.emitCurrentDoctype(e), this.state = L.DATA;
				break;
			}
			case D.EOF: {
				this._err(k.eofInDoctype), this._createDoctypeToken(null);
				let e = this.currentToken;
				e.forceQuirks = !0, this.emitCurrentDoctype(e), this._emitEOFToken();
				break;
			}
			default: this._createDoctypeToken(String.fromCodePoint(e)), this.state = L.DOCTYPE_NAME;
		}
	}
	_stateDoctypeName(e) {
		let t = this.currentToken;
		switch (e) {
			case D.SPACE:
			case D.LINE_FEED:
			case D.TABULATION:
			case D.FORM_FEED:
				this.state = L.AFTER_DOCTYPE_NAME;
				break;
			case D.GREATER_THAN_SIGN:
				this.state = L.DATA, this.emitCurrentDoctype(t);
				break;
			case D.NULL:
				this._err(k.unexpectedNullCharacter), t.name += "�";
				break;
			case D.EOF:
				this._err(k.eofInDoctype), t.forceQuirks = !0, this.emitCurrentDoctype(t), this._emitEOFToken();
				break;
			default: t.name += String.fromCodePoint(nn(e) ? sn(e) : e);
		}
	}
	_stateAfterDoctypeName(e) {
		let t = this.currentToken;
		switch (e) {
			case D.SPACE:
			case D.LINE_FEED:
			case D.TABULATION:
			case D.FORM_FEED: break;
			case D.GREATER_THAN_SIGN:
				this.state = L.DATA, this.emitCurrentDoctype(t);
				break;
			case D.EOF:
				this._err(k.eofInDoctype), t.forceQuirks = !0, this.emitCurrentDoctype(t), this._emitEOFToken();
				break;
			default: this._consumeSequenceIfMatch(O.PUBLIC, !1) ? this.state = L.AFTER_DOCTYPE_PUBLIC_KEYWORD : this._consumeSequenceIfMatch(O.SYSTEM, !1) ? this.state = L.AFTER_DOCTYPE_SYSTEM_KEYWORD : this._ensureHibernation() || (this._err(k.invalidCharacterSequenceAfterDoctypeName), t.forceQuirks = !0, this.state = L.BOGUS_DOCTYPE, this._stateBogusDoctype(e));
		}
	}
	_stateAfterDoctypePublicKeyword(e) {
		let t = this.currentToken;
		switch (e) {
			case D.SPACE:
			case D.LINE_FEED:
			case D.TABULATION:
			case D.FORM_FEED:
				this.state = L.BEFORE_DOCTYPE_PUBLIC_IDENTIFIER;
				break;
			case D.QUOTATION_MARK:
				this._err(k.missingWhitespaceAfterDoctypePublicKeyword), t.publicId = "", this.state = L.DOCTYPE_PUBLIC_IDENTIFIER_DOUBLE_QUOTED;
				break;
			case D.APOSTROPHE:
				this._err(k.missingWhitespaceAfterDoctypePublicKeyword), t.publicId = "", this.state = L.DOCTYPE_PUBLIC_IDENTIFIER_SINGLE_QUOTED;
				break;
			case D.GREATER_THAN_SIGN:
				this._err(k.missingDoctypePublicIdentifier), t.forceQuirks = !0, this.state = L.DATA, this.emitCurrentDoctype(t);
				break;
			case D.EOF:
				this._err(k.eofInDoctype), t.forceQuirks = !0, this.emitCurrentDoctype(t), this._emitEOFToken();
				break;
			default: this._err(k.missingQuoteBeforeDoctypePublicIdentifier), t.forceQuirks = !0, this.state = L.BOGUS_DOCTYPE, this._stateBogusDoctype(e);
		}
	}
	_stateBeforeDoctypePublicIdentifier(e) {
		let t = this.currentToken;
		switch (e) {
			case D.SPACE:
			case D.LINE_FEED:
			case D.TABULATION:
			case D.FORM_FEED: break;
			case D.QUOTATION_MARK:
				t.publicId = "", this.state = L.DOCTYPE_PUBLIC_IDENTIFIER_DOUBLE_QUOTED;
				break;
			case D.APOSTROPHE:
				t.publicId = "", this.state = L.DOCTYPE_PUBLIC_IDENTIFIER_SINGLE_QUOTED;
				break;
			case D.GREATER_THAN_SIGN:
				this._err(k.missingDoctypePublicIdentifier), t.forceQuirks = !0, this.state = L.DATA, this.emitCurrentDoctype(t);
				break;
			case D.EOF:
				this._err(k.eofInDoctype), t.forceQuirks = !0, this.emitCurrentDoctype(t), this._emitEOFToken();
				break;
			default: this._err(k.missingQuoteBeforeDoctypePublicIdentifier), t.forceQuirks = !0, this.state = L.BOGUS_DOCTYPE, this._stateBogusDoctype(e);
		}
	}
	_stateDoctypePublicIdentifierDoubleQuoted(e) {
		let t = this.currentToken;
		switch (e) {
			case D.QUOTATION_MARK:
				this.state = L.AFTER_DOCTYPE_PUBLIC_IDENTIFIER;
				break;
			case D.NULL:
				this._err(k.unexpectedNullCharacter), t.publicId += "�";
				break;
			case D.GREATER_THAN_SIGN:
				this._err(k.abruptDoctypePublicIdentifier), t.forceQuirks = !0, this.emitCurrentDoctype(t), this.state = L.DATA;
				break;
			case D.EOF:
				this._err(k.eofInDoctype), t.forceQuirks = !0, this.emitCurrentDoctype(t), this._emitEOFToken();
				break;
			default: t.publicId += String.fromCodePoint(e);
		}
	}
	_stateDoctypePublicIdentifierSingleQuoted(e) {
		let t = this.currentToken;
		switch (e) {
			case D.APOSTROPHE:
				this.state = L.AFTER_DOCTYPE_PUBLIC_IDENTIFIER;
				break;
			case D.NULL:
				this._err(k.unexpectedNullCharacter), t.publicId += "�";
				break;
			case D.GREATER_THAN_SIGN:
				this._err(k.abruptDoctypePublicIdentifier), t.forceQuirks = !0, this.emitCurrentDoctype(t), this.state = L.DATA;
				break;
			case D.EOF:
				this._err(k.eofInDoctype), t.forceQuirks = !0, this.emitCurrentDoctype(t), this._emitEOFToken();
				break;
			default: t.publicId += String.fromCodePoint(e);
		}
	}
	_stateAfterDoctypePublicIdentifier(e) {
		let t = this.currentToken;
		switch (e) {
			case D.SPACE:
			case D.LINE_FEED:
			case D.TABULATION:
			case D.FORM_FEED:
				this.state = L.BETWEEN_DOCTYPE_PUBLIC_AND_SYSTEM_IDENTIFIERS;
				break;
			case D.GREATER_THAN_SIGN:
				this.state = L.DATA, this.emitCurrentDoctype(t);
				break;
			case D.QUOTATION_MARK:
				this._err(k.missingWhitespaceBetweenDoctypePublicAndSystemIdentifiers), t.systemId = "", this.state = L.DOCTYPE_SYSTEM_IDENTIFIER_DOUBLE_QUOTED;
				break;
			case D.APOSTROPHE:
				this._err(k.missingWhitespaceBetweenDoctypePublicAndSystemIdentifiers), t.systemId = "", this.state = L.DOCTYPE_SYSTEM_IDENTIFIER_SINGLE_QUOTED;
				break;
			case D.EOF:
				this._err(k.eofInDoctype), t.forceQuirks = !0, this.emitCurrentDoctype(t), this._emitEOFToken();
				break;
			default: this._err(k.missingQuoteBeforeDoctypeSystemIdentifier), t.forceQuirks = !0, this.state = L.BOGUS_DOCTYPE, this._stateBogusDoctype(e);
		}
	}
	_stateBetweenDoctypePublicAndSystemIdentifiers(e) {
		let t = this.currentToken;
		switch (e) {
			case D.SPACE:
			case D.LINE_FEED:
			case D.TABULATION:
			case D.FORM_FEED: break;
			case D.GREATER_THAN_SIGN:
				this.emitCurrentDoctype(t), this.state = L.DATA;
				break;
			case D.QUOTATION_MARK:
				t.systemId = "", this.state = L.DOCTYPE_SYSTEM_IDENTIFIER_DOUBLE_QUOTED;
				break;
			case D.APOSTROPHE:
				t.systemId = "", this.state = L.DOCTYPE_SYSTEM_IDENTIFIER_SINGLE_QUOTED;
				break;
			case D.EOF:
				this._err(k.eofInDoctype), t.forceQuirks = !0, this.emitCurrentDoctype(t), this._emitEOFToken();
				break;
			default: this._err(k.missingQuoteBeforeDoctypeSystemIdentifier), t.forceQuirks = !0, this.state = L.BOGUS_DOCTYPE, this._stateBogusDoctype(e);
		}
	}
	_stateAfterDoctypeSystemKeyword(e) {
		let t = this.currentToken;
		switch (e) {
			case D.SPACE:
			case D.LINE_FEED:
			case D.TABULATION:
			case D.FORM_FEED:
				this.state = L.BEFORE_DOCTYPE_SYSTEM_IDENTIFIER;
				break;
			case D.QUOTATION_MARK:
				this._err(k.missingWhitespaceAfterDoctypeSystemKeyword), t.systemId = "", this.state = L.DOCTYPE_SYSTEM_IDENTIFIER_DOUBLE_QUOTED;
				break;
			case D.APOSTROPHE:
				this._err(k.missingWhitespaceAfterDoctypeSystemKeyword), t.systemId = "", this.state = L.DOCTYPE_SYSTEM_IDENTIFIER_SINGLE_QUOTED;
				break;
			case D.GREATER_THAN_SIGN:
				this._err(k.missingDoctypeSystemIdentifier), t.forceQuirks = !0, this.state = L.DATA, this.emitCurrentDoctype(t);
				break;
			case D.EOF:
				this._err(k.eofInDoctype), t.forceQuirks = !0, this.emitCurrentDoctype(t), this._emitEOFToken();
				break;
			default: this._err(k.missingQuoteBeforeDoctypeSystemIdentifier), t.forceQuirks = !0, this.state = L.BOGUS_DOCTYPE, this._stateBogusDoctype(e);
		}
	}
	_stateBeforeDoctypeSystemIdentifier(e) {
		let t = this.currentToken;
		switch (e) {
			case D.SPACE:
			case D.LINE_FEED:
			case D.TABULATION:
			case D.FORM_FEED: break;
			case D.QUOTATION_MARK:
				t.systemId = "", this.state = L.DOCTYPE_SYSTEM_IDENTIFIER_DOUBLE_QUOTED;
				break;
			case D.APOSTROPHE:
				t.systemId = "", this.state = L.DOCTYPE_SYSTEM_IDENTIFIER_SINGLE_QUOTED;
				break;
			case D.GREATER_THAN_SIGN:
				this._err(k.missingDoctypeSystemIdentifier), t.forceQuirks = !0, this.state = L.DATA, this.emitCurrentDoctype(t);
				break;
			case D.EOF:
				this._err(k.eofInDoctype), t.forceQuirks = !0, this.emitCurrentDoctype(t), this._emitEOFToken();
				break;
			default: this._err(k.missingQuoteBeforeDoctypeSystemIdentifier), t.forceQuirks = !0, this.state = L.BOGUS_DOCTYPE, this._stateBogusDoctype(e);
		}
	}
	_stateDoctypeSystemIdentifierDoubleQuoted(e) {
		let t = this.currentToken;
		switch (e) {
			case D.QUOTATION_MARK:
				this.state = L.AFTER_DOCTYPE_SYSTEM_IDENTIFIER;
				break;
			case D.NULL:
				this._err(k.unexpectedNullCharacter), t.systemId += "�";
				break;
			case D.GREATER_THAN_SIGN:
				this._err(k.abruptDoctypeSystemIdentifier), t.forceQuirks = !0, this.emitCurrentDoctype(t), this.state = L.DATA;
				break;
			case D.EOF:
				this._err(k.eofInDoctype), t.forceQuirks = !0, this.emitCurrentDoctype(t), this._emitEOFToken();
				break;
			default: t.systemId += String.fromCodePoint(e);
		}
	}
	_stateDoctypeSystemIdentifierSingleQuoted(e) {
		let t = this.currentToken;
		switch (e) {
			case D.APOSTROPHE:
				this.state = L.AFTER_DOCTYPE_SYSTEM_IDENTIFIER;
				break;
			case D.NULL:
				this._err(k.unexpectedNullCharacter), t.systemId += "�";
				break;
			case D.GREATER_THAN_SIGN:
				this._err(k.abruptDoctypeSystemIdentifier), t.forceQuirks = !0, this.emitCurrentDoctype(t), this.state = L.DATA;
				break;
			case D.EOF:
				this._err(k.eofInDoctype), t.forceQuirks = !0, this.emitCurrentDoctype(t), this._emitEOFToken();
				break;
			default: t.systemId += String.fromCodePoint(e);
		}
	}
	_stateAfterDoctypeSystemIdentifier(e) {
		let t = this.currentToken;
		switch (e) {
			case D.SPACE:
			case D.LINE_FEED:
			case D.TABULATION:
			case D.FORM_FEED: break;
			case D.GREATER_THAN_SIGN:
				this.emitCurrentDoctype(t), this.state = L.DATA;
				break;
			case D.EOF:
				this._err(k.eofInDoctype), t.forceQuirks = !0, this.emitCurrentDoctype(t), this._emitEOFToken();
				break;
			default: this._err(k.unexpectedCharacterAfterDoctypeSystemIdentifier), this.state = L.BOGUS_DOCTYPE, this._stateBogusDoctype(e);
		}
	}
	_stateBogusDoctype(e) {
		let t = this.currentToken;
		switch (e) {
			case D.GREATER_THAN_SIGN:
				this.emitCurrentDoctype(t), this.state = L.DATA;
				break;
			case D.NULL:
				this._err(k.unexpectedNullCharacter);
				break;
			case D.EOF: this.emitCurrentDoctype(t), this._emitEOFToken();
		}
	}
	_stateCdataSection(e) {
		switch (e) {
			case D.RIGHT_SQUARE_BRACKET:
				this.state = L.CDATA_SECTION_BRACKET;
				break;
			case D.EOF:
				this._err(k.eofInCdata), this._emitEOFToken();
				break;
			default: this._emitCodePoint(e);
		}
	}
	_stateCdataSectionBracket(e) {
		e === D.RIGHT_SQUARE_BRACKET ? this.state = L.CDATA_SECTION_END : (this._emitChars("]"), this.state = L.CDATA_SECTION, this._stateCdataSection(e));
	}
	_stateCdataSectionEnd(e) {
		switch (e) {
			case D.GREATER_THAN_SIGN:
				this.state = L.DATA;
				break;
			case D.RIGHT_SQUARE_BRACKET:
				this._emitChars("]");
				break;
			default: this._emitChars("]]"), this.state = L.CDATA_SECTION, this._stateCdataSection(e);
		}
	}
	_stateCharacterReference() {
		let e = this.entityDecoder.write(this.preprocessor.html, this.preprocessor.pos);
		if (e < 0) {
			if (this.preprocessor.lastChunkWritten) e = this.entityDecoder.end();
			else {
				this.active = !1, this.preprocessor.pos = this.preprocessor.html.length - 1, this.consumedAfterSnapshot = 0, this.preprocessor.endOfChunkHit = !0;
				return;
			}
		}
		e === 0 ? (this.preprocessor.pos = this.entityStartPos, this._flushCodePointConsumedAsCharacterReference(D.AMPERSAND), this.state = !this._isCharacterReferenceInAttribute() && on(this.preprocessor.peek(1)) ? L.AMBIGUOUS_AMPERSAND : this.returnState) : this.state = this.returnState;
	}
	_stateAmbiguousAmpersand(e) {
		on(e) ? this._flushCodePointConsumedAsCharacterReference(e) : (e === D.SEMICOLON && this._err(k.unknownNamedCharacterReference), this.state = this.returnState, this._callState(e));
	}
}, fn = /* @__PURE__ */ new Set([
	F.DD,
	F.DT,
	F.LI,
	F.OPTGROUP,
	F.OPTION,
	F.P,
	F.RB,
	F.RP,
	F.RT,
	F.RTC
]), pn = /* @__PURE__ */ new Set([
	...fn,
	F.CAPTION,
	F.COLGROUP,
	F.TBODY,
	F.TD,
	F.TFOOT,
	F.TH,
	F.THEAD,
	F.TR
]), mn = /* @__PURE__ */ new Set([
	F.APPLET,
	F.CAPTION,
	F.HTML,
	F.MARQUEE,
	F.OBJECT,
	F.TABLE,
	F.TD,
	F.TEMPLATE,
	F.TH
]), hn = /* @__PURE__ */ new Set([
	...mn,
	F.OL,
	F.UL
]), gn = /* @__PURE__ */ new Set([...mn, F.BUTTON]), _n = /* @__PURE__ */ new Set([
	F.ANNOTATION_XML,
	F.MI,
	F.MN,
	F.MO,
	F.MS,
	F.MTEXT
]), vn = /* @__PURE__ */ new Set([
	F.DESC,
	F.FOREIGN_OBJECT,
	F.TITLE
]), yn = /* @__PURE__ */ new Set([
	F.TR,
	F.TEMPLATE,
	F.HTML
]), bn = /* @__PURE__ */ new Set([
	F.TBODY,
	F.TFOOT,
	F.THEAD,
	F.TEMPLATE,
	F.HTML
]), xn = /* @__PURE__ */ new Set([
	F.TABLE,
	F.TEMPLATE,
	F.HTML
]), Sn = /* @__PURE__ */ new Set([F.TD, F.TH]), Cn = class {
	get currentTmplContentOrNode() {
		return this._isInTemplate() ? this.treeAdapter.getTemplateContent(this.current) : this.current;
	}
	constructor(e, t, n) {
		this.treeAdapter = t, this.handler = n, this.items = [], this.tagIDs = [], this.stackTop = -1, this.tmplCount = 0, this.currentTagId = F.UNKNOWN, this.current = e;
	}
	_indexOf(e) {
		return this.items.lastIndexOf(e, this.stackTop);
	}
	_isInTemplate() {
		return this.currentTagId === F.TEMPLATE && this.treeAdapter.getNamespaceURI(this.current) === N.HTML;
	}
	_updateCurrentElement() {
		this.current = this.items[this.stackTop], this.currentTagId = this.tagIDs[this.stackTop];
	}
	push(e, t) {
		this.stackTop++, this.items[this.stackTop] = e, this.current = e, this.tagIDs[this.stackTop] = t, this.currentTagId = t, this._isInTemplate() && this.tmplCount++, this.handler.onItemPush(e, t, !0);
	}
	pop() {
		let e = this.current;
		this.tmplCount > 0 && this._isInTemplate() && this.tmplCount--, this.stackTop--, this._updateCurrentElement(), this.handler.onItemPop(e, !0);
	}
	replace(e, t) {
		let n = this._indexOf(e);
		this.items[n] = t, n === this.stackTop && (this.current = t);
	}
	insertAfter(e, t, n) {
		let r = this._indexOf(e) + 1;
		this.items.splice(r, 0, t), this.tagIDs.splice(r, 0, n), this.stackTop++, r === this.stackTop && this._updateCurrentElement(), this.current && this.currentTagId !== void 0 && this.handler.onItemPush(this.current, this.currentTagId, r === this.stackTop);
	}
	popUntilTagNamePopped(e) {
		let t = this.stackTop + 1;
		do
			t = this.tagIDs.lastIndexOf(e, t - 1);
		while (t > 0 && this.treeAdapter.getNamespaceURI(this.items[t]) !== N.HTML);
		this.shortenToLength(Math.max(t, 0));
	}
	shortenToLength(e) {
		for (; this.stackTop >= e;) {
			let t = this.current;
			this.tmplCount > 0 && this._isInTemplate() && --this.tmplCount, this.stackTop--, this._updateCurrentElement(), this.handler.onItemPop(t, this.stackTop < e);
		}
	}
	popUntilElementPopped(e) {
		let t = this._indexOf(e);
		this.shortenToLength(Math.max(t, 0));
	}
	popUntilPopped(e, t) {
		let n = this._indexOfTagNames(e, t);
		this.shortenToLength(Math.max(n, 0));
	}
	popUntilNumberedHeaderPopped() {
		this.popUntilPopped(en, N.HTML);
	}
	popUntilTableCellPopped() {
		this.popUntilPopped(Sn, N.HTML);
	}
	popAllUpToHtmlElement() {
		this.tmplCount = 0, this.shortenToLength(1);
	}
	_indexOfTagNames(e, t) {
		for (let n = this.stackTop; n >= 0; n--) if (e.has(this.tagIDs[n]) && this.treeAdapter.getNamespaceURI(this.items[n]) === t) return n;
		return -1;
	}
	clearBackTo(e, t) {
		let n = this._indexOfTagNames(e, t);
		this.shortenToLength(n + 1);
	}
	clearBackToTableContext() {
		this.clearBackTo(xn, N.HTML);
	}
	clearBackToTableBodyContext() {
		this.clearBackTo(bn, N.HTML);
	}
	clearBackToTableRowContext() {
		this.clearBackTo(yn, N.HTML);
	}
	remove(e) {
		let t = this._indexOf(e);
		t >= 0 && (t === this.stackTop ? this.pop() : (this.items.splice(t, 1), this.tagIDs.splice(t, 1), this.stackTop--, this._updateCurrentElement(), this.handler.onItemPop(e, !1)));
	}
	tryPeekProperlyNestedBodyElement() {
		return this.stackTop >= 1 && this.tagIDs[1] === F.BODY ? this.items[1] : null;
	}
	contains(e) {
		return this._indexOf(e) > -1;
	}
	getCommonAncestor(e) {
		let t = this._indexOf(e) - 1;
		return t >= 0 ? this.items[t] : null;
	}
	isRootHtmlElementCurrent() {
		return this.stackTop === 0 && this.tagIDs[0] === F.HTML;
	}
	hasInDynamicScope(e, t) {
		for (let n = this.stackTop; n >= 0; n--) {
			let r = this.tagIDs[n];
			switch (this.treeAdapter.getNamespaceURI(this.items[n])) {
				case N.HTML:
					if (r === e) return !0;
					if (t.has(r)) return !1;
					break;
				case N.SVG:
					if (vn.has(r)) return !1;
					break;
				case N.MATHML: if (_n.has(r)) return !1;
			}
		}
		return !0;
	}
	hasInScope(e) {
		return this.hasInDynamicScope(e, mn);
	}
	hasInListItemScope(e) {
		return this.hasInDynamicScope(e, hn);
	}
	hasInButtonScope(e) {
		return this.hasInDynamicScope(e, gn);
	}
	hasNumberedHeaderInScope() {
		for (let e = this.stackTop; e >= 0; e--) {
			let t = this.tagIDs[e];
			switch (this.treeAdapter.getNamespaceURI(this.items[e])) {
				case N.HTML:
					if (en.has(t)) return !0;
					if (mn.has(t)) return !1;
					break;
				case N.SVG:
					if (vn.has(t)) return !1;
					break;
				case N.MATHML: if (_n.has(t)) return !1;
			}
		}
		return !0;
	}
	hasInTableScope(e) {
		for (let t = this.stackTop; t >= 0; t--) if (this.treeAdapter.getNamespaceURI(this.items[t]) === N.HTML) switch (this.tagIDs[t]) {
			case e: return !0;
			case F.TABLE:
			case F.HTML: return !1;
		}
		return !0;
	}
	hasTableBodyContextInTableScope() {
		for (let e = this.stackTop; e >= 0; e--) if (this.treeAdapter.getNamespaceURI(this.items[e]) === N.HTML) switch (this.tagIDs[e]) {
			case F.TBODY:
			case F.THEAD:
			case F.TFOOT: return !0;
			case F.TABLE:
			case F.HTML: return !1;
		}
		return !0;
	}
	hasInSelectScope(e) {
		for (let t = this.stackTop; t >= 0; t--) if (this.treeAdapter.getNamespaceURI(this.items[t]) === N.HTML) switch (this.tagIDs[t]) {
			case e: return !0;
			case F.OPTION:
			case F.OPTGROUP: break;
			default: return !1;
		}
		return !0;
	}
	generateImpliedEndTags() {
		for (; this.currentTagId !== void 0 && fn.has(this.currentTagId);) this.pop();
	}
	generateImpliedEndTagsThoroughly() {
		for (; this.currentTagId !== void 0 && pn.has(this.currentTagId);) this.pop();
	}
	generateImpliedEndTagsWithExclusion(e) {
		for (; this.currentTagId !== void 0 && this.currentTagId !== e && pn.has(this.currentTagId);) this.pop();
	}
}, wn = 3, Tn;
(function(e) {
	e[e.Marker = 0] = "Marker", e[e.Element = 1] = "Element";
})(Tn ||= {});
var En = { type: Tn.Marker }, Dn = class {
	constructor(e) {
		this.treeAdapter = e, this.entries = [], this.bookmark = null;
	}
	_getNoahArkConditionCandidates(e, t) {
		let n = [], r = t.length, i = this.treeAdapter.getTagName(e), a = this.treeAdapter.getNamespaceURI(e);
		for (let e = 0; e < this.entries.length; e++) {
			let t = this.entries[e];
			if (t.type === Tn.Marker) break;
			let { element: o } = t;
			if (this.treeAdapter.getTagName(o) === i && this.treeAdapter.getNamespaceURI(o) === a) {
				let t = this.treeAdapter.getAttrList(o);
				t.length === r && n.push({
					idx: e,
					attrs: t
				});
			}
		}
		return n;
	}
	_ensureNoahArkCondition(e) {
		if (this.entries.length < wn) return;
		let t = this.treeAdapter.getAttrList(e), n = this._getNoahArkConditionCandidates(e, t);
		if (n.length < wn) return;
		let r = new Map(t.map((e) => [e.name, e.value])), i = 0;
		for (let e = 0; e < n.length; e++) {
			let t = n[e];
			t.attrs.every((e) => r.get(e.name) === e.value) && (i += 1, i >= wn && this.entries.splice(t.idx, 1));
		}
	}
	insertMarker() {
		this.entries.unshift(En);
	}
	pushElement(e, t) {
		this._ensureNoahArkCondition(e), this.entries.unshift({
			type: Tn.Element,
			element: e,
			token: t
		});
	}
	insertElementAfterBookmark(e, t) {
		let n = this.entries.indexOf(this.bookmark);
		this.entries.splice(n, 0, {
			type: Tn.Element,
			element: e,
			token: t
		});
	}
	removeEntry(e) {
		let t = this.entries.indexOf(e);
		t !== -1 && this.entries.splice(t, 1);
	}
	clearToLastMarker() {
		let e = this.entries.indexOf(En);
		e === -1 ? this.entries.length = 0 : this.entries.splice(0, e + 1);
	}
	getElementEntryInScopeWithTagName(e) {
		let t = this.entries.find((t) => t.type === Tn.Marker || this.treeAdapter.getTagName(t.element) === e);
		return t && t.type === Tn.Element ? t : null;
	}
	getElementEntry(e) {
		return this.entries.find((t) => t.type === Tn.Element && t.element === e);
	}
}, On = {
	createDocument() {
		return {
			nodeName: "#document",
			mode: Xt.NO_QUIRKS,
			childNodes: []
		};
	},
	createDocumentFragment() {
		return {
			nodeName: "#document-fragment",
			childNodes: []
		};
	},
	createElement(e, t, n) {
		return {
			nodeName: e,
			tagName: e,
			attrs: n,
			namespaceURI: t,
			childNodes: [],
			parentNode: null
		};
	},
	createCommentNode(e) {
		return {
			nodeName: "#comment",
			data: e,
			parentNode: null
		};
	},
	createTextNode(e) {
		return {
			nodeName: "#text",
			value: e,
			parentNode: null
		};
	},
	appendChild(e, t) {
		e.childNodes.push(t), t.parentNode = e;
	},
	insertBefore(e, t, n) {
		let r = e.childNodes.indexOf(n);
		e.childNodes.splice(r, 0, t), t.parentNode = e;
	},
	setTemplateContent(e, t) {
		e.content = t;
	},
	getTemplateContent(e) {
		return e.content;
	},
	setDocumentType(e, t, n, r) {
		let i = e.childNodes.find((e) => e.nodeName === "#documentType");
		if (i) i.name = t, i.publicId = n, i.systemId = r;
		else {
			let i = {
				nodeName: "#documentType",
				name: t,
				publicId: n,
				systemId: r,
				parentNode: null
			};
			On.appendChild(e, i);
		}
	},
	setDocumentMode(e, t) {
		e.mode = t;
	},
	getDocumentMode(e) {
		return e.mode;
	},
	detachNode(e) {
		if (e.parentNode) {
			let t = e.parentNode.childNodes.indexOf(e);
			e.parentNode.childNodes.splice(t, 1), e.parentNode = null;
		}
	},
	insertText(e, t) {
		if (e.childNodes.length > 0) {
			let n = e.childNodes[e.childNodes.length - 1];
			if (On.isTextNode(n)) {
				n.value += t;
				return;
			}
		}
		On.appendChild(e, On.createTextNode(t));
	},
	insertTextBefore(e, t, n) {
		let r = e.childNodes[e.childNodes.indexOf(n) - 1];
		r && On.isTextNode(r) ? r.value += t : On.insertBefore(e, On.createTextNode(t), n);
	},
	adoptAttributes(e, t) {
		let n = new Set(e.attrs.map((e) => e.name));
		for (let r = 0; r < t.length; r++) n.has(t[r].name) || e.attrs.push(t[r]);
	},
	getFirstChild(e) {
		return e.childNodes[0];
	},
	getChildNodes(e) {
		return e.childNodes;
	},
	getParentNode(e) {
		return e.parentNode;
	},
	getAttrList(e) {
		return e.attrs;
	},
	getTagName(e) {
		return e.tagName;
	},
	getNamespaceURI(e) {
		return e.namespaceURI;
	},
	getTextNodeContent(e) {
		return e.value;
	},
	getCommentNodeContent(e) {
		return e.data;
	},
	getDocumentTypeNodeName(e) {
		return e.name;
	},
	getDocumentTypeNodePublicId(e) {
		return e.publicId;
	},
	getDocumentTypeNodeSystemId(e) {
		return e.systemId;
	},
	isTextNode(e) {
		return e.nodeName === "#text";
	},
	isCommentNode(e) {
		return e.nodeName === "#comment";
	},
	isDocumentTypeNode(e) {
		return e.nodeName === "#documentType";
	},
	isElementNode(e) {
		return Object.prototype.hasOwnProperty.call(e, "tagName");
	},
	setNodeSourceCodeLocation(e, t) {
		e.sourceCodeLocation = t;
	},
	getNodeSourceCodeLocation(e) {
		return e.sourceCodeLocation;
	},
	updateNodeSourceCodeLocation(e, t) {
		e.sourceCodeLocation = {
			...e.sourceCodeLocation,
			...t
		};
	}
}, kn = "html", An = "about:legacy-compat", jn = "http://www.ibm.com/data/dtd/v11/ibmxhtml1-transitional.dtd", Mn = /* @__PURE__ */ "+//silmaril//dtd html pro v0r11 19970101//,-//as//dtd html 3.0 aswedit + extensions//,-//advasoft ltd//dtd html 3.0 aswedit + extensions//,-//ietf//dtd html 2.0 level 1//,-//ietf//dtd html 2.0 level 2//,-//ietf//dtd html 2.0 strict level 1//,-//ietf//dtd html 2.0 strict level 2//,-//ietf//dtd html 2.0 strict//,-//ietf//dtd html 2.0//,-//ietf//dtd html 2.1e//,-//ietf//dtd html 3.0//,-//ietf//dtd html 3.2 final//,-//ietf//dtd html 3.2//,-//ietf//dtd html 3//,-//ietf//dtd html level 0//,-//ietf//dtd html level 1//,-//ietf//dtd html level 2//,-//ietf//dtd html level 3//,-//ietf//dtd html strict level 0//,-//ietf//dtd html strict level 1//,-//ietf//dtd html strict level 2//,-//ietf//dtd html strict level 3//,-//ietf//dtd html strict//,-//ietf//dtd html//,-//metrius//dtd metrius presentational//,-//microsoft//dtd internet explorer 2.0 html strict//,-//microsoft//dtd internet explorer 2.0 html//,-//microsoft//dtd internet explorer 2.0 tables//,-//microsoft//dtd internet explorer 3.0 html strict//,-//microsoft//dtd internet explorer 3.0 html//,-//microsoft//dtd internet explorer 3.0 tables//,-//netscape comm. corp.//dtd html//,-//netscape comm. corp.//dtd strict html//,-//o'reilly and associates//dtd html 2.0//,-//o'reilly and associates//dtd html extended 1.0//,-//o'reilly and associates//dtd html extended relaxed 1.0//,-//sq//dtd html 2.0 hotmetal + extensions//,-//softquad software//dtd hotmetal pro 6.0::19990601::extensions to html 4.0//,-//softquad//dtd hotmetal pro 4.0::19971010::extensions to html 4.0//,-//spyglass//dtd html 2.0 extended//,-//sun microsystems corp.//dtd hotjava html//,-//sun microsystems corp.//dtd hotjava strict html//,-//w3c//dtd html 3 1995-03-24//,-//w3c//dtd html 3.2 draft//,-//w3c//dtd html 3.2 final//,-//w3c//dtd html 3.2//,-//w3c//dtd html 3.2s draft//,-//w3c//dtd html 4.0 frameset//,-//w3c//dtd html 4.0 transitional//,-//w3c//dtd html experimental 19960712//,-//w3c//dtd html experimental 970421//,-//w3c//dtd w3 html//,-//w3o//dtd w3 html 3.0//,-//webtechs//dtd mozilla html 2.0//,-//webtechs//dtd mozilla html//".split(","), Nn = [
	...Mn,
	"-//w3c//dtd html 4.01 frameset//",
	"-//w3c//dtd html 4.01 transitional//"
], Pn = /* @__PURE__ */ new Set([
	"-//w3o//dtd w3 html strict 3.0//en//",
	"-/w3c/dtd html 4.0 transitional/en",
	"html"
]), Fn = ["-//w3c//dtd xhtml 1.0 frameset//", "-//w3c//dtd xhtml 1.0 transitional//"], In = [
	...Fn,
	"-//w3c//dtd html 4.01 frameset//",
	"-//w3c//dtd html 4.01 transitional//"
];
function Ln(e, t) {
	return t.some((t) => e.startsWith(t));
}
function Rn(e) {
	return e.name === kn && e.publicId === null && (e.systemId === null || e.systemId === An);
}
function zn(e) {
	if (e.name !== kn) return Xt.QUIRKS;
	let { systemId: t } = e;
	if (t && t.toLowerCase() === jn) return Xt.QUIRKS;
	let { publicId: n } = e;
	if (n !== null) {
		if (n = n.toLowerCase(), Pn.has(n)) return Xt.QUIRKS;
		let e = t === null ? Nn : Mn;
		if (Ln(n, e)) return Xt.QUIRKS;
		if (e = t === null ? Fn : In, Ln(n, e)) return Xt.LIMITED_QUIRKS;
	}
	return Xt.NO_QUIRKS;
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/parse5/7.3.0/4acbced08702cc2b068edf7ec6fbf0576d666aa2a4d5c21d9bdc1830acf8a8e2/node_modules/parse5/dist/common/foreign-content.js
var Bn = {
	TEXT_HTML: "text/html",
	APPLICATION_XML: "application/xhtml+xml"
}, Vn = "definitionurl", Hn = "definitionURL", Un = new Map((/* @__PURE__ */ "attributeName.attributeType.baseFrequency.baseProfile.calcMode.clipPathUnits.diffuseConstant.edgeMode.filterUnits.glyphRef.gradientTransform.gradientUnits.kernelMatrix.kernelUnitLength.keyPoints.keySplines.keyTimes.lengthAdjust.limitingConeAngle.markerHeight.markerUnits.markerWidth.maskContentUnits.maskUnits.numOctaves.pathLength.patternContentUnits.patternTransform.patternUnits.pointsAtX.pointsAtY.pointsAtZ.preserveAlpha.preserveAspectRatio.primitiveUnits.refX.refY.repeatCount.repeatDur.requiredExtensions.requiredFeatures.specularConstant.specularExponent.spreadMethod.startOffset.stdDeviation.stitchTiles.surfaceScale.systemLanguage.tableValues.targetX.targetY.textLength.viewBox.viewTarget.xChannelSelector.yChannelSelector.zoomAndPan".split(".")).map((e) => [e.toLowerCase(), e])), Wn = /* @__PURE__ */ new Map([
	["xlink:actuate", {
		prefix: "xlink",
		name: "actuate",
		namespace: N.XLINK
	}],
	["xlink:arcrole", {
		prefix: "xlink",
		name: "arcrole",
		namespace: N.XLINK
	}],
	["xlink:href", {
		prefix: "xlink",
		name: "href",
		namespace: N.XLINK
	}],
	["xlink:role", {
		prefix: "xlink",
		name: "role",
		namespace: N.XLINK
	}],
	["xlink:show", {
		prefix: "xlink",
		name: "show",
		namespace: N.XLINK
	}],
	["xlink:title", {
		prefix: "xlink",
		name: "title",
		namespace: N.XLINK
	}],
	["xlink:type", {
		prefix: "xlink",
		name: "type",
		namespace: N.XLINK
	}],
	["xml:lang", {
		prefix: "xml",
		name: "lang",
		namespace: N.XML
	}],
	["xml:space", {
		prefix: "xml",
		name: "space",
		namespace: N.XML
	}],
	["xmlns", {
		prefix: "",
		name: "xmlns",
		namespace: N.XMLNS
	}],
	["xmlns:xlink", {
		prefix: "xmlns",
		name: "xlink",
		namespace: N.XMLNS
	}]
]), Gn = new Map((/* @__PURE__ */ "altGlyph.altGlyphDef.altGlyphItem.animateColor.animateMotion.animateTransform.clipPath.feBlend.feColorMatrix.feComponentTransfer.feComposite.feConvolveMatrix.feDiffuseLighting.feDisplacementMap.feDistantLight.feFlood.feFuncA.feFuncB.feFuncG.feFuncR.feGaussianBlur.feImage.feMerge.feMergeNode.feMorphology.feOffset.fePointLight.feSpecularLighting.feSpotLight.feTile.feTurbulence.foreignObject.glyphRef.linearGradient.radialGradient.textPath".split(".")).map((e) => [e.toLowerCase(), e])), Kn = /* @__PURE__ */ new Set([
	F.B,
	F.BIG,
	F.BLOCKQUOTE,
	F.BODY,
	F.BR,
	F.CENTER,
	F.CODE,
	F.DD,
	F.DIV,
	F.DL,
	F.DT,
	F.EM,
	F.EMBED,
	F.H1,
	F.H2,
	F.H3,
	F.H4,
	F.H5,
	F.H6,
	F.HEAD,
	F.HR,
	F.I,
	F.IMG,
	F.LI,
	F.LISTING,
	F.MENU,
	F.META,
	F.NOBR,
	F.OL,
	F.P,
	F.PRE,
	F.RUBY,
	F.S,
	F.SMALL,
	F.SPAN,
	F.STRONG,
	F.STRIKE,
	F.SUB,
	F.SUP,
	F.TABLE,
	F.TT,
	F.U,
	F.UL,
	F.VAR
]);
function qn(e) {
	let t = e.tagID;
	return t === F.FONT && e.attrs.some(({ name: e }) => e === Yt.COLOR || e === Yt.SIZE || e === Yt.FACE) || Kn.has(t);
}
function Jn(e) {
	for (let t = 0; t < e.attrs.length; t++) if (e.attrs[t].name === Vn) {
		e.attrs[t].name = Hn;
		break;
	}
}
function Yn(e) {
	for (let t = 0; t < e.attrs.length; t++) {
		let n = Un.get(e.attrs[t].name);
		n != null && (e.attrs[t].name = n);
	}
}
function Xn(e) {
	for (let t = 0; t < e.attrs.length; t++) {
		let n = Wn.get(e.attrs[t].name);
		n && (e.attrs[t].prefix = n.prefix, e.attrs[t].name = n.name, e.attrs[t].namespace = n.namespace);
	}
}
function Zn(e) {
	let t = Gn.get(e.tagName);
	t != null && (e.tagName = t, e.tagID = Qt(e.tagName));
}
function Qn(e, t) {
	return t === N.MATHML && (e === F.MI || e === F.MO || e === F.MN || e === F.MS || e === F.MTEXT);
}
function $n(e, t, n) {
	if (t === N.MATHML && e === F.ANNOTATION_XML) {
		for (let e = 0; e < n.length; e++) if (n[e].name === Yt.ENCODING) {
			let t = n[e].value.toLowerCase();
			return t === Bn.TEXT_HTML || t === Bn.APPLICATION_XML;
		}
	}
	return t === N.SVG && (e === F.FOREIGN_OBJECT || e === F.DESC || e === F.TITLE);
}
function er(e, t, n, r) {
	return (!r || r === N.HTML) && $n(e, t, n) || (!r || r === N.MATHML) && Qn(e, t);
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/parse5/7.3.0/4acbced08702cc2b068edf7ec6fbf0576d666aa2a4d5c21d9bdc1830acf8a8e2/node_modules/parse5/dist/parser/index.js
var tr = "hidden", nr = 8, rr = 3, z;
(function(e) {
	e[e.INITIAL = 0] = "INITIAL", e[e.BEFORE_HTML = 1] = "BEFORE_HTML", e[e.BEFORE_HEAD = 2] = "BEFORE_HEAD", e[e.IN_HEAD = 3] = "IN_HEAD", e[e.IN_HEAD_NO_SCRIPT = 4] = "IN_HEAD_NO_SCRIPT", e[e.AFTER_HEAD = 5] = "AFTER_HEAD", e[e.IN_BODY = 6] = "IN_BODY", e[e.TEXT = 7] = "TEXT", e[e.IN_TABLE = 8] = "IN_TABLE", e[e.IN_TABLE_TEXT = 9] = "IN_TABLE_TEXT", e[e.IN_CAPTION = 10] = "IN_CAPTION", e[e.IN_COLUMN_GROUP = 11] = "IN_COLUMN_GROUP", e[e.IN_TABLE_BODY = 12] = "IN_TABLE_BODY", e[e.IN_ROW = 13] = "IN_ROW", e[e.IN_CELL = 14] = "IN_CELL", e[e.IN_SELECT = 15] = "IN_SELECT", e[e.IN_SELECT_IN_TABLE = 16] = "IN_SELECT_IN_TABLE", e[e.IN_TEMPLATE = 17] = "IN_TEMPLATE", e[e.AFTER_BODY = 18] = "AFTER_BODY", e[e.IN_FRAMESET = 19] = "IN_FRAMESET", e[e.AFTER_FRAMESET = 20] = "AFTER_FRAMESET", e[e.AFTER_AFTER_BODY = 21] = "AFTER_AFTER_BODY", e[e.AFTER_AFTER_FRAMESET = 22] = "AFTER_AFTER_FRAMESET";
})(z ||= {});
var ir = {
	startLine: -1,
	startCol: -1,
	startOffset: -1,
	endLine: -1,
	endCol: -1,
	endOffset: -1
}, ar = /* @__PURE__ */ new Set([
	F.TABLE,
	F.TBODY,
	F.TFOOT,
	F.THEAD,
	F.TR
]), or = {
	scriptingEnabled: !0,
	sourceCodeLocationInfo: !1,
	treeAdapter: On,
	onParseError: null
}, sr = class {
	constructor(e, t, n = null, r = null) {
		this.fragmentContext = n, this.scriptHandler = r, this.currentToken = null, this.stopped = !1, this.insertionMode = z.INITIAL, this.originalInsertionMode = z.INITIAL, this.headElement = null, this.formElement = null, this.currentNotInHTML = !1, this.tmplInsertionModeStack = [], this.pendingCharacterTokens = [], this.hasNonWhitespacePendingCharacterToken = !1, this.framesetOk = !0, this.skipNextNewLine = !1, this.fosterParentingEnabled = !1, this.options = {
			...or,
			...e
		}, this.treeAdapter = this.options.treeAdapter, this.onParseError = this.options.onParseError, this.onParseError && (this.options.sourceCodeLocationInfo = !0), this.document = t ?? this.treeAdapter.createDocument(), this.tokenizer = new dn(this.options, this), this.activeFormattingElements = new Dn(this.treeAdapter), this.fragmentContextID = n ? Qt(this.treeAdapter.getTagName(n)) : F.UNKNOWN, this._setContextModes(n ?? this.document, this.fragmentContextID), this.openElements = new Cn(this.document, this.treeAdapter, this);
	}
	static parse(e, t) {
		let n = new this(t);
		return n.tokenizer.write(e, !0), n.document;
	}
	static getFragmentParser(e, t) {
		let n = {
			...or,
			...t
		};
		e ??= n.treeAdapter.createElement(P.TEMPLATE, N.HTML, []);
		let r = n.treeAdapter.createElement("documentmock", N.HTML, []), i = new this(n, r, e);
		return i.fragmentContextID === F.TEMPLATE && i.tmplInsertionModeStack.unshift(z.IN_TEMPLATE), i._initTokenizerForFragmentParsing(), i._insertFakeRootElement(), i._resetInsertionMode(), i._findFormInFragmentContext(), i;
	}
	getFragment() {
		let e = this.treeAdapter.getFirstChild(this.document), t = this.treeAdapter.createDocumentFragment();
		return this._adoptNodes(e, t), t;
	}
	_err(e, t, n) {
		if (!this.onParseError) return;
		let r = e.location ?? ir, i = {
			code: t,
			startLine: r.startLine,
			startCol: r.startCol,
			startOffset: r.startOffset,
			endLine: n ? r.startLine : r.endLine,
			endCol: n ? r.startCol : r.endCol,
			endOffset: n ? r.startOffset : r.endOffset
		};
		this.onParseError(i);
	}
	onItemPush(e, t, n) {
		var r, i;
		(i = (r = this.treeAdapter).onItemPush) == null || i.call(r, e), n && this.openElements.stackTop > 0 && this._setContextModes(e, t);
	}
	onItemPop(e, t) {
		var n, r;
		if (this.options.sourceCodeLocationInfo && this._setEndLocation(e, this.currentToken), (r = (n = this.treeAdapter).onItemPop) == null || r.call(n, e, this.openElements.current), t) {
			let e, t;
			this.openElements.stackTop === 0 && this.fragmentContext ? (e = this.fragmentContext, t = this.fragmentContextID) : {current: e, currentTagId: t} = this.openElements, this._setContextModes(e, t);
		}
	}
	_setContextModes(e, t) {
		let n = e === this.document || e && this.treeAdapter.getNamespaceURI(e) === N.HTML;
		this.currentNotInHTML = !n, this.tokenizer.inForeignNode = !n && e !== void 0 && t !== void 0 && !this._isIntegrationPoint(t, e);
	}
	_switchToTextParsing(e, t) {
		this._insertElement(e, N.HTML), this.tokenizer.state = t, this.originalInsertionMode = this.insertionMode, this.insertionMode = z.TEXT;
	}
	switchToPlaintextParsing() {
		this.insertionMode = z.TEXT, this.originalInsertionMode = z.IN_BODY, this.tokenizer.state = R.PLAINTEXT;
	}
	_getAdjustedCurrentElement() {
		return this.openElements.stackTop === 0 && this.fragmentContext ? this.fragmentContext : this.openElements.current;
	}
	_findFormInFragmentContext() {
		let e = this.fragmentContext;
		for (; e;) {
			if (this.treeAdapter.getTagName(e) === P.FORM) {
				this.formElement = e;
				break;
			}
			e = this.treeAdapter.getParentNode(e);
		}
	}
	_initTokenizerForFragmentParsing() {
		if (this.fragmentContext && this.treeAdapter.getNamespaceURI(this.fragmentContext) === N.HTML) switch (this.fragmentContextID) {
			case F.TITLE:
			case F.TEXTAREA:
				this.tokenizer.state = R.RCDATA;
				break;
			case F.STYLE:
			case F.XMP:
			case F.IFRAME:
			case F.NOEMBED:
			case F.NOFRAMES:
			case F.NOSCRIPT:
				this.tokenizer.state = R.RAWTEXT;
				break;
			case F.SCRIPT:
				this.tokenizer.state = R.SCRIPT_DATA;
				break;
			case F.PLAINTEXT: this.tokenizer.state = R.PLAINTEXT;
		}
	}
	_setDocumentType(e) {
		let t = e.name || "", n = e.publicId || "", r = e.systemId || "";
		if (this.treeAdapter.setDocumentType(this.document, t, n, r), e.location) {
			let t = this.treeAdapter.getChildNodes(this.document).find((e) => this.treeAdapter.isDocumentTypeNode(e));
			t && this.treeAdapter.setNodeSourceCodeLocation(t, e.location);
		}
	}
	_attachElementToTree(e, t) {
		if (this.options.sourceCodeLocationInfo) {
			let n = t && {
				...t,
				startTag: t
			};
			this.treeAdapter.setNodeSourceCodeLocation(e, n);
		}
		if (this._shouldFosterParentOnInsertion()) this._fosterParentElement(e);
		else {
			let t = this.openElements.currentTmplContentOrNode;
			this.treeAdapter.appendChild(t ?? this.document, e);
		}
	}
	_appendElement(e, t) {
		let n = this.treeAdapter.createElement(e.tagName, t, e.attrs);
		this._attachElementToTree(n, e.location);
	}
	_insertElement(e, t) {
		let n = this.treeAdapter.createElement(e.tagName, t, e.attrs);
		this._attachElementToTree(n, e.location), this.openElements.push(n, e.tagID);
	}
	_insertFakeElement(e, t) {
		let n = this.treeAdapter.createElement(e, N.HTML, []);
		this._attachElementToTree(n, null), this.openElements.push(n, t);
	}
	_insertTemplate(e) {
		let t = this.treeAdapter.createElement(e.tagName, N.HTML, e.attrs), n = this.treeAdapter.createDocumentFragment();
		this.treeAdapter.setTemplateContent(t, n), this._attachElementToTree(t, e.location), this.openElements.push(t, e.tagID), this.options.sourceCodeLocationInfo && this.treeAdapter.setNodeSourceCodeLocation(n, null);
	}
	_insertFakeRootElement() {
		let e = this.treeAdapter.createElement(P.HTML, N.HTML, []);
		this.options.sourceCodeLocationInfo && this.treeAdapter.setNodeSourceCodeLocation(e, null), this.treeAdapter.appendChild(this.openElements.current, e), this.openElements.push(e, F.HTML);
	}
	_appendCommentNode(e, t) {
		let n = this.treeAdapter.createCommentNode(e.data);
		this.treeAdapter.appendChild(t, n), this.options.sourceCodeLocationInfo && this.treeAdapter.setNodeSourceCodeLocation(n, e.location);
	}
	_insertCharacters(e) {
		let t, n;
		if (this._shouldFosterParentOnInsertion() ? ({parent: t, beforeElement: n} = this._findFosterParentingLocation(), n ? this.treeAdapter.insertTextBefore(t, e.chars, n) : this.treeAdapter.insertText(t, e.chars)) : (t = this.openElements.currentTmplContentOrNode, this.treeAdapter.insertText(t, e.chars)), !e.location) return;
		let r = this.treeAdapter.getChildNodes(t), i = r[(n ? r.lastIndexOf(n) : r.length) - 1];
		if (this.treeAdapter.getNodeSourceCodeLocation(i)) {
			let { endLine: t, endCol: n, endOffset: r } = e.location;
			this.treeAdapter.updateNodeSourceCodeLocation(i, {
				endLine: t,
				endCol: n,
				endOffset: r
			});
		} else this.options.sourceCodeLocationInfo && this.treeAdapter.setNodeSourceCodeLocation(i, e.location);
	}
	_adoptNodes(e, t) {
		for (let n = this.treeAdapter.getFirstChild(e); n; n = this.treeAdapter.getFirstChild(e)) this.treeAdapter.detachNode(n), this.treeAdapter.appendChild(t, n);
	}
	_setEndLocation(e, t) {
		if (this.treeAdapter.getNodeSourceCodeLocation(e) && t.location) {
			let n = t.location, r = this.treeAdapter.getTagName(e), i = t.type === A.END_TAG && r === t.tagName ? {
				endTag: { ...n },
				endLine: n.endLine,
				endCol: n.endCol,
				endOffset: n.endOffset
			} : {
				endLine: n.startLine,
				endCol: n.startCol,
				endOffset: n.startOffset
			};
			this.treeAdapter.updateNodeSourceCodeLocation(e, i);
		}
	}
	shouldProcessStartTagTokenInForeignContent(e) {
		if (!this.currentNotInHTML) return !1;
		let t, n;
		return this.openElements.stackTop === 0 && this.fragmentContext ? (t = this.fragmentContext, n = this.fragmentContextID) : {current: t, currentTagId: n} = this.openElements, e.tagID === F.SVG && this.treeAdapter.getTagName(t) === P.ANNOTATION_XML && this.treeAdapter.getNamespaceURI(t) === N.MATHML ? !1 : this.tokenizer.inForeignNode || (e.tagID === F.MGLYPH || e.tagID === F.MALIGNMARK) && n !== void 0 && !this._isIntegrationPoint(n, t, N.HTML);
	}
	_processToken(e) {
		switch (e.type) {
			case A.CHARACTER:
				this.onCharacter(e);
				break;
			case A.NULL_CHARACTER:
				this.onNullCharacter(e);
				break;
			case A.COMMENT:
				this.onComment(e);
				break;
			case A.DOCTYPE:
				this.onDoctype(e);
				break;
			case A.START_TAG:
				this._processStartTag(e);
				break;
			case A.END_TAG:
				this.onEndTag(e);
				break;
			case A.EOF:
				this.onEof(e);
				break;
			case A.WHITESPACE_CHARACTER: this.onWhitespaceCharacter(e);
		}
	}
	_isIntegrationPoint(e, t, n) {
		return er(e, this.treeAdapter.getNamespaceURI(t), this.treeAdapter.getAttrList(t), n);
	}
	_reconstructActiveFormattingElements() {
		let e = this.activeFormattingElements.entries.length;
		if (e) {
			let t = this.activeFormattingElements.entries.findIndex((e) => e.type === Tn.Marker || this.openElements.contains(e.element)), n = t === -1 ? e - 1 : t - 1;
			for (let e = n; e >= 0; e--) {
				let t = this.activeFormattingElements.entries[e];
				this._insertElement(t.token, this.treeAdapter.getNamespaceURI(t.element)), t.element = this.openElements.current;
			}
		}
	}
	_closeTableCell() {
		this.openElements.generateImpliedEndTags(), this.openElements.popUntilTableCellPopped(), this.activeFormattingElements.clearToLastMarker(), this.insertionMode = z.IN_ROW;
	}
	_closePElement() {
		this.openElements.generateImpliedEndTagsWithExclusion(F.P), this.openElements.popUntilTagNamePopped(F.P);
	}
	_resetInsertionMode() {
		for (let e = this.openElements.stackTop; e >= 0; e--) switch (e === 0 && this.fragmentContext ? this.fragmentContextID : this.openElements.tagIDs[e]) {
			case F.TR:
				this.insertionMode = z.IN_ROW;
				return;
			case F.TBODY:
			case F.THEAD:
			case F.TFOOT:
				this.insertionMode = z.IN_TABLE_BODY;
				return;
			case F.CAPTION:
				this.insertionMode = z.IN_CAPTION;
				return;
			case F.COLGROUP:
				this.insertionMode = z.IN_COLUMN_GROUP;
				return;
			case F.TABLE:
				this.insertionMode = z.IN_TABLE;
				return;
			case F.BODY:
				this.insertionMode = z.IN_BODY;
				return;
			case F.FRAMESET:
				this.insertionMode = z.IN_FRAMESET;
				return;
			case F.SELECT:
				this._resetInsertionModeForSelect(e);
				return;
			case F.TEMPLATE:
				this.insertionMode = this.tmplInsertionModeStack[0];
				return;
			case F.HTML:
				this.insertionMode = this.headElement ? z.AFTER_HEAD : z.BEFORE_HEAD;
				return;
			case F.TD:
			case F.TH:
				if (e > 0) {
					this.insertionMode = z.IN_CELL;
					return;
				}
				break;
			case F.HEAD: if (e > 0) {
				this.insertionMode = z.IN_HEAD;
				return;
			}
		}
		this.insertionMode = z.IN_BODY;
	}
	_resetInsertionModeForSelect(e) {
		if (e > 0) for (let t = e - 1; t > 0; t--) {
			let e = this.openElements.tagIDs[t];
			if (e === F.TEMPLATE) break;
			if (e === F.TABLE) {
				this.insertionMode = z.IN_SELECT_IN_TABLE;
				return;
			}
		}
		this.insertionMode = z.IN_SELECT;
	}
	_isElementCausesFosterParenting(e) {
		return ar.has(e);
	}
	_shouldFosterParentOnInsertion() {
		return this.fosterParentingEnabled && this.openElements.currentTagId !== void 0 && this._isElementCausesFosterParenting(this.openElements.currentTagId);
	}
	_findFosterParentingLocation() {
		for (let e = this.openElements.stackTop; e >= 0; e--) {
			let t = this.openElements.items[e];
			switch (this.openElements.tagIDs[e]) {
				case F.TEMPLATE:
					if (this.treeAdapter.getNamespaceURI(t) === N.HTML) return {
						parent: this.treeAdapter.getTemplateContent(t),
						beforeElement: null
					};
					break;
				case F.TABLE: {
					let n = this.treeAdapter.getParentNode(t);
					return n ? {
						parent: n,
						beforeElement: t
					} : {
						parent: this.openElements.items[e - 1],
						beforeElement: null
					};
				}
			}
		}
		return {
			parent: this.openElements.items[0],
			beforeElement: null
		};
	}
	_fosterParentElement(e) {
		let t = this._findFosterParentingLocation();
		t.beforeElement ? this.treeAdapter.insertBefore(t.parent, e, t.beforeElement) : this.treeAdapter.appendChild(t.parent, e);
	}
	_isSpecialElement(e, t) {
		return $t[this.treeAdapter.getNamespaceURI(e)].has(t);
	}
	onCharacter(e) {
		if (this.skipNextNewLine = !1, this.tokenizer.inForeignNode) Ca(this, e);
		else switch (this.insertionMode) {
			case z.INITIAL:
				br(this, e);
				break;
			case z.BEFORE_HTML:
				Cr(this, e);
				break;
			case z.BEFORE_HEAD:
				Er(this, e);
				break;
			case z.IN_HEAD:
				kr(this, e);
				break;
			case z.IN_HEAD_NO_SCRIPT:
				Mr(this, e);
				break;
			case z.AFTER_HEAD:
				Fr(this, e);
				break;
			case z.IN_BODY:
			case z.IN_CAPTION:
			case z.IN_CELL:
			case z.IN_TEMPLATE:
				Rr(this, e);
				break;
			case z.TEXT:
			case z.IN_SELECT:
			case z.IN_SELECT_IN_TABLE:
				this._insertCharacters(e);
				break;
			case z.IN_TABLE:
			case z.IN_TABLE_BODY:
			case z.IN_ROW:
				Mi(this, e);
				break;
			case z.IN_TABLE_TEXT:
				Gi(this, e);
				break;
			case z.IN_COLUMN_GROUP:
				Qi(this, e);
				break;
			case z.AFTER_BODY:
				ma(this, e);
				break;
			case z.AFTER_AFTER_BODY: ba(this, e);
		}
	}
	onNullCharacter(e) {
		if (this.skipNextNewLine = !1, this.tokenizer.inForeignNode) Sa(this, e);
		else switch (this.insertionMode) {
			case z.INITIAL:
				br(this, e);
				break;
			case z.BEFORE_HTML:
				Cr(this, e);
				break;
			case z.BEFORE_HEAD:
				Er(this, e);
				break;
			case z.IN_HEAD:
				kr(this, e);
				break;
			case z.IN_HEAD_NO_SCRIPT:
				Mr(this, e);
				break;
			case z.AFTER_HEAD:
				Fr(this, e);
				break;
			case z.TEXT:
				this._insertCharacters(e);
				break;
			case z.IN_TABLE:
			case z.IN_TABLE_BODY:
			case z.IN_ROW:
				Mi(this, e);
				break;
			case z.IN_COLUMN_GROUP:
				Qi(this, e);
				break;
			case z.AFTER_BODY:
				ma(this, e);
				break;
			case z.AFTER_AFTER_BODY: ba(this, e);
		}
	}
	onComment(e) {
		if (this.skipNextNewLine = !1, this.currentNotInHTML) hr(this, e);
		else switch (this.insertionMode) {
			case z.INITIAL:
			case z.BEFORE_HTML:
			case z.BEFORE_HEAD:
			case z.IN_HEAD:
			case z.IN_HEAD_NO_SCRIPT:
			case z.AFTER_HEAD:
			case z.IN_BODY:
			case z.IN_TABLE:
			case z.IN_CAPTION:
			case z.IN_COLUMN_GROUP:
			case z.IN_TABLE_BODY:
			case z.IN_ROW:
			case z.IN_CELL:
			case z.IN_SELECT:
			case z.IN_SELECT_IN_TABLE:
			case z.IN_TEMPLATE:
			case z.IN_FRAMESET:
			case z.AFTER_FRAMESET:
				hr(this, e);
				break;
			case z.IN_TABLE_TEXT:
				Ki(this, e);
				break;
			case z.AFTER_BODY:
				gr(this, e);
				break;
			case z.AFTER_AFTER_BODY:
			case z.AFTER_AFTER_FRAMESET: _r(this, e);
		}
	}
	onDoctype(e) {
		switch (this.skipNextNewLine = !1, this.insertionMode) {
			case z.INITIAL:
				yr(this, e);
				break;
			case z.BEFORE_HEAD:
			case z.IN_HEAD:
			case z.IN_HEAD_NO_SCRIPT:
			case z.AFTER_HEAD:
				this._err(e, k.misplacedDoctype);
				break;
			case z.IN_TABLE_TEXT: Ki(this, e);
		}
	}
	onStartTag(e) {
		this.skipNextNewLine = !1, this.currentToken = e, this._processStartTag(e), e.selfClosing && !e.ackSelfClosing && this._err(e, k.nonVoidHtmlElementStartTagWithTrailingSolidus);
	}
	_processStartTag(e) {
		this.shouldProcessStartTagTokenInForeignContent(e) ? Ta(this, e) : this._startTagOutsideForeignContent(e);
	}
	_startTagOutsideForeignContent(e) {
		switch (this.insertionMode) {
			case z.INITIAL:
				br(this, e);
				break;
			case z.BEFORE_HTML:
				xr(this, e);
				break;
			case z.BEFORE_HEAD:
				wr(this, e);
				break;
			case z.IN_HEAD:
				B(this, e);
				break;
			case z.IN_HEAD_NO_SCRIPT:
				Ar(this, e);
				break;
			case z.AFTER_HEAD:
				Nr(this, e);
				break;
			case z.IN_BODY:
				V(this, e);
				break;
			case z.IN_TABLE:
				Vi(this, e);
				break;
			case z.IN_TABLE_TEXT:
				Ki(this, e);
				break;
			case z.IN_CAPTION:
				Ji(this, e);
				break;
			case z.IN_COLUMN_GROUP:
				Xi(this, e);
				break;
			case z.IN_TABLE_BODY:
				$i(this, e);
				break;
			case z.IN_ROW:
				ta(this, e);
				break;
			case z.IN_CELL:
				ra(this, e);
				break;
			case z.IN_SELECT:
				aa(this, e);
				break;
			case z.IN_SELECT_IN_TABLE:
				sa(this, e);
				break;
			case z.IN_TEMPLATE:
				la(this, e);
				break;
			case z.AFTER_BODY:
				fa(this, e);
				break;
			case z.IN_FRAMESET:
				ha(this, e);
				break;
			case z.AFTER_FRAMESET:
				_a(this, e);
				break;
			case z.AFTER_AFTER_BODY:
				ya(this, e);
				break;
			case z.AFTER_AFTER_FRAMESET: xa(this, e);
		}
	}
	onEndTag(e) {
		this.skipNextNewLine = !1, this.currentToken = e, this.currentNotInHTML ? Ea(this, e) : this._endTagOutsideForeignContent(e);
	}
	_endTagOutsideForeignContent(e) {
		switch (this.insertionMode) {
			case z.INITIAL:
				br(this, e);
				break;
			case z.BEFORE_HTML:
				Sr(this, e);
				break;
			case z.BEFORE_HEAD:
				Tr(this, e);
				break;
			case z.IN_HEAD:
				Dr(this, e);
				break;
			case z.IN_HEAD_NO_SCRIPT:
				jr(this, e);
				break;
			case z.AFTER_HEAD:
				Pr(this, e);
				break;
			case z.IN_BODY:
				Oi(this, e);
				break;
			case z.TEXT:
				Ai(this, e);
				break;
			case z.IN_TABLE:
				Hi(this, e);
				break;
			case z.IN_TABLE_TEXT:
				Ki(this, e);
				break;
			case z.IN_CAPTION:
				Yi(this, e);
				break;
			case z.IN_COLUMN_GROUP:
				Zi(this, e);
				break;
			case z.IN_TABLE_BODY:
				ea(this, e);
				break;
			case z.IN_ROW:
				na(this, e);
				break;
			case z.IN_CELL:
				ia(this, e);
				break;
			case z.IN_SELECT:
				oa(this, e);
				break;
			case z.IN_SELECT_IN_TABLE:
				ca(this, e);
				break;
			case z.IN_TEMPLATE:
				ua(this, e);
				break;
			case z.AFTER_BODY:
				pa(this, e);
				break;
			case z.IN_FRAMESET:
				ga(this, e);
				break;
			case z.AFTER_FRAMESET:
				va(this, e);
				break;
			case z.AFTER_AFTER_BODY: ba(this, e);
		}
	}
	onEof(e) {
		switch (this.insertionMode) {
			case z.INITIAL:
				br(this, e);
				break;
			case z.BEFORE_HTML:
				Cr(this, e);
				break;
			case z.BEFORE_HEAD:
				Er(this, e);
				break;
			case z.IN_HEAD:
				kr(this, e);
				break;
			case z.IN_HEAD_NO_SCRIPT:
				Mr(this, e);
				break;
			case z.AFTER_HEAD:
				Fr(this, e);
				break;
			case z.IN_BODY:
			case z.IN_TABLE:
			case z.IN_CAPTION:
			case z.IN_COLUMN_GROUP:
			case z.IN_TABLE_BODY:
			case z.IN_ROW:
			case z.IN_CELL:
			case z.IN_SELECT:
			case z.IN_SELECT_IN_TABLE:
				ki(this, e);
				break;
			case z.TEXT:
				ji(this, e);
				break;
			case z.IN_TABLE_TEXT:
				Ki(this, e);
				break;
			case z.IN_TEMPLATE:
				da(this, e);
				break;
			case z.AFTER_BODY:
			case z.IN_FRAMESET:
			case z.AFTER_FRAMESET:
			case z.AFTER_AFTER_BODY:
			case z.AFTER_AFTER_FRAMESET: vr(this, e);
		}
	}
	onWhitespaceCharacter(e) {
		if (this.skipNextNewLine && (this.skipNextNewLine = !1, e.chars.charCodeAt(0) === D.LINE_FEED)) {
			if (e.chars.length === 1) return;
			e.chars = e.chars.substr(1);
		}
		if (this.tokenizer.inForeignNode) this._insertCharacters(e);
		else switch (this.insertionMode) {
			case z.IN_HEAD:
			case z.IN_HEAD_NO_SCRIPT:
			case z.AFTER_HEAD:
			case z.TEXT:
			case z.IN_COLUMN_GROUP:
			case z.IN_SELECT:
			case z.IN_SELECT_IN_TABLE:
			case z.IN_FRAMESET:
			case z.AFTER_FRAMESET:
				this._insertCharacters(e);
				break;
			case z.IN_BODY:
			case z.IN_CAPTION:
			case z.IN_CELL:
			case z.IN_TEMPLATE:
			case z.AFTER_BODY:
			case z.AFTER_AFTER_BODY:
			case z.AFTER_AFTER_FRAMESET:
				Lr(this, e);
				break;
			case z.IN_TABLE:
			case z.IN_TABLE_BODY:
			case z.IN_ROW:
				Mi(this, e);
				break;
			case z.IN_TABLE_TEXT: Wi(this, e);
		}
	}
};
function cr(e, t) {
	let n = e.activeFormattingElements.getElementEntryInScopeWithTagName(t.tagName);
	return n ? e.openElements.contains(n.element) ? e.openElements.hasInScope(t.tagID) || (n = null) : (e.activeFormattingElements.removeEntry(n), n = null) : Di(e, t), n;
}
function lr(e, t) {
	let n = null, r = e.openElements.stackTop;
	for (; r >= 0; r--) {
		let i = e.openElements.items[r];
		if (i === t.element) break;
		e._isSpecialElement(i, e.openElements.tagIDs[r]) && (n = i);
	}
	return n || (e.openElements.shortenToLength(Math.max(r, 0)), e.activeFormattingElements.removeEntry(t)), n;
}
function ur(e, t, n) {
	let r = t, i = e.openElements.getCommonAncestor(t);
	for (let a = 0, o = i; o !== n; a++, o = i) {
		i = e.openElements.getCommonAncestor(o);
		let n = e.activeFormattingElements.getElementEntry(o), s = n && a >= rr;
		!n || s ? (s && e.activeFormattingElements.removeEntry(n), e.openElements.remove(o)) : (o = dr(e, n), r === t && (e.activeFormattingElements.bookmark = n), e.treeAdapter.detachNode(r), e.treeAdapter.appendChild(o, r), r = o);
	}
	return r;
}
function dr(e, t) {
	let n = e.treeAdapter.getNamespaceURI(t.element), r = e.treeAdapter.createElement(t.token.tagName, n, t.token.attrs);
	return e.openElements.replace(t.element, r), t.element = r, r;
}
function fr(e, t, n) {
	let r = Qt(e.treeAdapter.getTagName(t));
	if (e._isElementCausesFosterParenting(r)) e._fosterParentElement(n);
	else {
		let i = e.treeAdapter.getNamespaceURI(t);
		r === F.TEMPLATE && i === N.HTML && (t = e.treeAdapter.getTemplateContent(t)), e.treeAdapter.appendChild(t, n);
	}
}
function pr(e, t, n) {
	let r = e.treeAdapter.getNamespaceURI(n.element), { token: i } = n, a = e.treeAdapter.createElement(i.tagName, r, i.attrs);
	e._adoptNodes(t, a), e.treeAdapter.appendChild(t, a), e.activeFormattingElements.insertElementAfterBookmark(a, i), e.activeFormattingElements.removeEntry(n), e.openElements.remove(n.element), e.openElements.insertAfter(t, a, i.tagID);
}
function mr(e, t) {
	for (let n = 0; n < nr; n++) {
		let n = cr(e, t);
		if (!n) break;
		let r = lr(e, n);
		if (!r) break;
		e.activeFormattingElements.bookmark = n;
		let i = ur(e, r, n.element), a = e.openElements.getCommonAncestor(n.element);
		e.treeAdapter.detachNode(i), a && fr(e, a, i), pr(e, r, n);
	}
}
function hr(e, t) {
	e._appendCommentNode(t, e.openElements.currentTmplContentOrNode);
}
function gr(e, t) {
	e._appendCommentNode(t, e.openElements.items[0]);
}
function _r(e, t) {
	e._appendCommentNode(t, e.document);
}
function vr(e, t) {
	if (e.stopped = !0, t.location) {
		let n = e.fragmentContext ? 0 : 2;
		for (let r = e.openElements.stackTop; r >= n; r--) e._setEndLocation(e.openElements.items[r], t);
		if (!e.fragmentContext && e.openElements.stackTop >= 0) {
			let n = e.openElements.items[0], r = e.treeAdapter.getNodeSourceCodeLocation(n);
			if (r && !r.endTag && (e._setEndLocation(n, t), e.openElements.stackTop >= 1)) {
				let n = e.openElements.items[1], r = e.treeAdapter.getNodeSourceCodeLocation(n);
				r && !r.endTag && e._setEndLocation(n, t);
			}
		}
	}
}
function yr(e, t) {
	e._setDocumentType(t);
	let n = t.forceQuirks ? Xt.QUIRKS : zn(t);
	Rn(t) || e._err(t, k.nonConformingDoctype), e.treeAdapter.setDocumentMode(e.document, n), e.insertionMode = z.BEFORE_HTML;
}
function br(e, t) {
	e._err(t, k.missingDoctype, !0), e.treeAdapter.setDocumentMode(e.document, Xt.QUIRKS), e.insertionMode = z.BEFORE_HTML, e._processToken(t);
}
function xr(e, t) {
	t.tagID === F.HTML ? (e._insertElement(t, N.HTML), e.insertionMode = z.BEFORE_HEAD) : Cr(e, t);
}
function Sr(e, t) {
	let n = t.tagID;
	(n === F.HTML || n === F.HEAD || n === F.BODY || n === F.BR) && Cr(e, t);
}
function Cr(e, t) {
	e._insertFakeRootElement(), e.insertionMode = z.BEFORE_HEAD, e._processToken(t);
}
function wr(e, t) {
	switch (t.tagID) {
		case F.HTML:
			V(e, t);
			break;
		case F.HEAD:
			e._insertElement(t, N.HTML), e.headElement = e.openElements.current, e.insertionMode = z.IN_HEAD;
			break;
		default: Er(e, t);
	}
}
function Tr(e, t) {
	let n = t.tagID;
	n === F.HEAD || n === F.BODY || n === F.HTML || n === F.BR ? Er(e, t) : e._err(t, k.endTagWithoutMatchingOpenElement);
}
function Er(e, t) {
	e._insertFakeElement(P.HEAD, F.HEAD), e.headElement = e.openElements.current, e.insertionMode = z.IN_HEAD, e._processToken(t);
}
function B(e, t) {
	switch (t.tagID) {
		case F.HTML:
			V(e, t);
			break;
		case F.BASE:
		case F.BASEFONT:
		case F.BGSOUND:
		case F.LINK:
		case F.META:
			e._appendElement(t, N.HTML), t.ackSelfClosing = !0;
			break;
		case F.TITLE:
			e._switchToTextParsing(t, R.RCDATA);
			break;
		case F.NOSCRIPT:
			e.options.scriptingEnabled ? e._switchToTextParsing(t, R.RAWTEXT) : (e._insertElement(t, N.HTML), e.insertionMode = z.IN_HEAD_NO_SCRIPT);
			break;
		case F.NOFRAMES:
		case F.STYLE:
			e._switchToTextParsing(t, R.RAWTEXT);
			break;
		case F.SCRIPT:
			e._switchToTextParsing(t, R.SCRIPT_DATA);
			break;
		case F.TEMPLATE:
			e._insertTemplate(t), e.activeFormattingElements.insertMarker(), e.framesetOk = !1, e.insertionMode = z.IN_TEMPLATE, e.tmplInsertionModeStack.unshift(z.IN_TEMPLATE);
			break;
		case F.HEAD:
			e._err(t, k.misplacedStartTagForHeadElement);
			break;
		default: kr(e, t);
	}
}
function Dr(e, t) {
	switch (t.tagID) {
		case F.HEAD:
			e.openElements.pop(), e.insertionMode = z.AFTER_HEAD;
			break;
		case F.BODY:
		case F.BR:
		case F.HTML:
			kr(e, t);
			break;
		case F.TEMPLATE:
			Or(e, t);
			break;
		default: e._err(t, k.endTagWithoutMatchingOpenElement);
	}
}
function Or(e, t) {
	e.openElements.tmplCount > 0 ? (e.openElements.generateImpliedEndTagsThoroughly(), e.openElements.currentTagId !== F.TEMPLATE && e._err(t, k.closingOfElementWithOpenChildElements), e.openElements.popUntilTagNamePopped(F.TEMPLATE), e.activeFormattingElements.clearToLastMarker(), e.tmplInsertionModeStack.shift(), e._resetInsertionMode()) : e._err(t, k.endTagWithoutMatchingOpenElement);
}
function kr(e, t) {
	e.openElements.pop(), e.insertionMode = z.AFTER_HEAD, e._processToken(t);
}
function Ar(e, t) {
	switch (t.tagID) {
		case F.HTML:
			V(e, t);
			break;
		case F.BASEFONT:
		case F.BGSOUND:
		case F.HEAD:
		case F.LINK:
		case F.META:
		case F.NOFRAMES:
		case F.STYLE:
			B(e, t);
			break;
		case F.NOSCRIPT:
			e._err(t, k.nestedNoscriptInHead);
			break;
		default: Mr(e, t);
	}
}
function jr(e, t) {
	switch (t.tagID) {
		case F.NOSCRIPT:
			e.openElements.pop(), e.insertionMode = z.IN_HEAD;
			break;
		case F.BR:
			Mr(e, t);
			break;
		default: e._err(t, k.endTagWithoutMatchingOpenElement);
	}
}
function Mr(e, t) {
	let n = t.type === A.EOF ? k.openElementsLeftAfterEof : k.disallowedContentInNoscriptInHead;
	e._err(t, n), e.openElements.pop(), e.insertionMode = z.IN_HEAD, e._processToken(t);
}
function Nr(e, t) {
	switch (t.tagID) {
		case F.HTML:
			V(e, t);
			break;
		case F.BODY:
			e._insertElement(t, N.HTML), e.framesetOk = !1, e.insertionMode = z.IN_BODY;
			break;
		case F.FRAMESET:
			e._insertElement(t, N.HTML), e.insertionMode = z.IN_FRAMESET;
			break;
		case F.BASE:
		case F.BASEFONT:
		case F.BGSOUND:
		case F.LINK:
		case F.META:
		case F.NOFRAMES:
		case F.SCRIPT:
		case F.STYLE:
		case F.TEMPLATE:
		case F.TITLE:
			e._err(t, k.abandonedHeadElementChild), e.openElements.push(e.headElement, F.HEAD), B(e, t), e.openElements.remove(e.headElement);
			break;
		case F.HEAD:
			e._err(t, k.misplacedStartTagForHeadElement);
			break;
		default: Fr(e, t);
	}
}
function Pr(e, t) {
	switch (t.tagID) {
		case F.BODY:
		case F.HTML:
		case F.BR:
			Fr(e, t);
			break;
		case F.TEMPLATE:
			Or(e, t);
			break;
		default: e._err(t, k.endTagWithoutMatchingOpenElement);
	}
}
function Fr(e, t) {
	e._insertFakeElement(P.BODY, F.BODY), e.insertionMode = z.IN_BODY, Ir(e, t);
}
function Ir(e, t) {
	switch (t.type) {
		case A.CHARACTER:
			Rr(e, t);
			break;
		case A.WHITESPACE_CHARACTER:
			Lr(e, t);
			break;
		case A.COMMENT:
			hr(e, t);
			break;
		case A.START_TAG:
			V(e, t);
			break;
		case A.END_TAG:
			Oi(e, t);
			break;
		case A.EOF: ki(e, t);
	}
}
function Lr(e, t) {
	e._reconstructActiveFormattingElements(), e._insertCharacters(t);
}
function Rr(e, t) {
	e._reconstructActiveFormattingElements(), e._insertCharacters(t), e.framesetOk = !1;
}
function zr(e, t) {
	e.openElements.tmplCount === 0 && e.treeAdapter.adoptAttributes(e.openElements.items[0], t.attrs);
}
function Br(e, t) {
	let n = e.openElements.tryPeekProperlyNestedBodyElement();
	n && e.openElements.tmplCount === 0 && (e.framesetOk = !1, e.treeAdapter.adoptAttributes(n, t.attrs));
}
function Vr(e, t) {
	let n = e.openElements.tryPeekProperlyNestedBodyElement();
	e.framesetOk && n && (e.treeAdapter.detachNode(n), e.openElements.popAllUpToHtmlElement(), e._insertElement(t, N.HTML), e.insertionMode = z.IN_FRAMESET);
}
function Hr(e, t) {
	e.openElements.hasInButtonScope(F.P) && e._closePElement(), e._insertElement(t, N.HTML);
}
function Ur(e, t) {
	e.openElements.hasInButtonScope(F.P) && e._closePElement(), e.openElements.currentTagId !== void 0 && en.has(e.openElements.currentTagId) && e.openElements.pop(), e._insertElement(t, N.HTML);
}
function Wr(e, t) {
	e.openElements.hasInButtonScope(F.P) && e._closePElement(), e._insertElement(t, N.HTML), e.skipNextNewLine = !0, e.framesetOk = !1;
}
function Gr(e, t) {
	let n = e.openElements.tmplCount > 0;
	(!e.formElement || n) && (e.openElements.hasInButtonScope(F.P) && e._closePElement(), e._insertElement(t, N.HTML), n || (e.formElement = e.openElements.current));
}
function Kr(e, t) {
	e.framesetOk = !1;
	let n = t.tagID;
	for (let t = e.openElements.stackTop; t >= 0; t--) {
		let r = e.openElements.tagIDs[t];
		if (n === F.LI && r === F.LI || (n === F.DD || n === F.DT) && (r === F.DD || r === F.DT)) {
			e.openElements.generateImpliedEndTagsWithExclusion(r), e.openElements.popUntilTagNamePopped(r);
			break;
		}
		if (r !== F.ADDRESS && r !== F.DIV && r !== F.P && e._isSpecialElement(e.openElements.items[t], r)) break;
	}
	e.openElements.hasInButtonScope(F.P) && e._closePElement(), e._insertElement(t, N.HTML);
}
function qr(e, t) {
	e.openElements.hasInButtonScope(F.P) && e._closePElement(), e._insertElement(t, N.HTML), e.tokenizer.state = R.PLAINTEXT;
}
function Jr(e, t) {
	e.openElements.hasInScope(F.BUTTON) && (e.openElements.generateImpliedEndTags(), e.openElements.popUntilTagNamePopped(F.BUTTON)), e._reconstructActiveFormattingElements(), e._insertElement(t, N.HTML), e.framesetOk = !1;
}
function Yr(e, t) {
	let n = e.activeFormattingElements.getElementEntryInScopeWithTagName(P.A);
	n && (mr(e, t), e.openElements.remove(n.element), e.activeFormattingElements.removeEntry(n)), e._reconstructActiveFormattingElements(), e._insertElement(t, N.HTML), e.activeFormattingElements.pushElement(e.openElements.current, t);
}
function Xr(e, t) {
	e._reconstructActiveFormattingElements(), e._insertElement(t, N.HTML), e.activeFormattingElements.pushElement(e.openElements.current, t);
}
function Zr(e, t) {
	e._reconstructActiveFormattingElements(), e.openElements.hasInScope(F.NOBR) && (mr(e, t), e._reconstructActiveFormattingElements()), e._insertElement(t, N.HTML), e.activeFormattingElements.pushElement(e.openElements.current, t);
}
function Qr(e, t) {
	e._reconstructActiveFormattingElements(), e._insertElement(t, N.HTML), e.activeFormattingElements.insertMarker(), e.framesetOk = !1;
}
function $r(e, t) {
	e.treeAdapter.getDocumentMode(e.document) !== Xt.QUIRKS && e.openElements.hasInButtonScope(F.P) && e._closePElement(), e._insertElement(t, N.HTML), e.framesetOk = !1, e.insertionMode = z.IN_TABLE;
}
function ei(e, t) {
	e._reconstructActiveFormattingElements(), e._appendElement(t, N.HTML), e.framesetOk = !1, t.ackSelfClosing = !0;
}
function ti(e) {
	let t = It(e, Yt.TYPE);
	return t != null && t.toLowerCase() === tr;
}
function ni(e, t) {
	e._reconstructActiveFormattingElements(), e._appendElement(t, N.HTML), ti(t) || (e.framesetOk = !1), t.ackSelfClosing = !0;
}
function ri(e, t) {
	e._appendElement(t, N.HTML), t.ackSelfClosing = !0;
}
function ii(e, t) {
	e.openElements.hasInButtonScope(F.P) && e._closePElement(), e._appendElement(t, N.HTML), e.framesetOk = !1, t.ackSelfClosing = !0;
}
function ai(e, t) {
	t.tagName = P.IMG, t.tagID = F.IMG, ei(e, t);
}
function oi(e, t) {
	e._insertElement(t, N.HTML), e.skipNextNewLine = !0, e.tokenizer.state = R.RCDATA, e.originalInsertionMode = e.insertionMode, e.framesetOk = !1, e.insertionMode = z.TEXT;
}
function si(e, t) {
	e.openElements.hasInButtonScope(F.P) && e._closePElement(), e._reconstructActiveFormattingElements(), e.framesetOk = !1, e._switchToTextParsing(t, R.RAWTEXT);
}
function ci(e, t) {
	e.framesetOk = !1, e._switchToTextParsing(t, R.RAWTEXT);
}
function li(e, t) {
	e._switchToTextParsing(t, R.RAWTEXT);
}
function ui(e, t) {
	e._reconstructActiveFormattingElements(), e._insertElement(t, N.HTML), e.framesetOk = !1, e.insertionMode = e.insertionMode === z.IN_TABLE || e.insertionMode === z.IN_CAPTION || e.insertionMode === z.IN_TABLE_BODY || e.insertionMode === z.IN_ROW || e.insertionMode === z.IN_CELL ? z.IN_SELECT_IN_TABLE : z.IN_SELECT;
}
function di(e, t) {
	e.openElements.currentTagId === F.OPTION && e.openElements.pop(), e._reconstructActiveFormattingElements(), e._insertElement(t, N.HTML);
}
function fi(e, t) {
	e.openElements.hasInScope(F.RUBY) && e.openElements.generateImpliedEndTags(), e._insertElement(t, N.HTML);
}
function pi(e, t) {
	e.openElements.hasInScope(F.RUBY) && e.openElements.generateImpliedEndTagsWithExclusion(F.RTC), e._insertElement(t, N.HTML);
}
function mi(e, t) {
	e._reconstructActiveFormattingElements(), Jn(t), Xn(t), t.selfClosing ? e._appendElement(t, N.MATHML) : e._insertElement(t, N.MATHML), t.ackSelfClosing = !0;
}
function hi(e, t) {
	e._reconstructActiveFormattingElements(), Yn(t), Xn(t), t.selfClosing ? e._appendElement(t, N.SVG) : e._insertElement(t, N.SVG), t.ackSelfClosing = !0;
}
function gi(e, t) {
	e._reconstructActiveFormattingElements(), e._insertElement(t, N.HTML);
}
function V(e, t) {
	switch (t.tagID) {
		case F.I:
		case F.S:
		case F.B:
		case F.U:
		case F.EM:
		case F.TT:
		case F.BIG:
		case F.CODE:
		case F.FONT:
		case F.SMALL:
		case F.STRIKE:
		case F.STRONG:
			Xr(e, t);
			break;
		case F.A:
			Yr(e, t);
			break;
		case F.H1:
		case F.H2:
		case F.H3:
		case F.H4:
		case F.H5:
		case F.H6:
			Ur(e, t);
			break;
		case F.P:
		case F.DL:
		case F.OL:
		case F.UL:
		case F.DIV:
		case F.DIR:
		case F.NAV:
		case F.MAIN:
		case F.MENU:
		case F.ASIDE:
		case F.CENTER:
		case F.FIGURE:
		case F.FOOTER:
		case F.HEADER:
		case F.HGROUP:
		case F.DIALOG:
		case F.DETAILS:
		case F.ADDRESS:
		case F.ARTICLE:
		case F.SEARCH:
		case F.SECTION:
		case F.SUMMARY:
		case F.FIELDSET:
		case F.BLOCKQUOTE:
		case F.FIGCAPTION:
			Hr(e, t);
			break;
		case F.LI:
		case F.DD:
		case F.DT:
			Kr(e, t);
			break;
		case F.BR:
		case F.IMG:
		case F.WBR:
		case F.AREA:
		case F.EMBED:
		case F.KEYGEN:
			ei(e, t);
			break;
		case F.HR:
			ii(e, t);
			break;
		case F.RB:
		case F.RTC:
			fi(e, t);
			break;
		case F.RT:
		case F.RP:
			pi(e, t);
			break;
		case F.PRE:
		case F.LISTING:
			Wr(e, t);
			break;
		case F.XMP:
			si(e, t);
			break;
		case F.SVG:
			hi(e, t);
			break;
		case F.HTML:
			zr(e, t);
			break;
		case F.BASE:
		case F.LINK:
		case F.META:
		case F.STYLE:
		case F.TITLE:
		case F.SCRIPT:
		case F.BGSOUND:
		case F.BASEFONT:
		case F.TEMPLATE:
			B(e, t);
			break;
		case F.BODY:
			Br(e, t);
			break;
		case F.FORM:
			Gr(e, t);
			break;
		case F.NOBR:
			Zr(e, t);
			break;
		case F.MATH:
			mi(e, t);
			break;
		case F.TABLE:
			$r(e, t);
			break;
		case F.INPUT:
			ni(e, t);
			break;
		case F.PARAM:
		case F.TRACK:
		case F.SOURCE:
			ri(e, t);
			break;
		case F.IMAGE:
			ai(e, t);
			break;
		case F.BUTTON:
			Jr(e, t);
			break;
		case F.APPLET:
		case F.OBJECT:
		case F.MARQUEE:
			Qr(e, t);
			break;
		case F.IFRAME:
			ci(e, t);
			break;
		case F.SELECT:
			ui(e, t);
			break;
		case F.OPTION:
		case F.OPTGROUP:
			di(e, t);
			break;
		case F.NOEMBED:
		case F.NOFRAMES:
			li(e, t);
			break;
		case F.FRAMESET:
			Vr(e, t);
			break;
		case F.TEXTAREA:
			oi(e, t);
			break;
		case F.NOSCRIPT:
			e.options.scriptingEnabled ? li(e, t) : gi(e, t);
			break;
		case F.PLAINTEXT:
			qr(e, t);
			break;
		case F.COL:
		case F.TH:
		case F.TD:
		case F.TR:
		case F.HEAD:
		case F.FRAME:
		case F.TBODY:
		case F.TFOOT:
		case F.THEAD:
		case F.CAPTION:
		case F.COLGROUP: break;
		default: gi(e, t);
	}
}
function _i(e, t) {
	if (e.openElements.hasInScope(F.BODY) && (e.insertionMode = z.AFTER_BODY, e.options.sourceCodeLocationInfo)) {
		let n = e.openElements.tryPeekProperlyNestedBodyElement();
		n && e._setEndLocation(n, t);
	}
}
function vi(e, t) {
	e.openElements.hasInScope(F.BODY) && (e.insertionMode = z.AFTER_BODY, pa(e, t));
}
function yi(e, t) {
	let n = t.tagID;
	e.openElements.hasInScope(n) && (e.openElements.generateImpliedEndTags(), e.openElements.popUntilTagNamePopped(n));
}
function bi(e) {
	let t = e.openElements.tmplCount > 0, { formElement: n } = e;
	t || (e.formElement = null), (n || t) && e.openElements.hasInScope(F.FORM) && (e.openElements.generateImpliedEndTags(), t ? e.openElements.popUntilTagNamePopped(F.FORM) : n && e.openElements.remove(n));
}
function xi(e) {
	e.openElements.hasInButtonScope(F.P) || e._insertFakeElement(P.P, F.P), e._closePElement();
}
function Si(e) {
	e.openElements.hasInListItemScope(F.LI) && (e.openElements.generateImpliedEndTagsWithExclusion(F.LI), e.openElements.popUntilTagNamePopped(F.LI));
}
function Ci(e, t) {
	let n = t.tagID;
	e.openElements.hasInScope(n) && (e.openElements.generateImpliedEndTagsWithExclusion(n), e.openElements.popUntilTagNamePopped(n));
}
function wi(e) {
	e.openElements.hasNumberedHeaderInScope() && (e.openElements.generateImpliedEndTags(), e.openElements.popUntilNumberedHeaderPopped());
}
function Ti(e, t) {
	let n = t.tagID;
	e.openElements.hasInScope(n) && (e.openElements.generateImpliedEndTags(), e.openElements.popUntilTagNamePopped(n), e.activeFormattingElements.clearToLastMarker());
}
function Ei(e) {
	e._reconstructActiveFormattingElements(), e._insertFakeElement(P.BR, F.BR), e.openElements.pop(), e.framesetOk = !1;
}
function Di(e, t) {
	let n = t.tagName, r = t.tagID;
	for (let t = e.openElements.stackTop; t > 0; t--) {
		let i = e.openElements.items[t], a = e.openElements.tagIDs[t];
		if (r === a && (r !== F.UNKNOWN || e.treeAdapter.getTagName(i) === n)) {
			e.openElements.generateImpliedEndTagsWithExclusion(r), e.openElements.stackTop >= t && e.openElements.shortenToLength(t);
			break;
		}
		if (e._isSpecialElement(i, a)) break;
	}
}
function Oi(e, t) {
	switch (t.tagID) {
		case F.A:
		case F.B:
		case F.I:
		case F.S:
		case F.U:
		case F.EM:
		case F.TT:
		case F.BIG:
		case F.CODE:
		case F.FONT:
		case F.NOBR:
		case F.SMALL:
		case F.STRIKE:
		case F.STRONG:
			mr(e, t);
			break;
		case F.P:
			xi(e);
			break;
		case F.DL:
		case F.UL:
		case F.OL:
		case F.DIR:
		case F.DIV:
		case F.NAV:
		case F.PRE:
		case F.MAIN:
		case F.MENU:
		case F.ASIDE:
		case F.BUTTON:
		case F.CENTER:
		case F.FIGURE:
		case F.FOOTER:
		case F.HEADER:
		case F.HGROUP:
		case F.DIALOG:
		case F.ADDRESS:
		case F.ARTICLE:
		case F.DETAILS:
		case F.SEARCH:
		case F.SECTION:
		case F.SUMMARY:
		case F.LISTING:
		case F.FIELDSET:
		case F.BLOCKQUOTE:
		case F.FIGCAPTION:
			yi(e, t);
			break;
		case F.LI:
			Si(e);
			break;
		case F.DD:
		case F.DT:
			Ci(e, t);
			break;
		case F.H1:
		case F.H2:
		case F.H3:
		case F.H4:
		case F.H5:
		case F.H6:
			wi(e);
			break;
		case F.BR:
			Ei(e);
			break;
		case F.BODY:
			_i(e, t);
			break;
		case F.HTML:
			vi(e, t);
			break;
		case F.FORM:
			bi(e);
			break;
		case F.APPLET:
		case F.OBJECT:
		case F.MARQUEE:
			Ti(e, t);
			break;
		case F.TEMPLATE:
			Or(e, t);
			break;
		default: Di(e, t);
	}
}
function ki(e, t) {
	e.tmplInsertionModeStack.length > 0 ? da(e, t) : vr(e, t);
}
function Ai(e, t) {
	var n;
	t.tagID === F.SCRIPT && ((n = e.scriptHandler) == null || n.call(e, e.openElements.current)), e.openElements.pop(), e.insertionMode = e.originalInsertionMode;
}
function ji(e, t) {
	e._err(t, k.eofInElementThatCanContainOnlyText), e.openElements.pop(), e.insertionMode = e.originalInsertionMode, e.onEof(t);
}
function Mi(e, t) {
	if (e.openElements.currentTagId !== void 0 && ar.has(e.openElements.currentTagId)) switch (e.pendingCharacterTokens.length = 0, e.hasNonWhitespacePendingCharacterToken = !1, e.originalInsertionMode = e.insertionMode, e.insertionMode = z.IN_TABLE_TEXT, t.type) {
		case A.CHARACTER:
			Gi(e, t);
			break;
		case A.WHITESPACE_CHARACTER: Wi(e, t);
	}
	else Ui(e, t);
}
function Ni(e, t) {
	e.openElements.clearBackToTableContext(), e.activeFormattingElements.insertMarker(), e._insertElement(t, N.HTML), e.insertionMode = z.IN_CAPTION;
}
function Pi(e, t) {
	e.openElements.clearBackToTableContext(), e._insertElement(t, N.HTML), e.insertionMode = z.IN_COLUMN_GROUP;
}
function Fi(e, t) {
	e.openElements.clearBackToTableContext(), e._insertFakeElement(P.COLGROUP, F.COLGROUP), e.insertionMode = z.IN_COLUMN_GROUP, Xi(e, t);
}
function Ii(e, t) {
	e.openElements.clearBackToTableContext(), e._insertElement(t, N.HTML), e.insertionMode = z.IN_TABLE_BODY;
}
function Li(e, t) {
	e.openElements.clearBackToTableContext(), e._insertFakeElement(P.TBODY, F.TBODY), e.insertionMode = z.IN_TABLE_BODY, $i(e, t);
}
function Ri(e, t) {
	e.openElements.hasInTableScope(F.TABLE) && (e.openElements.popUntilTagNamePopped(F.TABLE), e._resetInsertionMode(), e._processStartTag(t));
}
function zi(e, t) {
	ti(t) ? e._appendElement(t, N.HTML) : Ui(e, t), t.ackSelfClosing = !0;
}
function Bi(e, t) {
	!e.formElement && e.openElements.tmplCount === 0 && (e._insertElement(t, N.HTML), e.formElement = e.openElements.current, e.openElements.pop());
}
function Vi(e, t) {
	switch (t.tagID) {
		case F.TD:
		case F.TH:
		case F.TR:
			Li(e, t);
			break;
		case F.STYLE:
		case F.SCRIPT:
		case F.TEMPLATE:
			B(e, t);
			break;
		case F.COL:
			Fi(e, t);
			break;
		case F.FORM:
			Bi(e, t);
			break;
		case F.TABLE:
			Ri(e, t);
			break;
		case F.TBODY:
		case F.TFOOT:
		case F.THEAD:
			Ii(e, t);
			break;
		case F.INPUT:
			zi(e, t);
			break;
		case F.CAPTION:
			Ni(e, t);
			break;
		case F.COLGROUP:
			Pi(e, t);
			break;
		default: Ui(e, t);
	}
}
function Hi(e, t) {
	switch (t.tagID) {
		case F.TABLE:
			e.openElements.hasInTableScope(F.TABLE) && (e.openElements.popUntilTagNamePopped(F.TABLE), e._resetInsertionMode());
			break;
		case F.TEMPLATE:
			Or(e, t);
			break;
		case F.BODY:
		case F.CAPTION:
		case F.COL:
		case F.COLGROUP:
		case F.HTML:
		case F.TBODY:
		case F.TD:
		case F.TFOOT:
		case F.TH:
		case F.THEAD:
		case F.TR: break;
		default: Ui(e, t);
	}
}
function Ui(e, t) {
	let n = e.fosterParentingEnabled;
	e.fosterParentingEnabled = !0, Ir(e, t), e.fosterParentingEnabled = n;
}
function Wi(e, t) {
	e.pendingCharacterTokens.push(t);
}
function Gi(e, t) {
	e.pendingCharacterTokens.push(t), e.hasNonWhitespacePendingCharacterToken = !0;
}
function Ki(e, t) {
	let n = 0;
	if (e.hasNonWhitespacePendingCharacterToken) for (; n < e.pendingCharacterTokens.length; n++) Ui(e, e.pendingCharacterTokens[n]);
	else for (; n < e.pendingCharacterTokens.length; n++) e._insertCharacters(e.pendingCharacterTokens[n]);
	e.insertionMode = e.originalInsertionMode, e._processToken(t);
}
var qi = /* @__PURE__ */ new Set([
	F.CAPTION,
	F.COL,
	F.COLGROUP,
	F.TBODY,
	F.TD,
	F.TFOOT,
	F.TH,
	F.THEAD,
	F.TR
]);
function Ji(e, t) {
	let n = t.tagID;
	qi.has(n) ? e.openElements.hasInTableScope(F.CAPTION) && (e.openElements.generateImpliedEndTags(), e.openElements.popUntilTagNamePopped(F.CAPTION), e.activeFormattingElements.clearToLastMarker(), e.insertionMode = z.IN_TABLE, Vi(e, t)) : V(e, t);
}
function Yi(e, t) {
	let n = t.tagID;
	switch (n) {
		case F.CAPTION:
		case F.TABLE:
			e.openElements.hasInTableScope(F.CAPTION) && (e.openElements.generateImpliedEndTags(), e.openElements.popUntilTagNamePopped(F.CAPTION), e.activeFormattingElements.clearToLastMarker(), e.insertionMode = z.IN_TABLE, n === F.TABLE && Hi(e, t));
			break;
		case F.BODY:
		case F.COL:
		case F.COLGROUP:
		case F.HTML:
		case F.TBODY:
		case F.TD:
		case F.TFOOT:
		case F.TH:
		case F.THEAD:
		case F.TR: break;
		default: Oi(e, t);
	}
}
function Xi(e, t) {
	switch (t.tagID) {
		case F.HTML:
			V(e, t);
			break;
		case F.COL:
			e._appendElement(t, N.HTML), t.ackSelfClosing = !0;
			break;
		case F.TEMPLATE:
			B(e, t);
			break;
		default: Qi(e, t);
	}
}
function Zi(e, t) {
	switch (t.tagID) {
		case F.COLGROUP:
			e.openElements.currentTagId === F.COLGROUP && (e.openElements.pop(), e.insertionMode = z.IN_TABLE);
			break;
		case F.TEMPLATE:
			Or(e, t);
			break;
		case F.COL: break;
		default: Qi(e, t);
	}
}
function Qi(e, t) {
	e.openElements.currentTagId === F.COLGROUP && (e.openElements.pop(), e.insertionMode = z.IN_TABLE, e._processToken(t));
}
function $i(e, t) {
	switch (t.tagID) {
		case F.TR:
			e.openElements.clearBackToTableBodyContext(), e._insertElement(t, N.HTML), e.insertionMode = z.IN_ROW;
			break;
		case F.TH:
		case F.TD:
			e.openElements.clearBackToTableBodyContext(), e._insertFakeElement(P.TR, F.TR), e.insertionMode = z.IN_ROW, ta(e, t);
			break;
		case F.CAPTION:
		case F.COL:
		case F.COLGROUP:
		case F.TBODY:
		case F.TFOOT:
		case F.THEAD:
			e.openElements.hasTableBodyContextInTableScope() && (e.openElements.clearBackToTableBodyContext(), e.openElements.pop(), e.insertionMode = z.IN_TABLE, Vi(e, t));
			break;
		default: Vi(e, t);
	}
}
function ea(e, t) {
	let n = t.tagID;
	switch (t.tagID) {
		case F.TBODY:
		case F.TFOOT:
		case F.THEAD:
			e.openElements.hasInTableScope(n) && (e.openElements.clearBackToTableBodyContext(), e.openElements.pop(), e.insertionMode = z.IN_TABLE);
			break;
		case F.TABLE:
			e.openElements.hasTableBodyContextInTableScope() && (e.openElements.clearBackToTableBodyContext(), e.openElements.pop(), e.insertionMode = z.IN_TABLE, Hi(e, t));
			break;
		case F.BODY:
		case F.CAPTION:
		case F.COL:
		case F.COLGROUP:
		case F.HTML:
		case F.TD:
		case F.TH:
		case F.TR: break;
		default: Hi(e, t);
	}
}
function ta(e, t) {
	switch (t.tagID) {
		case F.TH:
		case F.TD:
			e.openElements.clearBackToTableRowContext(), e._insertElement(t, N.HTML), e.insertionMode = z.IN_CELL, e.activeFormattingElements.insertMarker();
			break;
		case F.CAPTION:
		case F.COL:
		case F.COLGROUP:
		case F.TBODY:
		case F.TFOOT:
		case F.THEAD:
		case F.TR:
			e.openElements.hasInTableScope(F.TR) && (e.openElements.clearBackToTableRowContext(), e.openElements.pop(), e.insertionMode = z.IN_TABLE_BODY, $i(e, t));
			break;
		default: Vi(e, t);
	}
}
function na(e, t) {
	switch (t.tagID) {
		case F.TR:
			e.openElements.hasInTableScope(F.TR) && (e.openElements.clearBackToTableRowContext(), e.openElements.pop(), e.insertionMode = z.IN_TABLE_BODY);
			break;
		case F.TABLE:
			e.openElements.hasInTableScope(F.TR) && (e.openElements.clearBackToTableRowContext(), e.openElements.pop(), e.insertionMode = z.IN_TABLE_BODY, ea(e, t));
			break;
		case F.TBODY:
		case F.TFOOT:
		case F.THEAD:
			(e.openElements.hasInTableScope(t.tagID) || e.openElements.hasInTableScope(F.TR)) && (e.openElements.clearBackToTableRowContext(), e.openElements.pop(), e.insertionMode = z.IN_TABLE_BODY, ea(e, t));
			break;
		case F.BODY:
		case F.CAPTION:
		case F.COL:
		case F.COLGROUP:
		case F.HTML:
		case F.TD:
		case F.TH: break;
		default: Hi(e, t);
	}
}
function ra(e, t) {
	let n = t.tagID;
	qi.has(n) ? (e.openElements.hasInTableScope(F.TD) || e.openElements.hasInTableScope(F.TH)) && (e._closeTableCell(), ta(e, t)) : V(e, t);
}
function ia(e, t) {
	let n = t.tagID;
	switch (n) {
		case F.TD:
		case F.TH:
			e.openElements.hasInTableScope(n) && (e.openElements.generateImpliedEndTags(), e.openElements.popUntilTagNamePopped(n), e.activeFormattingElements.clearToLastMarker(), e.insertionMode = z.IN_ROW);
			break;
		case F.TABLE:
		case F.TBODY:
		case F.TFOOT:
		case F.THEAD:
		case F.TR:
			e.openElements.hasInTableScope(n) && (e._closeTableCell(), na(e, t));
			break;
		case F.BODY:
		case F.CAPTION:
		case F.COL:
		case F.COLGROUP:
		case F.HTML: break;
		default: Oi(e, t);
	}
}
function aa(e, t) {
	switch (t.tagID) {
		case F.HTML:
			V(e, t);
			break;
		case F.OPTION:
			e.openElements.currentTagId === F.OPTION && e.openElements.pop(), e._insertElement(t, N.HTML);
			break;
		case F.OPTGROUP:
			e.openElements.currentTagId === F.OPTION && e.openElements.pop(), e.openElements.currentTagId === F.OPTGROUP && e.openElements.pop(), e._insertElement(t, N.HTML);
			break;
		case F.HR:
			e.openElements.currentTagId === F.OPTION && e.openElements.pop(), e.openElements.currentTagId === F.OPTGROUP && e.openElements.pop(), e._appendElement(t, N.HTML), t.ackSelfClosing = !0;
			break;
		case F.INPUT:
		case F.KEYGEN:
		case F.TEXTAREA:
		case F.SELECT:
			e.openElements.hasInSelectScope(F.SELECT) && (e.openElements.popUntilTagNamePopped(F.SELECT), e._resetInsertionMode(), t.tagID !== F.SELECT && e._processStartTag(t));
			break;
		case F.SCRIPT:
		case F.TEMPLATE: B(e, t);
	}
}
function oa(e, t) {
	switch (t.tagID) {
		case F.OPTGROUP:
			e.openElements.stackTop > 0 && e.openElements.currentTagId === F.OPTION && e.openElements.tagIDs[e.openElements.stackTop - 1] === F.OPTGROUP && e.openElements.pop(), e.openElements.currentTagId === F.OPTGROUP && e.openElements.pop();
			break;
		case F.OPTION:
			e.openElements.currentTagId === F.OPTION && e.openElements.pop();
			break;
		case F.SELECT:
			e.openElements.hasInSelectScope(F.SELECT) && (e.openElements.popUntilTagNamePopped(F.SELECT), e._resetInsertionMode());
			break;
		case F.TEMPLATE: Or(e, t);
	}
}
function sa(e, t) {
	let n = t.tagID;
	n === F.CAPTION || n === F.TABLE || n === F.TBODY || n === F.TFOOT || n === F.THEAD || n === F.TR || n === F.TD || n === F.TH ? (e.openElements.popUntilTagNamePopped(F.SELECT), e._resetInsertionMode(), e._processStartTag(t)) : aa(e, t);
}
function ca(e, t) {
	let n = t.tagID;
	n === F.CAPTION || n === F.TABLE || n === F.TBODY || n === F.TFOOT || n === F.THEAD || n === F.TR || n === F.TD || n === F.TH ? e.openElements.hasInTableScope(n) && (e.openElements.popUntilTagNamePopped(F.SELECT), e._resetInsertionMode(), e.onEndTag(t)) : oa(e, t);
}
function la(e, t) {
	switch (t.tagID) {
		case F.BASE:
		case F.BASEFONT:
		case F.BGSOUND:
		case F.LINK:
		case F.META:
		case F.NOFRAMES:
		case F.SCRIPT:
		case F.STYLE:
		case F.TEMPLATE:
		case F.TITLE:
			B(e, t);
			break;
		case F.CAPTION:
		case F.COLGROUP:
		case F.TBODY:
		case F.TFOOT:
		case F.THEAD:
			e.tmplInsertionModeStack[0] = z.IN_TABLE, e.insertionMode = z.IN_TABLE, Vi(e, t);
			break;
		case F.COL:
			e.tmplInsertionModeStack[0] = z.IN_COLUMN_GROUP, e.insertionMode = z.IN_COLUMN_GROUP, Xi(e, t);
			break;
		case F.TR:
			e.tmplInsertionModeStack[0] = z.IN_TABLE_BODY, e.insertionMode = z.IN_TABLE_BODY, $i(e, t);
			break;
		case F.TD:
		case F.TH:
			e.tmplInsertionModeStack[0] = z.IN_ROW, e.insertionMode = z.IN_ROW, ta(e, t);
			break;
		default: e.tmplInsertionModeStack[0] = z.IN_BODY, e.insertionMode = z.IN_BODY, V(e, t);
	}
}
function ua(e, t) {
	t.tagID === F.TEMPLATE && Or(e, t);
}
function da(e, t) {
	e.openElements.tmplCount > 0 ? (e.openElements.popUntilTagNamePopped(F.TEMPLATE), e.activeFormattingElements.clearToLastMarker(), e.tmplInsertionModeStack.shift(), e._resetInsertionMode(), e.onEof(t)) : vr(e, t);
}
function fa(e, t) {
	t.tagID === F.HTML ? V(e, t) : ma(e, t);
}
function pa(e, t) {
	if (t.tagID === F.HTML) {
		if (e.fragmentContext || (e.insertionMode = z.AFTER_AFTER_BODY), e.options.sourceCodeLocationInfo && e.openElements.tagIDs[0] === F.HTML) {
			e._setEndLocation(e.openElements.items[0], t);
			let n = e.openElements.items[1];
			n && !e.treeAdapter.getNodeSourceCodeLocation(n)?.endTag && e._setEndLocation(n, t);
		}
	} else ma(e, t);
}
function ma(e, t) {
	e.insertionMode = z.IN_BODY, Ir(e, t);
}
function ha(e, t) {
	switch (t.tagID) {
		case F.HTML:
			V(e, t);
			break;
		case F.FRAMESET:
			e._insertElement(t, N.HTML);
			break;
		case F.FRAME:
			e._appendElement(t, N.HTML), t.ackSelfClosing = !0;
			break;
		case F.NOFRAMES: B(e, t);
	}
}
function ga(e, t) {
	t.tagID === F.FRAMESET && !e.openElements.isRootHtmlElementCurrent() && (e.openElements.pop(), !e.fragmentContext && e.openElements.currentTagId !== F.FRAMESET && (e.insertionMode = z.AFTER_FRAMESET));
}
function _a(e, t) {
	switch (t.tagID) {
		case F.HTML:
			V(e, t);
			break;
		case F.NOFRAMES: B(e, t);
	}
}
function va(e, t) {
	t.tagID === F.HTML && (e.insertionMode = z.AFTER_AFTER_FRAMESET);
}
function ya(e, t) {
	t.tagID === F.HTML ? V(e, t) : ba(e, t);
}
function ba(e, t) {
	e.insertionMode = z.IN_BODY, Ir(e, t);
}
function xa(e, t) {
	switch (t.tagID) {
		case F.HTML:
			V(e, t);
			break;
		case F.NOFRAMES: B(e, t);
	}
}
function Sa(e, t) {
	t.chars = "�", e._insertCharacters(t);
}
function Ca(e, t) {
	e._insertCharacters(t), e.framesetOk = !1;
}
function wa(e) {
	for (; e.treeAdapter.getNamespaceURI(e.openElements.current) !== N.HTML && e.openElements.currentTagId !== void 0 && !e._isIntegrationPoint(e.openElements.currentTagId, e.openElements.current);) e.openElements.pop();
}
function Ta(e, t) {
	if (qn(t)) wa(e), e._startTagOutsideForeignContent(t);
	else {
		let n = e._getAdjustedCurrentElement(), r = e.treeAdapter.getNamespaceURI(n);
		r === N.MATHML ? Jn(t) : r === N.SVG && (Zn(t), Yn(t)), Xn(t), t.selfClosing ? e._appendElement(t, r) : e._insertElement(t, r), t.ackSelfClosing = !0;
	}
}
function Ea(e, t) {
	if (t.tagID === F.P || t.tagID === F.BR) wa(e), e._endTagOutsideForeignContent(t);
	else for (let n = e.openElements.stackTop; n > 0; n--) {
		let r = e.openElements.items[n];
		if (e.treeAdapter.getNamespaceURI(r) === N.HTML) {
			e._endTagOutsideForeignContent(t);
			break;
		}
		let i = e.treeAdapter.getTagName(r);
		if (i.toLowerCase() === t.tagName) {
			t.tagName = i, e.openElements.shortenToLength(n);
			break;
		}
	}
}
P.AREA, P.BASE, P.BASEFONT, P.BGSOUND, P.BR, P.COL, P.EMBED, P.FRAME, P.HR, P.IMG, P.INPUT, P.KEYGEN, P.LINK, P.META, P.PARAM, P.SOURCE, P.TRACK, P.WBR;
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/unist-util-position/5.0.0/1d79e0203c9b065483fc8ef1df203f43199c23ab75fff695b368215f52aea4a0/node_modules/unist-util-position/lib/index.js
var Da = ka("end"), Oa = ka("start");
function ka(e) {
	return t;
	function t(t) {
		let n = t && t.position && t.position[e] || {};
		if (typeof n.line == "number" && n.line > 0 && typeof n.column == "number" && n.column > 0) return {
			line: n.line,
			column: n.column,
			offset: typeof n.offset == "number" && n.offset > -1 ? n.offset : void 0
		};
	}
}
function Aa(e) {
	let t = Oa(e), n = Da(e);
	if (t && n) return {
		start: t,
		end: n
	};
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/hast-util-raw/9.1.0/39eaa44d71036a37e1c8d2afa86885a645830a00e27763e58fa3b2085a15efb3/node_modules/hast-util-raw/lib/index.js
var ja = /<(\/?)(iframe|noembed|noframes|plaintext|script|style|textarea|title|xmp)(?=[\t\n\f\r />])/gi, Ma = /* @__PURE__ */ new Set([
	"mdxFlowExpression",
	"mdxJsxFlowElement",
	"mdxJsxTextElement",
	"mdxTextExpression",
	"mdxjsEsm"
]), Na = {
	sourceCodeLocationInfo: !0,
	scriptingEnabled: !1
};
function Pa(e, t) {
	let n = Ja(e), r = _("type", {
		handlers: {
			root: Ia,
			element: La,
			text: Ra,
			comment: Va,
			doctype: za,
			raw: Ha
		},
		unknown: Ua
	}), i = {
		parser: n ? new sr(Na) : sr.getFragmentParser(void 0, Na),
		handle(e) {
			r(e, i);
		},
		stitches: !1,
		options: t || {}
	};
	r(e, i), Wa(i, Oa());
	let a = st(n ? i.parser.document : i.parser.getFragment(), { file: i.options.file });
	return i.stitches && C(a, "comment", function(e, t, n) {
		let r = e;
		if (r.value.stitch && n && t !== void 0) {
			let e = n.children;
			return e[t] = r.value.stitch, t;
		}
	}), a.type === "root" && a.children.length === 1 && a.children[0].type === e.type ? a.children[0] : a;
}
function Fa(e, t) {
	let n = -1;
	/* istanbul ignore else - invalid nodes, see rehypejs/rehype-raw#7. */
	if (e) for (; ++n < e.length;) t.handle(e[n]);
}
function Ia(e, t) {
	Fa(e.children, t);
}
function La(e, t) {
	Ka(e, t), Fa(e.children, t), qa(e, t);
}
function Ra(e, t) {
	t.parser.tokenizer.state > 4 && (t.parser.tokenizer.state = 0);
	let n = {
		type: A.CHARACTER,
		chars: e.value,
		location: Ya(e)
	};
	Wa(t, Oa(e)), t.parser.currentToken = n, t.parser._processToken(t.parser.currentToken);
}
function za(e, t) {
	let n = {
		type: A.DOCTYPE,
		name: "html",
		forceQuirks: !1,
		publicId: "",
		systemId: "",
		location: Ya(e)
	};
	Wa(t, Oa(e)), t.parser.currentToken = n, t.parser._processToken(t.parser.currentToken);
}
function Ba(e, t) {
	t.stitches = !0;
	let n = Xa(e);
	"children" in e && "children" in n && (n.children = Pa({
		type: "root",
		children: e.children
	}, t.options).children), Va({
		type: "comment",
		value: { stitch: n }
	}, t);
}
function Va(e, t) {
	let n = e.value, r = {
		type: A.COMMENT,
		data: n,
		location: Ya(e)
	};
	Wa(t, Oa(e)), t.parser.currentToken = r, t.parser._processToken(t.parser.currentToken);
}
function Ha(e, t) {
	/* c8 ignore next 12 -- removed in <https://github.com/inikulin/parse5/pull/897> */
	if (t.parser.tokenizer.preprocessor.html = "", t.parser.tokenizer.preprocessor.pos = -1, t.parser.tokenizer.preprocessor.lastGapPos = -2, t.parser.tokenizer.preprocessor.gapStack = [], t.parser.tokenizer.preprocessor.skipNextNewLine = !1, t.parser.tokenizer.preprocessor.lastChunkWritten = !1, t.parser.tokenizer.preprocessor.endOfChunkHit = !1, t.parser.tokenizer.preprocessor.isEol = !1, Ga(t, Oa(e)), t.parser.tokenizer.write(t.options.tagfilter ? e.value.replace(ja, "&lt;$1$2") : e.value, !1), t.parser.tokenizer._runParsingLoop(), t.parser.tokenizer.state === 72 || t.parser.tokenizer.state === 78) {
		t.parser.tokenizer.preprocessor.lastChunkWritten = !0;
		let e = t.parser.tokenizer._consume();
		t.parser.tokenizer._callState(e);
	}
}
function Ua(e, t) {
	let n = e;
	if (t.options.passThrough && t.options.passThrough.includes(n.type)) Ba(n, t);
	else {
		let e = "";
		throw Ma.has(n.type) && (e = ". It looks like you are using MDX nodes with `hast-util-raw` (or `rehype-raw`). If you use this because you are using remark or rehype plugins that inject `'html'` nodes, then please raise an issue with that plugin, as its a bad and slow idea. If you use this because you are using markdown syntax, then you have to configure this utility (or plugin) to pass through these nodes (see `passThrough` in docs), but you can also migrate to use the MDX syntax"), Error("Cannot compile `" + n.type + "` node" + e);
	}
}
function Wa(e, t) {
	Ga(e, t);
	let n = e.parser.tokenizer.currentCharacterToken;
	n && n.location && (n.location.endLine = e.parser.tokenizer.preprocessor.line, n.location.endCol = e.parser.tokenizer.preprocessor.col + 1, n.location.endOffset = e.parser.tokenizer.preprocessor.offset + 1, e.parser.currentToken = n, e.parser._processToken(e.parser.currentToken)), e.parser.tokenizer.paused = !1, e.parser.tokenizer.inLoop = !1, e.parser.tokenizer.active = !1, e.parser.tokenizer.returnState = R.DATA, e.parser.tokenizer.charRefCode = -1, e.parser.tokenizer.consumedAfterSnapshot = -1, e.parser.tokenizer.currentLocation = null, e.parser.tokenizer.currentCharacterToken = null, e.parser.tokenizer.currentToken = null, e.parser.tokenizer.currentAttr = {
		name: "",
		value: ""
	};
}
function Ga(e, t) {
	if (t && t.offset !== void 0) {
		let n = {
			startLine: t.line,
			startCol: t.column,
			startOffset: t.offset,
			endLine: -1,
			endCol: -1,
			endOffset: -1
		};
		e.parser.tokenizer.preprocessor.lineStartPos = -t.column + 1, e.parser.tokenizer.preprocessor.droppedBufferSize = t.offset, e.parser.tokenizer.preprocessor.line = t.line, e.parser.tokenizer.currentLocation = n;
	}
}
function Ka(e, t) {
	let n = e.tagName.toLowerCase();
	if (t.parser.tokenizer.state === R.PLAINTEXT) return;
	Wa(t, Oa(e));
	let r = t.parser.openElements.current, i = "namespaceURI" in r ? r.namespaceURI : it.html;
	i === it.html && n === "svg" && (i = it.svg);
	let a = vt({
		...e,
		children: []
	}, { space: i === it.svg ? "svg" : "html" }), o = {
		type: A.START_TAG,
		tagName: n,
		tagID: Qt(n),
		selfClosing: !1,
		ackSelfClosing: !1,
		/* c8 ignore next */
		attrs: "attrs" in a ? a.attrs : [],
		location: Ya(e)
	};
	t.parser.currentToken = o, t.parser._processToken(t.parser.currentToken), t.parser.tokenizer.lastStartTagName = n;
}
function qa(e, t) {
	let n = e.tagName.toLowerCase();
	if (!t.parser.tokenizer.inForeignNode && g.includes(n) || t.parser.tokenizer.state === R.PLAINTEXT) return;
	Wa(t, Da(e));
	let r = {
		type: A.END_TAG,
		tagName: n,
		tagID: Qt(n),
		selfClosing: !1,
		ackSelfClosing: !1,
		attrs: [],
		location: Ya(e)
	};
	t.parser.currentToken = r, t.parser._processToken(t.parser.currentToken), n === t.parser.tokenizer.lastStartTagName && (t.parser.tokenizer.state === R.RCDATA || t.parser.tokenizer.state === R.RAWTEXT || t.parser.tokenizer.state === R.SCRIPT_DATA) && (t.parser.tokenizer.state = R.DATA);
}
function Ja(e) {
	let t = e.type === "root" ? e.children[0] : e;
	return !!(t && (t.type === "doctype" || t.type === "element" && t.tagName.toLowerCase() === "html"));
}
function Ya(e) {
	let t = Oa(e) || {
		line: void 0,
		column: void 0,
		offset: void 0
	}, n = Da(e) || {
		line: void 0,
		column: void 0,
		offset: void 0
	};
	return {
		startLine: t.line,
		startCol: t.column,
		startOffset: t.offset,
		endLine: n.line,
		endCol: n.column,
		endOffset: n.offset
	};
}
function Xa(e) {
	return "children" in e ? Ue({
		...e,
		children: []
	}) : Ue(e);
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/rehype-raw/7.0.0/e2f0b5cc74eccecb36110eb9dbd1fc83034a41f651e4e95097a392c9a0bd019b/node_modules/rehype-raw/lib/index.js
function Za(e) {
	return function(t, n) {
		return Pa(t, {
			...e,
			file: n
		});
	};
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/hast-util-sanitize/5.0.2/dae5cc6edd923d029a2e4b47c41abd46e179d73f2f6fd07696f9a483c29e0635/node_modules/hast-util-sanitize/lib/schema.js
var Qa = [
	"ariaDescribedBy",
	"ariaLabel",
	"ariaLabelledBy"
], $a = {
	ancestors: {
		tbody: ["table"],
		td: ["table"],
		th: ["table"],
		thead: ["table"],
		tfoot: ["table"],
		tr: ["table"]
	},
	attributes: {
		a: [
			...Qa,
			"dataFootnoteBackref",
			"dataFootnoteRef",
			["className", "data-footnote-backref"],
			"href"
		],
		blockquote: ["cite"],
		code: [["className", /^language-./]],
		del: ["cite"],
		div: ["itemScope", "itemType"],
		dl: [...Qa],
		h2: [["className", "sr-only"]],
		img: [
			...Qa,
			"longDesc",
			"src"
		],
		input: [["disabled", !0], ["type", "checkbox"]],
		ins: ["cite"],
		li: [["className", "task-list-item"]],
		ol: [...Qa, ["className", "contains-task-list"]],
		q: ["cite"],
		section: ["dataFootnotes", ["className", "footnotes"]],
		source: ["srcSet"],
		summary: [...Qa],
		table: [...Qa],
		ul: [...Qa, ["className", "contains-task-list"]],
		"*": /* @__PURE__ */ "abbr.accept.acceptCharset.accessKey.action.align.alt.axis.border.cellPadding.cellSpacing.char.charOff.charSet.checked.clear.colSpan.color.cols.compact.coords.dateTime.dir.encType.frame.hSpace.headers.height.hrefLang.htmlFor.id.isMap.itemProp.label.lang.maxLength.media.method.multiple.name.noHref.noShade.noWrap.open.prompt.readOnly.rev.rowSpan.rows.rules.scope.selected.shape.size.span.start.summary.tabIndex.title.useMap.vAlign.value.width".split(".")
	},
	clobber: [
		"ariaDescribedBy",
		"ariaLabelledBy",
		"id",
		"name"
	],
	clobberPrefix: "user-content-",
	protocols: {
		cite: ["http", "https"],
		href: [
			"http",
			"https",
			"irc",
			"ircs",
			"mailto",
			"xmpp"
		],
		longDesc: ["http", "https"],
		src: ["http", "https"]
	},
	required: { input: {
		disabled: !0,
		type: "checkbox"
	} },
	strip: ["script"],
	tagNames: /* @__PURE__ */ "a.b.blockquote.br.code.dd.del.details.div.dl.dt.em.h1.h2.h3.h4.h5.h6.hr.i.img.input.ins.kbd.li.ol.p.picture.pre.q.rp.rt.ruby.s.samp.section.source.span.strike.strong.sub.summary.sup.table.tbody.td.tfoot.th.thead.tr.tt.ul.var".split(".")
}, eo = {}.hasOwnProperty;
function to(e, t) {
	let n = {
		type: "root",
		children: []
	}, r = no({
		schema: t ? {
			...$a,
			...t
		} : $a,
		stack: []
	}, e);
	return r && (Array.isArray(r) ? r.length === 1 ? n = r[0] : n.children = r : n = r), n;
}
function no(e, t) {
	if (t && typeof t == "object") {
		let n = t;
		switch (typeof n.type == "string" ? n.type : "") {
			case "comment": return ro(e, n);
			case "doctype": return io(e, n);
			case "element": return ao(e, n);
			case "root": return oo(e, n);
			case "text": return so(e, n);
		}
	}
}
function ro(e, t) {
	if (e.schema.allowComments) {
		let e = typeof t.value == "string" ? t.value : "", n = e.indexOf("-->"), r = {
			type: "comment",
			value: n < 0 ? e : e.slice(0, n)
		};
		return ho(r, t), r;
	}
}
function io(e, t) {
	if (e.schema.allowDoctypes) {
		let e = { type: "doctype" };
		return ho(e, t), e;
	}
}
function ao(e, t) {
	let n = typeof t.tagName == "string" ? t.tagName : "";
	e.stack.push(n);
	let r = co(e, t.children), i = lo(e, t.properties);
	e.stack.pop();
	let a = !1;
	if (n && n !== "*" && (!e.schema.tagNames || e.schema.tagNames.includes(n)) && (a = !0, e.schema.ancestors && eo.call(e.schema.ancestors, n))) {
		let t = e.schema.ancestors[n], r = -1;
		for (a = !1; ++r < t.length;) e.stack.includes(t[r]) && (a = !0);
	}
	if (!a) return e.schema.strip && !e.schema.strip.includes(n) ? r : void 0;
	let o = {
		type: "element",
		tagName: n,
		properties: i,
		children: r
	};
	return ho(o, t), o;
}
function oo(e, t) {
	let n = {
		type: "root",
		children: co(e, t.children)
	};
	return ho(n, t), n;
}
function so(e, t) {
	let n = {
		type: "text",
		value: typeof t.value == "string" ? t.value : ""
	};
	return ho(n, t), n;
}
function co(e, t) {
	let n = [];
	if (Array.isArray(t)) {
		let r = t, i = -1;
		for (; ++i < r.length;) {
			let t = no(e, r[i]);
			t && (Array.isArray(t) ? n.push(...t) : n.push(t));
		}
	}
	return n;
}
function lo(e, t) {
	let n = e.stack[e.stack.length - 1], r = e.schema.attributes, i = e.schema.required, a = r && eo.call(r, n) ? r[n] : void 0, o = r && eo.call(r, "*") ? r["*"] : void 0, s = t && typeof t == "object" ? t : {}, c = {}, l;
	for (l in s) if (eo.call(s, l)) {
		let t = s[l], n = uo(e, go(a, l), l, t);
		n ??= uo(e, go(o, l), l, t), n != null && (c[l] = n);
	}
	if (i && eo.call(i, n)) {
		let e = i[n];
		for (l in e) eo.call(e, l) && !eo.call(c, l) && (c[l] = e[l]);
	}
	return c;
}
function uo(e, t, n, r) {
	return t ? Array.isArray(r) ? fo(e, t, n, r) : po(e, t, n, r) : void 0;
}
function fo(e, t, n, r) {
	let i = -1, a = [];
	for (; ++i < r.length;) {
		let o = po(e, t, n, r[i]);
		(typeof o == "number" || typeof o == "string") && a.push(o);
	}
	return a;
}
function po(e, t, n, r) {
	if ((typeof r == "boolean" || typeof r == "number" || typeof r == "string") && mo(e, n, r)) {
		if (typeof t == "object" && t.length > 1) {
			let e = !1, n = 0;
			for (; ++n < t.length;) {
				let i = t[n];
				if (i && typeof i == "object" && "flags" in i) {
					if (i.test(String(r))) {
						e = !0;
						break;
					}
				} else if (i === r) {
					e = !0;
					break;
				}
			}
			if (!e) return;
		}
		return e.schema.clobber && e.schema.clobberPrefix && e.schema.clobber.includes(n) ? e.schema.clobberPrefix + r : r;
	}
}
function mo(e, t, n) {
	let r = e.schema.protocols && eo.call(e.schema.protocols, t) ? e.schema.protocols[t] : void 0;
	if (!r || r.length === 0) return !0;
	let i = String(n), a = i.indexOf(":"), o = i.indexOf("?"), s = i.indexOf("#"), c = i.indexOf("/");
	if (a < 0 || c > -1 && a > c || o > -1 && a > o || s > -1 && a > s) return !0;
	let l = -1;
	for (; ++l < r.length;) {
		let e = r[l];
		if (a === e.length && i.slice(0, e.length) === e) return !0;
	}
	return !1;
}
function ho(e, t) {
	let n = Aa(t);
	t.data && (e.data = Ue(t.data)), n && (e.position = n);
}
function go(e, t) {
	let n, r = -1;
	if (e) for (; ++r < e.length;) {
		let i = e[r], a = typeof i == "string" ? i : i[0];
		if (a === t) return i;
		a === "data*" && (n = i);
	}
	if (t.length > 4 && t.slice(0, 4).toLowerCase() === "data") return n;
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/rehype-sanitize/6.0.0/a5fb9f442ca2faf093a61df8d8a98bd3fc316381236ad7e7a2de195a2d33df53/node_modules/rehype-sanitize/lib/index.js
function _o(e) {
	return function(t) {
		return to(t, e);
	};
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/mdast-util-find-and-replace/3.0.2/e8e9dd07825968186a98cb3f2d70f90d9c70a8db891b8487d6686d28aa3ee983/node_modules/mdast-util-find-and-replace/lib/index.js
function vo(e, t, n) {
	let r = ae((n || {}).ignore || []), i = yo(t), a = -1;
	for (; ++a < i.length;) ie(e, "text", o);
	function o(e, t) {
		let n = -1, i;
		for (; ++n < t.length;) {
			let e = t[n], a = i ? i.children : void 0;
			if (r(e, a ? a.indexOf(e) : void 0, i)) return;
			i = e;
		}
		if (i) return s(e, t);
	}
	function s(e, t) {
		let n = t[t.length - 1], r = i[a][0], o = i[a][1], s = 0, c = n.children.indexOf(e), l = !1, u = [];
		r.lastIndex = 0;
		let d = r.exec(e.value);
		for (; d;) {
			let n = d.index, i = {
				index: d.index,
				input: d.input,
				stack: [...t, e]
			}, a = o(...d, i);
			if (typeof a == "string" && (a = a.length > 0 ? {
				type: "text",
				value: a
			} : void 0), a === !1 ? r.lastIndex = n + 1 : (s !== n && u.push({
				type: "text",
				value: e.value.slice(s, n)
			}), Array.isArray(a) ? u.push(...a) : a && u.push(a), s = n + d[0].length, l = !0), !r.global) break;
			d = r.exec(e.value);
		}
		return l ? (s < e.value.length && u.push({
			type: "text",
			value: e.value.slice(s)
		}), n.children.splice(c, 1, ...u)) : u = [e], c + u.length;
	}
}
function yo(e) {
	let t = [];
	if (!Array.isArray(e)) throw TypeError("Expected find and replace tuple or list of tuples");
	let n = !e[0] || Array.isArray(e[0]) ? e : [e], r = -1;
	for (; ++r < n.length;) {
		let e = n[r];
		t.push([bo(e[0]), xo(e[1])]);
	}
	return t;
}
function bo(e) {
	return typeof e == "string" ? new RegExp(fe(e), "g") : e;
}
function xo(e) {
	return typeof e == "function" ? e : function() {
		return e;
	};
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/mdast-util-gfm-autolink-literal/2.0.1/253709983d58f4b2e3a16890a2ddc1facb66ce5f0b24c0acecafb22b2add4af0/node_modules/mdast-util-gfm-autolink-literal/lib/index.js
var So = "phrasing", Co = [
	"autolink",
	"link",
	"image",
	"label"
];
function wo() {
	return {
		transforms: [Mo],
		enter: {
			literalAutolink: Eo,
			literalAutolinkEmail: Do,
			literalAutolinkHttp: Do,
			literalAutolinkWww: Do
		},
		exit: {
			literalAutolink: jo,
			literalAutolinkEmail: Ao,
			literalAutolinkHttp: Oo,
			literalAutolinkWww: ko
		}
	};
}
function To() {
	return { unsafe: [
		{
			character: "@",
			before: "[+\\-.\\w]",
			after: "[\\-.\\w]",
			inConstruct: So,
			notInConstruct: Co
		},
		{
			character: ".",
			before: "[Ww]",
			after: "[\\-.\\w]",
			inConstruct: So,
			notInConstruct: Co
		},
		{
			character: ":",
			before: "[ps]",
			after: "\\/",
			inConstruct: So,
			notInConstruct: Co
		}
	] };
}
function Eo(e) {
	this.enter({
		type: "link",
		title: null,
		url: "",
		children: []
	}, e);
}
function Do(e) {
	this.config.enter.autolinkProtocol.call(this, e);
}
function Oo(e) {
	this.config.exit.autolinkProtocol.call(this, e);
}
function ko(e) {
	this.config.exit.data.call(this, e);
	let t = this.stack[this.stack.length - 1];
	t.type, t.url = "http://" + this.sliceSerialize(e);
}
function Ao(e) {
	this.config.exit.autolinkEmail.call(this, e);
}
function jo(e) {
	this.exit(e);
}
function Mo(e) {
	vo(e, [[/(https?:\/\/|www(?=\.))([-.\w]+)([^ \t\r\n]*)/gi, No], [/(?<=^|\s|\p{P}|\p{S})([-.\w+]+)@([-\w]+(?:\.[-\w]+)+)/gu, Po]], { ignore: ["link", "linkReference"] });
}
function No(e, t, n, r, i) {
	let a = "";
	if (!Lo(i) || (/^w/i.test(t) && (n = t + n, t = "", a = "http://"), !Fo(n))) return !1;
	let o = Io(n + r);
	if (!o[0]) return !1;
	let s = {
		type: "link",
		title: null,
		url: a + t + o[0],
		children: [{
			type: "text",
			value: t + o[0]
		}]
	};
	return o[1] ? [s, {
		type: "text",
		value: o[1]
	}] : s;
}
function Po(e, t, n, r) {
	return !Lo(r, !0) || /[-\d_]$/.test(n) ? !1 : {
		type: "link",
		title: null,
		url: "mailto:" + t + "@" + n,
		children: [{
			type: "text",
			value: t + "@" + n
		}]
	};
}
function Fo(e) {
	let t = e.split(".");
	return !(t.length < 2 || t[t.length - 1] && (/_/.test(t[t.length - 1]) || !/[a-zA-Z\d]/.test(t[t.length - 1])) || t[t.length - 2] && (/_/.test(t[t.length - 2]) || !/[a-zA-Z\d]/.test(t[t.length - 2])));
}
function Io(e) {
	let t = /[!"&'),.:;<>?\]}]+$/.exec(e);
	if (!t) return [e, void 0];
	e = e.slice(0, t.index);
	let n = t[0], r = n.indexOf(")"), i = v(e, "("), a = v(e, ")");
	for (; r !== -1 && i > a;) e += n.slice(0, r + 1), n = n.slice(r + 1), r = n.indexOf(")"), a++;
	return [e, n];
}
function Lo(e, t) {
	let n = e.input.charCodeAt(e.index - 1);
	return (e.index === 0 || T(n) || re(n)) && (!t || n !== 47);
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/mdast-util-gfm-footnote/2.1.0/c68b3efd3b641251b32406b42da019ff649376132a457071d05bc2068eea3f32/node_modules/mdast-util-gfm-footnote/lib/index.js
qo.peek = Ko;
function Ro() {
	this.buffer();
}
function zo(e) {
	this.enter({
		type: "footnoteReference",
		identifier: "",
		label: ""
	}, e);
}
function Bo() {
	this.buffer();
}
function Vo(e) {
	this.enter({
		type: "footnoteDefinition",
		identifier: "",
		label: "",
		children: []
	}, e);
}
function Ho(e) {
	let t = this.resume(), n = this.stack[this.stack.length - 1];
	n.type, n.identifier = y(this.sliceSerialize(e)).toLowerCase(), n.label = t;
}
function Uo(e) {
	this.exit(e);
}
function Wo(e) {
	let t = this.resume(), n = this.stack[this.stack.length - 1];
	n.type, n.identifier = y(this.sliceSerialize(e)).toLowerCase(), n.label = t;
}
function Go(e) {
	this.exit(e);
}
function Ko() {
	return "[";
}
function qo(e, t, n, r) {
	let i = n.createTracker(r), a = i.move("[^"), o = n.enter("footnoteReference"), s = n.enter("reference");
	return a += i.move(n.safe(n.associationId(e), {
		after: "]",
		before: a
	})), s(), o(), a += i.move("]"), a;
}
function Jo() {
	return {
		enter: {
			gfmFootnoteCallString: Ro,
			gfmFootnoteCall: zo,
			gfmFootnoteDefinitionLabelString: Bo,
			gfmFootnoteDefinition: Vo
		},
		exit: {
			gfmFootnoteCallString: Ho,
			gfmFootnoteCall: Uo,
			gfmFootnoteDefinitionLabelString: Wo,
			gfmFootnoteDefinition: Go
		}
	};
}
function Yo(e) {
	let t = !1;
	return e && e.firstLineBlank && (t = !0), {
		handlers: {
			footnoteDefinition: n,
			footnoteReference: qo
		},
		unsafe: [{
			character: "[",
			inConstruct: [
				"label",
				"phrasing",
				"reference"
			]
		}]
	};
	function n(e, n, r, i) {
		let a = r.createTracker(i), o = a.move("[^"), s = r.enter("footnoteDefinition"), c = r.enter("label");
		return o += a.move(r.safe(r.associationId(e), {
			before: o,
			after: "]"
		})), c(), o += a.move("]:"), e.children && e.children.length > 0 && (a.shift(4), o += a.move((t ? "\n" : " ") + r.indentLines(r.containerFlow(e, a.current()), t ? Zo : Xo))), s(), o;
	}
}
function Xo(e, t, n) {
	return t === 0 ? e : Zo(e, t, n);
}
function Zo(e, t, n) {
	return (n ? "" : "    ") + e;
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/mdast-util-gfm/3.1.0/dc215454bcdd9eadbdec1a600a9360808e553896fff298909b4def573e07fc11/node_modules/mdast-util-gfm/lib/index.js
function Qo() {
	return [
		wo(),
		Jo(),
		te(),
		ve(),
		de()
	];
}
function $o(e) {
	return { extensions: [
		To(),
		Yo(e),
		pe(),
		ye(e),
		oe()
	] };
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/micromark-extension-gfm-autolink-literal/2.1.0/25646f0ca472fed70ab10ebec494800588dfd9c95a0e3a643e72ebd0c815446b/node_modules/micromark-extension-gfm-autolink-literal/lib/syntax.js
var es = {
	tokenize: ms,
	partial: !0
}, ts = {
	tokenize: hs,
	partial: !0
}, ns = {
	tokenize: gs,
	partial: !0
}, rs = {
	tokenize: _s,
	partial: !0
}, is = {
	tokenize: vs,
	partial: !0
}, as = {
	name: "wwwAutolink",
	tokenize: fs,
	previous: ys
}, os = {
	name: "protocolAutolink",
	tokenize: ps,
	previous: bs
}, ss = {
	name: "emailAutolink",
	tokenize: ds,
	previous: xs
}, cs = {};
function ls() {
	return { text: cs };
}
for (var us = 48; us < 123;) cs[us] = ss, us++, us === 58 ? us = 65 : us === 91 && (us = 97);
cs[43] = ss, cs[45] = ss, cs[46] = ss, cs[95] = ss, cs[72] = [ss, os], cs[104] = [ss, os], cs[87] = [ss, as], cs[119] = [ss, as];
function ds(e, t, n) {
	let r = this, i, a;
	return o;
	function o(t) {
		return !Ss(t) || !xs.call(r, r.previous) || Cs(r.events) ? n(t) : (e.enter("literalAutolink"), e.enter("literalAutolinkEmail"), s(t));
	}
	function s(t) {
		return Ss(t) ? (e.consume(t), s) : t === 64 ? (e.consume(t), c) : n(t);
	}
	function c(t) {
		return t === 46 ? e.check(is, u, l)(t) : t === 45 || t === 95 || ee(t) ? (a = !0, e.consume(t), c) : u(t);
	}
	function l(t) {
		return e.consume(t), i = !0, c;
	}
	function u(o) {
		return a && i && S(r.previous) ? (e.exit("literalAutolinkEmail"), e.exit("literalAutolink"), t(o)) : n(o);
	}
}
function fs(e, t, n) {
	let r = this;
	return i;
	function i(t) {
		return t !== 87 && t !== 119 || !ys.call(r, r.previous) || Cs(r.events) ? n(t) : (e.enter("literalAutolink"), e.enter("literalAutolinkWww"), e.check(es, e.attempt(ts, e.attempt(ns, a), n), n)(t));
	}
	function a(n) {
		return e.exit("literalAutolinkWww"), e.exit("literalAutolink"), t(n);
	}
}
function ps(e, t, n) {
	let r = this, i = "", a = !1;
	return o;
	function o(t) {
		return (t === 72 || t === 104) && bs.call(r, r.previous) && !Cs(r.events) ? (e.enter("literalAutolink"), e.enter("literalAutolinkHttp"), i += String.fromCodePoint(t), e.consume(t), s) : n(t);
	}
	function s(t) {
		if (S(t) && i.length < 5) return i += String.fromCodePoint(t), e.consume(t), s;
		if (t === 58) {
			let n = i.toLowerCase();
			if (n === "http" || n === "https") return e.consume(t), c;
		}
		return n(t);
	}
	function c(t) {
		return t === 47 ? (e.consume(t), a ? l : (a = !0, c)) : n(t);
	}
	function l(t) {
		return t === null || ne(t) || x(t) || T(t) || re(t) ? n(t) : e.attempt(ts, e.attempt(ns, u), n)(t);
	}
	function u(n) {
		return e.exit("literalAutolinkHttp"), e.exit("literalAutolink"), t(n);
	}
}
function ms(e, t, n) {
	let r = 0;
	return i;
	function i(t) {
		return (t === 87 || t === 119) && r < 3 ? (r++, e.consume(t), i) : t === 46 && r === 3 ? (e.consume(t), a) : n(t);
	}
	function a(e) {
		return e === null ? n(e) : t(e);
	}
}
function hs(e, t, n) {
	let r, i, a;
	return o;
	function o(t) {
		return t === 46 || t === 95 ? e.check(rs, c, s)(t) : t === null || x(t) || T(t) || t !== 45 && re(t) ? c(t) : (a = !0, e.consume(t), o);
	}
	function s(t) {
		return t === 95 ? r = !0 : (i = r, r = void 0), e.consume(t), o;
	}
	function c(e) {
		return i || r || !a ? n(e) : t(e);
	}
}
function gs(e, t) {
	let n = 0, r = 0;
	return i;
	function i(o) {
		return o === 40 ? (n++, e.consume(o), i) : o === 41 && r < n ? a(o) : o === 33 || o === 34 || o === 38 || o === 39 || o === 41 || o === 42 || o === 44 || o === 46 || o === 58 || o === 59 || o === 60 || o === 63 || o === 93 || o === 95 || o === 126 ? e.check(rs, t, a)(o) : o === null || x(o) || T(o) ? t(o) : (e.consume(o), i);
	}
	function a(t) {
		return t === 41 && r++, e.consume(t), i;
	}
}
function _s(e, t, n) {
	return r;
	function r(o) {
		return o === 33 || o === 34 || o === 39 || o === 41 || o === 42 || o === 44 || o === 46 || o === 58 || o === 59 || o === 63 || o === 95 || o === 126 ? (e.consume(o), r) : o === 38 ? (e.consume(o), a) : o === 93 ? (e.consume(o), i) : o === 60 || o === null || x(o) || T(o) ? t(o) : n(o);
	}
	function i(e) {
		return e === null || e === 40 || e === 91 || x(e) || T(e) ? t(e) : r(e);
	}
	function a(e) {
		return S(e) ? o(e) : n(e);
	}
	function o(t) {
		return t === 59 ? (e.consume(t), r) : S(t) ? (e.consume(t), o) : n(t);
	}
}
function vs(e, t, n) {
	return r;
	function r(t) {
		return e.consume(t), i;
	}
	function i(e) {
		return ee(e) ? n(e) : t(e);
	}
}
function ys(e) {
	return e === null || e === 40 || e === 42 || e === 95 || e === 91 || e === 93 || e === 126 || x(e);
}
function bs(e) {
	return !S(e);
}
function xs(e) {
	return !(e === 47 || Ss(e));
}
function Ss(e) {
	return e === 43 || e === 45 || e === 46 || e === 95 || ee(e);
}
function Cs(e) {
	let t = e.length, n = !1;
	for (; t--;) {
		let r = e[t][1];
		if ((r.type === "labelLink" || r.type === "labelImage") && !r._balanced) {
			n = !0;
			break;
		}
		if (r._gfmAutolinkLiteralWalkedInto) {
			n = !1;
			break;
		}
	}
	return e.length > 0 && !n && (e[e.length - 1][1]._gfmAutolinkLiteralWalkedInto = !0), n;
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/micromark-util-sanitize-uri/2.0.1/640c60c1c8c5a3a6bb9e0f4ef20c289cd8e57f9c78da5a832b75c41d688f5d22/node_modules/micromark-util-sanitize-uri/index.js
function ws(e) {
	let t = [], n = -1, r = 0, i = 0;
	for (; ++n < e.length;) {
		let a = e.charCodeAt(n), o = "";
		if (a === 37 && ee(e.charCodeAt(n + 1)) && ee(e.charCodeAt(n + 2))) i = 2;
		else if (a < 128) /[!#$&-;=?-Z_a-z~]/.test(String.fromCharCode(a)) || (o = String.fromCharCode(a));
		else if (a > 55295 && a < 57344) {
			let t = e.charCodeAt(n + 1);
			a < 56320 && t > 56319 && t < 57344 ? (o = String.fromCharCode(a, t), i = 1) : o = "�";
		} else o = String.fromCharCode(a);
		o &&= (t.push(e.slice(r, n), encodeURIComponent(o)), r = n + i + 1, ""), i &&= (n += i, 0);
	}
	return t.join("") + e.slice(r);
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/micromark-extension-gfm-footnote/2.1.0/b44c8e84d93fa798639536629ef427f51491f682b55d6753baacd913323b1d0c/node_modules/micromark-extension-gfm-footnote/lib/syntax.js
var Ts = {
	tokenize: Ns,
	partial: !0
};
function Es() {
	return {
		document: { 91: {
			name: "gfmFootnoteDefinition",
			tokenize: As,
			continuation: { tokenize: js },
			exit: Ms
		} },
		text: {
			91: {
				name: "gfmFootnoteCall",
				tokenize: ks
			},
			93: {
				name: "gfmPotentialFootnoteCall",
				add: "after",
				tokenize: Ds,
				resolveTo: Os
			}
		}
	};
}
function Ds(e, t, n) {
	let r = this, i = r.events.length, a = r.parser.gfmFootnotes || (r.parser.gfmFootnotes = []), o;
	for (; i--;) {
		let e = r.events[i][1];
		if (e.type === "labelImage") {
			o = e;
			break;
		}
		if (e.type === "gfmFootnoteCall" || e.type === "labelLink" || e.type === "label" || e.type === "image" || e.type === "link") break;
	}
	return s;
	function s(i) {
		if (!o || !o._balanced) return n(i);
		let s = y(r.sliceSerialize({
			start: o.end,
			end: r.now()
		}));
		return s.codePointAt(0) !== 94 || !a.includes(s.slice(1)) ? n(i) : (e.enter("gfmFootnoteCallLabelMarker"), e.consume(i), e.exit("gfmFootnoteCallLabelMarker"), t(i));
	}
}
function Os(e, t) {
	let n = e.length;
	for (; n--;) if (e[n][1].type === "labelImage" && e[n][0] === "enter") {
		e[n][1];
		break;
	}
	e[n + 1][1].type = "data", e[n + 3][1].type = "gfmFootnoteCallLabelMarker";
	let r = {
		type: "gfmFootnoteCall",
		start: Object.assign({}, e[n + 3][1].start),
		end: Object.assign({}, e[e.length - 1][1].end)
	}, i = {
		type: "gfmFootnoteCallMarker",
		start: Object.assign({}, e[n + 3][1].end),
		end: Object.assign({}, e[n + 3][1].end)
	};
	i.end.column++, i.end.offset++, i.end._bufferIndex++;
	let a = {
		type: "gfmFootnoteCallString",
		start: Object.assign({}, i.end),
		end: Object.assign({}, e[e.length - 1][1].start)
	}, o = {
		type: "chunkString",
		contentType: "string",
		start: Object.assign({}, a.start),
		end: Object.assign({}, a.end)
	}, s = [
		e[n + 1],
		e[n + 2],
		[
			"enter",
			r,
			t
		],
		e[n + 3],
		e[n + 4],
		[
			"enter",
			i,
			t
		],
		[
			"exit",
			i,
			t
		],
		[
			"enter",
			a,
			t
		],
		[
			"enter",
			o,
			t
		],
		[
			"exit",
			o,
			t
		],
		[
			"exit",
			a,
			t
		],
		e[e.length - 2],
		e[e.length - 1],
		[
			"exit",
			r,
			t
		]
	];
	return e.splice(n, e.length - n + 1, ...s), e;
}
function ks(e, t, n) {
	let r = this, i = r.parser.gfmFootnotes || (r.parser.gfmFootnotes = []), a = 0, o;
	return s;
	function s(t) {
		return e.enter("gfmFootnoteCall"), e.enter("gfmFootnoteCallLabelMarker"), e.consume(t), e.exit("gfmFootnoteCallLabelMarker"), c;
	}
	function c(t) {
		return t === 94 ? (e.enter("gfmFootnoteCallMarker"), e.consume(t), e.exit("gfmFootnoteCallMarker"), e.enter("gfmFootnoteCallString"), e.enter("chunkString").contentType = "string", l) : n(t);
	}
	function l(s) {
		if (a > 999 || s === 93 && !o || s === null || s === 91 || x(s)) return n(s);
		if (s === 93) {
			e.exit("chunkString");
			let a = e.exit("gfmFootnoteCallString");
			return i.includes(y(r.sliceSerialize(a))) ? (e.enter("gfmFootnoteCallLabelMarker"), e.consume(s), e.exit("gfmFootnoteCallLabelMarker"), e.exit("gfmFootnoteCall"), t) : n(s);
		}
		return x(s) || (o = !0), a++, e.consume(s), s === 92 ? u : l;
	}
	function u(t) {
		return t === 91 || t === 92 || t === 93 ? (e.consume(t), a++, l) : l(t);
	}
}
function As(e, t, n) {
	let r = this, i = r.parser.gfmFootnotes || (r.parser.gfmFootnotes = []), a, o = 0, s;
	return c;
	function c(t) {
		return e.enter("gfmFootnoteDefinition")._container = !0, e.enter("gfmFootnoteDefinitionLabel"), e.enter("gfmFootnoteDefinitionLabelMarker"), e.consume(t), e.exit("gfmFootnoteDefinitionLabelMarker"), l;
	}
	function l(t) {
		return t === 94 ? (e.enter("gfmFootnoteDefinitionMarker"), e.consume(t), e.exit("gfmFootnoteDefinitionMarker"), e.enter("gfmFootnoteDefinitionLabelString"), e.enter("chunkString").contentType = "string", u) : n(t);
	}
	function u(t) {
		if (o > 999 || t === 93 && !s || t === null || t === 91 || x(t)) return n(t);
		if (t === 93) {
			e.exit("chunkString");
			let n = e.exit("gfmFootnoteDefinitionLabelString");
			return a = y(r.sliceSerialize(n)), e.enter("gfmFootnoteDefinitionLabelMarker"), e.consume(t), e.exit("gfmFootnoteDefinitionLabelMarker"), e.exit("gfmFootnoteDefinitionLabel"), f;
		}
		return x(t) || (s = !0), o++, e.consume(t), t === 92 ? d : u;
	}
	function d(t) {
		return t === 91 || t === 92 || t === 93 ? (e.consume(t), o++, u) : u(t);
	}
	function f(t) {
		return t === 58 ? (e.enter("definitionMarker"), e.consume(t), e.exit("definitionMarker"), i.includes(a) || i.push(a), ue(e, p, "gfmFootnoteDefinitionWhitespace")) : n(t);
	}
	function p(e) {
		return t(e);
	}
}
function js(e, t, n) {
	return e.check(le, t, e.attempt(Ts, t, n));
}
function Ms(e) {
	e.exit("gfmFootnoteDefinition");
}
function Ns(e, t, n) {
	let r = this;
	return ue(e, i, "gfmFootnoteDefinitionIndent", 5);
	function i(e) {
		let i = r.events[r.events.length - 1];
		return i && i[1].type === "gfmFootnoteDefinitionIndent" && i[2].sliceSerialize(i[1], !0).length === 4 ? t(e) : n(e);
	}
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/micromark-extension-gfm/3.0.0/ecc0d38f5a0d39e6498936ab24fa7414f69a55a425511808f8b7c5359e6d2fdf/node_modules/micromark-extension-gfm/index.js
function Ps(e) {
	return he([
		ls(),
		Es(),
		me(e),
		ce(),
		w()
	]);
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/remark-gfm/4.0.1/036151e017055832cf734f526db4e20cc63f4ce687d1103929754f287e671818/node_modules/remark-gfm/lib/index.js
var Fs = {};
function Is(e) {
	let t = this, n = e || Fs, r = t.data(), i = r.micromarkExtensions ||= [], a = r.fromMarkdownExtensions ||= [], o = r.toMarkdownExtensions ||= [];
	i.push(Ps(n)), a.push(Qo()), o.push($o(n));
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/remend/1.3.0/53664cbff19ab03c0fc5214f7a526d3f90357fb5926f9315e936a7eeb6fa86ef/node_modules/remend/dist/index.js
var Ls = Object.defineProperty, Rs = Object.defineProperties, zs = Object.getOwnPropertyDescriptors, Bs = Object.getOwnPropertySymbols, Vs = Object.prototype.hasOwnProperty, Hs = Object.prototype.propertyIsEnumerable, Us = (e, t, n) => t in e ? Ls(e, t, {
	enumerable: !0,
	configurable: !0,
	writable: !0,
	value: n
}) : e[t] = n, Ws = (e, t) => {
	for (var n in t ||= {}) Vs.call(t, n) && Us(e, n, t[n]);
	if (Bs) for (var n of Bs(t)) Hs.call(t, n) && Us(e, n, t[n]);
	return e;
}, Gs = (e, t) => Rs(e, zs(t)), H = (e, t) => {
	let n = !1, r = !1;
	for (let i = 0; i < t; i += 1) e[i] === "\\" && i + 1 < e.length && e[i + 1] === "`" ? i += 1 : e.substring(i, i + 3) === "```" ? (r = !r, i += 2) : !r && e[i] === "`" && (n = !n);
	return n || r;
}, Ks = (e, t) => {
	let n = e.substring(t, t + 3) === "```", r = t > 0 && e.substring(t - 1, t + 2) === "```", i = t > 1 && e.substring(t - 2, t + 1) === "```";
	return n || r || i;
}, qs = (e) => {
	let t = 0;
	for (let n = 0; n < e.length; n += 1) e[n] === "\\" && n + 1 < e.length && e[n + 1] === "`" ? n += 1 : e[n] === "`" && !Ks(e, n) && (t += 1);
	return t;
}, Js = (e, t) => {
	let n = !1, r = !1, i = -1;
	for (let a = 0; a < e.length; a += 1) if (e[a] === "\\" && a + 1 < e.length && e[a + 1] === "`") a += 1;
	else if (e.substring(a, a + 3) === "```") r = !r, a += 2;
	else if (!r && e[a] === "`") {
		if (n) {
			if (i < t && t < a) return !0;
			n = !1, i = -1;
		} else n = !0, i = a;
	}
	return !1;
}, Ys = /^(\s*(?:[-*+]|\d+[.)]) +)>(=?\s*[$]?\d)/gm, Xs = (e) => !e || typeof e != "string" || !e.includes(">") ? e : e.replace(Ys, (t, n, r, i) => H(e, i) ? t : `${n}\\>${r}`), Zs = /(\*\*)([^*]*\*?)$/, Qs = /(__)([^_]*?)$/, $s = /(\*\*\*)([^*]*?)$/, ec = /(\*)([^*]*?)$/, tc = /(_)([^_]*?)$/, nc = /(`)([^`]*?)$/, rc = /(~~)([^~]*?)$/, ic = /^[\s_~*`]*$/, ac = /^[\s]*[-*+][\s]+$/, oc = /[\p{L}\p{N}_]/u, sc = /^```[^`\n]*```?$/, cc = /^\*{4,}$/, lc = /(__)([^_]+)_$/, uc = /(~~)([^~]+)~$/, dc = /~~/g, fc = (e) => {
	if (!e) return !1;
	let t = e.charCodeAt(0);
	return t >= 48 && t <= 57 || t >= 65 && t <= 90 || t >= 97 && t <= 122 || t === 95 || oc.test(e);
}, pc = (e, t) => {
	let n = 1;
	for (let r = t - 1; r >= 0; --r) if (e[r] === "]") n += 1;
	else if (e[r] === "[" && (--n, n === 0)) return r;
	return -1;
}, mc = (e, t) => {
	let n = 1;
	for (let r = t + 1; r < e.length; r += 1) if (e[r] === "[") n += 1;
	else if (e[r] === "]" && (--n, n === 0)) return r;
	return -1;
}, hc = (e, t) => {
	let n = !1, r = !1;
	for (let i = 0; i < e.length && i < t; i += 1) e[i] === "\\" && e[i + 1] === "$" ? i += 1 : e[i] === "$" && (e[i + 1] === "$" ? (r = !r, i += 1, n = !1) : r || (n = !n));
	return n || r;
}, gc = (e, t) => {
	for (let n = t; n < e.length; n += 1) {
		if (e[n] === ")") return !0;
		if (e[n] === "\n") return !1;
	}
	return !1;
}, _c = (e, t) => {
	for (let n = t - 1; n >= 0; --n) {
		if (e[n] === ")") return !1;
		if (e[n] === "(") return n > 0 && e[n - 1] === "]" && gc(e, t);
		if (e[n] === "\n") return !1;
	}
	return !1;
}, vc = (e, t) => {
	for (let n = t - 1; n >= 0; --n) {
		if (e[n] === ">") return !1;
		if (e[n] === "<") {
			let t = n + 1 < e.length ? e[n + 1] : "";
			return t >= "a" && t <= "z" || t >= "A" && t <= "Z" || t === "/";
		}
		if (e[n] === "\n") return !1;
	}
	return !1;
}, yc = (e, t, n) => {
	let r = 0;
	for (let n = t - 1; n >= 0; --n) if (e[n] === "\n") {
		r = n + 1;
		break;
	}
	let i = e.length;
	for (let n = t; n < e.length; n += 1) if (e[n] === "\n") {
		i = n;
		break;
	}
	let a = e.substring(r, i), o = 0, s = !1;
	for (let e of a) if (e === n) o += 1;
	else if (e !== " " && e !== "	") {
		s = !0;
		break;
	}
	return o >= 3 && !s;
}, bc = (e, t, n, r) => n === "\\" || e.includes("$") && hc(e, t) ? !0 : n !== "*" && r === "*" ? (t < e.length - 2 ? e[t + 2] : "") !== "*" : !!(n === "*" || n && r && fc(n) && fc(r) || (!n || n === " " || n === "	" || n === "\n") && (!r || r === " " || r === "	" || r === "\n")), xc = (e) => {
	let t = 0, n = !1, r = e.length;
	for (let i = 0; i < r; i += 1) {
		if (e[i] === "`" && i + 2 < r && e[i + 1] === "`" && e[i + 2] === "`") {
			n = !n, i += 2;
			continue;
		}
		if (n || e[i] !== "*") continue;
		let a = i > 0 ? e[i - 1] : "", o = i < r - 1 ? e[i + 1] : "";
		bc(e, i, a, o) || (t += 1);
	}
	return t;
}, Sc = (e, t, n, r) => !!(n === "\\" || e.includes("$") && hc(e, t) || _c(e, t) || vc(e, t) || n === "_" || r === "_" || n && r && fc(n) && fc(r)), Cc = (e) => {
	let t = 0, n = !1, r = e.length;
	for (let i = 0; i < r; i += 1) {
		if (e[i] === "`" && i + 2 < r && e[i + 1] === "`" && e[i + 2] === "`") {
			n = !n, i += 2;
			continue;
		}
		if (n || e[i] !== "_") continue;
		let a = i > 0 ? e[i - 1] : "", o = i < r - 1 ? e[i + 1] : "";
		Sc(e, i, a, o) || (t += 1);
	}
	return t;
}, wc = (e) => {
	let t = 0, n = 0, r = !1;
	for (let i = 0; i < e.length; i += 1) e[i] === "`" && i + 2 < e.length && e[i + 1] === "`" && e[i + 2] === "`" ? (n >= 3 && (t += Math.floor(n / 3)), n = 0, r = !r, i += 2) : r || (e[i] === "*" ? n += 1 : (n >= 3 && (t += Math.floor(n / 3)), n = 0));
	return n >= 3 && (t += Math.floor(n / 3)), t;
}, Tc = (e) => {
	let t = 0, n = !1;
	for (let r = 0; r < e.length; r += 1) e[r] === "`" && r + 2 < e.length && e[r + 1] === "`" && e[r + 2] === "`" ? (n = !n, r += 2) : n || e[r] === "*" && r + 1 < e.length && e[r + 1] === "*" && (t += 1, r += 1);
	return t;
}, Ec = (e) => {
	let t = 0, n = !1;
	for (let r = 0; r < e.length; r += 1) e[r] === "`" && r + 2 < e.length && e[r + 1] === "`" && e[r + 2] === "`" ? (n = !n, r += 2) : n || e[r] === "_" && r + 1 < e.length && e[r + 1] === "_" && (t += 1, r += 1);
	return t;
}, Dc = (e, t, n) => {
	if (!t || ic.test(t)) return !0;
	let r = e.substring(0, n).lastIndexOf("\n"), i = r === -1 ? 0 : r + 1, a = e.substring(i, n);
	return ac.test(a) && t.includes("\n") ? !0 : yc(e, n, "*");
}, Oc = (e) => {
	let t = e.match(Zs);
	if (!t) return e;
	let n = t[2], r = e.lastIndexOf(t[1]);
	return H(e, r) || Js(e, r) || Dc(e, n, r) ? e : Tc(e) % 2 == 1 ? n.endsWith("*") ? `${e}*` : `${e}**` : e;
}, kc = (e, t, n) => {
	if (!t || ic.test(t)) return !0;
	let r = e.substring(0, n).lastIndexOf("\n"), i = r === -1 ? 0 : r + 1, a = e.substring(i, n);
	return ac.test(a) && t.includes("\n") ? !0 : yc(e, n, "_");
}, Ac = (e) => {
	let t = e.match(Qs);
	if (!t) {
		let t = e.match(lc);
		if (t) {
			let n = e.lastIndexOf(t[1]);
			if (!(H(e, n) || Js(e, n)) && Ec(e) % 2 == 1) return `${e}_`;
		}
		return e;
	}
	let n = t[2], r = e.lastIndexOf(t[1]);
	return H(e, r) || Js(e, r) || kc(e, n, r) ? e : Ec(e) % 2 == 1 ? `${e}__` : e;
}, jc = (e) => {
	let t = !1;
	for (let n = 0; n < e.length; n += 1) if (e[n] === "`" && n + 2 < e.length && e[n + 1] === "`" && e[n + 2] === "`") t = !t, n += 2;
	else if (!t && e[n] === "*" && e[n - 1] !== "*" && e[n + 1] !== "*" && e[n - 1] !== "\\" && !hc(e, n)) {
		let t = n > 0 ? e[n - 1] : "", r = n < e.length - 1 ? e[n + 1] : "";
		if ((!t || t === " " || t === "	" || t === "\n") && (!r || r === " " || r === "	" || r === "\n") || t && r && fc(t) && fc(r)) continue;
		return n;
	}
	return -1;
}, Mc = (e) => {
	if (!e.match(ec)) return e;
	let t = jc(e);
	if (t === -1 || H(e, t) || Js(e, t)) return e;
	let n = e.substring(t + 1);
	return !n || ic.test(n) ? e : xc(e) % 2 == 1 ? `${e}*` : e;
}, Nc = (e) => {
	let t = !1;
	for (let n = 0; n < e.length; n += 1) if (e[n] === "`" && n + 2 < e.length && e[n + 1] === "`" && e[n + 2] === "`") t = !t, n += 2;
	else if (!t && e[n] === "_" && e[n - 1] !== "_" && e[n + 1] !== "_" && e[n - 1] !== "\\" && !hc(e, n) && !_c(e, n)) {
		let t = n > 0 ? e[n - 1] : "", r = n < e.length - 1 ? e[n + 1] : "";
		if (t && r && fc(t) && fc(r)) continue;
		return n;
	}
	return -1;
}, Pc = (e) => {
	let t = e.length;
	for (; t > 0 && e[t - 1] === "\n";) --t;
	return t < e.length ? `${e.slice(0, t)}_${e.slice(t)}` : `${e}_`;
}, Fc = (e) => {
	if (!e.endsWith("**")) return null;
	let t = e.slice(0, -2);
	if (Tc(t) % 2 != 1) return null;
	let n = t.indexOf("**"), r = Nc(t);
	return n !== -1 && r !== -1 && n < r ? `${t}_**` : null;
}, Ic = (e) => {
	if (!e.match(tc)) return e;
	let t = Nc(e);
	if (t === -1) return e;
	let n = e.substring(t + 1);
	if (!n || ic.test(n) || H(e, t) || Js(e, t)) return e;
	if (Cc(e) % 2 == 1) {
		let t = Fc(e);
		return t === null ? Pc(e) : t;
	}
	return e;
}, Lc = (e) => {
	let t = Tc(e), n = xc(e);
	return t % 2 == 0 && n % 2 == 0;
}, Rc = (e, t, n) => !t || ic.test(t) || H(e, n) || Js(e, n) ? !0 : yc(e, n, "*"), zc = (e) => {
	if (cc.test(e)) return e;
	let t = e.match($s);
	if (!t) return e;
	let n = t[2];
	return Rc(e, n, e.lastIndexOf(t[1])) ? e : wc(e) % 2 == 1 ? Lc(e) ? e : `${e}***` : e;
}, Bc = /<[a-zA-Z/][^>]*$/, Vc = (e) => {
	let t = e.match(Bc);
	return !t || t.index === void 0 || H(e, t.index) ? e : e.substring(0, t.index).trimEnd();
}, Hc = (e) => !e.match(sc) || e.includes("\n") ? null : e.endsWith("``") && !e.endsWith("```") ? `${e}\`` : e, Uc = (e) => (e.match(/```/g) || []).length % 2 == 1, Wc = (e) => {
	let t = Hc(e);
	if (t !== null) return t;
	let n = e.match(nc);
	if (n && !Uc(e)) {
		let t = n[2];
		if (!t || ic.test(t)) return e;
		if (qs(e) % 2 == 1) return `${e}\``;
	}
	return e;
}, Gc = (e, t) => t >= 2 && e.substring(t - 2, t + 1) === "```" || t >= 1 && e.substring(t - 1, t + 2) === "```" || t <= e.length - 3 && e.substring(t, t + 3) === "```", Kc = (e) => {
	let t = 0, n = !1;
	for (let r = 0; r < e.length - 1; r += 1) e[r] === "`" && !Gc(e, r) && (n = !n), !n && e[r] === "$" && e[r + 1] === "$" && (t += 1, r += 1);
	return t;
}, qc = (e) => {
	let t = 0, n = !1;
	for (let r = 0; r < e.length; r += 1) e[r] === "\\" ? r += 1 : e[r] === "`" && !Gc(e, r) ? n = !n : !n && e[r] === "$" && (r + 1 < e.length && e[r + 1] === "$" ? r += 1 : t += 1);
	return t;
}, Jc = (e) => {
	if (e.endsWith("$") && !e.endsWith("$$")) return `${e}$`;
	let t = e.indexOf("$$");
	return t !== -1 && e.indexOf("\n", t) !== -1 && !e.endsWith("\n") ? `${e}
$$` : `${e}$$`;
}, Yc = (e) => Kc(e) % 2 == 0 ? e : Jc(e), Xc = (e) => qc(e) % 2 == 1 ? `${e}$` : e, Zc = (e, t, n) => {
	if (e.substring(t + 2).includes(")")) return null;
	let r = pc(e, t);
	if (r === -1 || H(e, r)) return null;
	let i = r > 0 && e[r - 1] === "!", a = i ? r - 1 : r, o = e.substring(0, a);
	if (i) return o;
	let s = e.substring(r + 1, t);
	return n === "text-only" ? `${o}${s}` : `${o}[${s}](streamdown:incomplete-link)`;
}, Qc = (e, t) => {
	for (let n = 0; n < t; n++) if (e[n] === "[" && !H(e, n)) {
		if (n > 0 && e[n - 1] === "!") continue;
		let t = mc(e, n);
		if (t === -1) return n;
		if (t + 1 < e.length && e[t + 1] === "(") {
			let r = e.indexOf(")", t + 2);
			r !== -1 && (n = r);
		}
	}
	return t;
}, $c = (e, t, n) => {
	let r = t > 0 && e[t - 1] === "!", i = r ? t - 1 : t;
	if (!e.substring(t + 1).includes("]") || mc(e, t) === -1) {
		let a = e.substring(0, i);
		if (r) return a;
		if (n === "text-only") {
			let n = Qc(e, t);
			return e.substring(0, n) + e.substring(n + 1);
		}
		return `${e}](streamdown:incomplete-link)`;
	}
	return null;
}, el = (e, t = "protocol") => {
	let n = e.lastIndexOf("](");
	if (n !== -1 && !H(e, n)) {
		let r = Zc(e, n, t);
		if (r !== null) return r;
	}
	for (let n = e.length - 1; n >= 0; --n) if (e[n] === "[" && !H(e, n)) {
		let r = $c(e, n, t);
		if (r !== null) return r;
	}
	return e;
}, tl = /^-{1,2}$/, nl = /^[\s]*-{1,2}[\s]+$/, rl = /^={1,2}$/, il = /^[\s]*={1,2}[\s]+$/, al = (e) => {
	if (!e || typeof e != "string") return e;
	let t = e.lastIndexOf("\n");
	if (t === -1) return e;
	let n = e.substring(t + 1), r = e.substring(0, t), i = n.trim();
	if (tl.test(i) && !n.match(nl)) {
		let t = r.split("\n").at(-1);
		if (t && t.trim().length > 0) return `${e}\u200B`;
	}
	if (rl.test(i) && !n.match(il)) {
		let t = r.split("\n").at(-1);
		if (t && t.trim().length > 0) return `${e}\u200B`;
	}
	return e;
}, ol = /* @__PURE__ */ RegExp("(?<=[\\p{L}\\p{N}_])~(?!~)(?=[\\p{L}\\p{N}_])", "gu"), sl = (e) => !e || typeof e != "string" || !e.includes("~") ? e : e.replace(ol, (t, n) => H(e, n) ? t : "\\~"), cl = (e) => {
	let t = e.match(rc);
	if (t) {
		let n = t[2];
		if (!n || ic.test(n)) return e;
		let r = e.lastIndexOf(t[1]);
		if (H(e, r) || Js(e, r)) return e;
		if (e.match(dc)?.length % 2 == 1) return `${e}~~`;
	} else {
		let t = e.match(uc);
		if (t) {
			let n = e.lastIndexOf(t[0].slice(0, 2));
			if (H(e, n) || Js(e, n)) return e;
			if (e.match(dc)?.length % 2 == 1) return `${e}~`;
		}
	}
	return e;
}, ll = (e) => e !== !1, ul = (e) => e === !0, U = {
	SINGLE_TILDE: 0,
	COMPARISON_OPERATORS: 5,
	HTML_TAGS: 10,
	SETEXT_HEADINGS: 15,
	LINKS: 20,
	BOLD_ITALIC: 30,
	BOLD: 35,
	ITALIC_DOUBLE_UNDERSCORE: 40,
	ITALIC_SINGLE_ASTERISK: 41,
	ITALIC_SINGLE_UNDERSCORE: 42,
	INLINE_CODE: 50,
	STRIKETHROUGH: 60,
	KATEX: 70,
	INLINE_KATEX: 75,
	DEFAULT: 100
}, dl = [
	{
		handler: {
			name: "singleTilde",
			handle: sl,
			priority: U.SINGLE_TILDE
		},
		optionKey: "singleTilde"
	},
	{
		handler: {
			name: "comparisonOperators",
			handle: Xs,
			priority: U.COMPARISON_OPERATORS
		},
		optionKey: "comparisonOperators"
	},
	{
		handler: {
			name: "htmlTags",
			handle: Vc,
			priority: U.HTML_TAGS
		},
		optionKey: "htmlTags"
	},
	{
		handler: {
			name: "setextHeadings",
			handle: al,
			priority: U.SETEXT_HEADINGS
		},
		optionKey: "setextHeadings"
	},
	{
		handler: {
			name: "links",
			handle: el,
			priority: U.LINKS
		},
		optionKey: "links",
		earlyReturn: (e) => e.endsWith("](streamdown:incomplete-link)")
	},
	{
		handler: {
			name: "boldItalic",
			handle: zc,
			priority: U.BOLD_ITALIC
		},
		optionKey: "boldItalic"
	},
	{
		handler: {
			name: "bold",
			handle: Oc,
			priority: U.BOLD
		},
		optionKey: "bold"
	},
	{
		handler: {
			name: "italicDoubleUnderscore",
			handle: Ac,
			priority: U.ITALIC_DOUBLE_UNDERSCORE
		},
		optionKey: "italic"
	},
	{
		handler: {
			name: "italicSingleAsterisk",
			handle: Mc,
			priority: U.ITALIC_SINGLE_ASTERISK
		},
		optionKey: "italic"
	},
	{
		handler: {
			name: "italicSingleUnderscore",
			handle: Ic,
			priority: U.ITALIC_SINGLE_UNDERSCORE
		},
		optionKey: "italic"
	},
	{
		handler: {
			name: "inlineCode",
			handle: Wc,
			priority: U.INLINE_CODE
		},
		optionKey: "inlineCode"
	},
	{
		handler: {
			name: "strikethrough",
			handle: cl,
			priority: U.STRIKETHROUGH
		},
		optionKey: "strikethrough"
	},
	{
		handler: {
			name: "katex",
			handle: Yc,
			priority: U.KATEX
		},
		optionKey: "katex"
	},
	{
		handler: {
			name: "inlineKatex",
			handle: Xc,
			priority: U.INLINE_KATEX
		},
		optionKey: "inlineKatex"
	}
], fl = (e) => {
	let t = e?.linkMode ?? "protocol";
	return dl.filter(({ handler: t, optionKey: n }) => t.name === "links" ? ll(e?.links) || ll(e?.images) : t.name === "inlineKatex" ? ul(e?.inlineKatex) : ll(e?.[n])).map(({ handler: e, earlyReturn: n }) => e.name === "links" ? {
		handler: Gs(Ws({}, e), { handle: (e) => el(e, t) }),
		earlyReturn: t === "protocol" ? n : void 0
	} : {
		handler: e,
		earlyReturn: n
	});
}, pl = (e, t) => {
	if (!e || typeof e != "string") return e;
	let n = e.endsWith(" ") && !e.endsWith("  ") ? e.slice(0, -1) : e, r = fl(t), i = (t?.handlers ?? []).map((e) => ({
		handler: Gs(Ws({}, e), { priority: e.priority ?? U.DEFAULT }),
		earlyReturn: void 0
	})), a = [...r, ...i].sort((e, t) => (e.handler.priority ?? 0) - (t.handler.priority ?? 0));
	for (let { handler: e, earlyReturn: t } of a) if (n = e.handle(n), t != null && t(n)) return n;
	return n;
}, ml = /* @__PURE__ */ t(((e, t) => {
	var n = /\/\*[^*]*\*+([^/*][^*]*\*+)*\//g, r = /\n/g, i = /^\s*/, a = /^(\*?[-#/*\\\w]+(\[[0-9a-z_-]+\])?)\s*/, o = /^:\s*/, s = /^((?:'(?:\\'|.)*?'|"(?:\\"|.)*?"|\([^)]*?\)|[^};])+)/, c = /^[;\s]*/, l = /^\s+|\s+$/g;
	function u(e, t) {
		if (typeof e != "string") throw TypeError("First argument must be a string");
		if (!e) return [];
		t ||= {};
		var l = 1, u = 1;
		function f(e) {
			var t = e.match(r);
			t && (l += t.length);
			var n = e.lastIndexOf("\n");
			u = ~n ? e.length - n : u + e.length;
		}
		function p() {
			var e = {
				line: l,
				column: u
			};
			return function(t) {
				return t.position = new m(e), _(), t;
			};
		}
		function m(e) {
			this.start = e, this.end = {
				line: l,
				column: u
			}, this.source = t.source;
		}
		m.prototype.content = e;
		function h(n) {
			var r = /* @__PURE__ */ Error(t.source + ":" + l + ":" + u + ": " + n);
			if (r.reason = n, r.filename = t.source, r.line = l, r.column = u, r.source = e, !t.silent) throw r;
		}
		function g(t) {
			var n = t.exec(e);
			if (n) {
				var r = n[0];
				return f(r), e = e.slice(r.length), n;
			}
		}
		function _() {
			g(i);
		}
		function v(e) {
			var t;
			for (e ||= []; t = y();) t !== !1 && e.push(t);
			return e;
		}
		function y() {
			var t = p();
			if (e.charAt(0) == "/" && e.charAt(1) == "*") {
				for (var n = 2; e.charAt(n) != "" && (e.charAt(n) != "*" || e.charAt(n + 1) != "/");) ++n;
				if (n += 2, e.charAt(n - 1) === "") return h("End of comment missing");
				var r = e.slice(2, n - 2);
				return u += 2, f(r), e = e.slice(n), u += 2, t({
					type: "comment",
					comment: r
				});
			}
		}
		function b() {
			var e = p(), t = g(a);
			if (t) {
				if (y(), !g(o)) return h("property missing ':'");
				var r = g(s), i = e({
					type: "declaration",
					property: d(t[0].replace(n, "")),
					value: r ? d(r[0].replace(n, "")) : ""
				});
				return g(c), i;
			}
		}
		function x() {
			var e = [];
			v(e);
			for (var t; t = b();) t !== !1 && (e.push(t), v(e));
			return e;
		}
		return _(), x();
	}
	function d(e) {
		return e ? e.replace(l, "") : "";
	}
	t.exports = u;
})), hl = /* @__PURE__ */ t(((e, t, n) => {
	var r = n && n.__importDefault || function(e) {
		return e && e.__esModule ? e : { default: e };
	};
	Object.defineProperty(e, "__esModule", { value: !0 }), e.default = a;
	var i = r(ml());
	function a(e, t) {
		let n = null;
		if (!e || typeof e != "string") return n;
		let r = (0, i.default)(e), a = typeof t == "function";
		return r.forEach((e) => {
			if (e.type !== "declaration") return;
			let { property: r, value: i } = e;
			a ? t(r, i, e) : i && (n ||= {}, n[r] = i);
		}), n;
	}
})), gl = /* @__PURE__ */ t(((e) => {
	Object.defineProperty(e, "__esModule", { value: !0 }), e.camelCase = void 0;
	var t = /^--[a-zA-Z0-9_-]+$/, n = /-([a-z])/g, r = /^[^-]+$/, i = /^-(webkit|moz|ms|o|khtml)-/, a = /^-(ms)-/, o = function(e) {
		return !e || r.test(e) || t.test(e);
	}, s = function(e, t) {
		return t.toUpperCase();
	}, c = function(e, t) {
		return `${t}-`;
	};
	e.camelCase = function(e, t) {
		return t === void 0 && (t = {}), o(e) ? e : (e = e.toLowerCase(), e = t.reactCompat ? e.replace(a, c) : e.replace(i, c), e.replace(n, s));
	};
})), _l = /* @__PURE__ */ e((/* @__PURE__ */ t(((e, t, n) => {
	var r = (n && n.__importDefault || function(e) {
		return e && e.__esModule ? e : { default: e };
	})(hl()), i = gl();
	function a(e, t) {
		var n = {};
		return !e || typeof e != "string" || (0, r.default)(e, function(e, r) {
			e && r && (n[(0, i.camelCase)(e, t)] = r);
		}), n;
	}
	a.default = a, t.exports = a;
})))(), 1), vl = {}.hasOwnProperty, yl = /* @__PURE__ */ new Map(), bl = /[A-Z]/g, xl = /* @__PURE__ */ new Set([
	"table",
	"tbody",
	"thead",
	"tfoot",
	"tr"
]), Sl = /* @__PURE__ */ new Set(["td", "th"]), Cl = "https://github.com/syntax-tree/hast-util-to-jsx-runtime";
function wl(e, t) {
	if (!t || t.Fragment === void 0) throw TypeError("Expected `Fragment` in options");
	let n = t.filePath || void 0, r;
	if (t.development) {
		if (typeof t.jsxDEV != "function") throw TypeError("Expected `jsxDEV` in options when `development: true`");
		r = Fl(n, t.jsxDEV);
	} else {
		if (typeof t.jsx != "function") throw TypeError("Expected `jsx` in production options");
		if (typeof t.jsxs != "function") throw TypeError("Expected `jsxs` in production options");
		r = Pl(n, t.jsx, t.jsxs);
	}
	let i = {
		Fragment: t.Fragment,
		ancestors: [],
		components: t.components || {},
		create: r,
		elementAttributeNameCase: t.elementAttributeNameCase || "react",
		evaluater: t.createEvaluater ? t.createEvaluater() : void 0,
		filePath: n,
		ignoreInvalidStyle: t.ignoreInvalidStyle || !1,
		passKeys: t.passKeys !== !1,
		passNode: t.passNode || !1,
		schema: t.space === "svg" ? m : f,
		stylePropertyNameCase: t.stylePropertyNameCase || "dom",
		tableCellAlignToStyle: t.tableCellAlignToStyle !== !1
	}, a = Tl(i, e, void 0);
	return a && typeof a != "string" ? a : i.create(e, i.Fragment, { children: a || void 0 }, void 0);
}
function Tl(e, t, n) {
	if (t.type === "element") return El(e, t, n);
	if (t.type === "mdxFlowExpression" || t.type === "mdxTextExpression") return Dl(e, t);
	if (t.type === "mdxJsxFlowElement" || t.type === "mdxJsxTextElement") return kl(e, t, n);
	if (t.type === "mdxjsEsm") return Ol(e, t);
	if (t.type === "root") return Al(e, t, n);
	if (t.type === "text") return jl(e, t);
}
function El(e, t, n) {
	let r = e.schema, i = r;
	t.tagName.toLowerCase() === "svg" && r.space === "html" && (i = m, e.schema = i), e.ancestors.push(t);
	let a = Vl(e, t.tagName, !1), o = Il(e, t), s = Rl(e, t);
	return xl.has(t.tagName) && (s = s.filter(function(e) {
		return typeof e != "string" || !h(e);
	})), Ml(e, o, a, t), Nl(o, s), e.ancestors.pop(), e.schema = r, e.create(t, a, o, n);
}
function Dl(e, t) {
	if (t.data && t.data.estree && e.evaluater) {
		let n = t.data.estree.body[0];
		return n.type, e.evaluater.evaluateExpression(n.expression);
	}
	Hl(e, t.position);
}
function Ol(e, t) {
	if (t.data && t.data.estree && e.evaluater) return e.evaluater.evaluateProgram(t.data.estree);
	Hl(e, t.position);
}
function kl(e, t, n) {
	let r = e.schema, i = r;
	t.name === "svg" && r.space === "html" && (i = m, e.schema = i), e.ancestors.push(t);
	let a = t.name === null ? e.Fragment : Vl(e, t.name, !0), o = Ll(e, t), s = Rl(e, t);
	return Ml(e, o, a, t), Nl(o, s), e.ancestors.pop(), e.schema = r, e.create(t, a, o, n);
}
function Al(e, t, n) {
	let r = {};
	return Nl(r, Rl(e, t)), e.create(t, e.Fragment, r, n);
}
function jl(e, t) {
	return t.value;
}
function Ml(e, t, n, r) {
	typeof n != "string" && n !== e.Fragment && e.passNode && (t.node = r);
}
function Nl(e, t) {
	if (t.length > 0) {
		let n = t.length > 1 ? t : t[0];
		n && (e.children = n);
	}
}
function Pl(e, t, n) {
	return r;
	function r(e, r, i, a) {
		let o = Array.isArray(i.children) ? n : t;
		return a ? o(r, i, a) : o(r, i);
	}
}
function Fl(e, t) {
	return n;
	function n(n, r, i, a) {
		let o = Array.isArray(i.children), s = Oa(n);
		return t(r, i, a, o, {
			columnNumber: s ? s.column - 1 : void 0,
			fileName: e,
			lineNumber: s ? s.line : void 0
		}, void 0);
	}
}
function Il(e, t) {
	let n = {}, r, i;
	for (i in t.properties) if (i !== "children" && vl.call(t.properties, i)) {
		let a = zl(e, i, t.properties[i]);
		if (a) {
			let [i, o] = a;
			e.tableCellAlignToStyle && i === "align" && typeof o == "string" && Sl.has(t.tagName) ? r = o : n[i] = o;
		}
	}
	if (r) {
		let t = n.style ||= {};
		t[e.stylePropertyNameCase === "css" ? "text-align" : "textAlign"] = r;
	}
	return n;
}
function Ll(e, t) {
	let n = {};
	for (let r of t.attributes) if (r.type === "mdxJsxExpressionAttribute") {
		if (r.data && r.data.estree && e.evaluater) {
			let t = r.data.estree.body[0];
			t.type;
			let i = t.expression;
			i.type;
			let a = i.properties[0];
			a.type, Object.assign(n, e.evaluater.evaluateExpression(a.argument));
		} else Hl(e, t.position);
	} else {
		let i = r.name, a;
		if (r.value && typeof r.value == "object") {
			if (r.value.data && r.value.data.estree && e.evaluater) {
				let t = r.value.data.estree.body[0];
				t.type, a = e.evaluater.evaluateExpression(t.expression);
			} else Hl(e, t.position);
		} else a = r.value === null || r.value;
		n[i] = a;
	}
	return n;
}
function Rl(e, t) {
	let n = [], r = -1, i = e.passKeys ? /* @__PURE__ */ new Map() : yl;
	for (; ++r < t.children.length;) {
		let a = t.children[r], o;
		if (e.passKeys) {
			let e = a.type === "element" ? a.tagName : a.type === "mdxJsxFlowElement" || a.type === "mdxJsxTextElement" ? a.name : void 0;
			if (e) {
				let t = i.get(e) || 0;
				o = e + "-" + t, i.set(e, t + 1);
			}
		}
		let s = Tl(e, a, o);
		s !== void 0 && n.push(s);
	}
	return n;
}
function zl(e, t, n) {
	let r = c(e.schema, t);
	if (!(n == null || typeof n == "number" && Number.isNaN(n))) {
		if (Array.isArray(n) && (n = r.commaSeparated ? s(n) : p(n)), r.property === "style") {
			let t = typeof n == "object" ? n : Bl(e, String(n));
			return e.stylePropertyNameCase === "css" && (t = Ul(t)), ["style", t];
		}
		return [e.elementAttributeNameCase === "react" && r.space ? be[r.property] || r.property : r.attribute, n];
	}
}
function Bl(e, t) {
	try {
		return (0, _l.default)(t, { reactCompat: !0 });
	} catch (t) {
		if (e.ignoreInvalidStyle) return {};
		let n = t, r = new ge("Cannot parse `style` attribute", {
			ancestors: e.ancestors,
			cause: n,
			ruleId: "style",
			source: "hast-util-to-jsx-runtime"
		});
		throw r.file = e.filePath || void 0, r.url = Cl + "#cannot-parse-style-attribute", r;
	}
}
function Vl(e, t, n) {
	let r;
	if (!n) r = {
		type: "Literal",
		value: t
	};
	else if (t.includes(".")) {
		let e = t.split("."), n = -1, i;
		for (; ++n < e.length;) {
			let t = se(e[n]) ? {
				type: "Identifier",
				name: e[n]
			} : {
				type: "Literal",
				value: e[n]
			};
			i = i ? {
				type: "MemberExpression",
				object: i,
				property: t,
				computed: !!(n && t.type === "Literal"),
				optional: !1
			} : t;
		}
		r = i;
	} else r = se(t) && !/^[a-z]/.test(t) ? {
		type: "Identifier",
		name: t
	} : {
		type: "Literal",
		value: t
	};
	if (r.type === "Literal") {
		let t = r.value;
		return vl.call(e.components, t) ? e.components[t] : t;
	}
	if (e.evaluater) return e.evaluater.evaluateExpression(r);
	Hl(e);
}
function Hl(e, t) {
	let n = new ge("Cannot handle MDX estrees without `createEvaluater`", {
		ancestors: e.ancestors,
		place: t,
		ruleId: "mdx-estree",
		source: "hast-util-to-jsx-runtime"
	});
	throw n.file = e.filePath || void 0, n.url = Cl + "#cannot-handle-mdx-estrees-without-createevaluater", n;
}
function Ul(e) {
	let t = {}, n;
	for (n in e) vl.call(e, n) && (t[Wl(n)] = e[n]);
	return t;
}
function Wl(e) {
	let t = e.replace(bl, Gl);
	return t.slice(0, 3) === "ms-" && (t = "-" + t), t;
}
function Gl(e) {
	return "-" + e.toLowerCase();
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/html-url-attributes/3.0.1/4ff74860c28cab7a302474ce2fd1dfeaa22ae8a9d7cc188f8eeff706084f44a8/node_modules/html-url-attributes/lib/index.js
var Kl = {
	action: ["form"],
	cite: [
		"blockquote",
		"del",
		"ins",
		"q"
	],
	data: ["object"],
	formAction: ["button", "input"],
	href: [
		"a",
		"area",
		"base",
		"link"
	],
	icon: ["menuitem"],
	itemId: null,
	manifest: ["html"],
	ping: ["a", "area"],
	poster: ["video"],
	src: [
		"audio",
		"embed",
		"iframe",
		"img",
		"input",
		"script",
		"source",
		"track",
		"video"
	]
};
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/remark-parse/11.0.0/59a9f7899d40718bc1cd4993d3ec5fda228d82bb6c86cdfcea654e3258fdad1e/node_modules/remark-parse/lib/index.js
function ql(e) {
	let t = this;
	t.parser = n;
	function n(n) {
		return _e(n, {
			...t.data("settings"),
			...e,
			extensions: t.data("micromarkExtensions") || [],
			mdastExtensions: t.data("fromMarkdownExtensions") || []
		});
	}
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/mdast-util-to-hast/13.2.1/9e2d4a6218823ab946e26ca11e111009b45d330ac891ab462284703470b69d9c/node_modules/mdast-util-to-hast/lib/handlers/blockquote.js
function Jl(e, t) {
	let n = {
		type: "element",
		tagName: "blockquote",
		properties: {},
		children: e.wrap(e.all(t), !0)
	};
	return e.patch(t, n), e.applyData(t, n);
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/mdast-util-to-hast/13.2.1/9e2d4a6218823ab946e26ca11e111009b45d330ac891ab462284703470b69d9c/node_modules/mdast-util-to-hast/lib/handlers/break.js
function Yl(e, t) {
	let n = {
		type: "element",
		tagName: "br",
		properties: {},
		children: []
	};
	return e.patch(t, n), [e.applyData(t, n), {
		type: "text",
		value: "\n"
	}];
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/mdast-util-to-hast/13.2.1/9e2d4a6218823ab946e26ca11e111009b45d330ac891ab462284703470b69d9c/node_modules/mdast-util-to-hast/lib/handlers/code.js
function Xl(e, t) {
	let n = t.value ? t.value + "\n" : "", r = {}, i = t.lang ? t.lang.split(/\s+/) : [];
	i.length > 0 && (r.className = ["language-" + i[0]]);
	let a = {
		type: "element",
		tagName: "code",
		properties: r,
		children: [{
			type: "text",
			value: n
		}]
	};
	return t.meta && (a.data = { meta: t.meta }), e.patch(t, a), a = e.applyData(t, a), a = {
		type: "element",
		tagName: "pre",
		properties: {},
		children: [a]
	}, e.patch(t, a), a;
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/mdast-util-to-hast/13.2.1/9e2d4a6218823ab946e26ca11e111009b45d330ac891ab462284703470b69d9c/node_modules/mdast-util-to-hast/lib/handlers/delete.js
function Zl(e, t) {
	let n = {
		type: "element",
		tagName: "del",
		properties: {},
		children: e.all(t)
	};
	return e.patch(t, n), e.applyData(t, n);
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/mdast-util-to-hast/13.2.1/9e2d4a6218823ab946e26ca11e111009b45d330ac891ab462284703470b69d9c/node_modules/mdast-util-to-hast/lib/handlers/emphasis.js
function Ql(e, t) {
	let n = {
		type: "element",
		tagName: "em",
		properties: {},
		children: e.all(t)
	};
	return e.patch(t, n), e.applyData(t, n);
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/mdast-util-to-hast/13.2.1/9e2d4a6218823ab946e26ca11e111009b45d330ac891ab462284703470b69d9c/node_modules/mdast-util-to-hast/lib/handlers/footnote-reference.js
function $l(e, t) {
	let n = typeof e.options.clobberPrefix == "string" ? e.options.clobberPrefix : "user-content-", r = String(t.identifier).toUpperCase(), i = ws(r.toLowerCase()), a = e.footnoteOrder.indexOf(r), o, s = e.footnoteCounts.get(r);
	s === void 0 ? (s = 0, e.footnoteOrder.push(r), o = e.footnoteOrder.length) : o = a + 1, s += 1, e.footnoteCounts.set(r, s);
	let c = {
		type: "element",
		tagName: "a",
		properties: {
			href: "#" + n + "fn-" + i,
			id: n + "fnref-" + i + (s > 1 ? "-" + s : ""),
			dataFootnoteRef: !0,
			ariaDescribedBy: ["footnote-label"]
		},
		children: [{
			type: "text",
			value: String(o)
		}]
	};
	e.patch(t, c);
	let l = {
		type: "element",
		tagName: "sup",
		properties: {},
		children: [c]
	};
	return e.patch(t, l), e.applyData(t, l);
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/mdast-util-to-hast/13.2.1/9e2d4a6218823ab946e26ca11e111009b45d330ac891ab462284703470b69d9c/node_modules/mdast-util-to-hast/lib/handlers/heading.js
function eu(e, t) {
	let n = {
		type: "element",
		tagName: "h" + t.depth,
		properties: {},
		children: e.all(t)
	};
	return e.patch(t, n), e.applyData(t, n);
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/mdast-util-to-hast/13.2.1/9e2d4a6218823ab946e26ca11e111009b45d330ac891ab462284703470b69d9c/node_modules/mdast-util-to-hast/lib/handlers/html.js
function tu(e, t) {
	if (e.options.allowDangerousHtml) {
		let n = {
			type: "raw",
			value: t.value
		};
		return e.patch(t, n), e.applyData(t, n);
	}
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/mdast-util-to-hast/13.2.1/9e2d4a6218823ab946e26ca11e111009b45d330ac891ab462284703470b69d9c/node_modules/mdast-util-to-hast/lib/revert.js
function nu(e, t) {
	let n = t.referenceType, r = "]";
	if (n === "collapsed" ? r += "[]" : n === "full" && (r += "[" + (t.label || t.identifier) + "]"), t.type === "imageReference") return [{
		type: "text",
		value: "![" + t.alt + r
	}];
	let i = e.all(t), a = i[0];
	a && a.type === "text" ? a.value = "[" + a.value : i.unshift({
		type: "text",
		value: "["
	});
	let o = i[i.length - 1];
	return o && o.type === "text" ? o.value += r : i.push({
		type: "text",
		value: r
	}), i;
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/mdast-util-to-hast/13.2.1/9e2d4a6218823ab946e26ca11e111009b45d330ac891ab462284703470b69d9c/node_modules/mdast-util-to-hast/lib/handlers/image-reference.js
function ru(e, t) {
	let n = String(t.identifier).toUpperCase(), r = e.definitionById.get(n);
	if (!r) return nu(e, t);
	let i = {
		src: ws(r.url || ""),
		alt: t.alt
	};
	r.title !== null && r.title !== void 0 && (i.title = r.title);
	let a = {
		type: "element",
		tagName: "img",
		properties: i,
		children: []
	};
	return e.patch(t, a), e.applyData(t, a);
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/mdast-util-to-hast/13.2.1/9e2d4a6218823ab946e26ca11e111009b45d330ac891ab462284703470b69d9c/node_modules/mdast-util-to-hast/lib/handlers/image.js
function iu(e, t) {
	let n = { src: ws(t.url) };
	t.alt !== null && t.alt !== void 0 && (n.alt = t.alt), t.title !== null && t.title !== void 0 && (n.title = t.title);
	let r = {
		type: "element",
		tagName: "img",
		properties: n,
		children: []
	};
	return e.patch(t, r), e.applyData(t, r);
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/mdast-util-to-hast/13.2.1/9e2d4a6218823ab946e26ca11e111009b45d330ac891ab462284703470b69d9c/node_modules/mdast-util-to-hast/lib/handlers/inline-code.js
function au(e, t) {
	let n = {
		type: "text",
		value: t.value.replace(/\r?\n|\r/g, " ")
	};
	e.patch(t, n);
	let r = {
		type: "element",
		tagName: "code",
		properties: {},
		children: [n]
	};
	return e.patch(t, r), e.applyData(t, r);
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/mdast-util-to-hast/13.2.1/9e2d4a6218823ab946e26ca11e111009b45d330ac891ab462284703470b69d9c/node_modules/mdast-util-to-hast/lib/handlers/link-reference.js
function ou(e, t) {
	let n = String(t.identifier).toUpperCase(), r = e.definitionById.get(n);
	if (!r) return nu(e, t);
	let i = { href: ws(r.url || "") };
	r.title !== null && r.title !== void 0 && (i.title = r.title);
	let a = {
		type: "element",
		tagName: "a",
		properties: i,
		children: e.all(t)
	};
	return e.patch(t, a), e.applyData(t, a);
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/mdast-util-to-hast/13.2.1/9e2d4a6218823ab946e26ca11e111009b45d330ac891ab462284703470b69d9c/node_modules/mdast-util-to-hast/lib/handlers/link.js
function su(e, t) {
	let n = { href: ws(t.url) };
	t.title !== null && t.title !== void 0 && (n.title = t.title);
	let r = {
		type: "element",
		tagName: "a",
		properties: n,
		children: e.all(t)
	};
	return e.patch(t, r), e.applyData(t, r);
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/mdast-util-to-hast/13.2.1/9e2d4a6218823ab946e26ca11e111009b45d330ac891ab462284703470b69d9c/node_modules/mdast-util-to-hast/lib/handlers/list-item.js
function cu(e, t, n) {
	let r = e.all(t), i = n ? lu(n) : uu(t), a = {}, o = [];
	if (typeof t.checked == "boolean") {
		let e = r[0], n;
		e && e.type === "element" && e.tagName === "p" ? n = e : (n = {
			type: "element",
			tagName: "p",
			properties: {},
			children: []
		}, r.unshift(n)), n.children.length > 0 && n.children.unshift({
			type: "text",
			value: " "
		}), n.children.unshift({
			type: "element",
			tagName: "input",
			properties: {
				type: "checkbox",
				checked: t.checked,
				disabled: !0
			},
			children: []
		}), a.className = ["task-list-item"];
	}
	let s = -1;
	for (; ++s < r.length;) {
		let e = r[s];
		(i || s !== 0 || e.type !== "element" || e.tagName !== "p") && o.push({
			type: "text",
			value: "\n"
		}), e.type === "element" && e.tagName === "p" && !i ? o.push(...e.children) : o.push(e);
	}
	let c = r[r.length - 1];
	c && (i || c.type !== "element" || c.tagName !== "p") && o.push({
		type: "text",
		value: "\n"
	});
	let l = {
		type: "element",
		tagName: "li",
		properties: a,
		children: o
	};
	return e.patch(t, l), e.applyData(t, l);
}
function lu(e) {
	let t = !1;
	if (e.type === "list") {
		t = e.spread || !1;
		let n = e.children, r = -1;
		for (; !t && ++r < n.length;) t = uu(n[r]);
	}
	return t;
}
function uu(e) {
	return e.spread ?? e.children.length > 1;
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/mdast-util-to-hast/13.2.1/9e2d4a6218823ab946e26ca11e111009b45d330ac891ab462284703470b69d9c/node_modules/mdast-util-to-hast/lib/handlers/list.js
function du(e, t) {
	let n = {}, r = e.all(t), i = -1;
	for (typeof t.start == "number" && t.start !== 1 && (n.start = t.start); ++i < r.length;) {
		let e = r[i];
		if (e.type === "element" && e.tagName === "li" && e.properties && Array.isArray(e.properties.className) && e.properties.className.includes("task-list-item")) {
			n.className = ["contains-task-list"];
			break;
		}
	}
	let a = {
		type: "element",
		tagName: t.ordered ? "ol" : "ul",
		properties: n,
		children: e.wrap(r, !0)
	};
	return e.patch(t, a), e.applyData(t, a);
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/mdast-util-to-hast/13.2.1/9e2d4a6218823ab946e26ca11e111009b45d330ac891ab462284703470b69d9c/node_modules/mdast-util-to-hast/lib/handlers/paragraph.js
function fu(e, t) {
	let n = {
		type: "element",
		tagName: "p",
		properties: {},
		children: e.all(t)
	};
	return e.patch(t, n), e.applyData(t, n);
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/mdast-util-to-hast/13.2.1/9e2d4a6218823ab946e26ca11e111009b45d330ac891ab462284703470b69d9c/node_modules/mdast-util-to-hast/lib/handlers/root.js
function pu(e, t) {
	let n = {
		type: "root",
		children: e.wrap(e.all(t))
	};
	return e.patch(t, n), e.applyData(t, n);
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/mdast-util-to-hast/13.2.1/9e2d4a6218823ab946e26ca11e111009b45d330ac891ab462284703470b69d9c/node_modules/mdast-util-to-hast/lib/handlers/strong.js
function mu(e, t) {
	let n = {
		type: "element",
		tagName: "strong",
		properties: {},
		children: e.all(t)
	};
	return e.patch(t, n), e.applyData(t, n);
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/mdast-util-to-hast/13.2.1/9e2d4a6218823ab946e26ca11e111009b45d330ac891ab462284703470b69d9c/node_modules/mdast-util-to-hast/lib/handlers/table.js
function hu(e, t) {
	let n = e.all(t), r = n.shift(), i = [];
	if (r) {
		let n = {
			type: "element",
			tagName: "thead",
			properties: {},
			children: e.wrap([r], !0)
		};
		e.patch(t.children[0], n), i.push(n);
	}
	if (n.length > 0) {
		let r = {
			type: "element",
			tagName: "tbody",
			properties: {},
			children: e.wrap(n, !0)
		}, a = Oa(t.children[1]), o = Da(t.children[t.children.length - 1]);
		a && o && (r.position = {
			start: a,
			end: o
		}), i.push(r);
	}
	let a = {
		type: "element",
		tagName: "table",
		properties: {},
		children: e.wrap(i, !0)
	};
	return e.patch(t, a), e.applyData(t, a);
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/mdast-util-to-hast/13.2.1/9e2d4a6218823ab946e26ca11e111009b45d330ac891ab462284703470b69d9c/node_modules/mdast-util-to-hast/lib/handlers/table-row.js
function gu(e, t, n) {
	let r = n ? n.children : void 0, i = (r ? r.indexOf(t) : 1) === 0 ? "th" : "td", a = n && n.type === "table" ? n.align : void 0, o = a ? a.length : t.children.length, s = -1, c = [];
	for (; ++s < o;) {
		let n = t.children[s], r = {}, o = a ? a[s] : void 0;
		o && (r.align = o);
		let l = {
			type: "element",
			tagName: i,
			properties: r,
			children: []
		};
		n && (l.children = e.all(n), e.patch(n, l), l = e.applyData(n, l)), c.push(l);
	}
	let l = {
		type: "element",
		tagName: "tr",
		properties: {},
		children: e.wrap(c, !0)
	};
	return e.patch(t, l), e.applyData(t, l);
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/mdast-util-to-hast/13.2.1/9e2d4a6218823ab946e26ca11e111009b45d330ac891ab462284703470b69d9c/node_modules/mdast-util-to-hast/lib/handlers/table-cell.js
function _u(e, t) {
	let n = {
		type: "element",
		tagName: "td",
		properties: {},
		children: e.all(t)
	};
	return e.patch(t, n), e.applyData(t, n);
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/trim-lines/3.0.1/3e05e1ebee3bd4d32745ba97f18d7b40732dc36fb1129ef08059a5ff4e875642/node_modules/trim-lines/index.js
var vu = 9, yu = 32;
function bu(e) {
	let t = String(e), n = /\r?\n|\r/g, r = n.exec(t), i = 0, a = [];
	for (; r;) a.push(xu(t.slice(i, r.index), i > 0, !0), r[0]), i = r.index + r[0].length, r = n.exec(t);
	return a.push(xu(t.slice(i), i > 0, !1)), a.join("");
}
function xu(e, t, n) {
	let r = 0, i = e.length;
	if (t) {
		let t = e.codePointAt(r);
		for (; t === vu || t === yu;) r++, t = e.codePointAt(r);
	}
	if (n) {
		let t = e.codePointAt(i - 1);
		for (; t === vu || t === yu;) i--, t = e.codePointAt(i - 1);
	}
	return i > r ? e.slice(r, i) : "";
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/mdast-util-to-hast/13.2.1/9e2d4a6218823ab946e26ca11e111009b45d330ac891ab462284703470b69d9c/node_modules/mdast-util-to-hast/lib/handlers/text.js
function Su(e, t) {
	let n = {
		type: "text",
		value: bu(String(t.value))
	};
	return e.patch(t, n), e.applyData(t, n);
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/mdast-util-to-hast/13.2.1/9e2d4a6218823ab946e26ca11e111009b45d330ac891ab462284703470b69d9c/node_modules/mdast-util-to-hast/lib/handlers/thematic-break.js
function Cu(e, t) {
	let n = {
		type: "element",
		tagName: "hr",
		properties: {},
		children: []
	};
	return e.patch(t, n), e.applyData(t, n);
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/mdast-util-to-hast/13.2.1/9e2d4a6218823ab946e26ca11e111009b45d330ac891ab462284703470b69d9c/node_modules/mdast-util-to-hast/lib/handlers/index.js
var wu = {
	blockquote: Jl,
	break: Yl,
	code: Xl,
	delete: Zl,
	emphasis: Ql,
	footnoteReference: $l,
	heading: eu,
	html: tu,
	imageReference: ru,
	image: iu,
	inlineCode: au,
	linkReference: ou,
	link: su,
	listItem: cu,
	list: du,
	paragraph: fu,
	root: pu,
	strong: mu,
	table: hu,
	tableCell: _u,
	tableRow: gu,
	text: Su,
	thematicBreak: Cu,
	toml: Tu,
	yaml: Tu,
	definition: Tu,
	footnoteDefinition: Tu
};
function Tu() {}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/mdast-util-to-hast/13.2.1/9e2d4a6218823ab946e26ca11e111009b45d330ac891ab462284703470b69d9c/node_modules/mdast-util-to-hast/lib/footer.js
function Eu(e, t) {
	let n = [{
		type: "text",
		value: "↩"
	}];
	return t > 1 && n.push({
		type: "element",
		tagName: "sup",
		properties: {},
		children: [{
			type: "text",
			value: String(t)
		}]
	}), n;
}
function Du(e, t) {
	return "Back to reference " + (e + 1) + (t > 1 ? "-" + t : "");
}
function Ou(e) {
	let t = typeof e.options.clobberPrefix == "string" ? e.options.clobberPrefix : "user-content-", n = e.options.footnoteBackContent || Eu, r = e.options.footnoteBackLabel || Du, i = e.options.footnoteLabel || "Footnotes", a = e.options.footnoteLabelTagName || "h2", o = e.options.footnoteLabelProperties || { className: ["sr-only"] }, s = [], c = -1;
	for (; ++c < e.footnoteOrder.length;) {
		let i = e.footnoteById.get(e.footnoteOrder[c]);
		if (!i) continue;
		let a = e.all(i), o = String(i.identifier).toUpperCase(), l = ws(o.toLowerCase()), u = 0, d = [], f = e.footnoteCounts.get(o);
		for (; f !== void 0 && ++u <= f;) {
			d.length > 0 && d.push({
				type: "text",
				value: " "
			});
			let e = typeof n == "string" ? n : n(c, u);
			typeof e == "string" && (e = {
				type: "text",
				value: e
			}), d.push({
				type: "element",
				tagName: "a",
				properties: {
					href: "#" + t + "fnref-" + l + (u > 1 ? "-" + u : ""),
					dataFootnoteBackref: "",
					ariaLabel: typeof r == "string" ? r : r(c, u),
					className: ["data-footnote-backref"]
				},
				children: Array.isArray(e) ? e : [e]
			});
		}
		let p = a[a.length - 1];
		if (p && p.type === "element" && p.tagName === "p") {
			let e = p.children[p.children.length - 1];
			e && e.type === "text" ? e.value += " " : p.children.push({
				type: "text",
				value: " "
			}), p.children.push(...d);
		} else a.push(...d);
		let m = {
			type: "element",
			tagName: "li",
			properties: { id: t + "fn-" + l },
			children: e.wrap(a, !0)
		};
		e.patch(i, m), s.push(m);
	}
	if (s.length !== 0) return {
		type: "element",
		tagName: "section",
		properties: {
			dataFootnotes: !0,
			className: ["footnotes"]
		},
		children: [
			{
				type: "element",
				tagName: a,
				properties: {
					...Ue(o),
					id: "footnote-label"
				},
				children: [{
					type: "text",
					value: i
				}]
			},
			{
				type: "text",
				value: "\n"
			},
			{
				type: "element",
				tagName: "ol",
				properties: {},
				children: e.wrap(s, !0)
			},
			{
				type: "text",
				value: "\n"
			}
		]
	};
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/mdast-util-to-hast/13.2.1/9e2d4a6218823ab946e26ca11e111009b45d330ac891ab462284703470b69d9c/node_modules/mdast-util-to-hast/lib/state.js
var ku = {}.hasOwnProperty, Au = {};
function ju(e, t) {
	let n = t || Au, r = /* @__PURE__ */ new Map(), i = /* @__PURE__ */ new Map(), a = {
		all: s,
		applyData: Nu,
		definitionById: r,
		footnoteById: i,
		footnoteCounts: /* @__PURE__ */ new Map(),
		footnoteOrder: [],
		handlers: {
			...wu,
			...n.handlers
		},
		one: o,
		options: n,
		patch: Mu,
		wrap: Fu
	};
	return C(e, function(e) {
		if (e.type === "definition" || e.type === "footnoteDefinition") {
			let t = e.type === "definition" ? r : i, n = String(e.identifier).toUpperCase();
			t.has(n) || t.set(n, e);
		}
	}), a;
	function o(e, t) {
		let n = e.type, r = a.handlers[n];
		if (ku.call(a.handlers, n) && r) return r(a, e, t);
		if (a.options.passThrough && a.options.passThrough.includes(n)) {
			if ("children" in e) {
				let { children: t, ...n } = e, r = Ue(n);
				return r.children = a.all(e), r;
			}
			return Ue(e);
		}
		return (a.options.unknownHandler || Pu)(a, e, t);
	}
	function s(e) {
		let t = [];
		if ("children" in e) {
			let n = e.children, r = -1;
			for (; ++r < n.length;) {
				let i = a.one(n[r], e);
				if (i) {
					if (r && n[r - 1].type === "break" && (!Array.isArray(i) && i.type === "text" && (i.value = Iu(i.value)), !Array.isArray(i) && i.type === "element")) {
						let e = i.children[0];
						e && e.type === "text" && (e.value = Iu(e.value));
					}
					Array.isArray(i) ? t.push(...i) : t.push(i);
				}
			}
		}
		return t;
	}
}
function Mu(e, t) {
	e.position && (t.position = Aa(e));
}
function Nu(e, t) {
	let n = t;
	if (e && e.data) {
		let t = e.data.hName, r = e.data.hChildren, i = e.data.hProperties;
		typeof t == "string" && (n.type === "element" ? n.tagName = t : n = {
			type: "element",
			tagName: t,
			properties: {},
			children: "children" in n ? n.children : [n]
		}), n.type === "element" && i && Object.assign(n.properties, Ue(i)), "children" in n && n.children && r != null && (n.children = r);
	}
	return n;
}
function Pu(e, t) {
	let n = t.data || {}, r = "value" in t && !(ku.call(n, "hProperties") || ku.call(n, "hChildren")) ? {
		type: "text",
		value: t.value
	} : {
		type: "element",
		tagName: "div",
		properties: {},
		children: e.all(t)
	};
	return e.patch(t, r), e.applyData(t, r);
}
function Fu(e, t) {
	let n = [], r = -1;
	for (t && n.push({
		type: "text",
		value: "\n"
	}); ++r < e.length;) r && n.push({
		type: "text",
		value: "\n"
	}), n.push(e[r]);
	return t && e.length > 0 && n.push({
		type: "text",
		value: "\n"
	}), n;
}
function Iu(e) {
	let t = 0, n = e.charCodeAt(t);
	for (; n === 9 || n === 32;) t++, n = e.charCodeAt(t);
	return e.slice(t);
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/mdast-util-to-hast/13.2.1/9e2d4a6218823ab946e26ca11e111009b45d330ac891ab462284703470b69d9c/node_modules/mdast-util-to-hast/lib/index.js
function Lu(e, t) {
	let n = ju(e, t), r = n.one(e, void 0), i = Ou(n), a = Array.isArray(r) ? {
		type: "root",
		children: r
	} : r || {
		type: "root",
		children: []
	};
	return i && ("children" in a, a.children.push({
		type: "text",
		value: "\n"
	}, i)), a;
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/remark-rehype/11.1.2/213d3ab44baeb4760347fb405e499f7ce7441f37956233b944b3f014542598b8/node_modules/remark-rehype/lib/index.js
function Ru(e, t) {
	return e && "run" in e ? async function(n, r) {
		let i = Lu(n, {
			file: r,
			...t
		});
		await e.run(i, r);
	} : function(n, r) {
		return Lu(n, {
			file: r,
			...e || t
		});
	};
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/bail/2.0.2/838d1c0654e3975d06f7e82c73b61b5a08f8fc8de6d14d08bf3bb0164fb24956/node_modules/bail/index.js
function zu(e) {
	if (e) throw e;
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/extend/3.0.2/5eb62444e2a85cf73742fe2e1f08618c58e4d2cc6aae35a1cde37dd81613cc59/node_modules/extend/index.js
var Bu = /* @__PURE__ */ t(((e, t) => {
	var n = Object.prototype.hasOwnProperty, r = Object.prototype.toString, i = Object.defineProperty, a = Object.getOwnPropertyDescriptor, o = function(e) {
		return typeof Array.isArray == "function" ? Array.isArray(e) : r.call(e) === "[object Array]";
	}, s = function(e) {
		if (!e || r.call(e) !== "[object Object]") return !1;
		var t = n.call(e, "constructor"), i = e.constructor && e.constructor.prototype && n.call(e.constructor.prototype, "isPrototypeOf");
		if (e.constructor && !t && !i) return !1;
		for (var a in e);
		return a === void 0 || n.call(e, a);
	}, c = function(e, t) {
		i && t.name === "__proto__" ? i(e, t.name, {
			enumerable: !0,
			configurable: !0,
			value: t.newValue,
			writable: !0
		}) : e[t.name] = t.newValue;
	}, l = function(e, t) {
		if (t === "__proto__") {
			if (!n.call(e, t)) return;
			if (a) return a(e, t).value;
		}
		return e[t];
	};
	t.exports = function e() {
		var t, n, r, i, a, u, d = arguments[0], f = 1, p = arguments.length, m = !1;
		for (typeof d == "boolean" && (m = d, d = arguments[1] || {}, f = 2), (d == null || typeof d != "object" && typeof d != "function") && (d = {}); f < p; ++f) if (t = arguments[f], t != null) for (n in t) r = l(d, n), i = l(t, n), d !== i && (m && i && (s(i) || (a = o(i))) ? (a ? (a = !1, u = r && o(r) ? r : []) : u = r && s(r) ? r : {}, c(d, {
			name: n,
			newValue: e(m, u, i)
		})) : i !== void 0 && c(d, {
			name: n,
			newValue: i
		}));
		return d;
	};
}));
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/is-plain-obj/4.1.0/d7fab263ffdd134b8e9e6130a60458baf7017c81e96947d5a2e6383c3d47b281/node_modules/is-plain-obj/index.js
function Vu(e) {
	if (typeof e != "object" || !e) return !1;
	let t = Object.getPrototypeOf(e);
	return (t === null || t === Object.prototype || Object.getPrototypeOf(t) === null) && !(Symbol.toStringTag in e) && !(Symbol.iterator in e);
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/trough/2.2.0/d3fda8c9b431ad5b37c76eca267b819119248e1cd71c2d1cb0490532cebf1f16/node_modules/trough/lib/index.js
function Hu() {
	let e = [], t = {
		run: n,
		use: r
	};
	return t;
	function n(...t) {
		let n = -1, r = t.pop();
		if (typeof r != "function") throw TypeError("Expected function as last argument, not " + r);
		i(null, ...t);
		function i(a, ...o) {
			let s = e[++n], c = -1;
			if (a) r(a);
			else {
				for (; ++c < t.length;) (o[c] === null || o[c] === void 0) && (o[c] = t[c]);
				t = o, s ? Uu(s, i)(...o) : r(null, ...o);
			}
		}
	}
	function r(n) {
		if (typeof n != "function") throw TypeError("Expected `middelware` to be a function, not " + n);
		return e.push(n), t;
	}
}
function Uu(e, t) {
	let n;
	return r;
	function r(...t) {
		let r = e.length > t.length, o;
		r && t.push(i);
		try {
			o = e.apply(this, t);
		} catch (e) {
			let t = e;
			if (r && n) throw t;
			return i(t);
		}
		r || (o && o.then && typeof o.then == "function" ? o.then(a, i) : o instanceof Error ? i(o) : a(o));
	}
	function i(e, ...r) {
		n || (n = !0, t(e, ...r));
	}
	function a(e) {
		i(null, e);
	}
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/vfile/6.0.3/c95f08038d7738c3693dc5b3e6949fc88e9e20afbbd4556e89a70570a86279d5/node_modules/vfile/lib/minpath.browser.js
var Wu = {
	basename: Gu,
	dirname: Ku,
	extname: qu,
	join: Ju,
	sep: "/"
};
function Gu(e, t) {
	if (t !== void 0 && typeof t != "string") throw TypeError("\"ext\" argument must be a string");
	Zu(e);
	let n = 0, r = -1, i = e.length, a;
	if (t === void 0 || t.length === 0 || t.length > e.length) {
		for (; i--;) if (e.codePointAt(i) === 47) {
			if (a) {
				n = i + 1;
				break;
			}
		} else r < 0 && (a = !0, r = i + 1);
		return r < 0 ? "" : e.slice(n, r);
	}
	if (t === e) return "";
	let o = -1, s = t.length - 1;
	for (; i--;) if (e.codePointAt(i) === 47) {
		if (a) {
			n = i + 1;
			break;
		}
	} else o < 0 && (a = !0, o = i + 1), s > -1 && (e.codePointAt(i) === t.codePointAt(s--) ? s < 0 && (r = i) : (s = -1, r = o));
	return n === r ? r = o : r < 0 && (r = e.length), e.slice(n, r);
}
function Ku(e) {
	if (Zu(e), e.length === 0) return ".";
	let t = -1, n = e.length, r;
	for (; --n;) if (e.codePointAt(n) === 47) {
		if (r) {
			t = n;
			break;
		}
	} else r ||= !0;
	return t < 0 ? e.codePointAt(0) === 47 ? "/" : "." : t === 1 && e.codePointAt(0) === 47 ? "//" : e.slice(0, t);
}
function qu(e) {
	Zu(e);
	let t = e.length, n = -1, r = 0, i = -1, a = 0, o;
	for (; t--;) {
		let s = e.codePointAt(t);
		if (s === 47) {
			if (o) {
				r = t + 1;
				break;
			}
		} else n < 0 && (o = !0, n = t + 1), s === 46 ? i < 0 ? i = t : a !== 1 && (a = 1) : i > -1 && (a = -1);
	}
	return i < 0 || n < 0 || a === 0 || a === 1 && i === n - 1 && i === r + 1 ? "" : e.slice(i, n);
}
function Ju(...e) {
	let t = -1, n;
	for (; ++t < e.length;) Zu(e[t]), e[t] && (n = n === void 0 ? e[t] : n + "/" + e[t]);
	return n === void 0 ? "." : Yu(n);
}
function Yu(e) {
	Zu(e);
	let t = e.codePointAt(0) === 47, n = Xu(e, !t);
	return n.length === 0 && !t && (n = "."), n.length > 0 && e.codePointAt(e.length - 1) === 47 && (n += "/"), t ? "/" + n : n;
}
function Xu(e, t) {
	let n = "", r = 0, i = -1, a = 0, o = -1, s, c;
	for (; ++o <= e.length;) {
		if (o < e.length) s = e.codePointAt(o);
		else if (s === 47) break;
		else s = 47;
		if (s === 47) {
			if (i !== o - 1 && a !== 1) {
				if (i !== o - 1 && a === 2) {
					if (n.length < 2 || r !== 2 || n.codePointAt(n.length - 1) !== 46 || n.codePointAt(n.length - 2) !== 46) {
						if (n.length > 2) {
							if (c = n.lastIndexOf("/"), c !== n.length - 1) {
								c < 0 ? (n = "", r = 0) : (n = n.slice(0, c), r = n.length - 1 - n.lastIndexOf("/")), i = o, a = 0;
								continue;
							}
						} else if (n.length > 0) {
							n = "", r = 0, i = o, a = 0;
							continue;
						}
					}
					t && (n = n.length > 0 ? n + "/.." : "..", r = 2);
				} else n.length > 0 ? n += "/" + e.slice(i + 1, o) : n = e.slice(i + 1, o), r = o - i - 1;
			}
			i = o, a = 0;
		} else s === 46 && a > -1 ? a++ : a = -1;
	}
	return n;
}
function Zu(e) {
	if (typeof e != "string") throw TypeError("Path must be a string. Received " + JSON.stringify(e));
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/vfile/6.0.3/c95f08038d7738c3693dc5b3e6949fc88e9e20afbbd4556e89a70570a86279d5/node_modules/vfile/lib/minproc.browser.js
var Qu = { cwd: $u };
function $u() {
	return "/";
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/vfile/6.0.3/c95f08038d7738c3693dc5b3e6949fc88e9e20afbbd4556e89a70570a86279d5/node_modules/vfile/lib/minurl.shared.js
function ed(e) {
	return !!(typeof e == "object" && e && "href" in e && e.href && "protocol" in e && e.protocol && e.auth === void 0);
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/vfile/6.0.3/c95f08038d7738c3693dc5b3e6949fc88e9e20afbbd4556e89a70570a86279d5/node_modules/vfile/lib/minurl.browser.js
function td(e) {
	if (typeof e == "string") e = new URL(e);
	else if (!ed(e)) {
		let t = /* @__PURE__ */ TypeError("The \"path\" argument must be of type string or an instance of URL. Received `" + e + "`");
		throw t.code = "ERR_INVALID_ARG_TYPE", t;
	}
	if (e.protocol !== "file:") {
		let e = /* @__PURE__ */ TypeError("The URL must be of scheme file");
		throw e.code = "ERR_INVALID_URL_SCHEME", e;
	}
	return nd(e);
}
function nd(e) {
	if (e.hostname !== "") {
		let e = /* @__PURE__ */ TypeError("File URL host must be \"localhost\" or empty on darwin");
		throw e.code = "ERR_INVALID_FILE_URL_HOST", e;
	}
	let t = e.pathname, n = -1;
	for (; ++n < t.length;) if (t.codePointAt(n) === 37 && t.codePointAt(n + 1) === 50) {
		let e = t.codePointAt(n + 2);
		if (e === 70 || e === 102) {
			let e = /* @__PURE__ */ TypeError("File URL path must not include encoded / characters");
			throw e.code = "ERR_INVALID_FILE_URL_PATH", e;
		}
	}
	return decodeURIComponent(t);
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/vfile/6.0.3/c95f08038d7738c3693dc5b3e6949fc88e9e20afbbd4556e89a70570a86279d5/node_modules/vfile/lib/index.js
var rd = [
	"history",
	"path",
	"basename",
	"stem",
	"extname",
	"dirname"
], id = class {
	constructor(e) {
		let t;
		t = e ? ed(e) ? { path: e } : typeof e == "string" || cd(e) ? { value: e } : e : {}, this.cwd = "cwd" in t ? "" : Qu.cwd(), this.data = {}, this.history = [], this.messages = [], this.value, this.map, this.result, this.stored;
		let n = -1;
		for (; ++n < rd.length;) {
			let e = rd[n];
			e in t && t[e] !== void 0 && t[e] !== null && (this[e] = e === "history" ? [...t[e]] : t[e]);
		}
		let r;
		for (r in t) rd.includes(r) || (this[r] = t[r]);
	}
	get basename() {
		return typeof this.path == "string" ? Wu.basename(this.path) : void 0;
	}
	set basename(e) {
		od(e, "basename"), ad(e, "basename"), this.path = Wu.join(this.dirname || "", e);
	}
	get dirname() {
		return typeof this.path == "string" ? Wu.dirname(this.path) : void 0;
	}
	set dirname(e) {
		sd(this.basename, "dirname"), this.path = Wu.join(e || "", this.basename);
	}
	get extname() {
		return typeof this.path == "string" ? Wu.extname(this.path) : void 0;
	}
	set extname(e) {
		if (ad(e, "extname"), sd(this.dirname, "extname"), e) {
			if (e.codePointAt(0) !== 46) throw Error("`extname` must start with `.`");
			if (e.includes(".", 1)) throw Error("`extname` cannot contain multiple dots");
		}
		this.path = Wu.join(this.dirname, this.stem + (e || ""));
	}
	get path() {
		return this.history[this.history.length - 1];
	}
	set path(e) {
		ed(e) && (e = td(e)), od(e, "path"), this.path !== e && this.history.push(e);
	}
	get stem() {
		return typeof this.path == "string" ? Wu.basename(this.path, this.extname) : void 0;
	}
	set stem(e) {
		od(e, "stem"), ad(e, "stem"), this.path = Wu.join(this.dirname || "", e + (this.extname || ""));
	}
	fail(e, t, n) {
		let r = this.message(e, t, n);
		throw r.fatal = !0, r;
	}
	info(e, t, n) {
		let r = this.message(e, t, n);
		return r.fatal = void 0, r;
	}
	message(e, t, n) {
		let r = new ge(e, t, n);
		return this.path && (r.name = this.path + ":" + r.name, r.file = this.path), r.fatal = !1, this.messages.push(r), r;
	}
	toString(e) {
		return this.value === void 0 ? "" : typeof this.value == "string" ? this.value : new TextDecoder(e || void 0).decode(this.value);
	}
};
function ad(e, t) {
	if (e && e.includes(Wu.sep)) throw Error("`" + t + "` cannot be a path: did not expect `" + Wu.sep + "`");
}
function od(e, t) {
	if (!e) throw Error("`" + t + "` cannot be empty");
}
function sd(e, t) {
	if (!e) throw Error("Setting `" + t + "` requires `path` to be set too");
}
function cd(e) {
	return !!(e && typeof e == "object" && "byteLength" in e && "byteOffset" in e);
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/unified/11.0.5/bc453fd1ce38b454018aa30afab98a43140ddb9d705be13a6e19f903ecb01c2d/node_modules/unified/lib/callable-instance.js
var ld = (function(e) {
	let t = this.constructor.prototype, n = t[e], r = function() {
		return n.apply(r, arguments);
	};
	return Object.setPrototypeOf(r, t), r;
}), ud = /* @__PURE__ */ e(Bu(), 1), dd = {}.hasOwnProperty, fd = new class e extends ld {
	constructor() {
		super("copy"), this.Compiler = void 0, this.Parser = void 0, this.attachers = [], this.compiler = void 0, this.freezeIndex = -1, this.frozen = void 0, this.namespace = {}, this.parser = void 0, this.transformers = Hu();
	}
	copy() {
		let t = new e(), n = -1;
		for (; ++n < this.attachers.length;) {
			let e = this.attachers[n];
			t.use(...e);
		}
		return t.data((0, ud.default)(!0, {}, this.namespace)), t;
	}
	data(e, t) {
		return typeof e == "string" ? arguments.length === 2 ? (hd("data", this.frozen), this.namespace[e] = t, this) : dd.call(this.namespace, e) && this.namespace[e] || void 0 : e ? (hd("data", this.frozen), this.namespace = e, this) : this.namespace;
	}
	freeze() {
		if (this.frozen) return this;
		let e = this;
		for (; ++this.freezeIndex < this.attachers.length;) {
			let [t, ...n] = this.attachers[this.freezeIndex];
			if (n[0] === !1) continue;
			n[0] === !0 && (n[0] = void 0);
			let r = t.call(e, ...n);
			typeof r == "function" && this.transformers.use(r);
		}
		return this.frozen = !0, this.freezeIndex = Infinity, this;
	}
	parse(e) {
		this.freeze();
		let t = vd(e), n = this.parser || this.Parser;
		return pd("parse", n), n(String(t), t);
	}
	process(e, t) {
		let n = this;
		return this.freeze(), pd("process", this.parser || this.Parser), md("process", this.compiler || this.Compiler), t ? r(void 0, t) : new Promise(r);
		function r(r, i) {
			let a = vd(e), o = n.parse(a);
			n.run(o, a, function(e, t, r) {
				if (e || !t || !r) return s(e);
				let i = t, a = n.stringify(i, r);
				bd(a) ? r.value = a : r.result = a, s(e, r);
			});
			function s(e, n) {
				e || !n ? i(e) : r ? r(n) : t(void 0, n);
			}
		}
	}
	processSync(e) {
		let t = !1, n;
		return this.freeze(), pd("processSync", this.parser || this.Parser), md("processSync", this.compiler || this.Compiler), this.process(e, r), _d("processSync", "process", t), n;
		function r(e, r) {
			t = !0, zu(e), n = r;
		}
	}
	run(e, t, n) {
		gd(e), this.freeze();
		let r = this.transformers;
		return !n && typeof t == "function" && (n = t, t = void 0), n ? i(void 0, n) : new Promise(i);
		function i(i, a) {
			let o = vd(t);
			r.run(e, o, s);
			function s(t, r, o) {
				let s = r || e;
				t ? a(t) : i ? i(s) : n(void 0, s, o);
			}
		}
	}
	runSync(e, t) {
		let n = !1, r;
		return this.run(e, t, i), _d("runSync", "run", n), r;
		function i(e, t) {
			zu(e), r = t, n = !0;
		}
	}
	stringify(e, t) {
		this.freeze();
		let n = vd(t), r = this.compiler || this.Compiler;
		return md("stringify", r), gd(e), r(e, n);
	}
	use(e, ...t) {
		let n = this.attachers, r = this.namespace;
		if (hd("use", this.frozen), e != null) {
			if (typeof e == "function") s(e, t);
			else if (typeof e == "object") Array.isArray(e) ? o(e) : a(e);
			else throw TypeError("Expected usable value, not `" + e + "`");
		}
		return this;
		function i(e) {
			if (typeof e == "function") s(e, []);
			else if (typeof e == "object") {
				if (Array.isArray(e)) {
					let [t, ...n] = e;
					s(t, n);
				} else a(e);
			} else throw TypeError("Expected usable value, not `" + e + "`");
		}
		function a(e) {
			if (!("plugins" in e) && !("settings" in e)) throw Error("Expected usable value but received an empty preset, which is probably a mistake: presets typically come with `plugins` and sometimes with `settings`, but this has neither");
			o(e.plugins), e.settings && (r.settings = (0, ud.default)(!0, r.settings, e.settings));
		}
		function o(e) {
			let t = -1;
			if (e != null) {
				if (Array.isArray(e)) for (; ++t < e.length;) {
					let n = e[t];
					i(n);
				}
				else throw TypeError("Expected a list of plugins, not `" + e + "`");
			}
		}
		function s(e, t) {
			let r = -1, i = -1;
			for (; ++r < n.length;) if (n[r][0] === e) {
				i = r;
				break;
			}
			if (i === -1) n.push([e, ...t]);
			else if (t.length > 0) {
				let [r, ...a] = t, o = n[i][1];
				Vu(o) && Vu(r) && (r = (0, ud.default)(!0, o, r)), n[i] = [
					e,
					r,
					...a
				];
			}
		}
	}
}().freeze();
function pd(e, t) {
	if (typeof t != "function") throw TypeError("Cannot `" + e + "` without `parser`");
}
function md(e, t) {
	if (typeof t != "function") throw TypeError("Cannot `" + e + "` without `compiler`");
}
function hd(e, t) {
	if (t) throw Error("Cannot call `" + e + "` on a frozen processor.\nCreate a new processor first, by calling it: use `processor()` instead of `processor`.");
}
function gd(e) {
	if (!Vu(e) || typeof e.type != "string") throw TypeError("Expected node, got `" + e + "`");
}
function _d(e, t, n) {
	if (!n) throw Error("`" + e + "` finished async. Use `" + t + "` instead");
}
function vd(e) {
	return yd(e) ? e : new id(e);
}
function yd(e) {
	return !!(e && typeof e == "object" && "message" in e && "messages" in e);
}
function bd(e) {
	return typeof e == "string" || xd(e);
}
function xd(e) {
	return !!(e && typeof e == "object" && "byteLength" in e && "byteOffset" in e);
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/marked/17.0.6/a3f28d42b68fc9a4f16859ea1c48d686d2adc98d620646c4493ba2bfdc6bad96/node_modules/marked/lib/marked.esm.js
var Sd = /* @__PURE__ */ e(i(), 1), W = r();
function Cd() {
	return {
		async: !1,
		breaks: !1,
		extensions: null,
		gfm: !0,
		hooks: null,
		pedantic: !1,
		renderer: null,
		silent: !1,
		tokenizer: null,
		walkTokens: null
	};
}
var wd = Cd();
function Td(e) {
	wd = e;
}
var Ed = { exec: () => null };
function G(e, t = "") {
	let n = typeof e == "string" ? e : e.source, r = {
		replace: (e, t) => {
			let i = typeof t == "string" ? t : t.source;
			return i = i.replace(K.caret, "$1"), n = n.replace(e, i), r;
		},
		getRegex: () => new RegExp(n, t)
	};
	return r;
}
var Dd = (() => {
	try {
		return !0;
	} catch {
		return !1;
	}
})(), K = {
	codeRemoveIndent: /^(?: {1,4}| {0,3}\t)/gm,
	outputLinkReplace: /\\([\[\]])/g,
	indentCodeCompensation: /^(\s+)(?:```)/,
	beginningSpace: /^\s+/,
	endingHash: /#$/,
	startingSpaceChar: /^ /,
	endingSpaceChar: / $/,
	nonSpaceChar: /[^ ]/,
	newLineCharGlobal: /\n/g,
	tabCharGlobal: /\t/g,
	multipleSpaceGlobal: /\s+/g,
	blankLine: /^[ \t]*$/,
	doubleBlankLine: /\n[ \t]*\n[ \t]*$/,
	blockquoteStart: /^ {0,3}>/,
	blockquoteSetextReplace: /\n {0,3}((?:=+|-+) *)(?=\n|$)/g,
	blockquoteSetextReplace2: /^ {0,3}>[ \t]?/gm,
	listReplaceNesting: /^ {1,4}(?=( {4})*[^ ])/g,
	listIsTask: /^\[[ xX]\] +\S/,
	listReplaceTask: /^\[[ xX]\] +/,
	listTaskCheckbox: /\[[ xX]\]/,
	anyLine: /\n.*\n/,
	hrefBrackets: /^<(.*)>$/,
	tableDelimiter: /[:|]/,
	tableAlignChars: /^\||\| *$/g,
	tableRowBlankLine: /\n[ \t]*$/,
	tableAlignRight: /^ *-+: *$/,
	tableAlignCenter: /^ *:-+: *$/,
	tableAlignLeft: /^ *:-+ *$/,
	startATag: /^<a /i,
	endATag: /^<\/a>/i,
	startPreScriptTag: /^<(pre|code|kbd|script)(\s|>)/i,
	endPreScriptTag: /^<\/(pre|code|kbd|script)(\s|>)/i,
	startAngleBracket: /^</,
	endAngleBracket: />$/,
	pedanticHrefTitle: /^([^'"]*[^\s])\s+(['"])(.*)\2/,
	unicodeAlphaNumeric: /[\p{L}\p{N}]/u,
	escapeTest: /[&<>"']/,
	escapeReplace: /[&<>"']/g,
	escapeTestNoEncode: /[<>"']|&(?!(#\d{1,7}|#[Xx][a-fA-F0-9]{1,6}|\w+);)/,
	escapeReplaceNoEncode: /[<>"']|&(?!(#\d{1,7}|#[Xx][a-fA-F0-9]{1,6}|\w+);)/g,
	caret: /(^|[^\[])\^/g,
	percentDecode: /%25/g,
	findPipe: /\|/g,
	splitPipe: / \|/,
	slashPipe: /\\\|/g,
	carriageReturn: /\r\n|\r/g,
	spaceLine: /^ +$/gm,
	notSpaceStart: /^\S*/,
	endingNewline: /\n$/,
	listItemRegex: (e) => RegExp(`^( {0,3}${e})((?:[	 ][^\\n]*)?(?:\\n|$))`),
	nextBulletRegex: (e) => RegExp(`^ {0,${Math.min(3, e - 1)}}(?:[*+-]|\\d{1,9}[.)])((?:[ 	][^\\n]*)?(?:\\n|$))`),
	hrRegex: (e) => RegExp(`^ {0,${Math.min(3, e - 1)}}((?:- *){3,}|(?:_ *){3,}|(?:\\* *){3,})(?:\\n+|$)`),
	fencesBeginRegex: (e) => RegExp(`^ {0,${Math.min(3, e - 1)}}(?:\`\`\`|~~~)`),
	headingBeginRegex: (e) => RegExp(`^ {0,${Math.min(3, e - 1)}}#`),
	htmlBeginRegex: (e) => RegExp(`^ {0,${Math.min(3, e - 1)}}<(?:[a-z].*>|!--)`, "i"),
	blockquoteBeginRegex: (e) => RegExp(`^ {0,${Math.min(3, e - 1)}}>`)
}, Od = /^(?:[ \t]*(?:\n|$))+/, kd = /^((?: {4}| {0,3}\t)[^\n]+(?:\n(?:[ \t]*(?:\n|$))*)?)+/, Ad = /^ {0,3}(`{3,}(?=[^`\n]*(?:\n|$))|~{3,})([^\n]*)(?:\n|$)(?:|([\s\S]*?)(?:\n|$))(?: {0,3}\1[~`]* *(?=\n|$)|$)/, jd = /^ {0,3}((?:-[\t ]*){3,}|(?:_[ \t]*){3,}|(?:\*[ \t]*){3,})(?:\n+|$)/, Md = /^ {0,3}(#{1,6})(?=\s|$)(.*)(?:\n+|$)/, Nd = / {0,3}(?:[*+-]|\d{1,9}[.)])/, Pd = /^(?!bull |blockCode|fences|blockquote|heading|html|table)((?:.|\n(?!\s*?\n|bull |blockCode|fences|blockquote|heading|html|table))+?)\n {0,3}(=+|-+) *(?:\n+|$)/, Fd = G(Pd).replace(/bull/g, Nd).replace(/blockCode/g, /(?: {4}| {0,3}\t)/).replace(/fences/g, / {0,3}(?:`{3,}|~{3,})/).replace(/blockquote/g, / {0,3}>/).replace(/heading/g, / {0,3}#{1,6}/).replace(/html/g, / {0,3}<[^\n>]+>\n/).replace(/\|table/g, "").getRegex(), Id = G(Pd).replace(/bull/g, Nd).replace(/blockCode/g, /(?: {4}| {0,3}\t)/).replace(/fences/g, / {0,3}(?:`{3,}|~{3,})/).replace(/blockquote/g, / {0,3}>/).replace(/heading/g, / {0,3}#{1,6}/).replace(/html/g, / {0,3}<[^\n>]+>\n/).replace(/table/g, / {0,3}\|?(?:[:\- ]*\|)+[\:\- ]*\n/).getRegex(), Ld = /^([^\n]+(?:\n(?!hr|heading|lheading|blockquote|fences|list|html|table| +\n)[^\n]+)*)/, Rd = /^[^\n]+/, zd = /(?!\s*\])(?:\\[\s\S]|[^\[\]\\])+/, Bd = G(/^ {0,3}\[(label)\]: *(?:\n[ \t]*)?([^<\s][^\s]*|<.*?>)(?:(?: +(?:\n[ \t]*)?| *\n[ \t]*)(title))? *(?:\n+|$)/).replace("label", zd).replace("title", /(?:"(?:\\"?|[^"\\])*"|'[^'\n]*(?:\n[^'\n]+)*\n?'|\([^()]*\))/).getRegex(), Vd = G(/^(bull)([ \t][^\n]+?)?(?:\n|$)/).replace(/bull/g, Nd).getRegex(), Hd = "address|article|aside|base|basefont|blockquote|body|caption|center|col|colgroup|dd|details|dialog|dir|div|dl|dt|fieldset|figcaption|figure|footer|form|frame|frameset|h[1-6]|head|header|hr|html|iframe|legend|li|link|main|menu|menuitem|meta|nav|noframes|ol|optgroup|option|p|param|search|section|summary|table|tbody|td|tfoot|th|thead|title|tr|track|ul", Ud = /<!--(?:-?>|[\s\S]*?(?:-->|$))/, Wd = G("^ {0,3}(?:<(script|pre|style|textarea)[\\s>][\\s\\S]*?(?:</\\1>[^\\n]*\\n+|$)|comment[^\\n]*(\\n+|$)|<\\?[\\s\\S]*?(?:\\?>\\n*|$)|<![A-Z][\\s\\S]*?(?:>\\n*|$)|<!\\[CDATA\\[[\\s\\S]*?(?:\\]\\]>\\n*|$)|</?(tag)(?: +|\\n|/?>)[\\s\\S]*?(?:(?:\\n[ 	]*)+\\n|$)|<(?!script|pre|style|textarea)([a-z][\\w-]*)(?:attribute)*? */?>(?=[ \\t]*(?:\\n|$))[\\s\\S]*?(?:(?:\\n[ 	]*)+\\n|$)|</(?!script|pre|style|textarea)[a-z][\\w-]*\\s*>(?=[ \\t]*(?:\\n|$))[\\s\\S]*?(?:(?:\\n[ 	]*)+\\n|$))", "i").replace("comment", Ud).replace("tag", Hd).replace("attribute", / +[a-zA-Z:_][\w.:-]*(?: *= *"[^"\n]*"| *= *'[^'\n]*'| *= *[^\s"'=<>`]+)?/).getRegex(), Gd = G(Ld).replace("hr", jd).replace("heading", " {0,3}#{1,6}(?:\\s|$)").replace("|lheading", "").replace("|table", "").replace("blockquote", " {0,3}>").replace("fences", " {0,3}(?:`{3,}(?=[^`\\n]*\\n)|~{3,})[^\\n]*\\n").replace("list", " {0,3}(?:[*+-]|1[.)])[ \\t]").replace("html", "</?(?:tag)(?: +|\\n|/?>)|<(?:script|pre|style|textarea|!--)").replace("tag", Hd).getRegex(), Kd = {
	blockquote: G(/^( {0,3}> ?(paragraph|[^\n]*)(?:\n|$))+/).replace("paragraph", Gd).getRegex(),
	code: kd,
	def: Bd,
	fences: Ad,
	heading: Md,
	hr: jd,
	html: Wd,
	lheading: Fd,
	list: Vd,
	newline: Od,
	paragraph: Gd,
	table: Ed,
	text: Rd
}, qd = G("^ *([^\\n ].*)\\n {0,3}((?:\\| *)?:?-+:? *(?:\\| *:?-+:? *)*(?:\\| *)?)(?:\\n((?:(?! *\\n|hr|heading|blockquote|code|fences|list|html).*(?:\\n|$))*)\\n*|$)").replace("hr", jd).replace("heading", " {0,3}#{1,6}(?:\\s|$)").replace("blockquote", " {0,3}>").replace("code", "(?: {4}| {0,3}	)[^\\n]").replace("fences", " {0,3}(?:`{3,}(?=[^`\\n]*\\n)|~{3,})[^\\n]*\\n").replace("list", " {0,3}(?:[*+-]|1[.)])[ \\t]").replace("html", "</?(?:tag)(?: +|\\n|/?>)|<(?:script|pre|style|textarea|!--)").replace("tag", Hd).getRegex(), Jd = {
	...Kd,
	lheading: Id,
	table: qd,
	paragraph: G(Ld).replace("hr", jd).replace("heading", " {0,3}#{1,6}(?:\\s|$)").replace("|lheading", "").replace("table", qd).replace("blockquote", " {0,3}>").replace("fences", " {0,3}(?:`{3,}(?=[^`\\n]*\\n)|~{3,})[^\\n]*\\n").replace("list", " {0,3}(?:[*+-]|1[.)])[ \\t]").replace("html", "</?(?:tag)(?: +|\\n|/?>)|<(?:script|pre|style|textarea|!--)").replace("tag", Hd).getRegex()
}, Yd = {
	...Kd,
	html: G("^ *(?:comment *(?:\\n|\\s*$)|<(tag)[\\s\\S]+?</\\1> *(?:\\n{2,}|\\s*$)|<tag(?:\"[^\"]*\"|'[^']*'|\\s[^'\"/>\\s]*)*?/?> *(?:\\n{2,}|\\s*$))").replace("comment", Ud).replace(/tag/g, "(?!(?:a|em|strong|small|s|cite|q|dfn|abbr|data|time|code|var|samp|kbd|sub|sup|i|b|u|mark|ruby|rt|rp|bdi|bdo|span|br|wbr|ins|del|img)\\b)\\w+(?!:|[^\\w\\s@]*@)\\b").getRegex(),
	def: /^ *\[([^\]]+)\]: *<?([^\s>]+)>?(?: +(["(][^\n]+[")]))? *(?:\n+|$)/,
	heading: /^(#{1,6})(.*)(?:\n+|$)/,
	fences: Ed,
	lheading: /^(.+?)\n {0,3}(=+|-+) *(?:\n+|$)/,
	paragraph: G(Ld).replace("hr", jd).replace("heading", " *#{1,6} *[^\n]").replace("lheading", Fd).replace("|table", "").replace("blockquote", " {0,3}>").replace("|fences", "").replace("|list", "").replace("|html", "").replace("|tag", "").getRegex()
}, Xd = /^\\([!"#$%&'()*+,\-./:;<=>?@\[\]\\^_`{|}~])/, Zd = /^(`+)([^`]|[^`][\s\S]*?[^`])\1(?!`)/, Qd = /^( {2,}|\\)\n(?!\s*$)/, $d = /^(`+|[^`])(?:(?= {2,}\n)|[\s\S]*?(?:(?=[\\<!\[`*_]|\b_|$)|[^ ](?= {2,}\n)))/, ef = /[\p{P}\p{S}]/u, tf = /[\s\p{P}\p{S}]/u, nf = /[^\s\p{P}\p{S}]/u, rf = G(/^((?![*_])punctSpace)/, "u").replace(/punctSpace/g, tf).getRegex(), af = /(?!~)[\p{P}\p{S}]/u, of = /(?!~)[\s\p{P}\p{S}]/u, sf = /(?:[^\s\p{P}\p{S}]|~)/u, cf = G(/link|precode-code|html/, "g").replace("link", /\[(?:[^\[\]`]|(?<a>`+)[^`]+\k<a>(?!`))*?\]\((?:\\[\s\S]|[^\\\(\)]|\((?:\\[\s\S]|[^\\\(\)])*\))*\)/).replace("precode-", Dd ? "(?<!`)()" : "(^^|[^`])").replace("code", /(?<b>`+)[^`]+\k<b>(?!`)/).replace("html", /<(?! )[^<>]*?>/).getRegex(), lf = /^(?:\*+(?:((?!\*)punct)|([^\s*]))?)|^_+(?:((?!_)punct)|([^\s_]))?/, uf = G(lf, "u").replace(/punct/g, ef).getRegex(), df = G(lf, "u").replace(/punct/g, af).getRegex(), ff = "^[^_*]*?__[^_*]*?\\*[^_*]*?(?=__)|[^*]+(?=[^*])|(?!\\*)punct(\\*+)(?=[\\s]|$)|notPunctSpace(\\*+)(?!\\*)(?=punctSpace|$)|(?!\\*)punctSpace(\\*+)(?=notPunctSpace)|[\\s](\\*+)(?!\\*)(?=punct)|(?!\\*)punct(\\*+)(?!\\*)(?=punct)|notPunctSpace(\\*+)(?=notPunctSpace)", pf = G(ff, "gu").replace(/notPunctSpace/g, nf).replace(/punctSpace/g, tf).replace(/punct/g, ef).getRegex(), mf = G(ff, "gu").replace(/notPunctSpace/g, sf).replace(/punctSpace/g, of).replace(/punct/g, af).getRegex(), hf = G("^[^_*]*?\\*\\*[^_*]*?_[^_*]*?(?=\\*\\*)|[^_]+(?=[^_])|(?!_)punct(_+)(?=[\\s]|$)|notPunctSpace(_+)(?!_)(?=punctSpace|$)|(?!_)punctSpace(_+)(?=notPunctSpace)|[\\s](_+)(?!_)(?=punct)|(?!_)punct(_+)(?!_)(?=punct)", "gu").replace(/notPunctSpace/g, nf).replace(/punctSpace/g, tf).replace(/punct/g, ef).getRegex(), gf = G(/^~~?(?:((?!~)punct)|[^\s~])/, "u").replace(/punct/g, ef).getRegex(), _f = G("^[^~]+(?=[^~])|(?!~)punct(~~?)(?=[\\s]|$)|notPunctSpace(~~?)(?!~)(?=punctSpace|$)|(?!~)punctSpace(~~?)(?=notPunctSpace)|[\\s](~~?)(?!~)(?=punct)|(?!~)punct(~~?)(?!~)(?=punct)|notPunctSpace(~~?)(?=notPunctSpace)", "gu").replace(/notPunctSpace/g, nf).replace(/punctSpace/g, tf).replace(/punct/g, ef).getRegex(), vf = G(/\\(punct)/, "gu").replace(/punct/g, ef).getRegex(), yf = G(/^<(scheme:[^\s\x00-\x1f<>]*|email)>/).replace("scheme", /[a-zA-Z][a-zA-Z0-9+.-]{1,31}/).replace("email", /[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+(@)[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)+(?![-_])/).getRegex(), bf = G(Ud).replace("(?:-->|$)", "-->").getRegex(), xf = G("^comment|^</[a-zA-Z][\\w:-]*\\s*>|^<[a-zA-Z][\\w-]*(?:attribute)*?\\s*/?>|^<\\?[\\s\\S]*?\\?>|^<![a-zA-Z]+\\s[\\s\\S]*?>|^<!\\[CDATA\\[[\\s\\S]*?\\]\\]>").replace("comment", bf).replace("attribute", /\s+[a-zA-Z:_][\w.:-]*(?:\s*=\s*"[^"]*"|\s*=\s*'[^']*'|\s*=\s*[^\s"'=<>`]+)?/).getRegex(), Sf = /(?:\[(?:\\[\s\S]|[^\[\]\\])*\]|\\[\s\S]|`+(?!`)[^`]*?`+(?!`)|``+(?=\])|[^\[\]\\`])*?/, Cf = G(/^!?\[(label)\]\(\s*(href)(?:(?:[ \t]+(?:\n[ \t]*)?|\n[ \t]*)(title))?\s*\)/).replace("label", Sf).replace("href", /<(?:\\.|[^\n<>\\])+>|[^ \t\n\x00-\x1f]*/).replace("title", /"(?:\\"?|[^"\\])*"|'(?:\\'?|[^'\\])*'|\((?:\\\)?|[^)\\])*\)/).getRegex(), wf = G(/^!?\[(label)\]\[(ref)\]/).replace("label", Sf).replace("ref", zd).getRegex(), Tf = G(/^!?\[(ref)\](?:\[\])?/).replace("ref", zd).getRegex(), Ef = G("reflink|nolink(?!\\()", "g").replace("reflink", wf).replace("nolink", Tf).getRegex(), Df = /[hH][tT][tT][pP][sS]?|[fF][tT][pP]/, Of = {
	_backpedal: Ed,
	anyPunctuation: vf,
	autolink: yf,
	blockSkip: cf,
	br: Qd,
	code: Zd,
	del: Ed,
	delLDelim: Ed,
	delRDelim: Ed,
	emStrongLDelim: uf,
	emStrongRDelimAst: pf,
	emStrongRDelimUnd: hf,
	escape: Xd,
	link: Cf,
	nolink: Tf,
	punctuation: rf,
	reflink: wf,
	reflinkSearch: Ef,
	tag: xf,
	text: $d,
	url: Ed
}, kf = {
	...Of,
	link: G(/^!?\[(label)\]\((.*?)\)/).replace("label", Sf).getRegex(),
	reflink: G(/^!?\[(label)\]\s*\[([^\]]*)\]/).replace("label", Sf).getRegex()
}, Af = {
	...Of,
	emStrongRDelimAst: mf,
	emStrongLDelim: df,
	delLDelim: gf,
	delRDelim: _f,
	url: G(/^((?:protocol):\/\/|www\.)(?:[a-zA-Z0-9\-]+\.?)+[^\s<]*|^email/).replace("protocol", Df).replace("email", /[A-Za-z0-9._+-]+(@)[a-zA-Z0-9-_]+(?:\.[a-zA-Z0-9-_]*[a-zA-Z0-9])+(?![-_])/).getRegex(),
	_backpedal: /(?:[^?!.,:;*_'"~()&]+|\([^)]*\)|&(?![a-zA-Z0-9]+;$)|[?!.,:;*_'"~)]+(?!$))+/,
	del: /^(~~?)(?=[^\s~])((?:\\[\s\S]|[^\\])*?(?:\\[\s\S]|[^\s~\\]))\1(?=[^~]|$)/,
	text: G(/^([`~]+|[^`~])(?:(?= {2,}\n)|(?=[a-zA-Z0-9.!#$%&'*+\/=?_`{\|}~-]+@)|[\s\S]*?(?:(?=[\\<!\[`*~_]|\b_|protocol:\/\/|www\.|$)|[^ ](?= {2,}\n)|[^a-zA-Z0-9.!#$%&'*+\/=?_`{\|}~-](?=[a-zA-Z0-9.!#$%&'*+\/=?_`{\|}~-]+@)))/).replace("protocol", Df).getRegex()
}, jf = {
	...Af,
	br: G(Qd).replace("{2,}", "*").getRegex(),
	text: G(Af.text).replace("\\b_", "\\b_| {2,}\\n").replace(/\{2,\}/g, "*").getRegex()
}, Mf = {
	normal: Kd,
	gfm: Jd,
	pedantic: Yd
}, Nf = {
	normal: Of,
	gfm: Af,
	breaks: jf,
	pedantic: kf
}, Pf = {
	"&": "&amp;",
	"<": "&lt;",
	">": "&gt;",
	"\"": "&quot;",
	"'": "&#39;"
}, Ff = (e) => Pf[e];
function If(e, t) {
	if (t) {
		if (K.escapeTest.test(e)) return e.replace(K.escapeReplace, Ff);
	} else if (K.escapeTestNoEncode.test(e)) return e.replace(K.escapeReplaceNoEncode, Ff);
	return e;
}
function Lf(e) {
	try {
		e = encodeURI(e).replace(K.percentDecode, "%");
	} catch {
		return null;
	}
	return e;
}
function Rf(e, t) {
	let n = e.replace(K.findPipe, (e, t, n) => {
		let r = !1, i = t;
		for (; --i >= 0 && n[i] === "\\";) r = !r;
		return r ? "|" : " |";
	}).split(K.splitPipe), r = 0;
	if (n[0].trim() || n.shift(), n.length > 0 && !n.at(-1)?.trim() && n.pop(), t) {
		if (n.length > t) n.splice(t);
		else for (; n.length < t;) n.push("");
	}
	for (; r < n.length; r++) n[r] = n[r].trim().replace(K.slashPipe, "|");
	return n;
}
function zf(e, t, n) {
	let r = e.length;
	if (r === 0) return "";
	let i = 0;
	for (; i < r;) {
		let a = e.charAt(r - i - 1);
		if (a === t && !n) i++;
		else if (a !== t && n) i++;
		else break;
	}
	return e.slice(0, r - i);
}
function Bf(e, t) {
	if (e.indexOf(t[1]) === -1) return -1;
	let n = 0;
	for (let r = 0; r < e.length; r++) if (e[r] === "\\") r++;
	else if (e[r] === t[0]) n++;
	else if (e[r] === t[1] && (n--, n < 0)) return r;
	return n > 0 ? -2 : -1;
}
function Vf(e, t = 0) {
	let n = t, r = "";
	for (let t of e) if (t === "	") {
		let e = 4 - n % 4;
		r += " ".repeat(e), n += e;
	} else r += t, n++;
	return r;
}
function Hf(e, t, n, r, i) {
	let a = t.href, o = t.title || null, s = e[1].replace(i.other.outputLinkReplace, "$1");
	r.state.inLink = !0;
	let c = {
		type: e[0].charAt(0) === "!" ? "image" : "link",
		raw: n,
		href: a,
		title: o,
		text: s,
		tokens: r.inlineTokens(s)
	};
	return r.state.inLink = !1, c;
}
function Uf(e, t, n) {
	let r = e.match(n.other.indentCodeCompensation);
	if (r === null) return t;
	let i = r[1];
	return t.split("\n").map((e) => {
		let t = e.match(n.other.beginningSpace);
		if (t === null) return e;
		let [r] = t;
		return r.length >= i.length ? e.slice(i.length) : e;
	}).join("\n");
}
var Wf = class {
	options;
	rules;
	lexer;
	constructor(e) {
		this.options = e || wd;
	}
	space(e) {
		let t = this.rules.block.newline.exec(e);
		if (t && t[0].length > 0) return {
			type: "space",
			raw: t[0]
		};
	}
	code(e) {
		let t = this.rules.block.code.exec(e);
		if (t) {
			let e = t[0].replace(this.rules.other.codeRemoveIndent, "");
			return {
				type: "code",
				raw: t[0],
				codeBlockStyle: "indented",
				text: this.options.pedantic ? e : zf(e, "\n")
			};
		}
	}
	fences(e) {
		let t = this.rules.block.fences.exec(e);
		if (t) {
			let e = t[0], n = Uf(e, t[3] || "", this.rules);
			return {
				type: "code",
				raw: e,
				lang: t[2] ? t[2].trim().replace(this.rules.inline.anyPunctuation, "$1") : t[2],
				text: n
			};
		}
	}
	heading(e) {
		let t = this.rules.block.heading.exec(e);
		if (t) {
			let e = t[2].trim();
			if (this.rules.other.endingHash.test(e)) {
				let t = zf(e, "#");
				(this.options.pedantic || !t || this.rules.other.endingSpaceChar.test(t)) && (e = t.trim());
			}
			return {
				type: "heading",
				raw: t[0],
				depth: t[1].length,
				text: e,
				tokens: this.lexer.inline(e)
			};
		}
	}
	hr(e) {
		let t = this.rules.block.hr.exec(e);
		if (t) return {
			type: "hr",
			raw: zf(t[0], "\n")
		};
	}
	blockquote(e) {
		let t = this.rules.block.blockquote.exec(e);
		if (t) {
			let e = zf(t[0], "\n").split("\n"), n = "", r = "", i = [];
			for (; e.length > 0;) {
				let t = !1, a = [], o = 0;
				for (; o < e.length; o++) if (this.rules.other.blockquoteStart.test(e[o])) a.push(e[o]), t = !0;
				else if (!t) a.push(e[o]);
				else break;
				e = e.slice(o);
				let s = a.join("\n"), c = s.replace(this.rules.other.blockquoteSetextReplace, "\n    $1").replace(this.rules.other.blockquoteSetextReplace2, "");
				n = n ? `${n}
${s}` : s, r = r ? `${r}
${c}` : c;
				let l = this.lexer.state.top;
				if (this.lexer.state.top = !0, this.lexer.blockTokens(c, i, !0), this.lexer.state.top = l, e.length === 0) break;
				let u = i.at(-1);
				if (u?.type === "code") break;
				if (u?.type === "blockquote") {
					let t = u, a = t.raw + "\n" + e.join("\n"), o = this.blockquote(a);
					i[i.length - 1] = o, n = n.substring(0, n.length - t.raw.length) + o.raw, r = r.substring(0, r.length - t.text.length) + o.text;
					break;
				}
				if (u?.type === "list") {
					let t = u, a = t.raw + "\n" + e.join("\n"), o = this.list(a);
					i[i.length - 1] = o, n = n.substring(0, n.length - u.raw.length) + o.raw, r = r.substring(0, r.length - t.raw.length) + o.raw, e = a.substring(i.at(-1).raw.length).split("\n");
				}
			}
			return {
				type: "blockquote",
				raw: n,
				tokens: i,
				text: r
			};
		}
	}
	list(e) {
		let t = this.rules.block.list.exec(e);
		if (t) {
			let n = t[1].trim(), r = n.length > 1, i = {
				type: "list",
				raw: "",
				ordered: r,
				start: r ? +n.slice(0, -1) : "",
				loose: !1,
				items: []
			};
			n = r ? `\\d{1,9}\\${n.slice(-1)}` : `\\${n}`, this.options.pedantic && (n = r ? n : "[*+-]");
			let a = this.rules.other.listItemRegex(n), o = !1;
			for (; e;) {
				let n = !1, r = "", s = "";
				if (!(t = a.exec(e)) || this.rules.block.hr.test(e)) break;
				r = t[0], e = e.substring(r.length);
				let c = Vf(t[2].split("\n", 1)[0], t[1].length), l = e.split("\n", 1)[0], u = !c.trim(), d = 0;
				if (this.options.pedantic ? (d = 2, s = c.trimStart()) : u ? d = t[1].length + 1 : (d = c.search(this.rules.other.nonSpaceChar), d = d > 4 ? 1 : d, s = c.slice(d), d += t[1].length), u && this.rules.other.blankLine.test(l) && (r += l + "\n", e = e.substring(l.length + 1), n = !0), !n) {
					let t = this.rules.other.nextBulletRegex(d), n = this.rules.other.hrRegex(d), i = this.rules.other.fencesBeginRegex(d), a = this.rules.other.headingBeginRegex(d), o = this.rules.other.htmlBeginRegex(d), f = this.rules.other.blockquoteBeginRegex(d);
					for (; e;) {
						let p = e.split("\n", 1)[0], m;
						if (l = p, this.options.pedantic ? (l = l.replace(this.rules.other.listReplaceNesting, "  "), m = l) : m = l.replace(this.rules.other.tabCharGlobal, "    "), i.test(l) || a.test(l) || o.test(l) || f.test(l) || t.test(l) || n.test(l)) break;
						if (m.search(this.rules.other.nonSpaceChar) >= d || !l.trim()) s += "\n" + m.slice(d);
						else {
							if (u || c.replace(this.rules.other.tabCharGlobal, "    ").search(this.rules.other.nonSpaceChar) >= 4 || i.test(c) || a.test(c) || n.test(c)) break;
							s += "\n" + l;
						}
						u = !l.trim(), r += p + "\n", e = e.substring(p.length + 1), c = m.slice(d);
					}
				}
				i.loose || (o ? i.loose = !0 : this.rules.other.doubleBlankLine.test(r) && (o = !0)), i.items.push({
					type: "list_item",
					raw: r,
					task: !!this.options.gfm && this.rules.other.listIsTask.test(s),
					loose: !1,
					text: s,
					tokens: []
				}), i.raw += r;
			}
			let s = i.items.at(-1);
			if (s) s.raw = s.raw.trimEnd(), s.text = s.text.trimEnd();
			else return;
			i.raw = i.raw.trimEnd();
			for (let e of i.items) {
				if (this.lexer.state.top = !1, e.tokens = this.lexer.blockTokens(e.text, []), e.task) {
					if (e.text = e.text.replace(this.rules.other.listReplaceTask, ""), e.tokens[0]?.type === "text" || e.tokens[0]?.type === "paragraph") {
						e.tokens[0].raw = e.tokens[0].raw.replace(this.rules.other.listReplaceTask, ""), e.tokens[0].text = e.tokens[0].text.replace(this.rules.other.listReplaceTask, "");
						for (let e = this.lexer.inlineQueue.length - 1; e >= 0; e--) if (this.rules.other.listIsTask.test(this.lexer.inlineQueue[e].src)) {
							this.lexer.inlineQueue[e].src = this.lexer.inlineQueue[e].src.replace(this.rules.other.listReplaceTask, "");
							break;
						}
					}
					let t = this.rules.other.listTaskCheckbox.exec(e.raw);
					if (t) {
						let n = {
							type: "checkbox",
							raw: t[0] + " ",
							checked: t[0] !== "[ ]"
						};
						e.checked = n.checked, i.loose ? e.tokens[0] && ["paragraph", "text"].includes(e.tokens[0].type) && "tokens" in e.tokens[0] && e.tokens[0].tokens ? (e.tokens[0].raw = n.raw + e.tokens[0].raw, e.tokens[0].text = n.raw + e.tokens[0].text, e.tokens[0].tokens.unshift(n)) : e.tokens.unshift({
							type: "paragraph",
							raw: n.raw,
							text: n.raw,
							tokens: [n]
						}) : e.tokens.unshift(n);
					}
				}
				if (!i.loose) {
					let t = e.tokens.filter((e) => e.type === "space");
					i.loose = t.length > 0 && t.some((e) => this.rules.other.anyLine.test(e.raw));
				}
			}
			if (i.loose) for (let e of i.items) {
				e.loose = !0;
				for (let t of e.tokens) t.type === "text" && (t.type = "paragraph");
			}
			return i;
		}
	}
	html(e) {
		let t = this.rules.block.html.exec(e);
		if (t) return {
			type: "html",
			block: !0,
			raw: t[0],
			pre: t[1] === "pre" || t[1] === "script" || t[1] === "style",
			text: t[0]
		};
	}
	def(e) {
		let t = this.rules.block.def.exec(e);
		if (t) {
			let e = t[1].toLowerCase().replace(this.rules.other.multipleSpaceGlobal, " "), n = t[2] ? t[2].replace(this.rules.other.hrefBrackets, "$1").replace(this.rules.inline.anyPunctuation, "$1") : "", r = t[3] ? t[3].substring(1, t[3].length - 1).replace(this.rules.inline.anyPunctuation, "$1") : t[3];
			return {
				type: "def",
				tag: e,
				raw: t[0],
				href: n,
				title: r
			};
		}
	}
	table(e) {
		let t = this.rules.block.table.exec(e);
		if (!t || !this.rules.other.tableDelimiter.test(t[2])) return;
		let n = Rf(t[1]), r = t[2].replace(this.rules.other.tableAlignChars, "").split("|"), i = t[3]?.trim() ? t[3].replace(this.rules.other.tableRowBlankLine, "").split("\n") : [], a = {
			type: "table",
			raw: t[0],
			header: [],
			align: [],
			rows: []
		};
		if (n.length === r.length) {
			for (let e of r) this.rules.other.tableAlignRight.test(e) ? a.align.push("right") : this.rules.other.tableAlignCenter.test(e) ? a.align.push("center") : this.rules.other.tableAlignLeft.test(e) ? a.align.push("left") : a.align.push(null);
			for (let e = 0; e < n.length; e++) a.header.push({
				text: n[e],
				tokens: this.lexer.inline(n[e]),
				header: !0,
				align: a.align[e]
			});
			for (let e of i) a.rows.push(Rf(e, a.header.length).map((e, t) => ({
				text: e,
				tokens: this.lexer.inline(e),
				header: !1,
				align: a.align[t]
			})));
			return a;
		}
	}
	lheading(e) {
		let t = this.rules.block.lheading.exec(e);
		if (t) {
			let e = t[1].trim();
			return {
				type: "heading",
				raw: t[0],
				depth: t[2].charAt(0) === "=" ? 1 : 2,
				text: e,
				tokens: this.lexer.inline(e)
			};
		}
	}
	paragraph(e) {
		let t = this.rules.block.paragraph.exec(e);
		if (t) {
			let e = t[1].charAt(t[1].length - 1) === "\n" ? t[1].slice(0, -1) : t[1];
			return {
				type: "paragraph",
				raw: t[0],
				text: e,
				tokens: this.lexer.inline(e)
			};
		}
	}
	text(e) {
		let t = this.rules.block.text.exec(e);
		if (t) return {
			type: "text",
			raw: t[0],
			text: t[0],
			tokens: this.lexer.inline(t[0])
		};
	}
	escape(e) {
		let t = this.rules.inline.escape.exec(e);
		if (t) return {
			type: "escape",
			raw: t[0],
			text: t[1]
		};
	}
	tag(e) {
		let t = this.rules.inline.tag.exec(e);
		if (t) return !this.lexer.state.inLink && this.rules.other.startATag.test(t[0]) ? this.lexer.state.inLink = !0 : this.lexer.state.inLink && this.rules.other.endATag.test(t[0]) && (this.lexer.state.inLink = !1), !this.lexer.state.inRawBlock && this.rules.other.startPreScriptTag.test(t[0]) ? this.lexer.state.inRawBlock = !0 : this.lexer.state.inRawBlock && this.rules.other.endPreScriptTag.test(t[0]) && (this.lexer.state.inRawBlock = !1), {
			type: "html",
			raw: t[0],
			inLink: this.lexer.state.inLink,
			inRawBlock: this.lexer.state.inRawBlock,
			block: !1,
			text: t[0]
		};
	}
	link(e) {
		let t = this.rules.inline.link.exec(e);
		if (t) {
			let e = t[2].trim();
			if (!this.options.pedantic && this.rules.other.startAngleBracket.test(e)) {
				if (!this.rules.other.endAngleBracket.test(e)) return;
				let t = zf(e.slice(0, -1), "\\");
				if ((e.length - t.length) % 2 == 0) return;
			} else {
				let e = Bf(t[2], "()");
				if (e === -2) return;
				if (e > -1) {
					let n = (t[0].indexOf("!") === 0 ? 5 : 4) + t[1].length + e;
					t[2] = t[2].substring(0, e), t[0] = t[0].substring(0, n).trim(), t[3] = "";
				}
			}
			let n = t[2], r = "";
			if (this.options.pedantic) {
				let e = this.rules.other.pedanticHrefTitle.exec(n);
				e && (n = e[1], r = e[3]);
			} else r = t[3] ? t[3].slice(1, -1) : "";
			return n = n.trim(), this.rules.other.startAngleBracket.test(n) && (n = this.options.pedantic && !this.rules.other.endAngleBracket.test(e) ? n.slice(1) : n.slice(1, -1)), Hf(t, {
				href: n && n.replace(this.rules.inline.anyPunctuation, "$1"),
				title: r && r.replace(this.rules.inline.anyPunctuation, "$1")
			}, t[0], this.lexer, this.rules);
		}
	}
	reflink(e, t) {
		let n;
		if ((n = this.rules.inline.reflink.exec(e)) || (n = this.rules.inline.nolink.exec(e))) {
			let e = t[(n[2] || n[1]).replace(this.rules.other.multipleSpaceGlobal, " ").toLowerCase()];
			if (!e) {
				let e = n[0].charAt(0);
				return {
					type: "text",
					raw: e,
					text: e
				};
			}
			return Hf(n, e, n[0], this.lexer, this.rules);
		}
	}
	emStrong(e, t, n = "") {
		let r = this.rules.inline.emStrongLDelim.exec(e);
		if (!(!r || !r[1] && !r[2] && !r[3] && !r[4] || r[4] && n.match(this.rules.other.unicodeAlphaNumeric)) && (!(r[1] || r[3]) || !n || this.rules.inline.punctuation.exec(n))) {
			let n = [...r[0]].length - 1, i, a, o = n, s = 0, c = r[0][0] === "*" ? this.rules.inline.emStrongRDelimAst : this.rules.inline.emStrongRDelimUnd;
			for (c.lastIndex = 0, t = t.slice(-1 * e.length + n); (r = c.exec(t)) !== null;) {
				if (i = r[1] || r[2] || r[3] || r[4] || r[5] || r[6], !i) continue;
				if (a = [...i].length, r[3] || r[4]) {
					o += a;
					continue;
				}
				if ((r[5] || r[6]) && n % 3 && !((n + a) % 3)) {
					s += a;
					continue;
				}
				if (o -= a, o > 0) continue;
				a = Math.min(a, a + o + s);
				let t = [...r[0]][0].length, c = e.slice(0, n + r.index + t + a);
				if (Math.min(n, a) % 2) {
					let e = c.slice(1, -1);
					return {
						type: "em",
						raw: c,
						text: e,
						tokens: this.lexer.inlineTokens(e)
					};
				}
				let l = c.slice(2, -2);
				return {
					type: "strong",
					raw: c,
					text: l,
					tokens: this.lexer.inlineTokens(l)
				};
			}
		}
	}
	codespan(e) {
		let t = this.rules.inline.code.exec(e);
		if (t) {
			let e = t[2].replace(this.rules.other.newLineCharGlobal, " "), n = this.rules.other.nonSpaceChar.test(e), r = this.rules.other.startingSpaceChar.test(e) && this.rules.other.endingSpaceChar.test(e);
			return n && r && (e = e.substring(1, e.length - 1)), {
				type: "codespan",
				raw: t[0],
				text: e
			};
		}
	}
	br(e) {
		let t = this.rules.inline.br.exec(e);
		if (t) return {
			type: "br",
			raw: t[0]
		};
	}
	del(e, t, n = "") {
		let r = this.rules.inline.delLDelim.exec(e);
		if (r && (!r[1] || !n || this.rules.inline.punctuation.exec(n))) {
			let n = [...r[0]].length - 1, i, a, o = n, s = this.rules.inline.delRDelim;
			for (s.lastIndex = 0, t = t.slice(-1 * e.length + n); (r = s.exec(t)) !== null;) {
				if (i = r[1] || r[2] || r[3] || r[4] || r[5] || r[6], !i || (a = [...i].length, a !== n)) continue;
				if (r[3] || r[4]) {
					o += a;
					continue;
				}
				if (o -= a, o > 0) continue;
				a = Math.min(a, a + o);
				let t = [...r[0]][0].length, s = e.slice(0, n + r.index + t + a), c = s.slice(n, -n);
				return {
					type: "del",
					raw: s,
					text: c,
					tokens: this.lexer.inlineTokens(c)
				};
			}
		}
	}
	autolink(e) {
		let t = this.rules.inline.autolink.exec(e);
		if (t) {
			let e, n;
			return t[2] === "@" ? (e = t[1], n = "mailto:" + e) : (e = t[1], n = e), {
				type: "link",
				raw: t[0],
				text: e,
				href: n,
				tokens: [{
					type: "text",
					raw: e,
					text: e
				}]
			};
		}
	}
	url(e) {
		let t;
		if (t = this.rules.inline.url.exec(e)) {
			let e, n;
			if (t[2] === "@") e = t[0], n = "mailto:" + e;
			else {
				let r;
				do
					r = t[0], t[0] = this.rules.inline._backpedal.exec(t[0])?.[0] ?? "";
				while (r !== t[0]);
				e = t[0], n = t[1] === "www." ? "http://" + t[0] : t[0];
			}
			return {
				type: "link",
				raw: t[0],
				text: e,
				href: n,
				tokens: [{
					type: "text",
					raw: e,
					text: e
				}]
			};
		}
	}
	inlineText(e) {
		let t = this.rules.inline.text.exec(e);
		if (t) {
			let e = this.lexer.state.inRawBlock;
			return {
				type: "text",
				raw: t[0],
				text: t[0],
				escaped: e
			};
		}
	}
}, q = class e {
	tokens;
	options;
	state;
	inlineQueue;
	tokenizer;
	constructor(e) {
		this.tokens = [], this.tokens.links = Object.create(null), this.options = e || wd, this.options.tokenizer = this.options.tokenizer || new Wf(), this.tokenizer = this.options.tokenizer, this.tokenizer.options = this.options, this.tokenizer.lexer = this, this.inlineQueue = [], this.state = {
			inLink: !1,
			inRawBlock: !1,
			top: !0
		};
		let t = {
			other: K,
			block: Mf.normal,
			inline: Nf.normal
		};
		this.options.pedantic ? (t.block = Mf.pedantic, t.inline = Nf.pedantic) : this.options.gfm && (t.block = Mf.gfm, t.inline = this.options.breaks ? Nf.breaks : Nf.gfm), this.tokenizer.rules = t;
	}
	static get rules() {
		return {
			block: Mf,
			inline: Nf
		};
	}
	static lex(t, n) {
		return new e(n).lex(t);
	}
	static lexInline(t, n) {
		return new e(n).inlineTokens(t);
	}
	lex(e) {
		e = e.replace(K.carriageReturn, "\n"), this.blockTokens(e, this.tokens);
		for (let e = 0; e < this.inlineQueue.length; e++) {
			let t = this.inlineQueue[e];
			this.inlineTokens(t.src, t.tokens);
		}
		return this.inlineQueue = [], this.tokens;
	}
	blockTokens(e, t = [], n = !1) {
		for (this.tokenizer.lexer = this, this.options.pedantic && (e = e.replace(K.tabCharGlobal, "    ").replace(K.spaceLine, "")); e;) {
			let r;
			if (this.options.extensions?.block?.some((n) => (r = n.call({ lexer: this }, e, t)) ? (e = e.substring(r.raw.length), t.push(r), !0) : !1)) continue;
			if (r = this.tokenizer.space(e)) {
				e = e.substring(r.raw.length);
				let n = t.at(-1);
				r.raw.length === 1 && n !== void 0 ? n.raw += "\n" : t.push(r);
				continue;
			}
			if (r = this.tokenizer.code(e)) {
				e = e.substring(r.raw.length);
				let n = t.at(-1);
				n?.type === "paragraph" || n?.type === "text" ? (n.raw += (n.raw.endsWith("\n") ? "" : "\n") + r.raw, n.text += "\n" + r.text, this.inlineQueue.at(-1).src = n.text) : t.push(r);
				continue;
			}
			if ((r = this.tokenizer.fences(e)) || (r = this.tokenizer.heading(e)) || (r = this.tokenizer.hr(e)) || (r = this.tokenizer.blockquote(e)) || (r = this.tokenizer.list(e)) || (r = this.tokenizer.html(e))) {
				e = e.substring(r.raw.length), t.push(r);
				continue;
			}
			if (r = this.tokenizer.def(e)) {
				e = e.substring(r.raw.length);
				let n = t.at(-1);
				n?.type === "paragraph" || n?.type === "text" ? (n.raw += (n.raw.endsWith("\n") ? "" : "\n") + r.raw, n.text += "\n" + r.raw, this.inlineQueue.at(-1).src = n.text) : this.tokens.links[r.tag] || (this.tokens.links[r.tag] = {
					href: r.href,
					title: r.title
				}, t.push(r));
				continue;
			}
			if ((r = this.tokenizer.table(e)) || (r = this.tokenizer.lheading(e))) {
				e = e.substring(r.raw.length), t.push(r);
				continue;
			}
			let i = e;
			if (this.options.extensions?.startBlock) {
				let t = 1 / 0, n = e.slice(1), r;
				this.options.extensions.startBlock.forEach((e) => {
					r = e.call({ lexer: this }, n), typeof r == "number" && r >= 0 && (t = Math.min(t, r));
				}), t < 1 / 0 && t >= 0 && (i = e.substring(0, t + 1));
			}
			if (this.state.top && (r = this.tokenizer.paragraph(i))) {
				let a = t.at(-1);
				n && a?.type === "paragraph" ? (a.raw += (a.raw.endsWith("\n") ? "" : "\n") + r.raw, a.text += "\n" + r.text, this.inlineQueue.pop(), this.inlineQueue.at(-1).src = a.text) : t.push(r), n = i.length !== e.length, e = e.substring(r.raw.length);
			} else if (r = this.tokenizer.text(e)) {
				e = e.substring(r.raw.length);
				let n = t.at(-1);
				n?.type === "text" ? (n.raw += (n.raw.endsWith("\n") ? "" : "\n") + r.raw, n.text += "\n" + r.text, this.inlineQueue.pop(), this.inlineQueue.at(-1).src = n.text) : t.push(r);
			} else if (e) {
				let t = "Infinite loop on byte: " + e.charCodeAt(0);
				if (this.options.silent) {
					console.error(t);
					break;
				}
				throw Error(t);
			}
		}
		return this.state.top = !0, t;
	}
	inline(e, t = []) {
		return this.inlineQueue.push({
			src: e,
			tokens: t
		}), t;
	}
	inlineTokens(e, t = []) {
		this.tokenizer.lexer = this;
		let n = e, r = null;
		if (this.tokens.links) {
			let e = Object.keys(this.tokens.links);
			if (e.length > 0) for (; (r = this.tokenizer.rules.inline.reflinkSearch.exec(n)) !== null;) e.includes(r[0].slice(r[0].lastIndexOf("[") + 1, -1)) && (n = n.slice(0, r.index) + "[" + "a".repeat(r[0].length - 2) + "]" + n.slice(this.tokenizer.rules.inline.reflinkSearch.lastIndex));
		}
		for (; (r = this.tokenizer.rules.inline.anyPunctuation.exec(n)) !== null;) n = n.slice(0, r.index) + "++" + n.slice(this.tokenizer.rules.inline.anyPunctuation.lastIndex);
		let i;
		for (; (r = this.tokenizer.rules.inline.blockSkip.exec(n)) !== null;) i = r[2] ? r[2].length : 0, n = n.slice(0, r.index + i) + "[" + "a".repeat(r[0].length - i - 2) + "]" + n.slice(this.tokenizer.rules.inline.blockSkip.lastIndex);
		n = this.options.hooks?.emStrongMask?.call({ lexer: this }, n) ?? n;
		let a = !1, o = "";
		for (; e;) {
			a || (o = ""), a = !1;
			let r;
			if (this.options.extensions?.inline?.some((n) => (r = n.call({ lexer: this }, e, t)) ? (e = e.substring(r.raw.length), t.push(r), !0) : !1)) continue;
			if ((r = this.tokenizer.escape(e)) || (r = this.tokenizer.tag(e)) || (r = this.tokenizer.link(e))) {
				e = e.substring(r.raw.length), t.push(r);
				continue;
			}
			if (r = this.tokenizer.reflink(e, this.tokens.links)) {
				e = e.substring(r.raw.length);
				let n = t.at(-1);
				r.type === "text" && n?.type === "text" ? (n.raw += r.raw, n.text += r.text) : t.push(r);
				continue;
			}
			if ((r = this.tokenizer.emStrong(e, n, o)) || (r = this.tokenizer.codespan(e)) || (r = this.tokenizer.br(e)) || (r = this.tokenizer.del(e, n, o)) || (r = this.tokenizer.autolink(e)) || !this.state.inLink && (r = this.tokenizer.url(e))) {
				e = e.substring(r.raw.length), t.push(r);
				continue;
			}
			let i = e;
			if (this.options.extensions?.startInline) {
				let t = 1 / 0, n = e.slice(1), r;
				this.options.extensions.startInline.forEach((e) => {
					r = e.call({ lexer: this }, n), typeof r == "number" && r >= 0 && (t = Math.min(t, r));
				}), t < 1 / 0 && t >= 0 && (i = e.substring(0, t + 1));
			}
			if (r = this.tokenizer.inlineText(i)) {
				e = e.substring(r.raw.length), r.raw.slice(-1) !== "_" && (o = r.raw.slice(-1)), a = !0;
				let n = t.at(-1);
				n?.type === "text" ? (n.raw += r.raw, n.text += r.text) : t.push(r);
			} else if (e) {
				let t = "Infinite loop on byte: " + e.charCodeAt(0);
				if (this.options.silent) {
					console.error(t);
					break;
				}
				throw Error(t);
			}
		}
		return t;
	}
}, Gf = class {
	options;
	parser;
	constructor(e) {
		this.options = e || wd;
	}
	space(e) {
		return "";
	}
	code({ text: e, lang: t, escaped: n }) {
		let r = (t || "").match(K.notSpaceStart)?.[0], i = e.replace(K.endingNewline, "") + "\n";
		return r ? "<pre><code class=\"language-" + If(r) + "\">" + (n ? i : If(i, !0)) + "</code></pre>\n" : "<pre><code>" + (n ? i : If(i, !0)) + "</code></pre>\n";
	}
	blockquote({ tokens: e }) {
		return `<blockquote>
${this.parser.parse(e)}</blockquote>
`;
	}
	html({ text: e }) {
		return e;
	}
	def(e) {
		return "";
	}
	heading({ tokens: e, depth: t }) {
		return `<h${t}>${this.parser.parseInline(e)}</h${t}>
`;
	}
	hr(e) {
		return "<hr>\n";
	}
	list(e) {
		let t = e.ordered, n = e.start, r = "";
		for (let t = 0; t < e.items.length; t++) {
			let n = e.items[t];
			r += this.listitem(n);
		}
		let i = t ? "ol" : "ul", a = t && n !== 1 ? " start=\"" + n + "\"" : "";
		return "<" + i + a + ">\n" + r + "</" + i + ">\n";
	}
	listitem(e) {
		return `<li>${this.parser.parse(e.tokens)}</li>
`;
	}
	checkbox({ checked: e }) {
		return "<input " + (e ? "checked=\"\" " : "") + "disabled=\"\" type=\"checkbox\"> ";
	}
	paragraph({ tokens: e }) {
		return `<p>${this.parser.parseInline(e)}</p>
`;
	}
	table(e) {
		let t = "", n = "";
		for (let t = 0; t < e.header.length; t++) n += this.tablecell(e.header[t]);
		t += this.tablerow({ text: n });
		let r = "";
		for (let t = 0; t < e.rows.length; t++) {
			let i = e.rows[t];
			n = "";
			for (let e = 0; e < i.length; e++) n += this.tablecell(i[e]);
			r += this.tablerow({ text: n });
		}
		return r &&= `<tbody>${r}</tbody>`, "<table>\n<thead>\n" + t + "</thead>\n" + r + "</table>\n";
	}
	tablerow({ text: e }) {
		return `<tr>
${e}</tr>
`;
	}
	tablecell(e) {
		let t = this.parser.parseInline(e.tokens), n = e.header ? "th" : "td";
		return (e.align ? `<${n} align="${e.align}">` : `<${n}>`) + t + `</${n}>
`;
	}
	strong({ tokens: e }) {
		return `<strong>${this.parser.parseInline(e)}</strong>`;
	}
	em({ tokens: e }) {
		return `<em>${this.parser.parseInline(e)}</em>`;
	}
	codespan({ text: e }) {
		return `<code>${If(e, !0)}</code>`;
	}
	br(e) {
		return "<br>";
	}
	del({ tokens: e }) {
		return `<del>${this.parser.parseInline(e)}</del>`;
	}
	link({ href: e, title: t, tokens: n }) {
		let r = this.parser.parseInline(n), i = Lf(e);
		if (i === null) return r;
		e = i;
		let a = "<a href=\"" + e + "\"";
		return t && (a += " title=\"" + If(t) + "\""), a += ">" + r + "</a>", a;
	}
	image({ href: e, title: t, text: n, tokens: r }) {
		r && (n = this.parser.parseInline(r, this.parser.textRenderer));
		let i = Lf(e);
		if (i === null) return If(n);
		e = i;
		let a = `<img src="${e}" alt="${If(n)}"`;
		return t && (a += ` title="${If(t)}"`), a += ">", a;
	}
	text(e) {
		return "tokens" in e && e.tokens ? this.parser.parseInline(e.tokens) : "escaped" in e && e.escaped ? e.text : If(e.text);
	}
}, Kf = class {
	strong({ text: e }) {
		return e;
	}
	em({ text: e }) {
		return e;
	}
	codespan({ text: e }) {
		return e;
	}
	del({ text: e }) {
		return e;
	}
	html({ text: e }) {
		return e;
	}
	text({ text: e }) {
		return e;
	}
	link({ text: e }) {
		return "" + e;
	}
	image({ text: e }) {
		return "" + e;
	}
	br() {
		return "";
	}
	checkbox({ raw: e }) {
		return e;
	}
}, J = class e {
	options;
	renderer;
	textRenderer;
	constructor(e) {
		this.options = e || wd, this.options.renderer = this.options.renderer || new Gf(), this.renderer = this.options.renderer, this.renderer.options = this.options, this.renderer.parser = this, this.textRenderer = new Kf();
	}
	static parse(t, n) {
		return new e(n).parse(t);
	}
	static parseInline(t, n) {
		return new e(n).parseInline(t);
	}
	parse(e) {
		this.renderer.parser = this;
		let t = "";
		for (let n = 0; n < e.length; n++) {
			let r = e[n];
			if (this.options.extensions?.renderers?.[r.type]) {
				let e = r, n = this.options.extensions.renderers[e.type].call({ parser: this }, e);
				if (n !== !1 || ![
					"space",
					"hr",
					"heading",
					"code",
					"table",
					"blockquote",
					"list",
					"html",
					"def",
					"paragraph",
					"text"
				].includes(e.type)) {
					t += n || "";
					continue;
				}
			}
			let i = r;
			switch (i.type) {
				case "space":
					t += this.renderer.space(i);
					break;
				case "hr":
					t += this.renderer.hr(i);
					break;
				case "heading":
					t += this.renderer.heading(i);
					break;
				case "code":
					t += this.renderer.code(i);
					break;
				case "table":
					t += this.renderer.table(i);
					break;
				case "blockquote":
					t += this.renderer.blockquote(i);
					break;
				case "list":
					t += this.renderer.list(i);
					break;
				case "checkbox":
					t += this.renderer.checkbox(i);
					break;
				case "html":
					t += this.renderer.html(i);
					break;
				case "def":
					t += this.renderer.def(i);
					break;
				case "paragraph":
					t += this.renderer.paragraph(i);
					break;
				case "text":
					t += this.renderer.text(i);
					break;
				default: {
					let e = "Token with \"" + i.type + "\" type was not found.";
					if (this.options.silent) return console.error(e), "";
					throw Error(e);
				}
			}
		}
		return t;
	}
	parseInline(e, t = this.renderer) {
		this.renderer.parser = this;
		let n = "";
		for (let r = 0; r < e.length; r++) {
			let i = e[r];
			if (this.options.extensions?.renderers?.[i.type]) {
				let e = this.options.extensions.renderers[i.type].call({ parser: this }, i);
				if (e !== !1 || ![
					"escape",
					"html",
					"link",
					"image",
					"strong",
					"em",
					"codespan",
					"br",
					"del",
					"text"
				].includes(i.type)) {
					n += e || "";
					continue;
				}
			}
			let a = i;
			switch (a.type) {
				case "escape":
					n += t.text(a);
					break;
				case "html":
					n += t.html(a);
					break;
				case "link":
					n += t.link(a);
					break;
				case "image":
					n += t.image(a);
					break;
				case "checkbox":
					n += t.checkbox(a);
					break;
				case "strong":
					n += t.strong(a);
					break;
				case "em":
					n += t.em(a);
					break;
				case "codespan":
					n += t.codespan(a);
					break;
				case "br":
					n += t.br(a);
					break;
				case "del":
					n += t.del(a);
					break;
				case "text":
					n += t.text(a);
					break;
				default: {
					let e = "Token with \"" + a.type + "\" type was not found.";
					if (this.options.silent) return console.error(e), "";
					throw Error(e);
				}
			}
		}
		return n;
	}
}, qf = class {
	options;
	block;
	constructor(e) {
		this.options = e || wd;
	}
	static passThroughHooks = /* @__PURE__ */ new Set([
		"preprocess",
		"postprocess",
		"processAllTokens",
		"emStrongMask"
	]);
	static passThroughHooksRespectAsync = /* @__PURE__ */ new Set([
		"preprocess",
		"postprocess",
		"processAllTokens"
	]);
	preprocess(e) {
		return e;
	}
	postprocess(e) {
		return e;
	}
	processAllTokens(e) {
		return e;
	}
	emStrongMask(e) {
		return e;
	}
	provideLexer(e = this.block) {
		return e ? q.lex : q.lexInline;
	}
	provideParser(e = this.block) {
		return e ? J.parse : J.parseInline;
	}
}, Jf = new class {
	defaults = Cd();
	options = this.setOptions;
	parse = this.parseMarkdown(!0);
	parseInline = this.parseMarkdown(!1);
	Parser = J;
	Renderer = Gf;
	TextRenderer = Kf;
	Lexer = q;
	Tokenizer = Wf;
	Hooks = qf;
	constructor(...e) {
		this.use(...e);
	}
	walkTokens(e, t) {
		let n = [];
		for (let r of e) switch (n = n.concat(t.call(this, r)), r.type) {
			case "table": {
				let e = r;
				for (let r of e.header) n = n.concat(this.walkTokens(r.tokens, t));
				for (let r of e.rows) for (let e of r) n = n.concat(this.walkTokens(e.tokens, t));
				break;
			}
			case "list": {
				let e = r;
				n = n.concat(this.walkTokens(e.items, t));
				break;
			}
			default: {
				let e = r;
				this.defaults.extensions?.childTokens?.[e.type] ? this.defaults.extensions.childTokens[e.type].forEach((r) => {
					let i = e[r].flat(1 / 0);
					n = n.concat(this.walkTokens(i, t));
				}) : e.tokens && (n = n.concat(this.walkTokens(e.tokens, t)));
			}
		}
		return n;
	}
	use(...e) {
		let t = this.defaults.extensions || {
			renderers: {},
			childTokens: {}
		};
		return e.forEach((e) => {
			let n = { ...e };
			if (n.async = this.defaults.async || n.async || !1, e.extensions && (e.extensions.forEach((e) => {
				if (!e.name) throw Error("extension name required");
				if ("renderer" in e) {
					let n = t.renderers[e.name];
					n ? t.renderers[e.name] = function(...t) {
						let r = e.renderer.apply(this, t);
						return r === !1 && (r = n.apply(this, t)), r;
					} : t.renderers[e.name] = e.renderer;
				}
				if ("tokenizer" in e) {
					if (!e.level || e.level !== "block" && e.level !== "inline") throw Error("extension level must be 'block' or 'inline'");
					let n = t[e.level];
					n ? n.unshift(e.tokenizer) : t[e.level] = [e.tokenizer], e.start && (e.level === "block" ? t.startBlock ? t.startBlock.push(e.start) : t.startBlock = [e.start] : e.level === "inline" && (t.startInline ? t.startInline.push(e.start) : t.startInline = [e.start]));
				}
				"childTokens" in e && e.childTokens && (t.childTokens[e.name] = e.childTokens);
			}), n.extensions = t), e.renderer) {
				let t = this.defaults.renderer || new Gf(this.defaults);
				for (let n in e.renderer) {
					if (!(n in t)) throw Error(`renderer '${n}' does not exist`);
					if (["options", "parser"].includes(n)) continue;
					let r = n, i = e.renderer[r], a = t[r];
					t[r] = (...e) => {
						let n = i.apply(t, e);
						return n === !1 && (n = a.apply(t, e)), n || "";
					};
				}
				n.renderer = t;
			}
			if (e.tokenizer) {
				let t = this.defaults.tokenizer || new Wf(this.defaults);
				for (let n in e.tokenizer) {
					if (!(n in t)) throw Error(`tokenizer '${n}' does not exist`);
					if ([
						"options",
						"rules",
						"lexer"
					].includes(n)) continue;
					let r = n, i = e.tokenizer[r], a = t[r];
					t[r] = (...e) => {
						let n = i.apply(t, e);
						return n === !1 && (n = a.apply(t, e)), n;
					};
				}
				n.tokenizer = t;
			}
			if (e.hooks) {
				let t = this.defaults.hooks || new qf();
				for (let n in e.hooks) {
					if (!(n in t)) throw Error(`hook '${n}' does not exist`);
					if (["options", "block"].includes(n)) continue;
					let r = n, i = e.hooks[r], a = t[r];
					t[r] = qf.passThroughHooks.has(n) ? (e) => {
						if (this.defaults.async && qf.passThroughHooksRespectAsync.has(n)) return (async () => {
							let n = await i.call(t, e);
							return a.call(t, n);
						})();
						let r = i.call(t, e);
						return a.call(t, r);
					} : (...e) => {
						if (this.defaults.async) return (async () => {
							let n = await i.apply(t, e);
							return n === !1 && (n = await a.apply(t, e)), n;
						})();
						let n = i.apply(t, e);
						return n === !1 && (n = a.apply(t, e)), n;
					};
				}
				n.hooks = t;
			}
			if (e.walkTokens) {
				let t = this.defaults.walkTokens, r = e.walkTokens;
				n.walkTokens = function(e) {
					let n = [];
					return n.push(r.call(this, e)), t && (n = n.concat(t.call(this, e))), n;
				};
			}
			this.defaults = {
				...this.defaults,
				...n
			};
		}), this;
	}
	setOptions(e) {
		return this.defaults = {
			...this.defaults,
			...e
		}, this;
	}
	lexer(e, t) {
		return q.lex(e, t ?? this.defaults);
	}
	parser(e, t) {
		return J.parse(e, t ?? this.defaults);
	}
	parseMarkdown(e) {
		return (t, n) => {
			let r = { ...n }, i = {
				...this.defaults,
				...r
			}, a = this.onError(!!i.silent, !!i.async);
			if (this.defaults.async === !0 && r.async === !1) return a(/* @__PURE__ */ Error("marked(): The async option was set to true by an extension. Remove async: false from the parse options object to return a Promise."));
			if (typeof t > "u" || t === null) return a(/* @__PURE__ */ Error("marked(): input parameter is undefined or null"));
			if (typeof t != "string") return a(/* @__PURE__ */ Error("marked(): input parameter is of type " + Object.prototype.toString.call(t) + ", string expected"));
			if (i.hooks && (i.hooks.options = i, i.hooks.block = e), i.async) return (async () => {
				let n = i.hooks ? await i.hooks.preprocess(t) : t, r = await (i.hooks ? await i.hooks.provideLexer(e) : e ? q.lex : q.lexInline)(n, i), a = i.hooks ? await i.hooks.processAllTokens(r) : r;
				i.walkTokens && await Promise.all(this.walkTokens(a, i.walkTokens));
				let o = await (i.hooks ? await i.hooks.provideParser(e) : e ? J.parse : J.parseInline)(a, i);
				return i.hooks ? await i.hooks.postprocess(o) : o;
			})().catch(a);
			try {
				i.hooks && (t = i.hooks.preprocess(t));
				let n = (i.hooks ? i.hooks.provideLexer(e) : e ? q.lex : q.lexInline)(t, i);
				i.hooks && (n = i.hooks.processAllTokens(n)), i.walkTokens && this.walkTokens(n, i.walkTokens);
				let r = (i.hooks ? i.hooks.provideParser(e) : e ? J.parse : J.parseInline)(n, i);
				return i.hooks && (r = i.hooks.postprocess(r)), r;
			} catch (e) {
				return a(e);
			}
		};
	}
	onError(e, t) {
		return (n) => {
			if (n.message += "\nPlease report this to https://github.com/markedjs/marked.", e) {
				let e = "<p>An error occurred:</p><pre>" + If(n.message + "", !0) + "</pre>";
				return t ? Promise.resolve(e) : e;
			}
			if (t) return Promise.reject(n);
			throw n;
		};
	}
}();
function Y(e, t) {
	return Jf.parse(e, t);
}
Y.options = Y.setOptions = function(e) {
	return Jf.setOptions(e), Y.defaults = Jf.defaults, Td(Y.defaults), Y;
}, Y.getDefaults = Cd, Y.defaults = wd, Y.use = function(...e) {
	return Jf.use(...e), Y.defaults = Jf.defaults, Td(Y.defaults), Y;
}, Y.walkTokens = function(e, t) {
	return Jf.walkTokens(e, t);
}, Y.parseInline = Jf.parseInline, Y.Parser = J, Y.parser = J.parse, Y.Renderer = Gf, Y.TextRenderer = Kf, Y.Lexer = q, Y.lexer = q.lex, Y.Tokenizer = Wf, Y.Hooks = qf, Y.parse = Y, Y.options, Y.setOptions, Y.use, Y.walkTokens, Y.parseInline, J.parse, q.lex;
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/streamdown/2.5.0/3bc14f2eb408a90c753497ac7cef03eb04d3869230b04a3264b6e3a53d1cc113/node_modules/streamdown/dist/chunk-BO2N2NFS.js
var Yf = 300, Xf = "300px", Zf = 500;
function Qf(e = {}) {
	let { immediate: t = !1, debounceDelay: n = Yf, rootMargin: r = Xf, idleTimeout: i = Zf } = e, [a, o] = (0, E.useState)(!1), s = (0, E.useRef)(null), c = (0, E.useRef)(null), l = (0, E.useRef)(null), u = (0, E.useMemo)(() => (e) => {
		let t = Date.now();
		return window.setTimeout(() => {
			e({
				didTimeout: !1,
				timeRemaining: () => Math.max(0, 50 - (Date.now() - t))
			});
		}, 1);
	}, []), d = (0, E.useMemo)(() => typeof window < "u" && window.requestIdleCallback ? (e, t) => window.requestIdleCallback(e, t) : u, [u]), f = (0, E.useMemo)(() => typeof window < "u" && window.cancelIdleCallback ? (e) => window.cancelIdleCallback(e) : (e) => {
		clearTimeout(e);
	}, []);
	return (0, E.useEffect)(() => {
		if (t) {
			o(!0);
			return;
		}
		let e = s.current;
		if (!e) return;
		c.current &&= (clearTimeout(c.current), null), l.current &&= (f(l.current), null);
		let a = () => {
			c.current &&= (clearTimeout(c.current), null), l.current &&= (f(l.current), null);
		}, u = (e) => {
			l.current = d((t) => {
				t.timeRemaining() > 0 || t.didTimeout ? (o(!0), e.disconnect()) : l.current = d(() => {
					o(!0), e.disconnect();
				}, { timeout: i / 2 });
			}, { timeout: i });
		}, p = (e) => {
			a(), c.current = window.setTimeout(() => {
				var t;
				let n = e.takeRecords();
				(n.length === 0 || (t = n.at(-1)?.isIntersecting) != null && t) && u(e);
			}, n);
		}, m = (e, t) => {
			e.isIntersecting ? p(t) : a();
		}, h = new IntersectionObserver((e) => {
			for (let t of e) m(t, h);
		}, {
			rootMargin: r,
			threshold: 0
		});
		return h.observe(e), () => {
			c.current && clearTimeout(c.current), l.current && f(l.current), h.disconnect();
		};
	}, [
		t,
		n,
		r,
		i,
		f,
		d
	]), {
		shouldRender: a,
		containerRef: s
	};
}
var $f = /\s/, ep = /^\s+$/, tp = /* @__PURE__ */ new Set([
	"code",
	"pre",
	"svg",
	"math",
	"annotation"
]), np = (e) => typeof e == "object" && !!e && "type" in e && e.type === "element", rp = (e) => e.some((e) => np(e) && tp.has(e.tagName)), ip = (e) => {
	let t = [], n = "", r = !1;
	for (let i of e) {
		let e = $f.test(i);
		e !== r && n && (t.push(n), n = ""), n += i, r = e;
	}
	return n && t.push(n), t;
}, ap = (e) => {
	let t = [], n = "";
	for (let r of e) $f.test(r) ? n += r : (n &&= (t.push(n), ""), t.push(r));
	return n && t.push(n), t;
}, op = (e, t, n, r, i, a) => {
	let o = `--sd-animation:sd-${t};--sd-duration:${i ? 0 : n}ms;--sd-easing:${r}`;
	return a && (o += `;--sd-delay:${a}ms`), {
		type: "element",
		tagName: "span",
		properties: {
			"data-sd-animate": !0,
			style: o
		},
		children: [{
			type: "text",
			value: e
		}]
	};
}, sp = (e, t, n, r, i) => {
	let a = t.at(-1);
	if (!(a && "children" in a)) return;
	if (rp(t)) return b;
	let o = a, s = o.children.indexOf(e);
	if (s === -1) return;
	let c = e.value;
	if (!c.trim()) {
		i.count += c.length;
		return;
	}
	let l = n.sep === "char" ? ap(c) : ip(c), u = r.prevContentLength, d = l.map((e) => {
		let t = i.count;
		if (i.count += e.length, ep.test(e)) return {
			type: "text",
			value: e
		};
		let r = u > 0 && t < u, a = r ? 0 : i.newIndex++ * n.stagger;
		return op(e, n.animation, n.duration, n.easing, r, a);
	});
	return o.children.splice(s, 1, ...d), s + d.length;
}, cp = 0;
function lp(e) {
	let t = {
		animation: e?.animation ?? "fadeIn",
		duration: e?.duration ?? 150,
		easing: e?.easing ?? "ease",
		sep: e?.sep ?? "word",
		stagger: e?.stagger ?? 40
	}, n = {
		prevContentLength: 0,
		lastRenderCharCount: 0
	}, r = cp++, i = () => (e) => {
		let r = {
			count: 0,
			newIndex: 0
		};
		ie(e, "text", (e, i) => sp(e, i, t, n, r)), n.lastRenderCharCount = r.count, n.prevContentLength = 0;
	};
	return Object.defineProperty(i, "name", { value: `rehypeAnimate$${r}` }), {
		name: "animate",
		type: "animate",
		rehypePlugin: i,
		setPrevContentLength(e) {
			n.prevContentLength = e;
		},
		getLastRenderCharCount() {
			let e = n.lastRenderCharCount;
			return n.lastRenderCharCount = 0, e;
		}
	};
}
lp();
var up = (0, E.createContext)(!1), dp = () => (0, E.useContext)(up), fp = (...e) => a(o(e)), pp = (e, t) => {
	if (!e || !t) return t;
	let n = `${e}:`;
	return t.split(/\s+/).filter(Boolean).map((t) => t.startsWith(n) ? t : `${e}:${t}`).join(" ");
}, mp = (e) => e ? (...t) => pp(e, a(o(t))) : fp, hp = (e, t, n) => {
	let r = typeof t == "string" && n.startsWith("text/csv") ? "﻿" : "", i = typeof t == "string" ? new Blob([r + t], { type: n }) : t, a = URL.createObjectURL(i), o = document.createElement("a");
	o.href = a, o.download = e, document.body.appendChild(o), o.click(), document.body.removeChild(o), URL.revokeObjectURL(a);
}, gp = (0, E.createContext)(fp), X = () => (0, E.useContext)(gp), _p = fp("block", "before:content-[counter(line)]", "before:inline-block", "before:[counter-increment:line]", "before:w-6", "before:mr-4", "before:text-[13px]", "before:text-right", "before:text-muted-foreground/50", "before:font-mono", "before:select-none"), vp = (e) => {
	let t = {};
	for (let n of e.split(";")) {
		let e = n.indexOf(":");
		if (e > 0) {
			let r = n.slice(0, e).trim(), i = n.slice(e + 1).trim();
			r && i && (t[r] = i);
		}
	}
	return t;
}, yp = (0, E.memo)(({ children: e, result: t, language: n, className: r, startLine: i, lineNumbers: a = !0, ...o }) => {
	let s = X(), c = (0, E.useMemo)(() => s(_p), [s]), l = (0, E.useMemo)(() => {
		let e = {};
		return t.bg && (e["--sdm-bg"] = t.bg), t.fg && (e["--sdm-fg"] = t.fg), t.rootStyle && Object.assign(e, vp(t.rootStyle)), e;
	}, [
		t.bg,
		t.fg,
		t.rootStyle
	]);
	return (0, W.jsx)("div", {
		className: s(r, "overflow-x-auto rounded-md border border-border bg-background p-4 text-sm"),
		"data-language": n,
		"data-streamdown": "code-block-body",
		...o,
		children: (0, W.jsx)("pre", {
			className: s(r, "bg-[var(--sdm-bg,inherit]", "dark:bg-[var(--shiki-dark-bg,var(--sdm-bg,inherit)]"),
			style: l,
			children: (0, W.jsx)("code", {
				className: a ? s("[counter-increment:line_0] [counter-reset:line]") : void 0,
				style: a && i && i > 1 ? { counterReset: `line ${i - 1}` } : void 0,
				children: t.tokens.map((e, t) => (0, W.jsx)("span", {
					className: a ? c : void 0,
					children: e.length === 0 || e.length === 1 && e[0].content === "" ? "\n" : e.map((e, t) => {
						let n = {}, r = !!e.bgColor;
						if (e.color && (n["--sdm-c"] = e.color), e.bgColor && (n["--sdm-tbg"] = e.bgColor), e.htmlStyle) for (let [t, i] of Object.entries(e.htmlStyle)) t === "color" ? n["--sdm-c"] = i : t === "background-color" ? (n["--sdm-tbg"] = i, r = !0) : n[t] = i;
						return (0, W.jsx)("span", {
							className: s("text-[var(--sdm-c,inherit)]", "dark:text-[var(--shiki-dark,var(--sdm-c,inherit))]", r && "bg-[var(--sdm-tbg)]", r && "dark:bg-[var(--shiki-dark-bg,var(--sdm-tbg))]"),
							style: n,
							...e.htmlAttrs,
							children: e.content
						}, t);
					})
				}, t))
			})
		})
	});
}, (e, t) => e.result === t.result && e.language === t.language && e.className === t.className && e.startLine === t.startLine && e.lineNumbers === t.lineNumbers), bp = ({ className: e, language: t, style: n, isIncomplete: r, ...i }) => {
	let a = X();
	return (0, W.jsx)("div", {
		className: a("my-4 flex w-full flex-col gap-2 rounded-xl border border-border bg-sidebar p-2", e),
		"data-incomplete": r || void 0,
		"data-language": t,
		"data-streamdown": "code-block",
		style: {
			contentVisibility: "auto",
			containIntrinsicSize: "auto 200px",
			...n
		},
		...i
	});
}, xp = (0, E.createContext)({ code: "" }), Sp = () => (0, E.useContext)(xp), Cp = ({ language: e }) => {
	let t = X();
	return (0, W.jsx)("div", {
		className: t("flex h-8 items-center text-muted-foreground text-xs"),
		"data-language": e,
		"data-streamdown": "code-block-header",
		children: (0, W.jsx)("span", {
			className: t("ml-1 font-mono lowercase"),
			children: e
		})
	});
}, wp = (e) => {
	let t = e.length;
	for (; t > 0 && e[t - 1] === "\n";) t--;
	return e.slice(0, t);
}, Tp = (0, E.lazy)(() => import("./highlighted-body-OFNGDK62-BA-3GrZL.js").then((e) => ({ default: e.HighlightedCodeBlockBody }))), Ep = ({ code: e, language: t, className: n, children: r, isIncomplete: i = !1, startLine: a, lineNumbers: o, ...s }) => {
	let c = X(), l = (0, E.useMemo)(() => wp(e), [e]), u = (0, E.useMemo)(() => ({
		bg: "transparent",
		fg: "inherit",
		tokens: l.split("\n").map((e) => [{
			content: e,
			color: "inherit",
			bgColor: "transparent",
			htmlStyle: {},
			offset: 0
		}])
	}), [l]);
	return (0, W.jsx)(xp.Provider, {
		value: { code: e },
		children: (0, W.jsxs)(bp, {
			isIncomplete: i,
			language: t,
			children: [
				(0, W.jsx)(Cp, { language: t }),
				r ? (0, W.jsx)("div", {
					className: c("pointer-events-none sticky top-2 z-10 -mt-10 flex h-8 items-center justify-end"),
					children: (0, W.jsx)("div", {
						className: c("pointer-events-auto flex shrink-0 items-center gap-2 rounded-md border border-sidebar bg-sidebar/80 px-1.5 py-1 supports-[backdrop-filter]:bg-sidebar/70 supports-[backdrop-filter]:backdrop-blur"),
						"data-streamdown": "code-block-actions",
						children: r
					})
				}) : null,
				(0, W.jsx)(E.Suspense, {
					fallback: (0, W.jsx)(yp, {
						className: n,
						language: t,
						lineNumbers: o,
						result: u,
						startLine: a,
						...s
					}),
					children: (0, W.jsx)(Tp, {
						className: n,
						code: l,
						language: t,
						lineNumbers: o,
						raw: u,
						startLine: a,
						...s
					})
				})
			]
		})
	});
}, Dp = {
	CheckIcon: (e) => (0, W.jsx)("svg", {
		color: "currentColor",
		height: 16,
		strokeLinejoin: "round",
		viewBox: "0 0 16 16",
		width: 16,
		...e,
		children: (0, W.jsx)("path", {
			clipRule: "evenodd",
			d: "M15.5607 3.99999L15.0303 4.53032L6.23744 13.3232C5.55403 14.0066 4.44599 14.0066 3.76257 13.3232L4.2929 12.7929L3.76257 13.3232L0.969676 10.5303L0.439346 9.99999L1.50001 8.93933L2.03034 9.46966L4.82323 12.2626C4.92086 12.3602 5.07915 12.3602 5.17678 12.2626L13.9697 3.46966L14.5 2.93933L15.5607 3.99999Z",
			fill: "currentColor",
			fillRule: "evenodd"
		})
	}),
	CopyIcon: (e) => (0, W.jsx)("svg", {
		color: "currentColor",
		height: 16,
		strokeLinejoin: "round",
		viewBox: "0 0 16 16",
		width: 16,
		...e,
		children: (0, W.jsx)("path", {
			clipRule: "evenodd",
			d: "M2.75 0.5C1.7835 0.5 1 1.2835 1 2.25V9.75C1 10.7165 1.7835 11.5 2.75 11.5H3.75H4.5V10H3.75H2.75C2.61193 10 2.5 9.88807 2.5 9.75V2.25C2.5 2.11193 2.61193 2 2.75 2H8.25C8.38807 2 8.5 2.11193 8.5 2.25V3H10V2.25C10 1.2835 9.2165 0.5 8.25 0.5H2.75ZM7.75 4.5C6.7835 4.5 6 5.2835 6 6.25V13.75C6 14.7165 6.7835 15.5 7.75 15.5H13.25C14.2165 15.5 15 14.7165 15 13.75V6.25C15 5.2835 14.2165 4.5 13.25 4.5H7.75ZM7.5 6.25C7.5 6.11193 7.61193 6 7.75 6H13.25C13.3881 6 13.5 6.11193 13.5 6.25V13.75C13.5 13.8881 13.3881 14 13.25 14H7.75C7.61193 14 7.5 13.8881 7.5 13.75V6.25Z",
			fill: "currentColor",
			fillRule: "evenodd"
		})
	}),
	DownloadIcon: (e) => (0, W.jsx)("svg", {
		color: "currentColor",
		height: 16,
		strokeLinejoin: "round",
		viewBox: "0 0 16 16",
		width: 16,
		...e,
		children: (0, W.jsx)("path", {
			clipRule: "evenodd",
			d: "M8.75 1V1.75V8.68934L10.7197 6.71967L11.25 6.18934L12.3107 7.25L11.7803 7.78033L8.70711 10.8536C8.31658 11.2441 7.68342 11.2441 7.29289 10.8536L4.21967 7.78033L3.68934 7.25L4.75 6.18934L5.28033 6.71967L7.25 8.68934V1.75V1H8.75ZM13.5 9.25V13.5H2.5V9.25V8.5H1V9.25V14C1 14.5523 1.44771 15 2 15H14C14.5523 15 15 14.5523 15 14V9.25V8.5H13.5V9.25Z",
			fill: "currentColor",
			fillRule: "evenodd"
		})
	}),
	ExternalLinkIcon: (e) => (0, W.jsx)("svg", {
		color: "currentColor",
		height: 16,
		strokeLinejoin: "round",
		viewBox: "0 0 16 16",
		width: 16,
		...e,
		children: (0, W.jsx)("path", {
			clipRule: "evenodd",
			d: "M13.5 10.25V13.25C13.5 13.3881 13.3881 13.5 13.25 13.5H2.75C2.61193 13.5 2.5 13.3881 2.5 13.25L2.5 2.75C2.5 2.61193 2.61193 2.5 2.75 2.5H5.75H6.5V1H5.75H2.75C1.7835 1 1 1.7835 1 2.75V13.25C1 14.2165 1.7835 15 2.75 15H13.25C14.2165 15 15 14.2165 15 13.25V10.25V9.5H13.5V10.25ZM9 1H9.75H14.2495C14.6637 1 14.9995 1.33579 14.9995 1.75V6.25V7H13.4995V6.25V3.56066L8.53033 8.52978L8 9.06011L6.93934 7.99945L7.46967 7.46912L12.4388 2.5H9.75H9V1Z",
			fill: "currentColor",
			fillRule: "evenodd"
		})
	}),
	Loader2Icon: (e) => (0, W.jsxs)("svg", {
		color: "currentColor",
		height: 16,
		strokeLinejoin: "round",
		viewBox: "0 0 16 16",
		width: 16,
		...e,
		children: [
			(0, W.jsx)("path", {
				d: "M8 0V4",
				stroke: "currentColor",
				strokeWidth: "1.5"
			}),
			(0, W.jsx)("path", {
				d: "M8 16V12",
				opacity: "0.5",
				stroke: "currentColor",
				strokeWidth: "1.5"
			}),
			(0, W.jsx)("path", {
				d: "M3.29773 1.52783L5.64887 4.7639",
				opacity: "0.9",
				stroke: "currentColor",
				strokeWidth: "1.5"
			}),
			(0, W.jsx)("path", {
				d: "M12.7023 1.52783L10.3511 4.7639",
				opacity: "0.1",
				stroke: "currentColor",
				strokeWidth: "1.5"
			}),
			(0, W.jsx)("path", {
				d: "M12.7023 14.472L10.3511 11.236",
				opacity: "0.4",
				stroke: "currentColor",
				strokeWidth: "1.5"
			}),
			(0, W.jsx)("path", {
				d: "M3.29773 14.472L5.64887 11.236",
				opacity: "0.6",
				stroke: "currentColor",
				strokeWidth: "1.5"
			}),
			(0, W.jsx)("path", {
				d: "M15.6085 5.52783L11.8043 6.7639",
				opacity: "0.2",
				stroke: "currentColor",
				strokeWidth: "1.5"
			}),
			(0, W.jsx)("path", {
				d: "M0.391602 10.472L4.19583 9.23598",
				opacity: "0.7",
				stroke: "currentColor",
				strokeWidth: "1.5"
			}),
			(0, W.jsx)("path", {
				d: "M15.6085 10.4722L11.8043 9.2361",
				opacity: "0.3",
				stroke: "currentColor",
				strokeWidth: "1.5"
			}),
			(0, W.jsx)("path", {
				d: "M0.391602 5.52783L4.19583 6.7639",
				opacity: "0.8",
				stroke: "currentColor",
				strokeWidth: "1.5"
			})
		]
	}),
	Maximize2Icon: (e) => (0, W.jsx)("svg", {
		color: "currentColor",
		height: 16,
		strokeLinejoin: "round",
		viewBox: "0 0 16 16",
		width: 16,
		...e,
		children: (0, W.jsx)("path", {
			clipRule: "evenodd",
			d: "M1 5.25V6H2.5V5.25V2.5H5.25H6V1H5.25H2C1.44772 1 1 1.44772 1 2V5.25ZM5.25 14.9994H6V13.4994H5.25H2.5V10.7494V9.99939H1V10.7494V13.9994C1 14.5517 1.44772 14.9994 2 14.9994H5.25ZM15 10V10.75V14C15 14.5523 14.5523 15 14 15H10.75H10V13.5H10.75H13.5V10.75V10H15ZM10.75 1H10V2.5H10.75H13.5V5.25V6H15V5.25V2C15 1.44772 14.5523 1 14 1H10.75Z",
			fill: "currentColor",
			fillRule: "evenodd"
		})
	}),
	RotateCcwIcon: (e) => (0, W.jsx)("svg", {
		color: "currentColor",
		height: 16,
		strokeLinejoin: "round",
		viewBox: "0 0 16 16",
		width: 16,
		...e,
		children: (0, W.jsx)("path", {
			clipRule: "evenodd",
			d: "M13.5 8C13.5 4.96643 11.0257 2.5 7.96452 2.5C5.42843 2.5 3.29365 4.19393 2.63724 6.5H5.25H6V8H5.25H0.75C0.335787 8 0 7.66421 0 7.25V2.75V2H1.5V2.75V5.23347C2.57851 2.74164 5.06835 1 7.96452 1C11.8461 1 15 4.13001 15 8C15 11.87 11.8461 15 7.96452 15C5.62368 15 3.54872 13.8617 2.27046 12.1122L1.828 11.5066L3.03915 10.6217L3.48161 11.2273C4.48831 12.6051 6.12055 13.5 7.96452 13.5C11.0257 13.5 13.5 11.0336 13.5 8Z",
			fill: "currentColor",
			fillRule: "evenodd"
		})
	}),
	XIcon: (e) => (0, W.jsx)("svg", {
		color: "currentColor",
		height: 16,
		strokeLinejoin: "round",
		viewBox: "0 0 16 16",
		width: 16,
		...e,
		children: (0, W.jsx)("path", {
			clipRule: "evenodd",
			d: "M12.4697 13.5303L13 14.0607L14.0607 13L13.5303 12.4697L9.06065 7.99999L13.5303 3.53032L14.0607 2.99999L13 1.93933L12.4697 2.46966L7.99999 6.93933L3.53032 2.46966L2.99999 1.93933L1.93933 2.99999L2.46966 3.53032L6.93933 7.99999L2.46966 12.4697L1.93933 13L2.99999 14.0607L3.53032 13.5303L7.99999 9.06065L12.4697 13.5303Z",
			fill: "currentColor",
			fillRule: "evenodd"
		})
	}),
	ZoomInIcon: (e) => (0, W.jsx)("svg", {
		color: "currentColor",
		height: 16,
		strokeLinejoin: "round",
		viewBox: "0 0 16 16",
		width: 16,
		...e,
		children: (0, W.jsx)("path", {
			clipRule: "evenodd",
			d: "M1.5 6.5C1.5 3.73858 3.73858 1.5 6.5 1.5C9.26142 1.5 11.5 3.73858 11.5 6.5C11.5 9.26142 9.26142 11.5 6.5 11.5C3.73858 11.5 1.5 9.26142 1.5 6.5ZM6.5 0C2.91015 0 0 2.91015 0 6.5C0 10.0899 2.91015 13 6.5 13C8.02469 13 9.42677 12.475 10.5353 11.596L13.9697 15.0303L14.5 15.5607L15.5607 14.5L15.0303 13.9697L11.596 10.5353C12.475 9.42677 13 8.02469 13 6.5C13 2.91015 10.0899 0 6.5 0ZM4.125 5.875H4.75H5.875V4.75V4.125H7.125V4.75V5.875H8.25H8.875V7.125H8.25H7.125V8.25V8.875H5.875V8.25V7.125H4.75H4.125V5.875Z",
			fill: "currentColor",
			fillRule: "evenodd"
		})
	}),
	ZoomOutIcon: (e) => (0, W.jsx)("svg", {
		color: "currentColor",
		height: 16,
		strokeLinejoin: "round",
		viewBox: "0 0 16 16",
		width: 16,
		...e,
		children: (0, W.jsx)("path", {
			clipRule: "evenodd",
			d: "M1.5 6.5C1.5 3.73858 3.73858 1.5 6.5 1.5C9.26142 1.5 11.5 3.73858 11.5 6.5C11.5 9.26142 9.26142 11.5 6.5 11.5C3.73858 11.5 1.5 9.26142 1.5 6.5ZM6.5 0C2.91015 0 0 2.91015 0 6.5C0 10.0899 2.91015 13 6.5 13C8.02469 13 9.42677 12.475 10.5353 11.596L13.9697 15.0303L14.5 15.5607L15.5607 14.5L15.0303 13.9697L11.596 10.5353C12.475 9.42677 13 8.02469 13 6.5C13 2.91015 10.0899 0 6.5 0ZM4.125 5.875H4.75H8.25H8.875V7.125H8.25H4.75H4.125V5.875Z",
			fill: "currentColor",
			fillRule: "evenodd"
		})
	})
}, Op = (0, E.createContext)(Dp), kp = (e, t) => {
	if (e === t) return !0;
	if (!(e && t)) return e === t;
	let n = Object.keys(e), r = Object.keys(t);
	return n.length === r.length && n.every((n) => e[n] === t[n]);
}, Ap = ({ icons: e, children: t }) => {
	let n = (0, E.useRef)(e), r = (0, E.useRef)(e ? {
		...Dp,
		...e
	} : Dp);
	kp(n.current, e) || (n.current = e, r.current = e ? {
		...Dp,
		...e
	} : Dp);
	let i = r.current;
	return (0, W.jsx)(Op.Provider, {
		value: i,
		children: t
	});
}, Z = () => (0, E.useContext)(Op), jp = {
	copyCode: "Copy Code",
	downloadFile: "Download file",
	downloadDiagram: "Download diagram",
	downloadDiagramAsSvg: "Download diagram as SVG",
	downloadDiagramAsPng: "Download diagram as PNG",
	downloadDiagramAsMmd: "Download diagram as MMD",
	viewFullscreen: "View fullscreen",
	exitFullscreen: "Exit fullscreen",
	mermaidFormatSvg: "SVG",
	mermaidFormatPng: "PNG",
	mermaidFormatMmd: "MMD",
	copyTable: "Copy table",
	copyTableAsMarkdown: "Copy table as Markdown",
	copyTableAsCsv: "Copy table as CSV",
	copyTableAsTsv: "Copy table as TSV",
	downloadTable: "Download table",
	downloadTableAsCsv: "Download table as CSV",
	downloadTableAsMarkdown: "Download table as Markdown",
	tableFormatMarkdown: "Markdown",
	tableFormatCsv: "CSV",
	tableFormatTsv: "TSV",
	imageNotAvailable: "Image not available",
	downloadImage: "Download image",
	openExternalLink: "Open external link?",
	externalLinkWarning: "You're about to visit an external website.",
	close: "Close",
	copyLink: "Copy link",
	copied: "Copied",
	openLink: "Open link"
}, Mp = (0, E.createContext)(jp), Np = () => (0, E.useContext)(Mp), Pp = ({ onCopy: e, onError: t, timeout: n = 2e3, children: r, className: i, code: a, ...o }) => {
	let s = X(), [c, l] = (0, E.useState)(!1), u = (0, E.useRef)(0), { code: d } = Sp(), { isAnimating: f } = (0, E.useContext)($), p = Np(), m = a ?? d, h = async () => {
		var r;
		if (typeof window > "u" || !((r = navigator == null ? void 0 : navigator.clipboard) != null && r.writeText)) t?.(/* @__PURE__ */ Error("Clipboard API not available"));
		else try {
			c || (await navigator.clipboard.writeText(m), l(!0), e?.(), u.current = window.setTimeout(() => l(!1), n));
		} catch (e) {
			t?.(e);
		}
	};
	(0, E.useEffect)(() => () => {
		window.clearTimeout(u.current);
	}, []);
	let g = Z(), _ = c ? g.CheckIcon : g.CopyIcon;
	return (0, W.jsx)("button", {
		className: s("cursor-pointer p-1 text-muted-foreground transition-all hover:text-foreground disabled:cursor-not-allowed disabled:opacity-50", i),
		"data-streamdown": "code-block-copy-button",
		disabled: f,
		onClick: h,
		title: p.copyCode,
		type: "button",
		...o,
		children: r ?? (0, W.jsx)(_, { size: 14 })
	});
}, Fp = {
	"1c": "1c",
	"1c-query": "1cq",
	abap: "abap",
	"actionscript-3": "as",
	ada: "ada",
	adoc: "adoc",
	"angular-html": "html",
	"angular-ts": "ts",
	apache: "conf",
	apex: "cls",
	apl: "apl",
	applescript: "applescript",
	ara: "ara",
	asciidoc: "adoc",
	asm: "asm",
	astro: "astro",
	awk: "awk",
	ballerina: "bal",
	bash: "sh",
	bat: "bat",
	batch: "bat",
	be: "be",
	beancount: "beancount",
	berry: "berry",
	bibtex: "bib",
	bicep: "bicep",
	blade: "blade.php",
	bsl: "bsl",
	c: "c",
	"c#": "cs",
	"c++": "cpp",
	cadence: "cdc",
	cairo: "cairo",
	cdc: "cdc",
	clarity: "clar",
	clj: "clj",
	clojure: "clj",
	"closure-templates": "soy",
	cmake: "cmake",
	cmd: "cmd",
	cobol: "cob",
	codeowners: "CODEOWNERS",
	codeql: "ql",
	coffee: "coffee",
	coffeescript: "coffee",
	"common-lisp": "lisp",
	console: "sh",
	coq: "v",
	cpp: "cpp",
	cql: "cql",
	crystal: "cr",
	cs: "cs",
	csharp: "cs",
	css: "css",
	csv: "csv",
	cue: "cue",
	cypher: "cql",
	d: "d",
	dart: "dart",
	dax: "dax",
	desktop: "desktop",
	diff: "diff",
	docker: "dockerfile",
	dockerfile: "dockerfile",
	dotenv: "env",
	"dream-maker": "dm",
	edge: "edge",
	elisp: "el",
	elixir: "ex",
	elm: "elm",
	"emacs-lisp": "el",
	erb: "erb",
	erl: "erl",
	erlang: "erl",
	f: "f",
	"f#": "fs",
	f03: "f03",
	f08: "f08",
	f18: "f18",
	f77: "f77",
	f90: "f90",
	f95: "f95",
	fennel: "fnl",
	fish: "fish",
	fluent: "ftl",
	for: "for",
	"fortran-fixed-form": "f",
	"fortran-free-form": "f90",
	fs: "fs",
	fsharp: "fs",
	fsl: "fsl",
	ftl: "ftl",
	gdresource: "tres",
	gdscript: "gd",
	gdshader: "gdshader",
	genie: "gs",
	gherkin: "feature",
	"git-commit": "gitcommit",
	"git-rebase": "gitrebase",
	gjs: "js",
	gleam: "gleam",
	"glimmer-js": "js",
	"glimmer-ts": "ts",
	glsl: "glsl",
	gnuplot: "plt",
	go: "go",
	gql: "gql",
	graphql: "graphql",
	groovy: "groovy",
	gts: "gts",
	hack: "hack",
	haml: "haml",
	handlebars: "hbs",
	haskell: "hs",
	haxe: "hx",
	hbs: "hbs",
	hcl: "hcl",
	hjson: "hjson",
	hlsl: "hlsl",
	hs: "hs",
	html: "html",
	"html-derivative": "html",
	http: "http",
	hxml: "hxml",
	hy: "hy",
	imba: "imba",
	ini: "ini",
	jade: "jade",
	java: "java",
	javascript: "js",
	jinja: "jinja",
	jison: "jison",
	jl: "jl",
	js: "js",
	json: "json",
	json5: "json5",
	jsonc: "jsonc",
	jsonl: "jsonl",
	jsonnet: "jsonnet",
	jssm: "jssm",
	jsx: "jsx",
	julia: "jl",
	kotlin: "kt",
	kql: "kql",
	kt: "kt",
	kts: "kts",
	kusto: "kql",
	latex: "tex",
	lean: "lean",
	lean4: "lean",
	less: "less",
	liquid: "liquid",
	lisp: "lisp",
	lit: "lit",
	llvm: "ll",
	log: "log",
	logo: "logo",
	lua: "lua",
	luau: "luau",
	make: "mak",
	makefile: "mak",
	markdown: "md",
	marko: "marko",
	matlab: "m",
	md: "md",
	mdc: "mdc",
	mdx: "mdx",
	mediawiki: "wiki",
	mermaid: "mmd",
	mips: "s",
	mipsasm: "s",
	mmd: "mmd",
	mojo: "mojo",
	move: "move",
	nar: "nar",
	narrat: "narrat",
	nextflow: "nf",
	nf: "nf",
	nginx: "conf",
	nim: "nim",
	nix: "nix",
	nu: "nu",
	nushell: "nu",
	objc: "m",
	"objective-c": "m",
	"objective-cpp": "mm",
	ocaml: "ml",
	pascal: "pas",
	perl: "pl",
	perl6: "p6",
	php: "php",
	plsql: "pls",
	po: "po",
	polar: "polar",
	postcss: "pcss",
	pot: "pot",
	potx: "potx",
	powerquery: "pq",
	powershell: "ps1",
	prisma: "prisma",
	prolog: "pl",
	properties: "properties",
	proto: "proto",
	protobuf: "proto",
	ps: "ps",
	ps1: "ps1",
	pug: "pug",
	puppet: "pp",
	purescript: "purs",
	py: "py",
	python: "py",
	ql: "ql",
	qml: "qml",
	qmldir: "qmldir",
	qss: "qss",
	r: "r",
	racket: "rkt",
	raku: "raku",
	razor: "cshtml",
	rb: "rb",
	reg: "reg",
	regex: "regex",
	regexp: "regexp",
	rel: "rel",
	riscv: "s",
	rs: "rs",
	rst: "rst",
	ruby: "rb",
	rust: "rs",
	sas: "sas",
	sass: "sass",
	scala: "scala",
	scheme: "scm",
	scss: "scss",
	sdbl: "sdbl",
	sh: "sh",
	shader: "shader",
	shaderlab: "shader",
	shell: "sh",
	shellscript: "sh",
	shellsession: "sh",
	smalltalk: "st",
	solidity: "sol",
	soy: "soy",
	sparql: "rq",
	spl: "spl",
	splunk: "spl",
	sql: "sql",
	"ssh-config": "config",
	stata: "do",
	styl: "styl",
	stylus: "styl",
	svelte: "svelte",
	swift: "swift",
	"system-verilog": "sv",
	systemd: "service",
	talon: "talon",
	talonscript: "talon",
	tasl: "tasl",
	tcl: "tcl",
	templ: "templ",
	terraform: "tf",
	tex: "tex",
	tf: "tf",
	tfvars: "tfvars",
	toml: "toml",
	ts: "ts",
	"ts-tags": "ts",
	tsp: "tsp",
	tsv: "tsv",
	tsx: "tsx",
	turtle: "ttl",
	twig: "twig",
	typ: "typ",
	typescript: "ts",
	typespec: "tsp",
	typst: "typ",
	v: "v",
	vala: "vala",
	vb: "vb",
	verilog: "v",
	vhdl: "vhdl",
	vim: "vim",
	viml: "vim",
	vimscript: "vim",
	vue: "vue",
	"vue-html": "html",
	"vue-vine": "vine",
	vy: "vy",
	vyper: "vy",
	wasm: "wasm",
	wenyan: "wy",
	wgsl: "wgsl",
	wiki: "wiki",
	wikitext: "wiki",
	wit: "wit",
	wl: "wl",
	wolfram: "wl",
	xml: "xml",
	xsl: "xsl",
	yaml: "yaml",
	yml: "yml",
	zenscript: "zs",
	zig: "zig",
	zsh: "zsh",
	文言: "wy"
}, Ip = ({ onDownload: e, onError: t, language: n, children: r, className: i, code: a, ...o }) => {
	let s = X(), { code: c } = Sp(), { isAnimating: l } = (0, E.useContext)($), u = Np(), d = Z(), f = a ?? c, p = `file.${n && n in Fp ? Fp[n] : "txt"}`;
	return (0, W.jsx)("button", {
		className: s("cursor-pointer p-1 text-muted-foreground transition-all hover:text-foreground disabled:cursor-not-allowed disabled:opacity-50", i),
		"data-streamdown": "code-block-download-button",
		disabled: l,
		onClick: () => {
			try {
				hp(p, f, "text/plain"), e?.();
			} catch (e) {
				t?.(e);
			}
		},
		title: u.downloadFile,
		type: "button",
		...o,
		children: r ?? (0, W.jsx)(d.DownloadIcon, { size: 14 })
	});
}, Lp = () => {
	let { Loader2Icon: e } = Z(), t = X();
	return (0, W.jsxs)("div", {
		className: t("w-full divide-y divide-border overflow-hidden rounded-xl border border-border"),
		children: [(0, W.jsx)("div", { className: t("h-[46px] w-full bg-muted/80") }), (0, W.jsx)("div", {
			className: t("flex w-full items-center justify-center p-4"),
			children: (0, W.jsx)(e, { className: t("size-4 animate-spin") })
		})]
	});
}, Rp = /\.[^/.]+$/, zp = ({ node: e, className: t, src: n, alt: r, onLoad: i, onError: a, ...o }) => {
	let { DownloadIcon: s } = Z(), c = X(), l = (0, E.useRef)(null), [u, d] = (0, E.useState)(!1), [f, p] = (0, E.useState)(!1), m = Np(), h = o.width != null || o.height != null, g = (u || h) && !f, _ = f && !h;
	(0, E.useEffect)(() => {
		let e = l.current;
		if (e != null && e.complete) {
			let t = e.naturalWidth > 0;
			d(t), p(!t);
		}
	}, []);
	let v = (0, E.useCallback)((e) => {
		d(!0), p(!1), i?.(e);
	}, [i]), y = (0, E.useCallback)((e) => {
		d(!1), p(!0), a?.(e);
	}, [a]), b = async () => {
		if (n) try {
			let e = await (await fetch(n)).blob(), t = new URL(n, window.location.origin).pathname.split("/").pop() || "", i = t.split(".").pop(), a = t.includes(".") && i !== void 0 && i.length <= 4, o = "";
			if (a) o = t;
			else {
				let n = e.type, i = "png";
				n.includes("jpeg") || n.includes("jpg") ? i = "jpg" : n.includes("png") ? i = "png" : n.includes("svg") ? i = "svg" : n.includes("gif") ? i = "gif" : n.includes("webp") && (i = "webp"), o = `${(r || t || "image").replace(Rp, "")}.${i}`;
			}
			hp(o, e, e.type);
		} catch {
			window.open(n, "_blank");
		}
	};
	return n ? (0, W.jsxs)("div", {
		className: c("group relative my-4 inline-block"),
		"data-streamdown": "image-wrapper",
		children: [
			(0, W.jsx)("img", {
				alt: r,
				className: c("max-w-full rounded-lg", _ && "hidden", t),
				"data-streamdown": "image",
				onError: y,
				onLoad: v,
				ref: l,
				src: n,
				...o
			}),
			_ && (0, W.jsx)("span", {
				className: c("text-muted-foreground text-xs italic"),
				"data-streamdown": "image-fallback",
				children: m.imageNotAvailable
			}),
			(0, W.jsx)("div", { className: c("pointer-events-none absolute inset-0 hidden rounded-lg bg-black/10 group-hover:block") }),
			g && (0, W.jsx)("button", {
				className: c("absolute right-2 bottom-2 flex h-8 w-8 cursor-pointer items-center justify-center rounded-md border border-border bg-background/90 shadow-sm backdrop-blur-sm transition-all duration-200 hover:bg-background", "opacity-0 group-hover:opacity-100"),
				onClick: b,
				title: m.downloadImage,
				type: "button",
				children: (0, W.jsx)(s, { size: 14 })
			})
		]
	}) : null;
}, Bp = 0, Vp = () => {
	Bp += 1, Bp === 1 && (document.body.style.overflow = "hidden");
}, Hp = () => {
	Bp = Math.max(0, Bp - 1), Bp === 0 && (document.body.style.overflow = "");
}, Up = ({ url: e, isOpen: t, onClose: n, onConfirm: r }) => {
	let { CheckIcon: i, CopyIcon: a, ExternalLinkIcon: o, XIcon: s } = Z(), c = X(), [l, u] = (0, E.useState)(!1), d = Np(), f = (0, E.useCallback)(async () => {
		try {
			await navigator.clipboard.writeText(e), u(!0), setTimeout(() => u(!1), 2e3);
		} catch {}
	}, [e]), p = (0, E.useCallback)(() => {
		r(), n();
	}, [r, n]);
	return (0, E.useEffect)(() => {
		if (t) {
			Vp();
			let e = (e) => {
				e.key === "Escape" && n();
			};
			return document.addEventListener("keydown", e), () => {
				document.removeEventListener("keydown", e), Hp();
			};
		}
	}, [t, n]), t ? (0, W.jsx)("div", {
		className: c("fixed inset-0 z-50 flex items-center justify-center bg-background/50 backdrop-blur-sm"),
		"data-streamdown": "link-safety-modal",
		onClick: n,
		onKeyDown: (e) => {
			e.key === "Escape" && n();
		},
		role: "button",
		tabIndex: 0,
		children: (0, W.jsxs)("div", {
			className: c("relative mx-4 flex w-full max-w-md flex-col gap-4 rounded-xl border bg-background p-6 shadow-lg"),
			onClick: (e) => e.stopPropagation(),
			onKeyDown: (e) => e.stopPropagation(),
			role: "presentation",
			children: [
				(0, W.jsx)("button", {
					className: c("absolute top-4 right-4 rounded-md p-1 text-muted-foreground transition-all hover:bg-muted hover:text-foreground"),
					onClick: n,
					title: d.close,
					type: "button",
					children: (0, W.jsx)(s, { size: 16 })
				}),
				(0, W.jsxs)("div", {
					className: c("flex flex-col gap-2"),
					children: [(0, W.jsxs)("div", {
						className: c("flex items-center gap-2 font-semibold text-lg"),
						children: [(0, W.jsx)(o, { size: 20 }), (0, W.jsx)("span", { children: d.openExternalLink })]
					}), (0, W.jsx)("p", {
						className: c("text-muted-foreground text-sm"),
						children: d.externalLinkWarning
					})]
				}),
				(0, W.jsx)("div", {
					className: c("break-all rounded-md bg-muted p-3 font-mono text-sm", e.length > 100 && "max-h-32 overflow-y-auto"),
					children: e
				}),
				(0, W.jsxs)("div", {
					className: c("flex gap-2"),
					children: [(0, W.jsx)("button", {
						className: c("flex flex-1 items-center justify-center gap-2 rounded-md border bg-background px-4 py-2 font-medium text-sm transition-all hover:bg-muted"),
						onClick: f,
						type: "button",
						children: l ? (0, W.jsxs)(W.Fragment, { children: [(0, W.jsx)(i, { size: 14 }), (0, W.jsx)("span", { children: d.copied })] }) : (0, W.jsxs)(W.Fragment, { children: [(0, W.jsx)(a, { size: 14 }), (0, W.jsx)("span", { children: d.copyLink })] })
					}), (0, W.jsxs)("button", {
						className: c("flex flex-1 items-center justify-center gap-2 rounded-md bg-primary px-4 py-2 font-medium text-primary-foreground text-sm transition-all hover:bg-primary/90"),
						onClick: p,
						type: "button",
						children: [(0, W.jsx)(o, { size: 14 }), (0, W.jsx)("span", { children: d.openLink })]
					})]
				})
			]
		})
	}) : null;
}, Wp = (0, E.createContext)(null), Gp = () => (0, E.useContext)(Wp), Kp = () => Gp()?.code ?? null, qp = () => Gp()?.mermaid ?? null, Jp = (e) => {
	var t;
	let n = Gp();
	return n != null && n.renderers && e && (t = n.renderers.find((t) => Array.isArray(t.language) ? t.language.includes(e) : t.language === e)) != null ? t : null;
}, Yp = (e, t) => new Promise((t, n) => {
	let r = "data:image/svg+xml;base64," + btoa(unescape(encodeURIComponent(e))), i = new Image();
	i.crossOrigin = "anonymous", i.onload = () => {
		let e = document.createElement("canvas"), r = i.width * 5, a = i.height * 5;
		e.width = r, e.height = a;
		let o = e.getContext("2d");
		o ? (o.drawImage(i, 0, 0, r, a), e.toBlob((e) => {
			e ? t(e) : n(/* @__PURE__ */ Error("Failed to create PNG blob"));
		}, "image/png")) : n(/* @__PURE__ */ Error("Failed to create 2D canvas context for PNG export"));
	}, i.onerror = () => n(/* @__PURE__ */ Error("Failed to load SVG image")), i.src = r;
}), Xp = ({ chart: e, children: t, className: n, onDownload: r, config: i, onError: a }) => {
	let o = X(), [s, c] = (0, E.useState)(!1), l = (0, E.useRef)(null), { isAnimating: u } = (0, E.useContext)($), d = Z(), f = qp(), p = Np(), m = async (t) => {
		try {
			if (t === "mmd") {
				hp("diagram.mmd", e, "text/plain"), c(!1), r?.(t);
				return;
			}
			if (!f) {
				a?.(/* @__PURE__ */ Error("Mermaid plugin not available"));
				return;
			}
			let n = f.getMermaid(i), o = e.split("").reduce((e, t) => (e << 5) - e + t.charCodeAt(0) | 0, 0), s = `mermaid-${Math.abs(o)}-${Date.now()}-${Math.random().toString(36).substring(2, 9)}`, { svg: l } = await n.render(s, e);
			if (!l) {
				a?.(/* @__PURE__ */ Error("SVG not found. Please wait for the diagram to render."));
				return;
			}
			if (t === "svg") {
				hp("diagram.svg", l, "image/svg+xml"), c(!1), r?.(t);
				return;
			}
			if (t === "png") {
				hp("diagram.png", await Yp(l), "image/png"), r?.(t), c(!1);
				return;
			}
		} catch (e) {
			a?.(e);
		}
	};
	return (0, E.useEffect)(() => {
		let e = (e) => {
			let t = e.composedPath();
			l.current && !t.includes(l.current) && c(!1);
		};
		return document.addEventListener("mousedown", e), () => {
			document.removeEventListener("mousedown", e);
		};
	}, []), (0, W.jsxs)("div", {
		className: o("relative"),
		ref: l,
		children: [(0, W.jsx)("button", {
			className: o("cursor-pointer p-1 text-muted-foreground transition-all hover:text-foreground disabled:cursor-not-allowed disabled:opacity-50", n),
			disabled: u,
			onClick: () => c(!s),
			title: p.downloadDiagram,
			type: "button",
			children: t ?? (0, W.jsx)(d.DownloadIcon, { size: 14 })
		}), s ? (0, W.jsxs)("div", {
			className: o("absolute top-full right-0 z-10 mt-1 min-w-[120px] overflow-hidden rounded-md border border-border bg-background shadow-lg"),
			children: [
				(0, W.jsx)("button", {
					className: o("w-full px-3 py-2 text-left text-sm transition-colors hover:bg-muted/40"),
					onClick: () => m("svg"),
					title: p.downloadDiagramAsSvg,
					type: "button",
					children: p.mermaidFormatSvg
				}),
				(0, W.jsx)("button", {
					className: o("w-full px-3 py-2 text-left text-sm transition-colors hover:bg-muted/40"),
					onClick: () => m("png"),
					title: p.downloadDiagramAsPng,
					type: "button",
					children: p.mermaidFormatPng
				}),
				(0, W.jsx)("button", {
					className: o("w-full px-3 py-2 text-left text-sm transition-colors hover:bg-muted/40"),
					onClick: () => m("mmd"),
					title: p.downloadDiagramAsMmd,
					type: "button",
					children: p.mermaidFormatMmd
				})
			]
		}) : null]
	});
}, Zp = ({ chart: e, config: t, onFullscreen: n, onExit: r, className: i, ...a }) => {
	let { Maximize2Icon: o, XIcon: s } = Z(), c = X(), [l, u] = (0, E.useState)(!1), { isAnimating: d, controls: f } = (0, E.useContext)($), p = Np(), m = (() => {
		if (typeof f == "boolean") return f;
		let e = f.mermaid;
		return e === !1 ? !1 : e === !0 || e === void 0 || e.panZoom !== !1;
	})(), h = () => {
		u(!l);
	};
	return (0, E.useEffect)(() => {
		if (l) {
			Vp();
			let e = (e) => {
				e.key === "Escape" && u(!1);
			};
			return document.addEventListener("keydown", e), () => {
				document.removeEventListener("keydown", e), Hp();
			};
		}
	}, [l]), (0, E.useEffect)(() => {
		l ? n?.() : r && r();
	}, [
		l,
		n,
		r
	]), (0, W.jsxs)(W.Fragment, { children: [(0, W.jsx)("button", {
		className: c("cursor-pointer p-1 text-muted-foreground transition-all hover:text-foreground disabled:cursor-not-allowed disabled:opacity-50", i),
		disabled: d,
		onClick: h,
		title: p.viewFullscreen,
		type: "button",
		...a,
		children: (0, W.jsx)(o, { size: 14 })
	}), l ? (0, Sd.createPortal)((0, W.jsxs)("div", {
		className: c("fixed inset-0 z-50 flex items-center justify-center bg-background/95 backdrop-blur-sm"),
		onClick: h,
		onKeyDown: (e) => {
			e.key === "Escape" && h();
		},
		role: "button",
		tabIndex: 0,
		children: [(0, W.jsx)("button", {
			className: c("absolute top-4 right-4 z-10 rounded-md p-2 text-muted-foreground transition-all hover:bg-muted hover:text-foreground"),
			onClick: h,
			title: p.exitFullscreen,
			type: "button",
			children: (0, W.jsx)(s, { size: 20 })
		}), (0, W.jsx)("div", {
			className: c("flex size-full items-center justify-center p-4"),
			onClick: (e) => e.stopPropagation(),
			onKeyDown: (e) => e.stopPropagation(),
			role: "presentation",
			children: (0, W.jsx)(Hh, {
				chart: e,
				className: c("size-full [&_svg]:h-auto [&_svg]:w-auto"),
				config: t,
				fullscreen: !0,
				showControls: m
			})
		})]
	}), document.body) : null] });
}, Qp = (e) => {
	let t = [], n = [], r = e.querySelectorAll("thead th");
	for (let e of r) t.push(e.textContent?.trim() || "");
	let i = e.querySelectorAll("tbody tr");
	for (let e of i) {
		let t = [], r = e.querySelectorAll("td");
		for (let e of r) t.push(e.textContent?.trim() || "");
		n.push(t);
	}
	return {
		headers: t,
		rows: n
	};
}, $p = (e) => {
	let { headers: t, rows: n } = e, r = (e) => {
		let t = !1, n = !1;
		for (let r of e) {
			if (r === "\"") {
				t = !0, n = !0;
				break;
			}
			(r === "," || r === "\n") && (t = !0);
		}
		return t ? n ? `"${e.replace(/"/g, "\"\"")}"` : `"${e}"` : e;
	}, i = t.length > 0 ? n.length + 1 : n.length, a = Array(i), o = 0;
	t.length > 0 && (a[o] = t.map(r).join(","), o += 1);
	for (let e of n) a[o] = e.map(r).join(","), o += 1;
	return a.join("\n");
}, em = (e) => {
	let { headers: t, rows: n } = e, r = (e) => {
		let t = !1;
		for (let n of e) if (n === "	" || n === "\n" || n === "\r") {
			t = !0;
			break;
		}
		if (!t) return e;
		let n = [];
		for (let t of e) t === "	" ? n.push("\\t") : t === "\n" ? n.push("\\n") : t === "\r" ? n.push("\\r") : n.push(t);
		return n.join("");
	}, i = t.length > 0 ? n.length + 1 : n.length, a = Array(i), o = 0;
	t.length > 0 && (a[o] = t.map(r).join("	"), o += 1);
	for (let e of n) a[o] = e.map(r).join("	"), o += 1;
	return a.join("\n");
}, tm = (e) => {
	let t = !1;
	for (let n of e) if (n === "\\" || n === "|") {
		t = !0;
		break;
	}
	if (!t) return e;
	let n = [];
	for (let t of e) t === "\\" ? n.push("\\\\") : t === "|" ? n.push("\\|") : n.push(t);
	return n.join("");
}, nm = (e) => {
	let { headers: t, rows: n } = e;
	if (t.length === 0) return "";
	let r = Array(n.length + 2), i = 0, a = t.map((e) => tm(e));
	r[i] = `| ${a.join(" | ")} |`, i += 1;
	let o = Array(t.length);
	for (let e = 0; e < t.length; e += 1) o[e] = "---";
	r[i] = `| ${o.join(" | ")} |`, i += 1;
	for (let e of n) if (e.length < t.length) {
		let n = Array(t.length);
		for (let r = 0; r < t.length; r += 1) n[r] = r < e.length ? tm(e[r]) : "";
		r[i] = `| ${n.join(" | ")} |`, i += 1;
	} else {
		let t = e.map((e) => tm(e));
		r[i] = `| ${t.join(" | ")} |`, i += 1;
	}
	return r.join("\n");
}, rm = ({ children: e, className: t, onCopy: n, onError: r, timeout: i = 2e3 }) => {
	let a = X(), [o, s] = (0, E.useState)(!1), [c, l] = (0, E.useState)(!1), u = (0, E.useRef)(null), d = (0, E.useRef)(0), { isAnimating: f } = (0, E.useContext)($), p = Np(), m = async (e) => {
		var t;
		if (typeof window > "u" || !((t = navigator == null ? void 0 : navigator.clipboard) != null && t.write)) r?.(/* @__PURE__ */ Error("Clipboard API not available"));
		else try {
			let t = (u.current?.closest("[data-streamdown=\"table-wrapper\"]"))?.querySelector("table");
			if (!t) {
				r?.(/* @__PURE__ */ Error("Table not found"));
				return;
			}
			let a = Qp(t), o = ({
				csv: $p,
				tsv: em,
				md: nm
			}[e] || nm)(a), c = new ClipboardItem({
				"text/plain": new Blob([o], { type: "text/plain" }),
				"text/html": new Blob([t.outerHTML], { type: "text/html" })
			});
			await navigator.clipboard.write([c]), l(!0), s(!1), n?.(e), d.current = window.setTimeout(() => l(!1), i);
		} catch (e) {
			r?.(e);
		}
	};
	(0, E.useEffect)(() => {
		let e = (e) => {
			let t = e.composedPath();
			u.current && !t.includes(u.current) && s(!1);
		};
		return document.addEventListener("mousedown", e), () => {
			document.removeEventListener("mousedown", e), window.clearTimeout(d.current);
		};
	}, []);
	let h = Z(), g = c ? h.CheckIcon : h.CopyIcon;
	return (0, W.jsxs)("div", {
		className: a("relative"),
		ref: u,
		children: [(0, W.jsx)("button", {
			className: a("cursor-pointer p-1 text-muted-foreground transition-all hover:text-foreground disabled:cursor-not-allowed disabled:opacity-50", t),
			disabled: f,
			onClick: () => s(!o),
			title: p.copyTable,
			type: "button",
			children: e ?? (0, W.jsx)(g, {
				height: 14,
				width: 14
			})
		}), o ? (0, W.jsxs)("div", {
			className: a("absolute top-full right-0 z-20 mt-1 min-w-[120px] overflow-hidden rounded-md border border-border bg-background shadow-lg"),
			children: [
				(0, W.jsx)("button", {
					className: a("w-full px-3 py-2 text-left text-sm transition-colors hover:bg-muted/40"),
					onClick: () => m("md"),
					title: p.copyTableAsMarkdown,
					type: "button",
					children: p.tableFormatMarkdown
				}),
				(0, W.jsx)("button", {
					className: a("w-full px-3 py-2 text-left text-sm transition-colors hover:bg-muted/40"),
					onClick: () => m("csv"),
					title: p.copyTableAsCsv,
					type: "button",
					children: p.tableFormatCsv
				}),
				(0, W.jsx)("button", {
					className: a("w-full px-3 py-2 text-left text-sm transition-colors hover:bg-muted/40"),
					onClick: () => m("tsv"),
					title: p.copyTableAsTsv,
					type: "button",
					children: p.tableFormatTsv
				})
			]
		}) : null]
	});
}, im = ({ children: e, className: t, onDownload: n, onError: r }) => {
	let i = X(), [a, o] = (0, E.useState)(!1), s = (0, E.useRef)(null), { isAnimating: c } = (0, E.useContext)($), l = Np(), u = Z(), d = (e) => {
		try {
			let t = (s.current?.closest("[data-streamdown=\"table-wrapper\"]"))?.querySelector("table");
			if (!t) {
				r?.(/* @__PURE__ */ Error("Table not found"));
				return;
			}
			let i = Qp(t), a = e === "csv" ? $p(i) : nm(i);
			hp(`table.${e === "csv" ? "csv" : "md"}`, a, e === "csv" ? "text/csv" : "text/markdown"), o(!1), n?.(e);
		} catch (e) {
			r?.(e);
		}
	};
	return (0, E.useEffect)(() => {
		let e = (e) => {
			let t = e.composedPath();
			s.current && !t.includes(s.current) && o(!1);
		};
		return document.addEventListener("mousedown", e), () => {
			document.removeEventListener("mousedown", e);
		};
	}, []), (0, W.jsxs)("div", {
		className: i("relative"),
		ref: s,
		children: [(0, W.jsx)("button", {
			className: i("cursor-pointer p-1 text-muted-foreground transition-all hover:text-foreground disabled:cursor-not-allowed disabled:opacity-50", t),
			disabled: c,
			onClick: () => o(!a),
			title: l.downloadTable,
			type: "button",
			children: e ?? (0, W.jsx)(u.DownloadIcon, { size: 14 })
		}), a ? (0, W.jsxs)("div", {
			className: i("absolute top-full right-0 z-20 mt-1 min-w-[120px] overflow-hidden rounded-md border border-border bg-background shadow-lg"),
			children: [(0, W.jsx)("button", {
				className: i("w-full px-3 py-2 text-left text-sm transition-colors hover:bg-muted/40"),
				onClick: () => d("csv"),
				title: l.downloadTableAsCsv,
				type: "button",
				children: l.tableFormatCsv
			}), (0, W.jsx)("button", {
				className: i("w-full px-3 py-2 text-left text-sm transition-colors hover:bg-muted/40"),
				onClick: () => d("markdown"),
				title: l.downloadTableAsMarkdown,
				type: "button",
				children: l.tableFormatMarkdown
			})]
		}) : null]
	});
}, am = ({ children: e, className: t, showCopy: n = !0, showDownload: r = !0 }) => {
	let { Maximize2Icon: i, XIcon: a } = Z(), o = X(), [s, c] = (0, E.useState)(!1), { isAnimating: l } = (0, E.useContext)($), u = Np(), d = () => {
		c(!0);
	}, f = () => {
		c(!1);
	};
	return (0, E.useEffect)(() => {
		if (s) {
			Vp();
			let e = (e) => {
				e.key === "Escape" && c(!1);
			};
			return document.addEventListener("keydown", e), () => {
				document.removeEventListener("keydown", e), Hp();
			};
		}
	}, [s]), (0, W.jsxs)(W.Fragment, { children: [(0, W.jsx)("button", {
		className: o("cursor-pointer p-1 text-muted-foreground transition-all hover:text-foreground disabled:cursor-not-allowed disabled:opacity-50", t),
		disabled: l,
		onClick: d,
		title: u.viewFullscreen,
		type: "button",
		children: (0, W.jsx)(i, { size: 14 })
	}), s ? (0, Sd.createPortal)((0, W.jsx)("div", {
		"aria-label": u.viewFullscreen,
		"aria-modal": "true",
		className: o("fixed inset-0 z-50 flex flex-col bg-background"),
		"data-streamdown": "table-fullscreen",
		onClick: f,
		onKeyDown: (e) => {
			e.key === "Escape" && f();
		},
		role: "dialog",
		children: (0, W.jsxs)("div", {
			className: o("flex h-full flex-col"),
			onClick: (e) => e.stopPropagation(),
			onKeyDown: (e) => e.stopPropagation(),
			role: "presentation",
			children: [(0, W.jsxs)("div", {
				className: o("flex items-center justify-end gap-1 p-4"),
				children: [
					n ? (0, W.jsx)(rm, {}) : null,
					r ? (0, W.jsx)(im, {}) : null,
					(0, W.jsx)("button", {
						className: o("rounded-md p-1 text-muted-foreground transition-all hover:bg-muted hover:text-foreground"),
						onClick: f,
						title: u.exitFullscreen,
						type: "button",
						children: (0, W.jsx)(a, { size: 20 })
					})
				]
			}), (0, W.jsx)("div", {
				className: o("flex-1 overflow-auto p-4 pt-0 [&_thead]:sticky [&_thead]:top-0 [&_thead]:z-10"),
				children: (0, W.jsx)("table", {
					className: o("w-full border-collapse border border-border"),
					"data-streamdown": "table",
					children: e
				})
			})]
		})
	}), document.body) : null] });
}, om = ({ children: e, className: t, showControls: n, showCopy: r = !0, showDownload: i = !0, showFullscreen: a = !0, ...o }) => {
	let s = X(), c = n && r, l = n && i, u = n && a, d = c || l || u;
	return (0, W.jsxs)("div", {
		className: s("my-4 flex flex-col gap-2 rounded-lg border border-border bg-sidebar p-2"),
		"data-streamdown": "table-wrapper",
		children: [d ? (0, W.jsxs)("div", {
			className: s("flex items-center justify-end gap-1"),
			children: [
				c ? (0, W.jsx)(rm, {}) : null,
				l ? (0, W.jsx)(im, {}) : null,
				u ? (0, W.jsx)(am, {
					showCopy: c,
					showDownload: l,
					children: e
				}) : null
			]
		}) : null, (0, W.jsx)("div", {
			className: s("border-collapse overflow-x-auto overflow-y-auto rounded-md border border-border bg-background"),
			children: (0, W.jsx)("table", {
				className: s("w-full divide-y divide-border", t),
				"data-streamdown": "table",
				...o,
				children: e
			})
		})]
	});
}, sm = /startLine=(\d+)/, cm = /\bnoLineNumbers\b/, lm = (0, E.lazy)(() => import("./mermaid-GHXKKRXX-9BQfOLnE.js").then((e) => ({ default: e.Mermaid }))), um = /language-([^\s]+)/;
function dm(e, t) {
	if (!(e != null && e.position || t != null && t.position)) return !0;
	if (!(e != null && e.position && t != null && t.position)) return !1;
	let n = e.position.start, r = t.position.start, i = e.position.end, a = t.position.end;
	return n?.line === r?.line && n?.column === r?.column && i?.line === a?.line && i?.column === a?.column;
}
function Q(e, t) {
	return e.className === t.className && dm(e.node, t.node);
}
var fm = (e, t) => typeof e == "boolean" ? e : e[t] !== !1, pm = (e, t) => {
	if (typeof e == "boolean") return e;
	let n = e.table;
	return n === !1 ? !1 : n === !0 || n === void 0 || n[t] !== !1;
}, mm = (e, t) => {
	if (typeof e == "boolean") return e;
	let n = e.code;
	return n === !1 ? !1 : n === !0 || n === void 0 || n[t] !== !1;
}, hm = (e, t) => {
	if (typeof e == "boolean") return e;
	let n = e.mermaid;
	return n === !1 ? !1 : n === !0 || n === void 0 || n[t] !== !1;
}, gm = (0, E.memo)(({ children: e, className: t, node: n, ...r }) => {
	let i = X();
	return (0, W.jsx)("ol", {
		className: i("list-inside list-decimal whitespace-normal [li_&]:pl-6", t),
		"data-streamdown": "ordered-list",
		...r,
		children: e
	});
}, (e, t) => Q(e, t));
gm.displayName = "MarkdownOl";
var _m = (0, E.memo)(({ children: e, className: t, node: n, ...r }) => {
	let i = X();
	return (0, W.jsx)("li", {
		className: i("py-1 [&>p]:inline", t),
		"data-streamdown": "list-item",
		...r,
		children: e
	});
}, (e, t) => e.className === t.className && dm(e.node, t.node));
_m.displayName = "MarkdownLi";
var vm = (0, E.memo)(({ children: e, className: t, node: n, ...r }) => {
	let i = X();
	return (0, W.jsx)("ul", {
		className: i("list-inside list-disc whitespace-normal [li_&]:pl-6", t),
		"data-streamdown": "unordered-list",
		...r,
		children: e
	});
}, (e, t) => Q(e, t));
vm.displayName = "MarkdownUl";
var ym = (0, E.memo)(({ className: e, node: t, ...n }) => {
	let r = X();
	return (0, W.jsx)("hr", {
		className: r("my-6 border-border", e),
		"data-streamdown": "horizontal-rule",
		...n
	});
}, (e, t) => Q(e, t));
ym.displayName = "MarkdownHr";
var bm = (0, E.memo)(({ children: e, className: t, node: n, ...r }) => {
	let i = X();
	return (0, W.jsx)("span", {
		className: i("font-semibold", t),
		"data-streamdown": "strong",
		...r,
		children: e
	});
}, (e, t) => Q(e, t));
bm.displayName = "MarkdownStrong";
var xm = (0, E.memo)(({ children: e, className: t, href: n, node: r, ...i }) => {
	let a = X(), { linkSafety: o } = (0, E.useContext)($), [s, c] = (0, E.useState)(!1), l = n === "streamdown:incomplete-link", u = (0, E.useCallback)(async (e) => {
		if (o != null && o.enabled && n && !l) {
			if (e.preventDefault(), o.onLinkCheck && await o.onLinkCheck(n)) {
				window.open(n, "_blank", "noreferrer");
				return;
			}
			c(!0);
		}
	}, [
		o,
		n,
		l
	]), d = (0, E.useCallback)(() => {
		n && window.open(n, "_blank", "noreferrer");
	}, [n]), f = (0, E.useCallback)(() => {
		c(!1);
	}, []), p = {
		url: n ?? "",
		isOpen: s,
		onClose: f,
		onConfirm: d
	};
	return o != null && o.enabled && n ? (0, W.jsxs)(W.Fragment, { children: [(0, W.jsx)("button", {
		className: a("wrap-anywhere appearance-none text-left font-medium text-primary underline", t),
		"data-incomplete": l,
		"data-streamdown": "link",
		onClick: u,
		type: "button",
		children: e
	}), o.renderModal ? o.renderModal(p) : (0, W.jsx)(Up, { ...p })] }) : (0, W.jsx)("a", {
		className: a("wrap-anywhere font-medium text-primary underline", t),
		"data-incomplete": l,
		"data-streamdown": "link",
		href: n,
		rel: "noreferrer",
		target: "_blank",
		...i,
		children: e
	});
}, (e, t) => Q(e, t) && e.href === t.href);
xm.displayName = "MarkdownA";
var Sm = (0, E.memo)(({ children: e, className: t, node: n, ...r }) => {
	let i = X();
	return (0, W.jsx)("h1", {
		className: i("mt-6 mb-2 font-semibold text-3xl", t),
		"data-streamdown": "heading-1",
		...r,
		children: e
	});
}, (e, t) => Q(e, t));
Sm.displayName = "MarkdownH1";
var Cm = (0, E.memo)(({ children: e, className: t, node: n, ...r }) => {
	let i = X();
	return (0, W.jsx)("h2", {
		className: i("mt-6 mb-2 font-semibold text-2xl", t),
		"data-streamdown": "heading-2",
		...r,
		children: e
	});
}, (e, t) => Q(e, t));
Cm.displayName = "MarkdownH2";
var wm = (0, E.memo)(({ children: e, className: t, node: n, ...r }) => {
	let i = X();
	return (0, W.jsx)("h3", {
		className: i("mt-6 mb-2 font-semibold text-xl", t),
		"data-streamdown": "heading-3",
		...r,
		children: e
	});
}, (e, t) => Q(e, t));
wm.displayName = "MarkdownH3";
var Tm = (0, E.memo)(({ children: e, className: t, node: n, ...r }) => {
	let i = X();
	return (0, W.jsx)("h4", {
		className: i("mt-6 mb-2 font-semibold text-lg", t),
		"data-streamdown": "heading-4",
		...r,
		children: e
	});
}, (e, t) => Q(e, t));
Tm.displayName = "MarkdownH4";
var Em = (0, E.memo)(({ children: e, className: t, node: n, ...r }) => {
	let i = X();
	return (0, W.jsx)("h5", {
		className: i("mt-6 mb-2 font-semibold text-base", t),
		"data-streamdown": "heading-5",
		...r,
		children: e
	});
}, (e, t) => Q(e, t));
Em.displayName = "MarkdownH5";
var Dm = (0, E.memo)(({ children: e, className: t, node: n, ...r }) => {
	let i = X();
	return (0, W.jsx)("h6", {
		className: i("mt-6 mb-2 font-semibold text-sm", t),
		"data-streamdown": "heading-6",
		...r,
		children: e
	});
}, (e, t) => Q(e, t));
Dm.displayName = "MarkdownH6";
var Om = (0, E.memo)(({ children: e, className: t, node: n, ...r }) => {
	let { controls: i } = (0, E.useContext)($), a = fm(i, "table"), o = pm(i, "copy"), s = pm(i, "download"), c = pm(i, "fullscreen");
	return (0, W.jsx)(om, {
		className: t,
		showControls: a,
		showCopy: o,
		showDownload: s,
		showFullscreen: c,
		...r,
		children: e
	});
}, (e, t) => Q(e, t));
Om.displayName = "MarkdownTable";
var km = (0, E.memo)(({ children: e, className: t, node: n, ...r }) => {
	let i = X();
	return (0, W.jsx)("thead", {
		className: i("bg-muted/80", t),
		"data-streamdown": "table-header",
		...r,
		children: e
	});
}, (e, t) => Q(e, t));
km.displayName = "MarkdownThead";
var Am = (0, E.memo)(({ children: e, className: t, node: n, ...r }) => {
	let i = X();
	return (0, W.jsx)("tbody", {
		className: i("divide-y divide-border", t),
		"data-streamdown": "table-body",
		...r,
		children: e
	});
}, (e, t) => Q(e, t));
Am.displayName = "MarkdownTbody";
var jm = (0, E.memo)(({ children: e, className: t, node: n, ...r }) => {
	let i = X();
	return (0, W.jsx)("tr", {
		className: i("border-border", t),
		"data-streamdown": "table-row",
		...r,
		children: e
	});
}, (e, t) => Q(e, t));
jm.displayName = "MarkdownTr";
var Mm = (0, E.memo)(({ children: e, className: t, node: n, ...r }) => {
	let i = X();
	return (0, W.jsx)("th", {
		className: i("whitespace-nowrap px-4 py-2 text-left font-semibold text-sm", t),
		"data-streamdown": "table-header-cell",
		...r,
		children: e
	});
}, (e, t) => Q(e, t));
Mm.displayName = "MarkdownTh";
var Nm = (0, E.memo)(({ children: e, className: t, node: n, ...r }) => {
	let i = X();
	return (0, W.jsx)("td", {
		className: i("px-4 py-2 text-sm", t),
		"data-streamdown": "table-cell",
		...r,
		children: e
	});
}, (e, t) => Q(e, t));
Nm.displayName = "MarkdownTd";
var Pm = (0, E.memo)(({ children: e, className: t, node: n, ...r }) => {
	let i = X();
	return (0, W.jsx)("blockquote", {
		className: i("my-4 border-muted-foreground/30 border-l-4 pl-4 text-muted-foreground italic", t),
		"data-streamdown": "blockquote",
		...r,
		children: e
	});
}, (e, t) => Q(e, t));
Pm.displayName = "MarkdownBlockquote";
var Fm = (0, E.memo)(({ children: e, className: t, node: n, ...r }) => {
	let i = X();
	return (0, W.jsx)("sup", {
		className: i("text-sm", t),
		"data-streamdown": "superscript",
		...r,
		children: e
	});
}, (e, t) => Q(e, t));
Fm.displayName = "MarkdownSup";
var Im = (0, E.memo)(({ children: e, className: t, node: n, ...r }) => {
	let i = X();
	return (0, W.jsx)("sub", {
		className: i("text-sm", t),
		"data-streamdown": "subscript",
		...r,
		children: e
	});
}, (e, t) => Q(e, t));
Im.displayName = "MarkdownSub";
var Lm = (0, E.memo)(({ children: e, className: t, node: n, ...r }) => {
	if ("data-footnotes" in r) {
		let n = (e) => {
			if (!(0, E.isValidElement)(e)) return !1;
			let t = Array.isArray(e.props.children) ? e.props.children : [e.props.children], n = !1, r = !1;
			for (let e of t) if (e) {
				if (typeof e == "string") e.trim() !== "" && (n = !0);
				else if ((0, E.isValidElement)(e)) {
					if (e.props?.["data-footnote-backref"] !== void 0) r = !0;
					else {
						let t = Array.isArray(e.props.children) ? e.props.children : [e.props.children];
						for (let e of t) if (typeof e == "string" && e.trim() !== "" || (0, E.isValidElement)(e) && e.props?.["data-footnote-backref"] === void 0) {
							n = !0;
							break;
						}
					}
				}
			}
			return r && !n;
		}, i = Array.isArray(e) ? e.map((e) => {
			if (!(0, E.isValidElement)(e)) return e;
			if (e.type === gm) {
				let t = (Array.isArray(e.props.children) ? e.props.children : [e.props.children]).filter((e) => !n(e));
				return t.length === 0 ? null : {
					...e,
					props: {
						...e.props,
						children: t
					}
				};
			}
			return e;
		}) : e;
		return (Array.isArray(i) ? i.some((e) => e !== null) : i !== null) ? (0, W.jsx)("section", {
			className: t,
			...r,
			children: i
		}) : null;
	}
	return (0, W.jsx)("section", {
		className: t,
		...r,
		children: e
	});
}, (e, t) => Q(e, t));
Lm.displayName = "MarkdownSection";
var Rm = (0, E.memo)(({ node: e, className: t, children: n, ...r }) => {
	let i = X(), a = !("data-block" in r), { mermaid: o, controls: s, lineNumbers: c } = (0, E.useContext)($), l = qp(), u = dp(), d = (t?.match(um))?.at(1) ?? "", f = Jp(d);
	if (a) return (0, W.jsx)("code", {
		className: i("rounded bg-muted px-1.5 py-0.5 font-mono text-sm", t),
		"data-streamdown": "inline-code",
		...r,
		children: n
	});
	let p = e?.properties?.metastring, m = p?.match(sm), h = m ? Number.parseInt(m[1], 10) : void 0, g = h !== void 0 && h >= 1 ? h : void 0, _ = !(p && cm.test(p)) && c !== !1, v = "";
	if ((0, E.isValidElement)(n) && n.props && typeof n.props == "object" && "children" in n.props && typeof n.props.children == "string" ? v = n.props.children : typeof n == "string" && (v = n), f) {
		let e = f.component;
		return (0, W.jsx)(E.Suspense, {
			fallback: (0, W.jsx)(Lp, {}),
			children: (0, W.jsx)(e, {
				code: v,
				isIncomplete: u,
				language: d,
				meta: p
			})
		});
	}
	if (d === "mermaid" && l) {
		let e = fm(s, "mermaid"), n = hm(s, "download"), r = hm(s, "copy"), a = hm(s, "fullscreen"), c = hm(s, "panZoom"), l = e && (n || r || a);
		return (0, W.jsx)(E.Suspense, {
			fallback: (0, W.jsx)(Lp, {}),
			children: (0, W.jsxs)("div", {
				className: i("group relative my-4 flex w-full flex-col gap-2 rounded-xl border border-border bg-sidebar p-2", t),
				"data-streamdown": "mermaid-block",
				children: [
					(0, W.jsx)("div", {
						className: i("flex h-8 items-center text-muted-foreground text-xs"),
						children: (0, W.jsx)("span", {
							className: i("ml-1 font-mono lowercase"),
							children: "mermaid"
						})
					}),
					l ? (0, W.jsx)("div", {
						className: i("pointer-events-none sticky top-2 z-10 -mt-10 flex h-8 items-center justify-end"),
						children: (0, W.jsxs)("div", {
							className: i("pointer-events-auto flex shrink-0 items-center gap-2 rounded-md border border-sidebar bg-sidebar/80 px-1.5 py-1 supports-[backdrop-filter]:bg-sidebar/70 supports-[backdrop-filter]:backdrop-blur"),
							"data-streamdown": "mermaid-block-actions",
							children: [
								n ? (0, W.jsx)(Xp, {
									chart: v,
									config: o?.config
								}) : null,
								r ? (0, W.jsx)(Pp, { code: v }) : null,
								a ? (0, W.jsx)(Zp, {
									chart: v,
									config: o?.config
								}) : null
							]
						})
					}) : null,
					(0, W.jsx)("div", {
						className: i("rounded-md border border-border bg-background"),
						children: (0, W.jsx)(lm, {
							chart: v,
							config: o?.config,
							showControls: c
						})
					})
				]
			})
		});
	}
	let y = fm(s, "code"), b = mm(s, "download"), x = mm(s, "copy");
	return (0, W.jsx)(Ep, {
		className: t,
		code: v,
		isIncomplete: u,
		language: d,
		lineNumbers: _,
		startLine: g,
		children: y ? (0, W.jsxs)(W.Fragment, { children: [b ? (0, W.jsx)(Ip, {
			code: v,
			language: d
		}) : null, x ? (0, W.jsx)(Pp, {}) : null] }) : null
	});
}, (e, t) => e.className === t.className && dm(e.node, t.node));
Rm.displayName = "MarkdownCode";
var zm = (0, E.memo)(zp, (e, t) => e.className === t.className && dm(e.node, t.node));
zm.displayName = "MarkdownImg";
var Bm = (0, E.memo)(({ children: e, node: t, ...n }) => {
	let r = (Array.isArray(e) ? e : [e]).filter((e) => e != null && e !== "");
	if (r.length === 1 && (0, E.isValidElement)(r[0])) {
		let t = r[0].props.node?.tagName;
		if (t === "img" || t === "code" && "data-block" in r[0].props) return (0, W.jsx)(W.Fragment, { children: e });
	}
	return (0, W.jsx)("p", {
		...n,
		children: e
	});
}, (e, t) => Q(e, t));
Bm.displayName = "MarkdownParagraph";
var Vm = {
	ol: gm,
	li: _m,
	ul: vm,
	hr: ym,
	strong: bm,
	a: xm,
	h1: Sm,
	h2: Cm,
	h3: wm,
	h4: Tm,
	h5: Em,
	h6: Dm,
	table: Om,
	thead: km,
	tbody: Am,
	tr: jm,
	th: Mm,
	td: Nm,
	blockquote: Pm,
	code: Rm,
	img: zm,
	pre: ({ children: e }) => (0, E.isValidElement)(e) ? (0, E.cloneElement)(e, { "data-block": "true" }) : e,
	sup: Fm,
	sub: Im,
	p: Bm,
	section: Lm
}, Hm = /[\u0590-\u08FF\uFB1D-\uFDFF\uFE70-\uFEFF]/, Um = /\p{L}/u;
function Wm(e) {
	let t = e.replace(/^#{1,6}\s+/gm, "").replace(/(\*{1,3}|_{1,3})/g, "").replace(/`[^`]*`/g, "").replace(/\[([^\]]*)\]\([^)]*\)/g, "$1").replace(/^[\s>*\-+\d.]+/gm, "");
	for (let e of t) {
		if (Hm.test(e)) return "rtl";
		if (Um.test(e)) return "ltr";
	}
	return "ltr";
}
var Gm = /^[ \t]{0,3}(`{3,}|~{3,})/, Km = /^\|?[ \t]*:?-{1,}:?[ \t]*(\|[ \t]*:?-{1,}:?[ \t]*)*\|?$/, qm = (e) => {
	let t = e.split("\n"), n = null, r = 0;
	for (let e of t) {
		let t = Gm.exec(e);
		if (n === null) {
			if (t) {
				let e = t[1];
				n = e[0], r = e.length;
			}
		} else if (t) {
			let e = t[1], i = e[0], a = e.length;
			i === n && a >= r && (n = null, r = 0);
		}
	}
	return n !== null;
}, Jm = (e) => {
	let t = e.split("\n");
	for (let e of t) {
		let t = e.trim();
		if (t.length > 0 && t.includes("|") && Km.test(t)) return !0;
	}
	return !1;
}, Ym = () => (e) => {
	C(e, "html", (e, t, n) => {
		!n || typeof t != "number" || (n.children[t] = {
			type: "text",
			value: e.value
		});
	});
}, Xm = [], Zm = { allowDangerousHtml: !0 }, Qm = /* @__PURE__ */ new WeakMap(), $m = new class {
	constructor() {
		this.cache = /* @__PURE__ */ new Map(), this.keyCache = /* @__PURE__ */ new WeakMap(), this.maxSize = 100;
	}
	generateCacheKey(e) {
		let t = this.keyCache.get(e);
		if (t) return t;
		let n = e.rehypePlugins, r = e.remarkPlugins, i = e.remarkRehypeOptions;
		if (!(n || r || i)) {
			let t = "default";
			return this.keyCache.set(e, t), t;
		}
		let a = (e) => {
			if (!e || e.length === 0) return "";
			let t = "";
			for (let n = 0; n < e.length; n += 1) {
				let r = e[n];
				if (n > 0 && (t += ","), Array.isArray(r)) {
					let [e, n] = r;
					if (typeof e == "function") {
						let n = Qm.get(e);
						n || (n = e.name, Qm.set(e, n)), t += n;
					} else t += String(e);
					t += ":", t += JSON.stringify(n);
				} else if (typeof r == "function") {
					let e = Qm.get(r);
					e || (e = r.name, Qm.set(r, e)), t += e;
				} else t += String(r);
			}
			return t;
		}, o = a(n), s = `${a(r)}::${o}::${i ? JSON.stringify(i) : ""}`;
		return this.keyCache.set(e, s), s;
	}
	get(e) {
		let t = this.generateCacheKey(e), n = this.cache.get(t);
		return n && (this.cache.delete(t), this.cache.set(t, n)), n;
	}
	set(e, t) {
		let n = this.generateCacheKey(e);
		if (this.cache.size >= this.maxSize) {
			let e = this.cache.keys().next().value;
			e && this.cache.delete(e);
		}
		this.cache.set(n, t);
	}
	clear() {
		this.cache.clear();
	}
}(), eh = (e) => {
	let t = th(e), n = e.children || "";
	return ch(t.runSync(t.parse(n), n), e);
}, th = (e) => {
	let t = $m.get(e);
	if (t) return t;
	let n = rh(e);
	return $m.set(e, n), n;
}, nh = (e) => e.some((e) => Array.isArray(e) ? e[0] === Za : e === Za), rh = (e) => {
	let t = e.rehypePlugins || Xm, n = e.remarkPlugins || Xm, r = nh(t) ? n : [...n, Ym], i = e.remarkRehypeOptions ? {
		...Zm,
		...e.remarkRehypeOptions
	} : Zm;
	return fd().use(ql).use(r).use(Ru, i).use(t);
}, ih = (e) => e, ah = (e, t, n, r) => {
	n ? e.children.splice(t, 1) : e.children[t] = {
		type: "text",
		value: r
	};
}, oh = (e, t) => {
	for (let n in Kl) if (Object.hasOwn(Kl, n) && Object.hasOwn(e.properties, n)) {
		let r = e.properties[n], i = Kl[n];
		(i === null || i.includes(e.tagName)) && (e.properties[n] = t(String(r || ""), n, e) ?? void 0);
	}
}, sh = (e, t, n, r, i, a) => {
	let o = !1;
	return r ? o = !r.includes(e.tagName) : i && (o = i.includes(e.tagName)), !o && a && typeof t == "number" && (o = !a(e, t, n)), o;
}, ch = (e, t) => {
	let { allowElement: n, allowedElements: r, disallowedElements: i, skipHtml: a, unwrapDisallowed: o, urlTransform: s } = t;
	if (n || r || i || a || s) {
		let t = s || ih;
		C(e, (e, s, c) => {
			if (e.type === "raw" && c && typeof s == "number") return ah(c, s, a, e.value), s;
			if (e.type === "element" && (oh(e, t), sh(e, s, c, r, i, n) && c && typeof s == "number")) return o && e.children ? c.children.splice(s, 1, ...e.children) : c.children.splice(s, 1), s;
		});
	}
	return wl(e, {
		Fragment: W.Fragment,
		components: t.components,
		ignoreInvalidStyle: !0,
		jsx: W.jsx,
		jsxs: W.jsxs,
		passKeys: !0,
		passNode: !0
	});
}, lh = /\[\^[\w-]{1,200}\](?!:)/, uh = /\[\^[\w-]{1,200}\]:/, dh = /<(\w+)[\s>]/, fh = /* @__PURE__ */ new Set([
	"area",
	"base",
	"br",
	"col",
	"embed",
	"hr",
	"img",
	"input",
	"link",
	"meta",
	"param",
	"source",
	"track",
	"wbr"
]), ph = /* @__PURE__ */ new Map(), mh = /* @__PURE__ */ new Map(), hh = (e) => {
	let t = e.toLowerCase(), n = ph.get(t);
	if (n) return n;
	let r = RegExp(`<${t}(?=[\\s>/])[^>]*>`, "gi");
	return ph.set(t, r), r;
}, gh = (e) => {
	let t = e.toLowerCase(), n = mh.get(t);
	if (n) return n;
	let r = RegExp(`</${t}(?=[\\s>])[^>]*>`, "gi");
	return mh.set(t, r), r;
}, _h = (e, t) => {
	if (fh.has(t.toLowerCase())) return 0;
	let n = e.match(hh(t));
	if (!n) return 0;
	let r = 0;
	for (let e of n) e.trimEnd().endsWith("/>") || (r += 1);
	return r;
}, vh = (e, t) => {
	let n = e.match(gh(t));
	return n ? n.length : 0;
}, yh = (e) => {
	let t = 0;
	for (let n = 0; n < e.length - 1; n += 1) e[n] === "$" && e[n + 1] === "$" && (t += 1, n += 1);
	return t;
}, bh = (e) => {
	let t = lh.test(e), n = uh.test(e);
	if (t || n) return [e];
	let r = q.lex(e, { gfm: !0 }), i = [], a = [], o = !1;
	for (let e of r) {
		let t = e.raw, n = i.length;
		if (a.length > 0) {
			i[n - 1] += t;
			let e = a.at(-1), r = _h(t, e), o = vh(t, e);
			for (let t = 0; t < r; t += 1) a.push(e);
			for (let t = 0; t < o; t += 1) a.length > 0 && a.at(-1) === e && a.pop();
		} else {
			if (e.type === "html" && e.block) {
				let e = t.match(dh);
				if (e) {
					let n = e[1];
					_h(t, n) > vh(t, n) && a.push(n);
				}
			}
			if (n > 0 && !o) {
				let e = i[n - 1];
				if (yh(e) % 2 == 1) {
					i[n - 1] = e + t;
					continue;
				}
			}
			i.push(t), e.type !== "space" && (o = e.type === "code");
		}
	}
	return i;
}, xh = (e, t) => {
	if (!t.length) return e;
	let n = e;
	for (let e of t) {
		let t = RegExp(`(<${e}(?=[\\s>/])[^>]*>)([\\s\\S]*?)(</${e}\\s*>)`, "gi");
		n = n.replace(t, (e, t, n, r) => {
			if (!n.includes("\n\n")) return t + n + r;
			let i = n.replace(/\n\n/g, "\n<!---->\n");
			return `${t}${(i.startsWith("\n") ? "" : "\n") + i + (i.endsWith("\n") ? "" : "\n")}${r}

`;
		});
	}
	return n;
}, Sh = /([\\`*_~[\]|])/g, Ch = (e) => e.replace(Sh, "\\$1"), wh = (e, t) => {
	if (!t.length) return e;
	let n = e;
	for (let e of t) {
		let t = RegExp(`(<${e}(?=[\\s>/])[^>]*>)([\\s\\S]*?)(</${e}\\s*>)`, "gi");
		n = n.replace(t, (e, t, n, r) => t + Ch(n).replace(/\n\n/g, "&#10;&#10;") + r);
	}
	return n;
}, Th = (e) => e.type === "text" ? e.value : "children" in e && Array.isArray(e.children) ? e.children.map(Th).join("") : "", Eh = (e) => (t) => {
	if (!e || e.length === 0) return;
	let n = new Set(e.map((e) => e.toLowerCase()));
	C(t, "element", (e) => {
		if (n.has(e.tagName.toLowerCase())) {
			let t = Th(e);
			e.children = t ? [{
				type: "text",
				value: t
			}] : [];
		}
	});
}, Dh = () => (e) => {
	C(e, "code", (e) => {
		e.meta && (e.data = e.data ?? {}, e.data.hProperties = {
			...e.data.hProperties ?? {},
			metastring: e.meta
		});
	});
}, Oh = /^[ \t]*<[\w!/?-]/, kh = /(^|\n)[ \t]{4,}(?=<[\w!/?-])/g, Ah = (e) => typeof e != "string" || e.length === 0 || !Oh.test(e) ? e : e.replace(kh, "$1"), jh = {
	...$a,
	protocols: {
		...$a.protocols,
		href: [...$a.protocols?.href ?? [], "tel"]
	},
	attributes: {
		...$a.attributes,
		code: [...$a.attributes?.code ?? [], "metastring"]
	}
}, Mh = {
	raw: Za,
	sanitize: [_o, jh],
	harden: [Se, {
		allowedImagePrefixes: ["*"],
		allowedLinkPrefixes: ["*"],
		allowedProtocols: ["*"],
		defaultOrigin: void 0,
		allowDataImages: !0
	}]
}, Nh = {
	gfm: [Is, {}],
	codeMeta: Dh
}, Ph = Object.values(Mh), Fh = Object.values(Nh), Ih = {
	block: " ▋",
	circle: " ●"
}, Lh = ["github-light", "github-dark"], Rh = { enabled: !0 }, $ = (0, E.createContext)({
	shikiTheme: Lh,
	controls: !0,
	isAnimating: !1,
	lineNumbers: !0,
	mode: "streaming",
	mermaid: void 0,
	linkSafety: Rh
}), zh = (0, E.memo)(({ content: e, shouldParseIncompleteMarkdown: t, shouldNormalizeHtmlIndentation: n, index: r, isIncomplete: i, dir: a, animatePlugin: o, ...s }) => {
	if (o) {
		let e = o.getLastRenderCharCount();
		o.setPrevContentLength(e);
	}
	let c = typeof e == "string" && n ? Ah(e) : e, l = (0, W.jsx)(eh, {
		...s,
		children: c
	});
	return (0, W.jsx)(up.Provider, {
		value: i,
		children: a ? (0, W.jsx)("div", {
			dir: a,
			style: { display: "contents" },
			children: l
		}) : l
	});
}, (e, t) => {
	if (e.content !== t.content || e.shouldNormalizeHtmlIndentation !== t.shouldNormalizeHtmlIndentation || e.index !== t.index || e.isIncomplete !== t.isIncomplete || e.dir !== t.dir) return !1;
	if (e.components !== t.components) {
		let n = Object.keys(e.components || {}), r = Object.keys(t.components || {});
		if (n.length !== r.length || n.some((n) => e.components?.[n] !== t.components?.[n])) return !1;
	}
	return e.rehypePlugins === t.rehypePlugins && e.remarkPlugins === t.remarkPlugins;
});
zh.displayName = "Block";
var Bh = (0, E.memo)(({ children: e, mode: t = "streaming", dir: n, parseIncompleteMarkdown: r = !0, normalizeHtmlIndentation: i = !1, components: a, rehypePlugins: o = Ph, remarkPlugins: s = Fh, className: c, shikiTheme: l = Lh, mermaid: u, controls: d = !0, isAnimating: f = !1, animated: p, BlockComponent: m = zh, parseMarkdownIntoBlocksFn: h = bh, caret: g, plugins: _, remend: v, linkSafety: y = Rh, lineNumbers: b = !0, allowedTags: x, literalTagContent: S, translations: ee, icons: te, prefix: ne, onAnimationStart: re, onAnimationEnd: ie, ...C }) => {
	let ae = (0, E.useId)(), [oe, se] = (0, E.useTransition)(), ce = (0, E.useMemo)(() => mp(ne), [ne]), le = (0, E.useRef)(null), ue = (0, E.useRef)(re), de = (0, E.useRef)(ie);
	ue.current = re, de.current = ie, (0, E.useEffect)(() => {
		var e, n, r;
		if (t === "static") return;
		let i = le.current;
		le.current = f, i === null ? f && ((e = ue.current) == null || e.call(ue)) : f && !i ? (n = ue.current) == null || n.call(ue) : !f && i && ((r = de.current) == null || r.call(de));
	}, [f, t]);
	let fe = (0, E.useMemo)(() => x ? Object.keys(x) : [], [x]), pe = (0, E.useMemo)(() => {
		if (typeof e != "string") return "";
		let n = t === "streaming" && r ? pl(e, v) : e;
		return S && S.length > 0 && (n = wh(n, S)), fe.length > 0 && (n = xh(n, fe)), n;
	}, [
		e,
		t,
		r,
		v,
		fe,
		S
	]), me = (0, E.useMemo)(() => h(pe), [pe, h]), [he, ge] = (0, E.useState)(me);
	(0, E.useEffect)(() => {
		t === "streaming" && !T ? se(() => {
			ge(me);
		}) : ge(me);
	}, [me, t]);
	let w = t === "streaming" ? he : me, _e = (0, E.useMemo)(() => n === "auto" ? w.map(Wm) : void 0, [w, n]), ve = (0, E.useMemo)(() => w.map((e, t) => `${ae}-${t}`), [w.length, ae]), ye = (0, E.useMemo)(() => p === !0 ? "true" : p ? JSON.stringify(p) : "", [p]), T = (0, E.useMemo)(() => ye ? ye === "true" ? lp() : lp(p) : null, [ye]), be = (0, E.useMemo)(() => ({
		shikiTheme: (_?.code)?.getThemes() ?? l,
		controls: d,
		isAnimating: f,
		lineNumbers: b,
		mode: t,
		mermaid: u,
		linkSafety: y
	}), [
		l,
		d,
		f,
		b,
		t,
		u,
		y,
		_?.code
	]), xe = (0, E.useMemo)(() => ee ? JSON.stringify(ee) : "", [ee]), Se = (0, E.useMemo)(() => ({
		...jp,
		...ee
	}), [xe]), Ce = (0, E.useMemo)(() => {
		let { inlineCode: e, ...t } = a ?? {}, n = {
			...Vm,
			...t
		};
		if (e) {
			let t = n.code;
			n.code = (n) => "data-block" in n ? t ? (0, E.createElement)(t, n) : null : (0, E.createElement)(e, n);
		}
		return n;
	}, [a]), we = (0, E.useMemo)(() => {
		let e = [];
		return _ != null && _.cjk && (e = [...e, ..._.cjk.remarkPluginsBefore]), e = [...e, ...s], _ != null && _.cjk && (e = [...e, ..._.cjk.remarkPluginsAfter]), _ != null && _.math && (e = [...e, _.math.remarkPlugin]), e;
	}, [
		s,
		_?.math,
		_?.cjk
	]), Te = (0, E.useMemo)(() => {
		let e = o;
		if (x && Object.keys(x).length > 0 && o === Ph) {
			let t = {
				...jh,
				tagNames: [...jh.tagNames ?? [], ...Object.keys(x)],
				attributes: {
					...jh.attributes,
					...x
				}
			};
			e = [
				Mh.raw,
				[_o, t],
				Mh.harden
			];
		}
		return S && S.length > 0 && (e = [...e, [Eh, S]]), _ != null && _.math && (e = [...e, _.math.rehypePlugin]), T && f && (e = [...e, T.rehypePlugin]), e;
	}, [
		o,
		_?.math,
		T,
		f,
		x,
		S
	]), Ee = (0, E.useMemo)(() => {
		if (!f || w.length === 0) return !1;
		let e = w.at(-1);
		return qm(e) || Jm(e);
	}, [f, w]), De = (0, E.useMemo)(() => g && f && !Ee ? { "--streamdown-caret": `"${Ih[g]}"` } : void 0, [
		g,
		f,
		Ee
	]);
	return t === "static" ? (0, W.jsx)(Mp.Provider, {
		value: Se,
		children: (0, W.jsx)(Wp.Provider, {
			value: _ ?? null,
			children: (0, W.jsx)($.Provider, {
				value: be,
				children: (0, W.jsx)(Ap, {
					icons: te,
					children: (0, W.jsx)(gp.Provider, {
						value: ce,
						children: (0, W.jsx)("div", {
							className: ce("space-y-4 whitespace-normal [&>*:first-child]:mt-0 [&>*:last-child]:mb-0", c),
							dir: n === "auto" ? Wm(pe) : n,
							children: (0, W.jsx)(eh, {
								components: Ce,
								rehypePlugins: Te,
								remarkPlugins: we,
								...C,
								children: pe
							})
						})
					})
				})
			})
		})
	}) : (0, W.jsx)(Mp.Provider, {
		value: Se,
		children: (0, W.jsx)(Wp.Provider, {
			value: _ ?? null,
			children: (0, W.jsx)($.Provider, {
				value: be,
				children: (0, W.jsx)(Ap, {
					icons: te,
					children: (0, W.jsx)(gp.Provider, {
						value: ce,
						children: (0, W.jsxs)("div", {
							className: ce("space-y-4 whitespace-normal [&>*:first-child]:mt-0 [&>*:last-child]:mb-0", g && !Ee ? "[&>*:last-child]:after:inline [&>*:last-child]:after:align-baseline [&>*:last-child]:after:content-[var(--streamdown-caret)]" : null, c),
							style: De,
							children: [w.length === 0 && g && f && (0, W.jsx)("span", {}), w.map((e, t) => {
								let a = t === w.length - 1, o = f && a && qm(e);
								return (0, W.jsx)(m, {
									animatePlugin: T,
									components: Ce,
									content: e,
									dir: _e?.[t] ?? (n === "auto" ? void 0 : n),
									index: t,
									isIncomplete: o,
									rehypePlugins: Te,
									remarkPlugins: we,
									shouldNormalizeHtmlIndentation: i,
									shouldParseIncompleteMarkdown: r,
									...C
								}, ve[t]);
							})]
						})
					})
				})
			})
		})
	});
}, (e, t) => e.children === t.children && e.shikiTheme === t.shikiTheme && e.isAnimating === t.isAnimating && e.animated === t.animated && e.mode === t.mode && e.plugins === t.plugins && e.className === t.className && e.linkSafety === t.linkSafety && e.lineNumbers === t.lineNumbers && e.normalizeHtmlIndentation === t.normalizeHtmlIndentation && e.literalTagContent === t.literalTagContent && JSON.stringify(e.translations) === JSON.stringify(t.translations) && e.prefix === t.prefix && e.dir === t.dir);
Bh.displayName = "Streamdown";
var Vh = ({ children: e, className: t, minZoom: n = .5, maxZoom: r = 3, zoomStep: i = .1, showControls: a = !0, initialZoom: o = 1, fullscreen: s = !1 }) => {
	let { RotateCcwIcon: c, ZoomInIcon: l, ZoomOutIcon: u } = Z(), d = X(), f = (0, E.useRef)(null), p = (0, E.useRef)(null), [m, h] = (0, E.useState)(o), [g, _] = (0, E.useState)({
		x: 0,
		y: 0
	}), [v, y] = (0, E.useState)(!1), [b, x] = (0, E.useState)({
		x: 0,
		y: 0
	}), [S, ee] = (0, E.useState)({
		x: 0,
		y: 0
	}), te = (0, E.useCallback)((e) => {
		h((t) => Math.max(n, Math.min(r, t + e)));
	}, [n, r]), ne = (0, E.useCallback)(() => {
		te(i);
	}, [te, i]), re = (0, E.useCallback)(() => {
		te(-i);
	}, [te, i]), ie = (0, E.useCallback)(() => {
		h(o), _({
			x: 0,
			y: 0
		});
	}, [o]), C = (0, E.useCallback)((e) => {
		e.preventDefault();
		let t = e.deltaY > 0 ? -i : i;
		te(t);
	}, [te, i]), ae = (0, E.useCallback)((e) => {
		if (e.button !== 0 || e.isPrimary === !1) return;
		y(!0), x({
			x: e.clientX,
			y: e.clientY
		}), ee(g);
		let t = e.currentTarget;
		t instanceof HTMLElement && t.setPointerCapture(e.pointerId);
	}, [g]), oe = (0, E.useCallback)((e) => {
		if (!v) return;
		e.preventDefault();
		let t = e.clientX - b.x, n = e.clientY - b.y;
		_({
			x: S.x + t,
			y: S.y + n
		});
	}, [
		v,
		b,
		S
	]), se = (0, E.useCallback)((e) => {
		y(!1);
		let t = e.currentTarget;
		t instanceof HTMLElement && t.releasePointerCapture(e.pointerId);
	}, []);
	return (0, E.useEffect)(() => {
		let e = f.current;
		if (e) return e.addEventListener("wheel", C, { passive: !1 }), () => {
			e.removeEventListener("wheel", C);
		};
	}, [C]), (0, E.useEffect)(() => {
		let e = p.current;
		if (e && v) return document.body.style.userSelect = "none", e.addEventListener("pointermove", oe, { passive: !1 }), e.addEventListener("pointerup", se), e.addEventListener("pointercancel", se), () => {
			document.body.style.userSelect = "", e.removeEventListener("pointermove", oe), e.removeEventListener("pointerup", se), e.removeEventListener("pointercancel", se);
		};
	}, [
		v,
		oe,
		se
	]), (0, W.jsxs)("div", {
		className: d("relative flex flex-col", s ? "h-full w-full" : "min-h-28 w-full", t),
		ref: f,
		style: { cursor: v ? "grabbing" : "grab" },
		children: [a ? (0, W.jsxs)("div", {
			className: d("absolute z-10 flex flex-col gap-1 rounded-md border border-border bg-background/80 p-1 supports-[backdrop-filter]:bg-background/70 supports-[backdrop-filter]:backdrop-blur-sm", s ? "bottom-4 left-4" : "bottom-2 left-2"),
			children: [
				(0, W.jsx)("button", {
					className: d("flex items-center justify-center rounded p-1.5 text-muted-foreground transition-colors hover:bg-muted hover:text-foreground disabled:cursor-not-allowed disabled:opacity-50"),
					disabled: m >= r,
					onClick: ne,
					title: "Zoom in",
					type: "button",
					children: (0, W.jsx)(l, { size: 16 })
				}),
				(0, W.jsx)("button", {
					className: d("flex items-center justify-center rounded p-1.5 text-muted-foreground transition-colors hover:bg-muted hover:text-foreground disabled:cursor-not-allowed disabled:opacity-50"),
					disabled: m <= n,
					onClick: re,
					title: "Zoom out",
					type: "button",
					children: (0, W.jsx)(u, { size: 16 })
				}),
				(0, W.jsx)("button", {
					className: d("flex items-center justify-center rounded p-1.5 text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"),
					onClick: ie,
					title: "Reset zoom and pan",
					type: "button",
					children: (0, W.jsx)(c, { size: 16 })
				})
			]
		}) : null, (0, W.jsx)("div", {
			className: d("flex-1 origin-center transition-transform duration-150 ease-out", s ? "flex h-full w-full items-center justify-center" : "flex w-full items-center justify-center"),
			onPointerDown: ae,
			ref: p,
			role: "application",
			style: {
				transform: `translate(${g.x}px, ${g.y}px) scale(${m})`,
				transformOrigin: "center center",
				touchAction: "none",
				willChange: "transform"
			},
			children: e
		})]
	});
}, Hh = ({ chart: e, className: t, config: n, fullscreen: r = !1, showControls: i = !0 }) => {
	let a = X(), [o, s] = (0, E.useState)(null), [c, l] = (0, E.useState)(!1), [u, d] = (0, E.useState)(""), [f, p] = (0, E.useState)(""), [m, h] = (0, E.useState)(0), { mermaid: g } = (0, E.useContext)($), _ = qp(), v = g?.errorComponent, { shouldRender: y, containerRef: b } = Qf({ immediate: r });
	if ((0, E.useEffect)(() => {
		y && (_ ? (async () => {
			try {
				s(null), l(!0);
				let t = _.getMermaid(n), r = e.split("").reduce((e, t) => (e << 5) - e + t.charCodeAt(0) | 0, 0), i = `mermaid-${Math.abs(r)}-${Date.now()}-${Math.random().toString(36).substring(2, 9)}`, { svg: a } = await t.render(i, e);
				d(a), p(a);
			} catch (e) {
				if (!(f || u)) {
					let t = e instanceof Error ? e.message : "Failed to render Mermaid chart";
					s(t);
				}
			} finally {
				l(!1);
			}
		})() : s("Mermaid plugin not available. Please add the mermaid plugin to enable diagram rendering."));
	}, [
		e,
		n,
		m,
		y,
		_
	]), !(y || u || f)) return (0, W.jsx)("div", {
		className: a("my-4 min-h-[200px]", t),
		ref: b
	});
	if (c && !u && !f) return (0, W.jsx)("div", {
		className: a("my-4 flex justify-center p-4", t),
		ref: b,
		children: (0, W.jsxs)("div", {
			className: a("flex items-center space-x-2 text-muted-foreground"),
			children: [(0, W.jsx)("div", { className: a("h-4 w-4 animate-spin rounded-full border-current border-b-2") }), (0, W.jsx)("span", {
				className: a("text-sm"),
				children: "Loading diagram..."
			})]
		})
	});
	if (o && !u && !f) return v ? (0, W.jsx)("div", {
		ref: b,
		children: (0, W.jsx)(v, {
			chart: e,
			error: o,
			retry: () => h((e) => e + 1)
		})
	}) : (0, W.jsxs)("div", {
		className: a("rounded-md bg-red-50 p-4", t),
		ref: b,
		children: [(0, W.jsxs)("p", {
			className: a("font-mono text-red-700 text-sm"),
			children: ["Mermaid Error: ", o]
		}), (0, W.jsxs)("details", {
			className: a("mt-2"),
			children: [(0, W.jsx)("summary", {
				className: a("cursor-pointer text-red-600 text-xs"),
				children: "Show Code"
			}), (0, W.jsx)("pre", {
				className: a("mt-2 overflow-x-auto rounded bg-red-100 p-2 text-red-800 text-xs"),
				children: e
			})]
		})]
	});
	let x = u || f;
	return (0, W.jsx)("div", {
		className: a("size-full", t),
		"data-streamdown": "mermaid",
		ref: b,
		children: (0, W.jsx)(Vh, {
			className: a(r ? "size-full overflow-hidden" : "overflow-hidden", t),
			fullscreen: r,
			maxZoom: 3,
			minZoom: .5,
			showControls: i,
			zoomStep: .1,
			children: (0, W.jsx)("div", {
				"aria-label": "Mermaid chart",
				className: a("flex justify-center", r ? "size-full items-center" : null),
				dangerouslySetInnerHTML: { __html: x },
				role: "img"
			})
		})
	});
};
export { Bh as Streamdown, Mh as defaultRehypePlugins, Hh as i, Kp as n, $ as r, yp as t };
