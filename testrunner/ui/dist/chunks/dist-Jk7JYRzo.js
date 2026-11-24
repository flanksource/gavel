import { a as e, t } from "./rolldown-runtime-DUcC5eAW.js";
import { n, t as r } from "./jsx-runtime-B4NxTvNR.js";
import { t as i } from "./react-dom-BBpWcFtU.js";
import { n as a, r as o } from "./bundle-mjs-BzYtwfp9.js";
import { a as s, c, i as l, l as u, n as d, o as f, r as p, s as m, t as h, u as g } from "./lib-CSaF9pPm.js";
import { n as _, t as v } from "./ccount-nXP_mlOV.js";
import { B as y, F as b, G as x, H as ee, I as S, J as C, K as w, L as T, M as E, N as te, P as D, R as ne, U as re, _ as ie, a as ae, c as oe, d as se, f as ce, g as le, l as ue, m as de, n as fe, q as pe, s as O, t as me, v as he, y as ge } from "./lib-D-bqMcv1.js";
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/property-information/7.2.0/ae2783c0859d66dbd6da8f2e2132f5776838f0da409524a90d251a0decc70a69/node_modules/property-information/lib/hast-to-react.js
var _e = {
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
}, k = /* @__PURE__ */ e(n(), 1), ve = {
	indicator: "indicator",
	textOnly: "text-only",
	remove: "remove"
};
function ye({ defaultOrigin: e = "", allowedLinkPrefixes: t = [], allowedImagePrefixes: n = [], allowDataImages: r = !1, allowedProtocols: i = [], blockedImageClass: a = "inline-block bg-gray-200 dark:bg-gray-700 text-gray-600 dark:text-gray-400 px-3 py-1 rounded text-sm", blockedLinkClass: o = "text-gray-500", linkBlockPolicy: s = ve.indicator, imageBlockPolicy: c = ve.indicator }) {
	let l = t.length && !t.every((e) => e === "*"), u = n.length && !n.every((e) => e === "*");
	if (!e && (l || u)) throw Error("defaultOrigin is required when allowedLinkPrefixes or allowedImagePrefixes are provided");
	return (l) => {
		let u = Ae(e, t, n, r, i, a, o, s, c);
		Ee(l), x(l, u);
	};
}
function be(e, t) {
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
function xe(e) {
	return typeof e == "string" ? e.startsWith("/") || e.startsWith("./") || e.startsWith("../") : !1;
}
function Se(e) {
	let t = e.pathname + e.search + e.hash;
	return e.pathname.startsWith("//") ? "/." + t : t;
}
var Ce = /* @__PURE__ */ new Set([
	"https:",
	"http:",
	"irc:",
	"ircs:",
	"mailto:",
	"xmpp:",
	"blob:"
]), we = /* @__PURE__ */ new Set([
	"javascript:",
	"data:",
	"file:",
	"vbscript:"
]);
function Te(e, t, n, r = !1, i = !1, a = []) {
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
	let o = be(e, n);
	if (!o || we.has(o.protocol) || !(Ce.has(o.protocol) || a.includes(o.protocol) || a.includes("*"))) return null;
	if (o.protocol === "mailto:" || !o.protocol.match(/^https?:$/)) return o.href;
	let s = xe(e);
	return o && t.some((e) => {
		let t = be(e, n);
		return !t || t.origin !== o.origin ? !1 : o.href.startsWith(t.href);
	}) ? s ? Se(o) : o.href : t.includes("*") ? o.protocol !== "https:" && o.protocol !== "http:" ? null : s ? Se(o) : o.href : null;
}
function Ee(e) {
	if ("children" in e && Array.isArray(e.children)) {
		e.children = e.children.filter((e) => e != null);
		for (let t of e.children) Ee(t);
	}
}
var De = Symbol("node-seen");
function Oe(e, t, n) {
	return t === ve.remove ? { type: "remove" } : t === ve.textOnly ? {
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
function ke(e, t, n) {
	if (t === ve.remove) return { type: "remove" };
	if (t === ve.textOnly) {
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
var Ae = (e, t, n, r, i, a, o, s, c) => {
	let l = (u, d, f) => {
		if (u.type !== "element" || u[De]) return !0;
		if (u.tagName === "a") {
			let n = Te(u.properties.href, t, e, !1, !1, i);
			if (n === null) {
				if (u[De] = !0, x(u, l), f && typeof d == "number") {
					let e = Oe(u, s, o);
					if (e.type === "remove") return f.children.splice(d, 1), [w, d];
					f.children[d] = e.element;
				}
				return w;
			}
			return u.properties.href = n, u.properties.target = "_blank", u.properties.rel = "noopener noreferrer", !0;
		}
		if (u.tagName === "img") {
			let t = Te(u.properties.src, n, e, r, !0, i);
			if (t === null) {
				if (u[De] = !0, x(u, l), f && typeof d == "number") {
					let e = ke(u, c, a);
					if (e.type === "remove") return f.children.splice(d, 1), [w, d];
					f.children[d] = e.element;
				}
				return w;
			}
			return u.properties.src = t, !0;
		}
		return !0;
	};
	return l;
}, { defineProperty: je } = Object, Me = typeof self == "object" ? self : globalThis, Ne = (e, t) => {
	switch (e) {
		case "Function":
		case "SharedWorker":
		case "Worker":
		case "eval":
		case "setInterval":
		case "setTimeout": throw TypeError("unable to deserialize " + e);
	}
	return new Me[e](t);
}, Pe = (e, t) => {
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
				for (let [t, n] of o) {
					let i = r(t), a = r(n);
					i === "__proto__" ? je(e, i, {
						value: a,
						configurable: !0,
						enumerable: !0,
						writable: !0
					}) : e[i] = a;
				}
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
				return n(typeof Me[e] == "function" ? Ne(e, t) : Error(t), i);
			}
			case 8: return n(BigInt(o), i);
			case "BigInt": return n(Object(BigInt(o)), i);
			case "ArrayBuffer": return n(new Uint8Array(o).buffer, o);
			case "DataView": {
				let { buffer: e } = new Uint8Array(o);
				return n(new DataView(e), o);
			}
			case "-0": return -0;
		}
		return n(Ne(a, o), i);
	};
	return r;
}, Fe = (e) => Pe(/* @__PURE__ */ new Map(), e)(0), Ie = "", { toString: Le } = {}, { keys: Re, is: ze } = Object, Be = (e) => {
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
	return n.includes("Array") ? [1, n] : e instanceof Error ? [7, e.name || "Error"] : [2, n];
}, Ve = ([e, t]) => e === 0 && (t === "function" || t === "symbol"), He = (e, t, n, r) => {
	let i = (e, t) => {
		let i = r.push(e) - 1;
		return n.set(t, i), i;
	}, a = (o) => {
		if (n.has(o)) return n.get(o);
		let [s, c] = Be(o);
		switch (s) {
			case 0: {
				let t = o;
				switch (c) {
					case "bigint":
						s = 8, t = o.toString();
						break;
					case "number":
						if (!o && ze(o, -0)) return r.push(["-0"]) - 1;
						break;
					case "function":
					case "symbol":
						if (e) throw TypeError("unable to serialize " + c);
						t = null;
						break;
					case "undefined": return i([-1], o);
				}
				return i([s, t], o);
			}
			case 1: {
				if (c) {
					let e = o;
					return c === "DataView" ? e = new Uint8Array(o.buffer) : c === "ArrayBuffer" && (e = new Uint8Array(o)), i([c, [...e]], o);
				}
				let e = [], t = i([s, e], o);
				for (let t of o) e.push(a(t));
				return t;
			}
			case 2: {
				if (c) switch (c) {
					case "BigInt": return i([c, o.toString()], o);
					case "Boolean":
					case "Number":
					case "String": return i([c, o.valueOf()], o);
				}
				if (t && "toJSON" in o) return a(o.toJSON());
				let n = [], r = i([s, n], o);
				for (let t of Re(o)) (e || !Ve(Be(o[t]))) && n.push([a(t), a(o[t])]);
				return r;
			}
			case 3: return i([s, isNaN(o.getTime()) ? Ie : o.toISOString()], o);
			case 4: {
				let { source: e, flags: t } = o;
				return i([s, {
					source: e,
					flags: t
				}], o);
			}
			case 5: {
				let t = [], n = i([s, t], o);
				for (let [n, r] of o) (e || !(Ve(Be(n)) || Ve(Be(r)))) && t.push([a(n), a(r)]);
				return n;
			}
			case 6: {
				let t = [], n = i([s, t], o);
				for (let n of o) (e || !Ve(Be(n))) && t.push(a(n));
				return n;
			}
		}
		let { message: l } = o;
		return i([s, {
			name: c,
			message: l
		}], o);
	};
	return a;
}, Ue = (e, { json: t, lossy: n } = {}) => {
	let r = [];
	return He(!(t || n), !!t, /* @__PURE__ */ new Map(), r)(e), r;
}, We = typeof structuredClone == "function" ? 
/* c8 ignore start */
(e, t) => t && ("json" in t || "lossy" in t) ? Fe(Ue(e, t)) : structuredClone(e) : (e, t) => Fe(Ue(e, t)), Ge = /[#.]/g;
function Ke(e, t) {
	let n = e || "", r = {}, i = 0, a, o;
	for (; i < n.length;) {
		Ge.lastIndex = i;
		let e = Ge.exec(n), t = n.slice(i, e ? e.index : n.length);
		t && (a ? a === "#" ? r.id = t : Array.isArray(r.className) ? r.className.push(t) : r.className = [t] : o = t, i += t.length), e && (a = e[0], i++);
	}
	return {
		type: "element",
		tagName: o || t || "div",
		properties: r,
		children: []
	};
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/hastscript/9.0.3/cd3ef7ed9d120a550633de5ea73d412adfebf796c935fa62a9f34e4ca3eb1662/node_modules/hastscript/lib/create-h.js
function qe(e, t, n) {
	let r = n ? $e(n) : void 0;
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
			o = Ke(n, t);
			let s = o.tagName.toLowerCase(), c = r ? r.get(s) : void 0;
			if (o.tagName = c || s, Je(i)) a.unshift(i);
			else for (let [t, n] of Object.entries(i)) Ye(e, o.properties, t, n);
		}
		for (let e of a) Xe(o.children, e);
		return o.type === "element" && o.tagName === "template" && e.space === "html" && (o.content = {
			type: "root",
			children: o.children
		}, o.children = []), o;
	}
	return i;
}
function Je(e) {
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
function Ye(e, t, n, r) {
	if (r == null) return;
	let i = c(e, n), a;
	if (typeof r == "number") {
		if (Number.isNaN(r)) return;
		a = r;
	} else a = typeof r == "boolean" ? r : typeof r == "string" ? i.spaceSeparated ? d(r) : i.commaSeparated ? l(r) : i.commaOrSpaceSeparated ? d(l(r).join(" ")) : Ze(i, i.property, r) : Array.isArray(r) ? [...r] : i.property === "style" ? Qe(r) : String(r);
	if (Array.isArray(a)) {
		let e = [];
		for (let t of a) e.push(Ze(i, i.property, t));
		a = e;
	}
	if (i.property === "className" && Array.isArray(t.className)) {
		let e = a;
		a = [...t.className, ...Array.isArray(e) ? e : [e]];
	}
	t[i.property] = a;
}
function Xe(e, t) {
	if (t != null) {
		if (typeof t == "number" || typeof t == "string") e.push({
			type: "text",
			value: String(t)
		});
		else if (Array.isArray(t)) for (let n of t) Xe(e, n);
		else if (typeof t == "object" && "type" in t) t.type === "root" ? Xe(e, t.children) : e.push(t);
		else throw Error("Expected node, nodes, or string, got `" + t + "`");
	}
}
function Ze(e, t, n) {
	if (typeof n == "string") {
		if (e.number && n && !Number.isNaN(Number(n))) return Number(n);
		if ((e.boolean || e.overloadedBoolean) && (n === "" || u(n) === u(t))) return !0;
	}
	return n;
}
function Qe(e) {
	let t = [];
	for (let [n, r] of Object.entries(e)) t.push([n, r].join(": "));
	return t.join("; ");
}
function $e(e) {
	let t = /* @__PURE__ */ new Map();
	for (let n of e) t.set(n.toLowerCase(), n);
	return t;
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/hastscript/9.0.3/cd3ef7ed9d120a550633de5ea73d412adfebf796c935fa62a9f34e4ca3eb1662/node_modules/hastscript/lib/svg-case-sensitive-tag-names.js
var et = /* @__PURE__ */ "altGlyph.altGlyphDef.altGlyphItem.animateColor.animateMotion.animateTransform.clipPath.feBlend.feColorMatrix.feComponentTransfer.feComposite.feConvolveMatrix.feDiffuseLighting.feDisplacementMap.feDistantLight.feDropShadow.feFlood.feFuncA.feFuncB.feFuncG.feFuncR.feGaussianBlur.feImage.feMerge.feMergeNode.feMorphology.feOffset.fePointLight.feSpecularLighting.feSpotLight.feTile.feTurbulence.foreignObject.glyphRef.linearGradient.radialGradient.solidColor.textArea.textPath".split("."), tt = qe(f, "div"), nt = qe(m, "g", et);
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/vfile-location/5.0.4/70c6c1f6fe19837905163968e2c6416fddbd29d4947f39a1526afc80b236be88/node_modules/vfile-location/lib/index.js
function rt(e) {
	let t = String(e), n = [], r = -1, i = -1;
	return {
		toOffset: s,
		toPoint: o
	};
	function a() {
		let e = n.at(-1) ?? 0;
		r < e && (r = t.indexOf("\r", e), r === -1 && (r = Infinity)), i < e && (i = t.indexOf("\n", e), i === -1 && (i = Infinity));
		let a = r + 1 === i ? i : Math.min(r, i);
		n.push((a === Infinity ? t.length : a) + 1);
	}
	function o(e) {
		if (typeof e == "number" && e > -1 && e <= t.length) {
			for (; (n.at(-1) ?? 0) <= e;) a();
			let t = 0, r = n.length - 1;
			for (; t < r;) {
				let i = Math.floor((t + r) / 2);
				n[i] > e ? r = i : t = i + 1;
			}
			return {
				line: t + 1,
				column: e - (t > 0 ? n[t - 1] : 0) + 1,
				offset: e
			};
		}
	}
	function s(e) {
		if (e && typeof e.line == "number" && typeof e.column == "number" && !Number.isNaN(e.line) && e.column > 0) {
			for (; n.length < e.line && (n.at(-1) ?? 0) <= t.length;) a();
			let r = (e.line > 1 ? n[e.line - 2] : 0) + e.column - 1;
			if (r < n[e.line - 1]) return r;
		}
	}
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/web-namespaces/2.0.1/1e7c04fc430e08cefe5b44c2483610eceb3a425439cbb1ed916eb11aca6d9599/node_modules/web-namespaces/index.js
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
				let t = String(e.file), r = rt(t), i = r.toPoint(0), a = r.toPoint(t.length);
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
	let { schema: n } = e;
	e.schema = t.namespaceURI === it.svg ? m : f;
	let r = -1, i = {};
	for (; ++r < t.attrs.length;) {
		let e = t.attrs[r], n = (e.prefix ? e.prefix + ":" : "") + e.name;
		at.call(ot, n) || (i[n] = e.value);
	}
	let a = (e.schema.space === "svg" ? nt : tt)(t.tagName, i, lt(e, t.childNodes));
	if (dt(e, t, a), a.tagName === "template") {
		if (t.namespaceURI === it.html) {
			let n = t, r = n.sourceCodeLocation, i = r && r.startTag && pt(r.startTag), o = r && r.endTag && pt(r.endTag), s = ct(e, n.content);
			i && o && e.file && (s.position = {
				start: i.end,
				end: o.start
			}), a.content = s;
		} else a.content && (a.children = a.content.children, delete a.content);
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
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/hast-util-to-parse5/8.0.1/921a42925cfb74e1d0163e3011d24c4d762ae34e51a1d8f95346ca71a3b74827/node_modules/hast-util-to-parse5/lib/index.js
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
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/parse5/8.1.0/38d3691d20fe3ab4780a57194f923f73d434662a4115031b2d4cf9cab8ac9fb9/node_modules/parse5/dist/common/unicode.js
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
]), A;
(function(e) {
	e[e.EOF = -1] = "EOF", e[e.NULL = 0] = "NULL", e[e.TABULATION = 9] = "TABULATION", e[e.CARRIAGE_RETURN = 13] = "CARRIAGE_RETURN", e[e.LINE_FEED = 10] = "LINE_FEED", e[e.FORM_FEED = 12] = "FORM_FEED", e[e.SPACE = 32] = "SPACE", e[e.EXCLAMATION_MARK = 33] = "EXCLAMATION_MARK", e[e.QUOTATION_MARK = 34] = "QUOTATION_MARK", e[e.AMPERSAND = 38] = "AMPERSAND", e[e.APOSTROPHE = 39] = "APOSTROPHE", e[e.HYPHEN_MINUS = 45] = "HYPHEN_MINUS", e[e.SOLIDUS = 47] = "SOLIDUS", e[e.DIGIT_0 = 48] = "DIGIT_0", e[e.DIGIT_9 = 57] = "DIGIT_9", e[e.SEMICOLON = 59] = "SEMICOLON", e[e.LESS_THAN_SIGN = 60] = "LESS_THAN_SIGN", e[e.EQUALS_SIGN = 61] = "EQUALS_SIGN", e[e.GREATER_THAN_SIGN = 62] = "GREATER_THAN_SIGN", e[e.QUESTION_MARK = 63] = "QUESTION_MARK", e[e.LATIN_CAPITAL_A = 65] = "LATIN_CAPITAL_A", e[e.LATIN_CAPITAL_Z = 90] = "LATIN_CAPITAL_Z", e[e.RIGHT_SQUARE_BRACKET = 93] = "RIGHT_SQUARE_BRACKET", e[e.GRAVE_ACCENT = 96] = "GRAVE_ACCENT", e[e.LATIN_SMALL_A = 97] = "LATIN_SMALL_A", e[e.LATIN_SMALL_Z = 122] = "LATIN_SMALL_Z";
})(A ||= {});
var j = {
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
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/parse5/8.1.0/38d3691d20fe3ab4780a57194f923f73d434662a4115031b2d4cf9cab8ac9fb9/node_modules/parse5/dist/common/error-codes.js
var M;
(function(e) {
	e.controlCharacterInInputStream = "control-character-in-input-stream", e.noncharacterInInputStream = "noncharacter-in-input-stream", e.surrogateInInputStream = "surrogate-in-input-stream", e.nonVoidHtmlElementStartTagWithTrailingSolidus = "non-void-html-element-start-tag-with-trailing-solidus", e.endTagWithAttributes = "end-tag-with-attributes", e.endTagWithTrailingSolidus = "end-tag-with-trailing-solidus", e.unexpectedSolidusInTag = "unexpected-solidus-in-tag", e.unexpectedNullCharacter = "unexpected-null-character", e.unexpectedQuestionMarkInsteadOfTagName = "unexpected-question-mark-instead-of-tag-name", e.invalidFirstCharacterOfTagName = "invalid-first-character-of-tag-name", e.unexpectedEqualsSignBeforeAttributeName = "unexpected-equals-sign-before-attribute-name", e.missingEndTagName = "missing-end-tag-name", e.unexpectedCharacterInAttributeName = "unexpected-character-in-attribute-name", e.unknownNamedCharacterReference = "unknown-named-character-reference", e.missingSemicolonAfterCharacterReference = "missing-semicolon-after-character-reference", e.unexpectedCharacterAfterDoctypeSystemIdentifier = "unexpected-character-after-doctype-system-identifier", e.unexpectedCharacterInUnquotedAttributeValue = "unexpected-character-in-unquoted-attribute-value", e.eofBeforeTagName = "eof-before-tag-name", e.eofInTag = "eof-in-tag", e.missingAttributeValue = "missing-attribute-value", e.missingWhitespaceBetweenAttributes = "missing-whitespace-between-attributes", e.missingWhitespaceAfterDoctypePublicKeyword = "missing-whitespace-after-doctype-public-keyword", e.missingWhitespaceBetweenDoctypePublicAndSystemIdentifiers = "missing-whitespace-between-doctype-public-and-system-identifiers", e.missingWhitespaceAfterDoctypeSystemKeyword = "missing-whitespace-after-doctype-system-keyword", e.missingQuoteBeforeDoctypePublicIdentifier = "missing-quote-before-doctype-public-identifier", e.missingQuoteBeforeDoctypeSystemIdentifier = "missing-quote-before-doctype-system-identifier", e.missingDoctypePublicIdentifier = "missing-doctype-public-identifier", e.missingDoctypeSystemIdentifier = "missing-doctype-system-identifier", e.abruptDoctypePublicIdentifier = "abrupt-doctype-public-identifier", e.abruptDoctypeSystemIdentifier = "abrupt-doctype-system-identifier", e.cdataInHtmlContent = "cdata-in-html-content", e.incorrectlyOpenedComment = "incorrectly-opened-comment", e.eofInScriptHtmlCommentLikeText = "eof-in-script-html-comment-like-text", e.eofInDoctype = "eof-in-doctype", e.nestedComment = "nested-comment", e.abruptClosingOfEmptyComment = "abrupt-closing-of-empty-comment", e.eofInComment = "eof-in-comment", e.incorrectlyClosedComment = "incorrectly-closed-comment", e.eofInCdata = "eof-in-cdata", e.absenceOfDigitsInNumericCharacterReference = "absence-of-digits-in-numeric-character-reference", e.nullCharacterReference = "null-character-reference", e.surrogateCharacterReference = "surrogate-character-reference", e.characterReferenceOutsideUnicodeRange = "character-reference-outside-unicode-range", e.controlCharacterReference = "control-character-reference", e.noncharacterCharacterReference = "noncharacter-character-reference", e.missingWhitespaceBeforeDoctypeName = "missing-whitespace-before-doctype-name", e.missingDoctypeName = "missing-doctype-name", e.invalidCharacterSequenceAfterDoctypeName = "invalid-character-sequence-after-doctype-name", e.duplicateAttribute = "duplicate-attribute", e.nonConformingDoctype = "non-conforming-doctype", e.missingDoctype = "missing-doctype", e.misplacedDoctype = "misplaced-doctype", e.endTagWithoutMatchingOpenElement = "end-tag-without-matching-open-element", e.closingOfElementWithOpenChildElements = "closing-of-element-with-open-child-elements", e.disallowedContentInNoscriptInHead = "disallowed-content-in-noscript-in-head", e.openElementsLeftAfterEof = "open-elements-left-after-eof", e.abandonedHeadElementChild = "abandoned-head-element-child", e.misplacedStartTagForHeadElement = "misplaced-start-tag-for-head-element", e.nestedNoscriptInHead = "nested-noscript-in-head", e.eofInElementThatCanContainOnlyText = "eof-in-element-that-can-contain-only-text";
})(M ||= {});
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/parse5/8.1.0/38d3691d20fe3ab4780a57194f923f73d434662a4115031b2d4cf9cab8ac9fb9/node_modules/parse5/dist/tokenizer/preprocessor.js
var Pt = 65536, Ft = class {
	constructor(e) {
		this.handler = e, this.html = "", this.pos = -1, this.lastGapPos = -2, this.gapStack = [], this.skipNextNewLine = !1, this.lastChunkWritten = !1, this.endOfChunkHit = !1, this.bufferWaterline = Pt, this.isEol = !1, this.lineStartPos = 0, this.droppedBufferSize = 0, this.line = 1, this.lastErrOffset = -1;
	}
	get col() {
		return this.pos - this.lineStartPos + Number(this.lastGapPos !== this.pos);
	}
	get codePointStart() {
		return this.pos - Number(this.lastGapPos === this.pos);
	}
	get offset() {
		return this.droppedBufferSize + this.codePointStart;
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
		} else if (!this.lastChunkWritten) return this.endOfChunkHit = !0, A.EOF;
		return this._err(M.surrogateInInputStream), e;
	}
	willDropParsedChunk() {
		return this.pos > this.bufferWaterline;
	}
	dropParsedChunk() {
		if (this.willDropParsedChunk()) {
			let e = this.offset - this.droppedBufferSize;
			this.html = this.html.substring(e), this.lineStartPos -= e, this.droppedBufferSize += e, this.pos -= e, this.lastGapPos = -2, this.gapStack.length = 0, this.pos > 0 && this._addGap();
		}
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
		if (t >= this.html.length) return this.endOfChunkHit = !this.lastChunkWritten, A.EOF;
		let n = this.html.charCodeAt(t);
		return n === A.CARRIAGE_RETURN ? A.LINE_FEED : n;
	}
	advance() {
		if (this.pos++, this.isEol && (this.isEol = !1, this.line++, this.lineStartPos = this.pos), this.pos >= this.html.length) return this.endOfChunkHit = !this.lastChunkWritten, A.EOF;
		let e = this.html.charCodeAt(this.pos);
		return e === A.CARRIAGE_RETURN ? (this.isEol = !0, this.skipNextNewLine = !0, A.LINE_FEED) : e === A.LINE_FEED && (this.isEol = !0, this.skipNextNewLine) ? (this.line--, this.skipNextNewLine = !1, this._addGap(), this.advance()) : (this.skipNextNewLine = !1, kt(e) && (e = this._processSurrogate(e)), this.handler.onParseError === null || e > 31 && e < 127 || e === A.LINE_FEED || e === A.CARRIAGE_RETURN || e > 159 && e < 64976 || this._checkForProblematicCharacters(e), e);
	}
	_checkForProblematicCharacters(e) {
		Mt(e) ? this._err(M.controlCharacterInInputStream) : Nt(e) && this._err(M.noncharacterInInputStream);
	}
	retreat(e) {
		for (this.pos -= e; this.pos < this.lastGapPos;) this.lastGapPos = this.gapStack.pop(), this.pos--;
		this.isEol = !1;
	}
}, N;
(function(e) {
	e[e.CHARACTER = 0] = "CHARACTER", e[e.NULL_CHARACTER = 1] = "NULL_CHARACTER", e[e.WHITESPACE_CHARACTER = 2] = "WHITESPACE_CHARACTER", e[e.START_TAG = 3] = "START_TAG", e[e.END_TAG = 4] = "END_TAG", e[e.COMMENT = 5] = "COMMENT", e[e.DOCTYPE = 6] = "DOCTYPE", e[e.EOF = 7] = "EOF", e[e.HIBERNATION = 8] = "HIBERNATION";
})(N ||= {});
function It(e, t) {
	for (let n = e.attrs.length - 1; n >= 0; n--) if (e.attrs[n].name === t) return e.attrs[n].value;
	return null;
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/entities/9.0.0/206fdf6804b40d193967eec89f28a97160a4760aac8416d76e1e48f74066f8fd/node_modules/entities/dist/decode-codepoint.js
var Lt = [
	8364,
	0,
	8218,
	402,
	8222,
	8230,
	8224,
	8225,
	710,
	8240,
	352,
	8249,
	338,
	0,
	381,
	0,
	0,
	8216,
	8217,
	8220,
	8221,
	8226,
	8211,
	8212,
	732,
	8482,
	353,
	8250,
	339,
	0,
	382,
	376
];
function Rt(e) {
	return e === 0 || e >= 55296 && e <= 57343 || e > 1114111;
}
function zt(e) {
	return Rt(e) ? 65533 : e >= 128 && e <= 159 && Lt[e - 128] || e;
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/entities/9.0.0/206fdf6804b40d193967eec89f28a97160a4760aac8416d76e1e48f74066f8fd/node_modules/entities/dist/generated/decode-data-html.js
var Bt = ["Q=¨`p0AEligMPacutebrevecircyfrgravelphamacrndogonpfpplyFunctionringscrsigntildeumlBackslashrvwedcyecausernoullistafropfrevescrumpeqCHcyOPYacutepitalDifferentialDyleyscaronedilirconintdotedillanterDotfrhiircleDotMinusPlusTimeslockwiseContourIntegralseCurlyDoubleQuoteQuoteolonengruentinttourIntegralpfroductunterClockwiseContourIntegralrossscrupCapDDotrahdJcyScyZcyaggerrrshvcaronyeltafriacriticalAcuteDotubleAcuteGraveTildemondfferentialDopftDotEqualubleContourIntegralDotwnArrowLeftArrowRightArrowTeeongLeftArrowRightArrowRightArrowRightArrowTeeUpArrowDownArrowVerticalBarwnArrowBarUpArrowBreveLeftRightVectorTeeVectorVectorBarRightTeeVectorVectorBarTeeArrowarrowscrtrokENGTHacutecaronircydotfrgravelementmacrptySmallSquareVerySmallSquareogonpfpsilonqualTildeilibriumscrimtaumlxistsponentialEFcyfrilledSmallSquareVerySmallSquareopfrAlluriertrfscrGJcyTammadbrevecedilircydotfrgopfreaterEqualLessFullEqualGreaterLessSlantEqualTildescrtHARDcyacektcircfrilbertSpaceopfrizontalLinescrtrokumpDownHumpEqualIEcyJligOcyacutecircydotfrgravemacrginaryIpliesntegralrsectionvisibleCommaTimesogonpftascrtildeukcymlJcircyfropfscrercyukcyKHcyJcyappacedilyfropfscrLJcyTacutembdangplacetrfrrcaronedilyeftAngleBracketrrowBarRightArrowCeilingDoubleBracketwnTeeVectorVectorBarFloorRightArrowVectorTeeArrowVectorriangleBarEqualUpDownVectorTeeVectorVectorBarVectorBararrowrightarrowssEqualGreaterFullEqualGreaterLessSlantEqualTildefrleftarrowmidotongLeftArrowRightArrowRightArrowleftarrowrightarrowrightarrowpfwerLeftArrowRightArrowscrhtroktMapcyediumSpacellintrffrinusPlusopfscruNJcyacutecaronedilyegativeMediumSpaceThickSpacenSpaceVeryThinSpacestedGreaterGreaterLessLesswLinefroBreaknBreakingSpacepftCongruentupCapDoubleVerticalBarElementqualTildexistsGreaterEqualFullEqualGreaterLessSlantEqualTildeHumpDownHumpEqualLeftTriangleBarEqualssEqualGreaterLessSlantEqualTildeNestedGreaterGreaterLessLessPrecedesEqualSlantEqualReverseElementightTriangleBarEqualSquareSubsetEqualpersetEqualubsetEqualcceedsEqualSlantEqualTildepersetEqualTildeEqualFullEqualTildeVerticalBarscrtildeuOEligacutecircydblacfrgravemacregaicronopfpenCurlyDoubleQuoteQuoterscrlashtildemesumlverBarraceketParenthesisPartialDcyfrhiilusMinusoincareplanepfrecedesEqualSlantEqualTildeimeoductportionalscriQUOTfropfscrRBarrEGacutengrrtlcaronedilyeverseElementquilibriumUpEquilibriumfrhoightAngleBracketrrowBarLeftArrowCeilingDoubleBracketwnTeeVectorVectorBarFloorTeeArrowVectorriangleBarEqualUpDownVectorTeeVectorVectorBarVectorBararrowopfundImpliesrightarrowscrhuleDelayedSHCHcycyOFTcyacutecaronedilircyfrhortDownArrowLeftArrowRightArrowUpArrowigmamallCircleopfqrtuareIntersectionSubsetEqualpersetEqualUnionscrtarubsetEqualcceedsEqualSlantEqualTildehThatmpersetEqualsetTHORNRADESHcycyabucaronedilyfrhereforetaickSpacenSpaceildeEqualFullEqualTildeopfripleDotscrtrokUacuterrocirbrcyevecircydblacfrgravemacrnderBarraceketParenthesisionPlusogonpfpArrowBarDownArrowDownArrowEquilibriumTeeArrowarrowdownarrowperLeftArrowRightArrowsilonringscrtildeumlVDashbarcydashleerbarticalBarLineSeparatorTildeyThinSpacefropfscrvdashWcircedgefropfscrXfriopfscrYAcyIcyUcyacutecircyfropfscrumlZHcyacutecaronydoteroWidthSpacetafropfscraacutebrevecEdircuteyeligfrgravelefsymphphamacrlgpndanddslopevgelemsdaabcdefghrtvbdsphtzarrogonpfpEacireidosproxeqringscrtympeqtildeumlwconintintbNotackcongepsilonprimesimeqrveewedgebrktbrkcongydquoecausemptyvpsirnoutahweenfrigcapircupodotplustimessqcuptartriangledownupuplusveewedgekarowlacklozengesquaretriangledownleftrightnkk12434ocknequivotopfttomwtiexDLRlrHDUduULRlrVHLRhlrboxdLRlrhDUduminusplustimesuLRlrvHLRhlrprimerevevbarscremiimeolbhsubulletmpEeqcacutepandbrcupcapupdotsretoncapsronedilircupssmdotedilmptyvnterdotfrhcyeckmarkiirEceqlearrowleftrightdRSastcircdashefnintmidscirlubsuitoloneqmmatpfnlementxesngdotintpfrodysrrarrossscrubepetdotudarrlreprsclarrppbrcapcapupdotorsrarrmlyeqprecsuccveewedgerenvearrowleftrightveewedwconintintylctydArrHaraggerlethrrshvbkarowlaccaronydaggerrrotseqegltamptyvfishtrharlriamondsuitsegammasinvideontimesonxjcylcornropollarpfteqdotminusplussquareublebarwedgewnarrowdownarrowsharpoonleftrightrbkarowcornropscryoltroktdotrifuarrharwanglezcyigrarreDDototacutestercaronircolonydotefDotrgravesdotlinterslsdotmacrptysetvsp134ngspogonpfparsllussilonvqcircolonsimlantgtrlessualsestivDDvparslrDotarrscrdotimtahumlroxclistpectationonentialefallingdotseqcyemalefiligligligriligjliglatligtnsnofopfrallkvpartintrac123456823534584556878slownscrgElacutemmadpbrevecircydotelqqslantsccdotollesfrggimeljcylEajnEapproxeqqsimopfravescrimeltccirdotlParquestrapproxrrdoteqlessqlesslesssimvertneqqnEhArrairsplfmiltrdcyrcirwbarcirceartsuitlliprconfrksearowwarowoarrmthtokleftarrowrightarrowpfrbarscrlashtrokybullpheniacutecircyecyxclffrgraveiiintntnfinotajligmacrgelinepartthofpedncarefintieodottcalegersrcallarhkprodocygonpftaprodquestscrinEdotsvvtildeukcymljcircyfrmathopfscrercyukcykappavcedilyfrgreenhcyjcyopfscrlAarrrrtailBarrEgHaracuteemptyvgranmbdangdlepquorrbfsfshklpplsimtltailesbarrbrkracekkeslducaronedililubydcaquorrdharusharsheftarrowtailharpoondownupleftarrowsrightarrowsharpoonssquigarrowthreetimesgqqslantsccdotorgessapproxdoteqgtrqgtrgtrsimfishtloorrgEhardulblkjcylarrcornerhardtrimidotoustachenEapproxeqqsimoangrrbrkngleftarrowrightarrowmapstorightarrowoparrowleftrightparflustimeswastbarzengefparltrarrcornerhardmtrisaquocrhimegqbuortroktccirdothreeimeslarrquestrPariefurdsharuharvertneqqnEmDDotacrletesepstodownleftuprkercommaydasheasuredanglefrhoicrodastcirdotnusbdulcpdrnplusodelspfpscrtposultimapmapnGgtvLeftarrowrightarrowltvRightarrowVDashdashablacutengpEidosproxturalsbspumpecapronedilongdotupydasheArrarhkrowdotquivsearimxistsfrgEeqqslantssimtrhArrarrparisdvjcylArrEarrdreftarrowrightarrowqqslantsssimtriemidopftinEdotvabcnivabcparallelsltolintrcueeceqrArrarrcwightarrowtriesccueerhortmidparallelimeqmidparqsubepeubEeseteqqcceqpEeseteqqtglildelgrianglelefteqrightequmerospvDashHarrapdashgetinfinlArretrierArrtriesimwArrarhkrownearoSacutestcircydashblacivotsoldeligfcirrgonravethbarmintlarrcirrossinetmacregaicrondnusopfparerplusrarrderoffmigoforslopevscrlasholtildemesasumlvbarparallelsimltcyercntiodmilptenkfrhivmmatoneitchforkvlanckhkvusacirbcirdouemnsimtwomointintpfundrEapcueecapproxcurlyeqeqnapproxeqqsimsimimesnEapsimodfalarlinesurfptosimurelscriuncspqfrintopfprimescruaternionsintesteqotrAarrrrtailBarrHaraceutedicemptyvngdelequorrapbfscfshklpplsimtlwtailionalsbarrbrkracekkeslducaronedililubydcaldharquorshealinepartsctgfishtloorrhardulovightarrowtailharpoondownupleftarrowsharpoonsrightarrowssquigarrowthreetimesngsingdotseqlarrharmmoustachenmidoangrrbrkparflustimespargtpolintrarrsaquocrhqbuorthreeimesriefltriuluharxsacutebquocEaproncueedilircnEapsimpolintsimydotbeeArrarhkrowctmiswartminusnxtfrownharpchcyyortmidparallelyigmafvmdoteqgElEneplusrarrlarrmallsetminusshpeparslidletesoftcylbarpfpadesuitrqcapsupssubeseteqpesetequareffrarrscretmnmiletarftarfraightepsilonphinsubEdotedotmultnEeplusrarrseteqqneqqimubpccapproxcurlyeqeqnapproxeqqsimsimmngp123EdotsubedothsolublarrmultnEeplusseteqqneqqimubpwArrarhkrownwarzligtargetubrkcaronedilydotelrecfrhere4foretasymvickapproxsimnspkapsimornildemesbardntoeapbotcirforksaprimeradeiangledownlefteqqrighteqdoteminusplussbtimepeziumscryhcytrokwixtoheadleftarrowrightarrowuArrHaracuterrbrcyevecircydarrblacharfishtrgraveharlrblklcornerroptrimacrlogonpfparrowdownarrowharpoonleftrightlussihlonuparrowsrcornerropingtriscrtdotilderifuarrmlwanglevArrBarvDashangrtrepsilonkappanothingphiiroptorhosigmaubsetneqqpsetneqqthetariangleleftrightcydasheebareqlliprbartfrltrinsubpopfproprtriscrubnEepnEezigzagwcircedbargeqierpfropfpreathscrxcapircupdtrifrhArrarrilArrarrmapnisodotpflustimerArrarrscrqcupuplustriveewedgeyacuteycircyenfricyopfscrucymlzacutecaronydoteetrftafrhcyigrarropfscrwjnjC(D%D($(D'%$$ED($'$'$%$'%E$/D'$F%'D(D&#,%%&&$%$)%+%%$E$F$'$&$(#'D&$(%$&4%($(E'%&%($&$)%*$%$%$+)()')($:&5-($'($%*&&&/%%&)%@$'$F$%&&#%%)$&$&$&$(%%%&$(%$$%&%$E$2.&0,.(.(&'%.$F%$&&&(%6)&+*),---&*/1---)-.&)*+,).%*,&,*'('2+,+)1&'1,)2&'&*('($F%'#&D%D($(E&%$$&$ED($)$'%1(2$'%E$)$'((&+$&%%$%D&$(%-#&$E$3)2$F%'%+$&#'D$$'($$($(%&%$$&$E$$$F$./'*,***'*-*($F$$#)$'%$$'$%$.$&%/$&%'$.'(#'$'$&D(D'%$$&$%D($$%&&*%($%&('+%/,($'%E%%$&$($'E%#(%$$E$F$F%'$'#'$&$'$(%$$E$F$F#'D$$(%'%%%+%%$(%'%$$2(',&,-'*'0).+)1&'('-,)'&*(*)(*/&/('/),))/&')-&'('-%1','*'''-'($E$$%+$($/+-'-',+-'-%E%/(-$&%$%'$$#&$%$-%*$E$+$F$&$$#'$($(%'%$$5+-.)+0%5)+%($E$)%1%%%$&,'(&4&*''+H'(&*-(-L-J-'-M-(&O*H&/2F2((%*(***G*M*(&W,K&+.H.-&1'/3F3(&O2(.I4('H,(().H.-.H(I.(&(+(+,+(&.$FD($$#(D(D'%$$($ED($'%&%($F$6,($$$FE'D(&&D&$)('+&'.#+$%$E$%$$$+$/%%$$%)+(+-+(%&%(&*-%$F%$C'$E$&$F#(D%$(%%%%'%$(%'%$$$%/+-*0$%$%$3)'-&-,(*(0*.,)2&(((&+(+))*0&0((/*,*)0&().&(($&%-$-$&%$$-#)%%$($($$%'%'%&%$$E$0(,(-(*$'$-$F$&%')/))/(+)1()($F$&$%&&)(%)+(+-+(&(%$%$&(+(&&C($'$'%%$%%$$(%'%$$E$+&%%K&)$'(((,(($F$+$F%'C)%%''$'&&D'%$$($ED($'$*)',&(.%&('$'%E$)*&*,%,%.%&((%(%,%/(-%%'&$'$F$(D&#($&$%$'($$%%'&$'*+'+,+(&-$E$F$F$(#($'$E$F$F#F$$$F$F#'$&$&D($'%$$E$F$F$&#'$($(%$$&$0%%$%$&$FC)$($$%D%$E&E&%$D'$$%DD($)&%%&$'&%E$$%&&&$&(&$%$&$&%&&)%*$*$*$*$*$*$*$&%(%*$&&'$&'$'%E$$%$%'%$%%%%%')%D'$F%$%&(%D(D&$*%&#'$*'*'('&*%%'&&)%$&''$'%$$'$()$%(%&%'%%&$&'$E$('&'%&''''(&('&&/.%&(&&&($($.()(+0'0'0(&%%&'$&%%&$E&G%%$F%$&&%'%&'$'$'$&$'$'$'$'$&%'$'$'$&$'$'$'$'$'$'$&&&%'$'$'$&$'$'$'$'$&(&'&(&%'$'$'$&$'$'$'$'$'$'$$($'E'$F%&%%'$%%'$''$&'%%%'$'$($#)%$&&&(&&'%&&&D%&&%$'&&E'%&%&(%$&D'%(E%'($E$&%&('%$$%&$&$'%'..()%*$*&*'*'&$&(&&&'$'(&$'($)$%&($&$'%'))&%%'&&&%E&&F$'%$'%&$F%%'$&$'$$'$))$%&&%%')$%$&(&&'%&&&%&D%')$&+*'(&((F&&.-(%&%&$*%&$(#'$&$(%'%%%%'$$)%&$(%$$$%(&%%(D%%&%($(%D$''$$&'&*''$%$%(%&%$F&)*&&$&$(&&$(%E%$&%(&&(&'&)%/%*'-'..($*%'&&$F&$%%%'$'%%'$$'%&$)$&%)#(%%D(%'$(%%G$%'%$$&$$$'%D$$E'%$&&$$%)%$%$&&$'%&(&($%%'%($$%%%$'%E$&'%%&%%'&'$$(&'%&&**'%'&&&%(%%)$'%&$&%&%%$%E$D&%%$&%&%,&,#0$%$($(%&&&%D$'$G$&%&%&$&$F%'&$'$$*D(($H$($($($'%($G%($($'%'%($'%&%%&$F#%%$$(%&($%$$($'%$$&$$%$%$&$&(%$&%&&)$*$&D'%$E$$%$$'$&$$%$%$%$$%%%''%$&$'$%&$F$'$&%%'$'$D$%%&%%&%'%(%*'%&&&)((&'&&$K%E#'$(%%%'%'&$'&'$$&$'$()&%'%'$E$*&($'%'%.'-%E%'$F%'%'$(%'C)$$E&%$$&E&$%%DD($$%'&%%'%&$'$'&%''''&%%%%&$$%'%&(&%'%$&&&(''&(&'$&%&%E%%$'D($F%%'$'&'$($'$$$%'$'E%#(%$$E$'$F$F%'$'#(($$(%$$E$($&$&$F$F#(%%%'$'$$%$$&$(%)%'%'%%'$'%%$E&%%'$(%'%'%'%'%'&'%%$&&&$'D$'%&%'($&%'&)$$(%'&%%%%$$&%&($%(&(%%$$%*,''..%'-'-1$,+,-'-%$%$&$&(%$&%&&)$*$&D'%&*'&'()''&'&$(%'%D$$%$$''$($%&$&$$%&%)%'%&$(%')'$%%%''%$&$'$%&$'&%%&%.+-')'-%.,(%&&D&&%(%'&&%$&'&$$&'%$'%)%&($%$%&$(%E%$%%'$'$%%&%($%'D$%%&%%&%'%'%'%(%'&$'$'$$*&'$K%E#(D&%%&$'&%$&&)')')%%'$(%$$'$/$E$%D'%$&&&&F&%&($($)$$&%%$($(%E$$$F%'$$%)%&#F%D&D$,(-%D%D&D$-$(%'$'%'%E%$&D&E&%&'%&(%*$D&%F(D$&&&%'%&(F%%%$$'$$%&%''$(%%F%'%'&E%')$$E$E%$&$'D'H&D%&%$&$$'%&%&$$%$&$%$$&$'%D%&%%%$&*(-&$'D'H&D'$%&%$&%($$&$FE$&%(D(F(%)$)$&%(%)$)$$&'('E'D%(%$&&&D'$(E$'%&(D(D%,%&($$%&&&D&D%*)+%%'$($%&%&%((%%%'D'$'F*%,D&%(E&$'D'$'F*%,D$&E'%%%.0%,(1%$$%$&&&%$(%'%E%'%E&D%(%'&D&D'F%'&G%F$'%''$(%%'#%D(%%$&G$%$$'%'%%%%%'$'$'%D$&E'%$$'%$$&$'%&&'%&%$$'%&%(&$&&$F$&%&%&$$%&%$&%(%F$F$%'%%%(%$$&E'%%D(&&)%D&$'#&F$''&&'$&$$%$(&&&&&$&'$E$%&$%'%&$$%*%$$()$'%%%'''$'&'%($'$G%'&'&$$$*%EE&$$%$%%%&%$&$')'*'%'*(&(&'&%&($%%&%&&%%&(''''&$'%%&%'$F%$$(#F$&$F$($F$-'&%&(%E%#(%%%'$'$&$F&&%&%)%%'$'$'%E&%%'%'$(%'$'%'%'%'%'&'%'$%'&%('$'%&%'($&%'&)$$(%'&%%%%$$&%(%&($%%$&'&'''$%%E$$(%'%D$''$($%$&$$,-'(./%(-,+(.(-(-%%%-$'%&%$$()'$'$'&%%&%&&D&&%($&'%%)$'$(%E%$%%&%($$(%'%%'$'$''$)$$#)$'$$%$%%&&%&%$&&%&%%&%&&%)%&%$$&'$'$$'%''$(%E%%%%'%)'$%%$E&&$'%'&$%)(+E$$'($($%$&&&$'$&$'$&$'$&%&'&'$'$/&&%)%%&%%$&$'D$(%$&$'%%E$()&&$$'(D&%(D%&($(&+%'$($(&+%%$&&($&$$'$F%'%'%'$&'$%0+&&%$%&$&&&$'&&'&%'$&'&'&&)%+$)&,$'%'%($%%')'*'%'*(&(&'&%$%%%$F$F$F$&$&&'&&$'&&'(%&'&'&%'$&'&&)%+$)&,$'%'%($$'%''$(%%'D'#)%$$&$(%'%$$&$($E$(('&%(&($%,(&&&%&&&E&$'E&($)%($%%$&%$&&&&&D'&%%$($'%)+'+'/%+$+(0%&&&$&(&'&%&'%)$F&$%&%'$'%1*-#'$&D(%%$'&&D'%$$'%'%&$(%DD($''$%&$()%&&%&$'E$$'%E$)%,%.,(%&%%'$'&%+$()%&&%&%&$F$'%'%%'$$'E%$)#'$&'$$'$(%+&(&*&&'$'(&$'%&('K/D(J/D&('..($%$'$%&&&%%'%'&$$E$'$G'D$F$'$'$F%G(D&F(D$)#($(&%($%'$E$F$$$$%'$F#'%&%%$'$E$'%&$$$'%&$&$&$'%E&&%'$'%&$F%'$(%&$&$(C)&$$'%$D%$E$&$F$F$&E%#)$(%$$&$(%%$E$&$)$F$F$%%%0@0000H8H40H000047P00P@001000@0B0HP@:H9T8P000LB8d@800@I0<110000QQ8400P10030005940@410020000Q02000@0200000@P0101@8000440400000000808000000000@00000000P`0000000040400@0004<000000000008000030000000000P020T400020@0@00000000000000000000000000P0000000400400000400220080000100000000000000000P000000000000P000000@00402000000000@00000000000000000000040000000000000", "Æ&ÁĂÂА𝔄ÀΑĀ⩓Ą𝔸⁡Å𝒜≔ÃÄ∖⫧⌆Б∵ℬΒ𝔅𝔹˘ℬ≎Ч©Ć⋒ⅅℭČÇĈ∰Ċ¸·ℭΧ⊙⊖⊕⊗∲”’∷⩴≡∯∮ℂ∐∳⨯𝒞⋓≍ⅅ⤑ЂЅЏ‡↡⫤ĎД∇Δ𝔇´˙˝`˜⋄ⅆ𝔻¨⃜≐∯¨⇓⇐⇔⫤⟸⟺⟹⇒⊨⇑⇕∥↓⤓⇵̑⥐⥞↽⥖⥟⇁⥗⊤↧⇓𝒟ĐŊÐÉĚÊЭĖ𝔈È∈Ē◻▫Ę𝔼Ε⩵≂⇌ℰ⩳ΗË∃ⅇФ𝔉◼▪𝔽∀ℱℱЃ>ΓϜĞĢĜГĠ𝔊⋙𝔾≥⋛≧⪢≷⩾≳𝒢≫Ъˇ^Ĥℌℋℍ─ℋĦ≎≏ЕĲЁÍÎИİℑÌℑĪⅈ⇒∬∫⋂⁣⁢Į𝕀ΙℐĨІÏĴЙ𝔍𝕁𝒥ЈЄХЌΚĶК𝔎𝕂𝒦Љ<ĹΛ⟪ℒ↞ĽĻЛ⟨←⇤⇆⌈⟦⥡⇃⥙⌊↔⥎⊣↤⥚⊲⧏⊴⥑⥠↿⥘↼⥒⇐⇔⋚≦≶⪡⩽≲𝔏⋘⇚Ŀ⟵⟷⟶⟸⟺⟹𝕃↙↘ℒ↰Ł≪⤅М ℳ𝔐∓𝕄ℳΜЊŃŇŅН​​​​≫≪\n𝔑⁠\xA0ℕ⫬≢≭∦∉≠≂̸∄≯≱≧̸≫̸≹⩾̸≵≎̸≏̸⋪⧏̸⋬≮≰≸≪̸⩽̸≴⪢̸⪡̸⊀⪯̸⋠∌⋫⧐̸⋭⊏̸⋢⊐̸⋣⊂⃒⊈⊁⪰̸⋡≿̸⊃⃒⊉≁≄≇≉∤𝒩ÑΝŒÓÔОŐ𝔒ÒŌΩΟ𝕆“‘⩔𝒪ØÕ⨷Ö‾⏞⎴⏜∂П𝔓ΦΠ±ℌℙ⪻≺⪯≼≾″∏∷∝𝒫Ψ\"𝔔ℚ𝒬⤐®Ŕ⟫↠⤖ŘŖРℜ∋⇋⥯ℜΡ⟩→⇥⇄⌉⟧⥝⇂⥕⌋⊢↦⥛⊳⧐⊵⥏⥜↾⥔⇀⥓⇒ℝ⥰⇛ℛ↱⧴ЩШЬŚ⪼ŠŞŜС𝔖↓←→↑Σ∘𝕊√□⊓⊏⊑⊐⊒⊔𝒮⋆⋐⋐⊆≻⪰≽≿∋∑⋑⊃⊇⋑Þ™ЋЦ	ΤŤŢТ𝔗∴Θ   ∼≃≅≈𝕋⃛𝒯ŦÚ↟⥉ЎŬÛУŰ𝔘ÙŪ_⏟⎵⏝⋃⊎Ų𝕌↑⤒⇅↕⥮⊥↥⇑⇕↖↗ϒΥŮ𝒰ŨÜ⊫⫫В⊩⫦⋁‖‖∣|❘≀ 𝔙𝕍𝒱⊪Ŵ⋀𝔚𝕎𝒲𝔛Ξ𝕏𝒳ЯЇЮÝŶЫ𝔜𝕐𝒴ŸЖŹŽЗŻ​Ζℨℤ𝒵áă∾∾̳∿â´аæ⁡𝔞àℵℵαā⨿&∧⩕⩜⩘⩚∠⦤∠∡⦨⦩⦪⦫⦬⦭⦮⦯∟⊾⦝∢Å⍼ą𝕒≈⩰⩯≊≋'≈≊å𝒶*≈≍ãä∳⨑⫭≌϶‵∽⋍⊽⌅⌅⎵⎶≌б„∵∵⦰϶ℬβℶ≬𝔟⋂◯⋃⨀⨁⨂⨆★▽△⨄⋁⋀⤍⧫▪▴▾◂▸␣▒░▓█=⃥≡⃥⌐𝕓⊥⊥⋈╗╔╖╓═╦╩╤╧╝╚╜╙║╬╣╠╫╢╟⧉╕╒┐┌─╥╨┬┴⊟⊞⊠╛╘┘└│╪╡╞┼┤├‵˘¦𝒷⁏∽⋍\\⧅⟈••≎⪮≏≏ć∩⩄⩉⩋⩇⩀∩︀⁁ˇ⩍čçĉ⩌⩐ċ¸⦲¢·𝔠ч✓✓χ○⧃ˆ≗↺↻®Ⓢ⊛⊚⊝≗⨐⫯⧂♣♣:≔≔,@∁∘∁ℂ≅⩭∮𝕔∐©℗↵✗𝒸⫏⫑⫐⫒⋯⤸⤵⋞⋟↶⤽∪⩈⩆⩊⊍⩅∪︀↷⤼⋞⋟⋎⋏¤↶↷⋎⋏∲∱⌭⇓⥥†ℸ↓‐⊣⤏˝ďдⅆ‡⇊⩷°δ⦱⥿𝔡⇃⇂⋄⋄♦♦¨ϝ⋲÷÷⋇⋇ђ⌞⌍$𝕕˙≐≑∸∔⊡⌆↓⇊⇃⇂⤐⌟⌌𝒹ѕ⧶đ⋱▿▾⇵⥯⦦џ⟿⩷≑é⩮ě≖ê≕эėⅇ≒𝔢⪚è⪖⪘⪙⏧ℓ⪕⪗ē∅∅∅   ŋ ę𝕖⋕⧣⩱εεϵ≖≕≂⪖⪕=≟≡⩸⧥≓⥱ℯ≐≂ηðë€!∃ℰⅇ≒ф♀ﬃﬀﬄ𝔣ﬁfj♭ﬂ▱ƒ𝕗∀⋔⫙⨍½⅓¼⅕⅙⅛⅔⅖¾⅗⅜⅘⅚⅝⅞⁄⌢𝒻≧⪌ǵγϝ⪆ğĝгġ≥⋛≥≧⩾⩾⪩⪀⪂⪄⋛︀⪔𝔤≫⋙ℷѓ≷⪒⪥⪤≩⪊⪊⪈⪈≩⋧𝕘`ℊ≳⪎⪐>⪧⩺⋗⦕⩼⪆⥸⋗⋛⪌≷≳≩︀≩︀⇔ ½ℋъ↔⥈↭ℏĥ♥♥…⊹𝔥⤥⤦⇿∻↩↪𝕙―𝒽ℏħ⁃‐í⁣îие¡⇔𝔦ìⅈ⨌∭⧜℩ĳīℑℐℑı⊷Ƶ∈℅∞⧝ı∫⊺ℤ⊺⨗⨼ёį𝕚ι⨼¿𝒾∈⋹⋵⋴⋳∈⁢ĩіïĵй𝔧ȷ𝕛𝒿јєκϰķк𝔨ĸхќ𝕜𝓀⇚⇐⤛⤎≦⪋⥢ĺ⦴ℒλ⟨⦑⟨⪅«←⇤⤟⤝↩↫⤹⥳↢⪫⤙⪭⪭︀⤌❲{[⦋⦏⦍ľļ⌈{л⤶“„⥧⥋↲≤←↢↽↼⇇↔⇆⇋↭⋋⋚≤≦⩽⩽⪨⩿⪁⪃⋚︀⪓⪅⋖⋚⪋≶≲⥼⌊𝔩≶⪑↽↼⥪▄љ≪⇇⌞⥫◺ŀ⎰⎰≨⪉⪉⪇⪇≨⋦⟬⇽⟦⟵⟷⟼⟶↫↬⦅𝕝⨭⨴∗_◊◊⧫(⦓⇆⌟⇋⥭‎⊿‹𝓁↰≲⪍⪏[‘‚ł<⪦⩹⋖⋋⋉⥶⩻⦖◃⊴◂⥊⥦≨︀≨︀∺¯♂✠✠↦↦↧↤↥▮⨩м—∡𝔪℧µ∣*⫰·−⊟∸⨪⫛…∓⊧𝕞∓𝓂∾μ⊸⊸⋙̸≫⃒≫̸⇍⇎⋘̸≪⃒≪̸⇏⊯⊮∇ń∠⃒≉⩰̸≋̸ŉ≉♮♮ℕ\xA0≎̸≏̸⩃ňņ≇⩭̸⩂н–≠⇗⤤↗↗≐̸≢⤨≂̸∄∄𝔫≧̸≱≱≧̸⩾̸⩾̸≵≯≯⇎↮⫲∋⋼⋺∋њ⇍≦̸↚‥≰↚↮≰≦̸⩽̸⩽̸≮≴≮⋪⋬∤𝕟¬∉⋹̸⋵̸∉⋷⋶∌∌⋾⋽∦∦⫽⃥∂̸⨔⊀⋠⪯̸⊀⪯̸⇏↛⤳̸↝̸↛⋫⋭⊁⋡⪰̸𝓃∤∦≁≄≄∤∦⋢⋣⊄⫅̸⊈⊂⃒⊈⫅̸⊁⪰̸⊅⫆̸⊉⊃⃒⊉⫆̸≹ñ≸⋪⋬⋫⋭ν#№ ⊭⤄≍⃒⊬≥⃒>⃒⧞⤂≤⃒<⃒⊴⃒⤃⊵⃒∼⃒⇖⤣↖↖⤧Ⓢó⊛⊚ôо⊝ő⨸⊙⦼œ⦿𝔬˛ò⧁⦵Ω∮↺⦾⦻‾⧀ōωο⦶⊖𝕠⦷⦹⊕∨↻⩝ℴℴªº⊶⩖⩗⩛ℴø⊘õ⊗⨶ö⌽∥¶∥⫳⫽∂п%.‰⊥‱𝔭φϕℳ☎π⋔ϖℏℎℏ+⨣⊞⨢∔⨥⩲±⨦⨧±⨕𝕡£≺⪳⪷≼⪯≺⪷≼⪯⪹⪵⋨≾′ℙ⪵⪹⋨∏⌮⌒⌓∝∝≾⊰𝓅ψ 𝔮⨌𝕢⁗𝓆ℍ⨖?≟\"⇛⇒⤜⤏⥤∽̱ŕ√⦳⟩⦒⦥⟩»→⥵⇥⤠⤳⤞↪↬⥅⥴↣↝⤚∶ℚ⤍❳}]⦌⦎⦐řŗ⌉}р⤷⥩””↳ℜℛℜℝ▭®⥽⌋𝔯⇁⇀⥬ρϱ→↣⇁⇀⇄⇌⇉↝⋌˚≓⇄⇌‏⎱⎱⫮⟭⇾⟧⦆𝕣⨮⨵)⦔⨒⇉›𝓇↱]’’⋌⋊▹⊵▸⧎⥨℞ś‚≻⪴⪸š≽⪰şŝ⪶⪺⋩⨓≿с⋅⊡⩦⇘⤥↘↘§;⤩∖∖✶𝔰⌢♯щш∣∥­σςς∼⩪≃≃⪞⪠⪝⪟≆⨤⥲←∖⨳⧤∣⌣⪪⪬⪬︀ь/⧄⌿𝕤♠♠∥⊓⊓︀⊔⊔︀⊏⊑⊏⊑⊐⊒⊐⊒□□▪▪→𝓈∖⌣⋆☆★ϵϕ¯⊂⫅⪽⊆⫃⫁⫋⊊⪿⥹⊂⊆⫅⊊⫋⫇⫕⫓≻⪸≽⪰⪺⪶⋩≿∑♪⊃¹²³⫆⪾⫘⊇⫄⟉⫗⥻⫂⫌⊋⫀⊃⊇⫆⊋⫌⫈⫔⫖⇙⤦↙↙⤪ß⌖τ⎴ťţт⃛⌕𝔱∴∴θϑϑ≈∼ ≈∼þ˜×⊠⨱⨰∭⤨⊤⌶⫱𝕥⫚⤩‴™▵▿◃⊴≜▹⊵◬≜⨺⨹⧍⨻⏢𝓉цћŧ≬↞↠⇑⥣ú↑ўŭûу⇅ű⥮⥾𝔲ù↿↾▀⌜⌜⌏◸ū¨ų𝕦↑↕↿↾⊎υϒυ⇈⌝⌝⌎ů◹𝓊⋰ũ▵▴⇈ü⦧⇕⫨⫩⊨⦜ϵϰ∅ϕϖ∝↕ϱς⊊︀⫋︀⊋︀⫌︀ϑ⊲⊳в⊢∨⊻≚⋮||𝔳⊲⊂⃒⊃⃒𝕧∝⊳𝓋⫋︀⊊︀⫌︀⊋︀⦚ŵ⩟∧≙℘𝔴𝕨℘≀≀𝓌⋂◯⋃▽𝔵⟺⟷ξ⟸⟵⟼⋻⨀𝕩⨁⨂⟹⟶𝓍⨆⨄△⋁⋀ýяŷы¥𝔶ї𝕪𝓎юÿźžзżℨζ𝔷ж⇝𝕫𝓏‍‌"], Vt = 2654435761, Ht = 2246822507, Ut = 1024, Wt = (() => {
	let e = /* @__PURE__ */ new Uint8Array(128);
	for (let t = 48; t <= 57; t++) e[t] = t - 47;
	for (let t = 65; t <= 90; t++) e[t] = t - 54;
	for (let t = 97; t <= 122; t++) e[t] = t - 60;
	return e;
})();
function Gt(e, t, n) {
	return e.charCodeAt(t) << 25 | e.charCodeAt(t + 1) << 18 | e.charCodeAt(t + n - 2) << 11 | Wt[e.charCodeAt(t + n - 1)] << 5 | n;
}
function Kt(e, t) {
	return (e * 3 << 3 ^ t) & 1023;
}
function qt(e, t) {
	return (Math.imul(e, Vt) >>> 16) * t >>> 16;
}
function Jt(e, t) {
	return (Math.imul(e, Ht) >>> 16) * t >>> 16;
}
function Yt(e) {
	let [t, n] = e, r = t.charCodeAt(0) - 48 << 6 | t.charCodeAt(1) - 48, i = t.charCodeAt(2) - 48 << 6 | t.charCodeAt(3) - 48, a = t.charCodeAt(4) - 48 << 6 | t.charCodeAt(5) - 48, o = 6 + i, s = o + 2 * r, c = 2 * a, l = new Int32Array(c), u = new Uint16Array(c), d = new Uint16Array(c), f = new Uint32Array(Ut), p = /* @__PURE__ */ new Map(), m = [], h = "", g = 6, _ = 0;
	for (let e = 0; e < r; e++) {
		let n = t.charCodeAt(o + 2 * e) - 35, r = t.charCodeAt(o + 2 * e + 1) - 35, i = n & 31, c = i + (r & 31), v = (r >> 5) + 1;
		h = h.slice(0, i) + t.slice(g, g + c - i), g += c - i;
		let y = Gt(h, 0, c), b = 2 * (t.charCodeAt(s + Math.floor(e / 6)) - 48 & 1 << e % 6 ? Jt : qt)(y, a);
		if (l[b] !== 0 && (b += 1), l[b] = y, c > 4) {
			let e = h.slice(2, c - 2), t = p.get(e);
			if (t === void 0) {
				p.set(e, m.length), u[b] = m.length;
				for (let t = 0; t < e.length; t += 2) m.push(e.charCodeAt(t) | e.charCodeAt(t + 1) << 16);
			} else u[b] = t;
		}
		let x = Kt(h.charCodeAt(0), h.charCodeAt(1));
		f[x] |= c <= 16 ? 1 << c - 2 : 32768, n & 32 && (f[x] |= 1 << c - 2 + 16), d[b] = _ << 2 | v - 1 | (n & 32) >> 4, _ += v;
	}
	return {
		keys: l,
		buckets: a,
		slotMidOff: u,
		slotValue: d,
		values: n,
		lengthBits: f,
		middles: new Uint32Array(m)
	};
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/entities/9.0.0/206fdf6804b40d193967eec89f28a97160a4760aac8416d76e1e48f74066f8fd/node_modules/entities/dist/decode.js
var Xt = 35, Zt = 59, Qt = 61, $t = 48, en = 97, tn = 120, nn = 65, rn = 32;
function an(e) {
	return e - $t >>> 0 <= 9;
}
function on(e) {
	return an(e) || (e | rn) - en >>> 0 <= 25;
}
function sn(e) {
	return e === Qt || on(e);
}
var P;
(function(e) {
	e[e.Legacy = 0] = "Legacy", e[e.Strict = 1] = "Strict", e[e.Attribute = 2] = "Attribute";
})(P ||= {});
var cn = /* #__PURE__ */ Yt(Bt), ln = cn.keys, un = cn.buckets, dn = cn.slotMidOff, fn = cn.lengthBits, pn = cn.middles, mn = cn.slotValue, hn = cn.values;
function gn(e, t, n) {
	let r = e.charCodeAt(t), i = e.charCodeAt(t + 1), a = e.charCodeAt(t + n - 2), o = e.charCodeAt(t + n - 1);
	if ((r | i | a | o) > 127) return -1;
	let s = r << 25 | i << 18 | a << 11 | Wt[o] << 5 | n, c = 2 * ((Math.imul(s, Vt) >>> 16) * un >>> 16);
	for (let r = 0;; r++) {
		if (ln[c] === s && (n <= 4 || _n(c, e, t, n))) return c;
		if (ln[c + 1] === s && (n <= 4 || _n(c + 1, e, t, n))) return c + 1;
		if (r === 1) return -1;
		c = 2 * ((Math.imul(s, Ht) >>> 16) * un >>> 16);
	}
}
function _n(e, t, n, r) {
	let i = dn[e], a = n + 2, o = n + r - 2;
	for (; a + 1 < o;) {
		if ((t.charCodeAt(a) | t.charCodeAt(a + 1) << 16) !== pn[i++]) return !1;
		a += 2;
	}
	return a === o || t.charCodeAt(a) === (pn[i] & 65535);
}
var vn = /* #__PURE__ */ (() => {
	let e = (/* @__PURE__ */ new Uint8Array(128)).fill(255);
	for (let t = 0; t < 10; t++) e[$t + t] = t;
	for (let t = 0; t < 6; t++) e[nn + t] = t + 10, e[en + t] = t + 10;
	return e;
})();
function yn(e, t, n) {
	let r = fn[Kt(e.charCodeAt(t), e.charCodeAt(t + 1))] >>> 16 & 31;
	for (; r !== 0;) {
		let i = 31 - Math.clz32(r);
		if (r ^= 1 << i, i + 2 > n) continue;
		let a = gn(e, t, i + 2);
		if (a >= 0 && mn[a] & 2) return a << 3 | i + 2;
	}
	return -1;
}
function bn(e, t, n) {
	let r = e.slice(t, t + 32).indexOf(";");
	if (r - 2 >>> 0 <= 29) {
		let n = gn(e, t, r);
		if (n >= 0) return n << 6 | r + 1;
	}
	if (n !== P.Strict) {
		let r = yn(e, t, Math.min(31, e.length - t));
		if (r >= 0) {
			let i = r & 7;
			if (n !== P.Attribute || t + i >= e.length || !sn(e.charCodeAt(t + i))) return r >> 3 << 6 | i;
		}
	}
	return -1;
}
var xn = 0, Sn = 1, Cn = 2, wn = 3, Tn = 4, En = class {
	emitCodePoint;
	errors;
	state = xn;
	consumed = 1;
	result = 0;
	decodeMode = P.Strict;
	replaceNumericCodePoint = zt;
	constructor(e, t) {
		this.emitCodePoint = e, this.errors = t;
	}
	stateNumericStart(e, t) {
		return t >= e.length ? -1 : (e.charCodeAt(t) | rn) === tn ? (this.state = wn, this.consumed += 1, this.stateNumericHex(e, t + 1)) : (this.state = Cn, this.stateNumericDecimal(e, t));
	}
	stateNumericHex(e, t) {
		let n = e.length, { result: r, consumed: i } = this, a = t;
		for (; a < n;) {
			let t = e.charCodeAt(a), n = vn[t];
			if (n <= 15) r = r * 16 + n;
			else return this.result = r, this.consumed = i, this.emitNumericEntity(t, 3);
			i += 1, a += 1;
		}
		return this.result = r, this.consumed = i, -1;
	}
	stateNumericDecimal(e, t) {
		let n = e.length, { result: r, consumed: i } = this, a = t;
		for (; a < n;) {
			let t = e.charCodeAt(a) - $t;
			if (t >>> 0 > 9) return this.result = r, this.consumed = i, this.emitNumericEntity(t + $t, 2);
			r = r * 10 + t, i += 1, a += 1;
		}
		return this.result = r, this.consumed = i, -1;
	}
	emitNumericEntity(e, t) {
		if (this.consumed <= t) return this.errors?.absenceOfDigitsInNumericCharacterReference(this.consumed), 0;
		if (e === Zt) this.consumed += 1;
		else if (this.decodeMode === P.Strict) return 0;
		return this.emitCodePoint(this.replaceNumericCodePoint(this.result), this.consumed), this.errors && (e !== Zt && this.errors.missingSemicolonAfterCharacterReference(), this.errors.validateNumericCharacterReference(this.result)), this.consumed;
	}
	startEntity(e) {
		this.decodeMode = e, this.state = xn, this.result = 0, this.consumed = 1;
	}
	write(e, t) {
		switch (this.state) {
			case xn: return e.charCodeAt(t) === Xt ? (this.state = Sn, this.consumed += 1, this.stateNumericStart(e, t + 1)) : (this.state = Tn, this.stateNamedEntity(e, t));
			case Sn: return this.stateNumericStart(e, t);
			case Cn: return this.stateNumericDecimal(e, t);
			case wn: return this.stateNumericHex(e, t);
			default: return this.stateNamedEntity(e, t);
		}
	}
	end() {
		switch (this.state) {
			case Tn: return this.endNamedEntity();
			case Cn: return this.emitNumericEntity(0, 2);
			case wn: return this.emitNumericEntity(0, 3);
			case Sn: return this.errors?.absenceOfDigitsInNumericCharacterReference(this.consumed), 0;
			default: return 0;
		}
	}
}, Dn = class extends En {
	runLength = 0;
	nameBuffer = /* @__PURE__ */ new Uint8Array(32);
	emitSlot(e) {
		let t = mn[e], n = t >> 2;
		this.emitCodePoint(hn.charCodeAt(n), this.consumed), t & 1 && this.emitCodePoint(hn.charCodeAt(n + 1), this.consumed);
	}
	startEntity(e) {
		super.startEntity(e), this.runLength = 0;
	}
	stateNamedEntity(e, t) {
		let n = e.length;
		if (this.runLength === 0 && t + 17 <= n) {
			let r = fn[Kt(e.charCodeAt(t), e.charCodeAt(t + 1))];
			if (r & 32768 && t + 32 <= n) {
				let n = bn(e, t, this.decodeMode);
				if (n >= 0) {
					let r = n & 63;
					return this.consumed = r + 1, this.emitSlot(n >> 6), e.charCodeAt(t + r - 1) !== Zt && this.errors?.missingSemicolonAfterCharacterReference(), this.consumed;
				}
				return 0;
			}
			let i = r & 32767;
			for (; i !== 0;) {
				let n = i & -i;
				i ^= n;
				let r = 33 - Math.clz32(n);
				if (e.charCodeAt(t + r) === Zt) {
					let n = gn(e, t, r);
					if (n >= 0) return this.consumed = r + 2, this.emitSlot(n), this.consumed;
				}
			}
			if (!(r & 32768)) {
				if (this.decodeMode !== P.Strict) {
					let n = yn(e, t, 31);
					if (n >= 0) {
						let r = n & 7, i = e.charCodeAt(t + r);
						return this.decodeMode === P.Attribute && sn(i) ? 0 : (this.consumed = r + 1, this.emitSlot(n >> 3), this.errors?.missingSemicolonAfterCharacterReference(), this.consumed);
					}
				}
				return 0;
			}
		}
		let r = t, { runLength: i } = this, a = -1, o = Math.min(n, t + 32 - i);
		for (; r < o;) {
			let t = e.charCodeAt(r);
			if (!on(t)) {
				a = t;
				break;
			}
			this.nameBuffer[i++] = t, r++;
		}
		if (a < 0 && i < 32) return this.runLength = i, -1;
		if (a === Zt && i - 2 >>> 0 <= 29) {
			let e = On(this.nameBuffer, i);
			if (e >= 0) return this.consumed = i + 2, this.emitSlot(e), this.consumed;
		}
		if (this.decodeMode !== P.Strict && i >= 2) {
			let e = An(this.nameBuffer, i);
			if (e >= 0) {
				let t = e & 7, n = t < i ? this.nameBuffer[t] : a;
				return this.decodeMode === P.Attribute && sn(n) ? 0 : (this.consumed = t + 1, this.emitSlot(e >> 3), this.errors?.missingSemicolonAfterCharacterReference(), this.consumed);
			}
		}
		return 0;
	}
	endNamedEntity() {
		if (this.decodeMode === P.Strict || this.runLength < 2) return 0;
		let e = An(this.nameBuffer, this.runLength);
		return e < 0 || this.decodeMode === P.Attribute && (e & 7) < this.runLength ? 0 : (this.consumed = (e & 7) + 1, this.emitSlot(e >> 3), this.errors?.missingSemicolonAfterCharacterReference(), this.consumed);
	}
};
function On(e, t) {
	let n = e[0] << 25 | e[1] << 18 | e[t - 2] << 11 | Wt[e[t - 1]] << 5 | t, r = 2 * ((Math.imul(n, Vt) >>> 16) * un >>> 16);
	for (let i = 0;; i++) {
		if (ln[r] === n && (t <= 4 || kn(r, e, t))) return r;
		if (ln[r + 1] === n && (t <= 4 || kn(r + 1, e, t))) return r + 1;
		if (i === 1) return -1;
		r = 2 * ((Math.imul(n, Ht) >>> 16) * un >>> 16);
	}
}
function kn(e, t, n) {
	let r = dn[e], i = 2, a = n - 2;
	for (; i + 1 < a;) {
		if ((t[i] | t[i + 1] << 16) !== pn[r++]) return !1;
		i += 2;
	}
	return i === a || t[i] === (pn[r] & 65535);
}
function An(e, t) {
	let n = fn[Kt(e[0], e[1])] >>> 16 & 31;
	for (; n !== 0;) {
		let r = 31 - Math.clz32(n);
		if (n ^= 1 << r, r + 2 > t) continue;
		let i = On(e, r + 2);
		if (i >= 0 && mn[i] & 2) return i << 3 | r + 2;
	}
	return -1;
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/parse5/8.1.0/38d3691d20fe3ab4780a57194f923f73d434662a4115031b2d4cf9cab8ac9fb9/node_modules/parse5/dist/common/html.js
var F;
(function(e) {
	e.HTML = "http://www.w3.org/1999/xhtml", e.MATHML = "http://www.w3.org/1998/Math/MathML", e.SVG = "http://www.w3.org/2000/svg", e.XLINK = "http://www.w3.org/1999/xlink", e.XML = "http://www.w3.org/XML/1998/namespace", e.XMLNS = "http://www.w3.org/2000/xmlns/";
})(F ||= {});
var jn;
(function(e) {
	e.TYPE = "type", e.ACTION = "action", e.ENCODING = "encoding", e.PROMPT = "prompt", e.NAME = "name", e.COLOR = "color", e.FACE = "face", e.SIZE = "size";
})(jn ||= {});
var Mn;
(function(e) {
	e.NO_QUIRKS = "no-quirks", e.QUIRKS = "quirks", e.LIMITED_QUIRKS = "limited-quirks";
})(Mn ||= {});
var I;
(function(e) {
	e.A = "a", e.ADDRESS = "address", e.ANNOTATION_XML = "annotation-xml", e.APPLET = "applet", e.AREA = "area", e.ARTICLE = "article", e.ASIDE = "aside", e.B = "b", e.BASE = "base", e.BASEFONT = "basefont", e.BGSOUND = "bgsound", e.BIG = "big", e.BLOCKQUOTE = "blockquote", e.BODY = "body", e.BR = "br", e.BUTTON = "button", e.CAPTION = "caption", e.CENTER = "center", e.CODE = "code", e.COL = "col", e.COLGROUP = "colgroup", e.DD = "dd", e.DESC = "desc", e.DETAILS = "details", e.DIALOG = "dialog", e.DIR = "dir", e.DIV = "div", e.DL = "dl", e.DT = "dt", e.EM = "em", e.EMBED = "embed", e.FIELDSET = "fieldset", e.FIGCAPTION = "figcaption", e.FIGURE = "figure", e.FONT = "font", e.FOOTER = "footer", e.FOREIGN_OBJECT = "foreignObject", e.FORM = "form", e.FRAME = "frame", e.FRAMESET = "frameset", e.H1 = "h1", e.H2 = "h2", e.H3 = "h3", e.H4 = "h4", e.H5 = "h5", e.H6 = "h6", e.HEAD = "head", e.HEADER = "header", e.HGROUP = "hgroup", e.HR = "hr", e.HTML = "html", e.I = "i", e.IMG = "img", e.IMAGE = "image", e.INPUT = "input", e.IFRAME = "iframe", e.KEYGEN = "keygen", e.LABEL = "label", e.LI = "li", e.LINK = "link", e.LISTING = "listing", e.MAIN = "main", e.MALIGNMARK = "malignmark", e.MARQUEE = "marquee", e.MATH = "math", e.MENU = "menu", e.META = "meta", e.MGLYPH = "mglyph", e.MI = "mi", e.MO = "mo", e.MN = "mn", e.MS = "ms", e.MTEXT = "mtext", e.NAV = "nav", e.NOBR = "nobr", e.NOFRAMES = "noframes", e.NOEMBED = "noembed", e.NOSCRIPT = "noscript", e.OBJECT = "object", e.OL = "ol", e.OPTGROUP = "optgroup", e.OPTION = "option", e.P = "p", e.PARAM = "param", e.PLAINTEXT = "plaintext", e.PRE = "pre", e.RB = "rb", e.RP = "rp", e.RT = "rt", e.RTC = "rtc", e.RUBY = "ruby", e.S = "s", e.SCRIPT = "script", e.SEARCH = "search", e.SECTION = "section", e.SELECT = "select", e.SOURCE = "source", e.SMALL = "small", e.SPAN = "span", e.STRIKE = "strike", e.STRONG = "strong", e.STYLE = "style", e.SUB = "sub", e.SUMMARY = "summary", e.SUP = "sup", e.TABLE = "table", e.TBODY = "tbody", e.TEMPLATE = "template", e.TEXTAREA = "textarea", e.TFOOT = "tfoot", e.TD = "td", e.TH = "th", e.THEAD = "thead", e.TITLE = "title", e.TR = "tr", e.TRACK = "track", e.TT = "tt", e.U = "u", e.UL = "ul", e.SVG = "svg", e.VAR = "var", e.WBR = "wbr", e.XMP = "xmp";
})(I ||= {});
var L;
(function(e) {
	e[e.UNKNOWN = 0] = "UNKNOWN", e[e.A = 1] = "A", e[e.ADDRESS = 2] = "ADDRESS", e[e.ANNOTATION_XML = 3] = "ANNOTATION_XML", e[e.APPLET = 4] = "APPLET", e[e.AREA = 5] = "AREA", e[e.ARTICLE = 6] = "ARTICLE", e[e.ASIDE = 7] = "ASIDE", e[e.B = 8] = "B", e[e.BASE = 9] = "BASE", e[e.BASEFONT = 10] = "BASEFONT", e[e.BGSOUND = 11] = "BGSOUND", e[e.BIG = 12] = "BIG", e[e.BLOCKQUOTE = 13] = "BLOCKQUOTE", e[e.BODY = 14] = "BODY", e[e.BR = 15] = "BR", e[e.BUTTON = 16] = "BUTTON", e[e.CAPTION = 17] = "CAPTION", e[e.CENTER = 18] = "CENTER", e[e.CODE = 19] = "CODE", e[e.COL = 20] = "COL", e[e.COLGROUP = 21] = "COLGROUP", e[e.DD = 22] = "DD", e[e.DESC = 23] = "DESC", e[e.DETAILS = 24] = "DETAILS", e[e.DIALOG = 25] = "DIALOG", e[e.DIR = 26] = "DIR", e[e.DIV = 27] = "DIV", e[e.DL = 28] = "DL", e[e.DT = 29] = "DT", e[e.EM = 30] = "EM", e[e.EMBED = 31] = "EMBED", e[e.FIELDSET = 32] = "FIELDSET", e[e.FIGCAPTION = 33] = "FIGCAPTION", e[e.FIGURE = 34] = "FIGURE", e[e.FONT = 35] = "FONT", e[e.FOOTER = 36] = "FOOTER", e[e.FOREIGN_OBJECT = 37] = "FOREIGN_OBJECT", e[e.FORM = 38] = "FORM", e[e.FRAME = 39] = "FRAME", e[e.FRAMESET = 40] = "FRAMESET", e[e.H1 = 41] = "H1", e[e.H2 = 42] = "H2", e[e.H3 = 43] = "H3", e[e.H4 = 44] = "H4", e[e.H5 = 45] = "H5", e[e.H6 = 46] = "H6", e[e.HEAD = 47] = "HEAD", e[e.HEADER = 48] = "HEADER", e[e.HGROUP = 49] = "HGROUP", e[e.HR = 50] = "HR", e[e.HTML = 51] = "HTML", e[e.I = 52] = "I", e[e.IMG = 53] = "IMG", e[e.IMAGE = 54] = "IMAGE", e[e.INPUT = 55] = "INPUT", e[e.IFRAME = 56] = "IFRAME", e[e.KEYGEN = 57] = "KEYGEN", e[e.LABEL = 58] = "LABEL", e[e.LI = 59] = "LI", e[e.LINK = 60] = "LINK", e[e.LISTING = 61] = "LISTING", e[e.MAIN = 62] = "MAIN", e[e.MALIGNMARK = 63] = "MALIGNMARK", e[e.MARQUEE = 64] = "MARQUEE", e[e.MATH = 65] = "MATH", e[e.MENU = 66] = "MENU", e[e.META = 67] = "META", e[e.MGLYPH = 68] = "MGLYPH", e[e.MI = 69] = "MI", e[e.MO = 70] = "MO", e[e.MN = 71] = "MN", e[e.MS = 72] = "MS", e[e.MTEXT = 73] = "MTEXT", e[e.NAV = 74] = "NAV", e[e.NOBR = 75] = "NOBR", e[e.NOFRAMES = 76] = "NOFRAMES", e[e.NOEMBED = 77] = "NOEMBED", e[e.NOSCRIPT = 78] = "NOSCRIPT", e[e.OBJECT = 79] = "OBJECT", e[e.OL = 80] = "OL", e[e.OPTGROUP = 81] = "OPTGROUP", e[e.OPTION = 82] = "OPTION", e[e.P = 83] = "P", e[e.PARAM = 84] = "PARAM", e[e.PLAINTEXT = 85] = "PLAINTEXT", e[e.PRE = 86] = "PRE", e[e.RB = 87] = "RB", e[e.RP = 88] = "RP", e[e.RT = 89] = "RT", e[e.RTC = 90] = "RTC", e[e.RUBY = 91] = "RUBY", e[e.S = 92] = "S", e[e.SCRIPT = 93] = "SCRIPT", e[e.SEARCH = 94] = "SEARCH", e[e.SECTION = 95] = "SECTION", e[e.SELECT = 96] = "SELECT", e[e.SOURCE = 97] = "SOURCE", e[e.SMALL = 98] = "SMALL", e[e.SPAN = 99] = "SPAN", e[e.STRIKE = 100] = "STRIKE", e[e.STRONG = 101] = "STRONG", e[e.STYLE = 102] = "STYLE", e[e.SUB = 103] = "SUB", e[e.SUMMARY = 104] = "SUMMARY", e[e.SUP = 105] = "SUP", e[e.TABLE = 106] = "TABLE", e[e.TBODY = 107] = "TBODY", e[e.TEMPLATE = 108] = "TEMPLATE", e[e.TEXTAREA = 109] = "TEXTAREA", e[e.TFOOT = 110] = "TFOOT", e[e.TD = 111] = "TD", e[e.TH = 112] = "TH", e[e.THEAD = 113] = "THEAD", e[e.TITLE = 114] = "TITLE", e[e.TR = 115] = "TR", e[e.TRACK = 116] = "TRACK", e[e.TT = 117] = "TT", e[e.U = 118] = "U", e[e.UL = 119] = "UL", e[e.SVG = 120] = "SVG", e[e.VAR = 121] = "VAR", e[e.WBR = 122] = "WBR", e[e.XMP = 123] = "XMP";
})(L ||= {});
var Nn = /* @__PURE__ */ new Map([
	[I.A, L.A],
	[I.ADDRESS, L.ADDRESS],
	[I.ANNOTATION_XML, L.ANNOTATION_XML],
	[I.APPLET, L.APPLET],
	[I.AREA, L.AREA],
	[I.ARTICLE, L.ARTICLE],
	[I.ASIDE, L.ASIDE],
	[I.B, L.B],
	[I.BASE, L.BASE],
	[I.BASEFONT, L.BASEFONT],
	[I.BGSOUND, L.BGSOUND],
	[I.BIG, L.BIG],
	[I.BLOCKQUOTE, L.BLOCKQUOTE],
	[I.BODY, L.BODY],
	[I.BR, L.BR],
	[I.BUTTON, L.BUTTON],
	[I.CAPTION, L.CAPTION],
	[I.CENTER, L.CENTER],
	[I.CODE, L.CODE],
	[I.COL, L.COL],
	[I.COLGROUP, L.COLGROUP],
	[I.DD, L.DD],
	[I.DESC, L.DESC],
	[I.DETAILS, L.DETAILS],
	[I.DIALOG, L.DIALOG],
	[I.DIR, L.DIR],
	[I.DIV, L.DIV],
	[I.DL, L.DL],
	[I.DT, L.DT],
	[I.EM, L.EM],
	[I.EMBED, L.EMBED],
	[I.FIELDSET, L.FIELDSET],
	[I.FIGCAPTION, L.FIGCAPTION],
	[I.FIGURE, L.FIGURE],
	[I.FONT, L.FONT],
	[I.FOOTER, L.FOOTER],
	[I.FOREIGN_OBJECT, L.FOREIGN_OBJECT],
	[I.FORM, L.FORM],
	[I.FRAME, L.FRAME],
	[I.FRAMESET, L.FRAMESET],
	[I.H1, L.H1],
	[I.H2, L.H2],
	[I.H3, L.H3],
	[I.H4, L.H4],
	[I.H5, L.H5],
	[I.H6, L.H6],
	[I.HEAD, L.HEAD],
	[I.HEADER, L.HEADER],
	[I.HGROUP, L.HGROUP],
	[I.HR, L.HR],
	[I.HTML, L.HTML],
	[I.I, L.I],
	[I.IMG, L.IMG],
	[I.IMAGE, L.IMAGE],
	[I.INPUT, L.INPUT],
	[I.IFRAME, L.IFRAME],
	[I.KEYGEN, L.KEYGEN],
	[I.LABEL, L.LABEL],
	[I.LI, L.LI],
	[I.LINK, L.LINK],
	[I.LISTING, L.LISTING],
	[I.MAIN, L.MAIN],
	[I.MALIGNMARK, L.MALIGNMARK],
	[I.MARQUEE, L.MARQUEE],
	[I.MATH, L.MATH],
	[I.MENU, L.MENU],
	[I.META, L.META],
	[I.MGLYPH, L.MGLYPH],
	[I.MI, L.MI],
	[I.MO, L.MO],
	[I.MN, L.MN],
	[I.MS, L.MS],
	[I.MTEXT, L.MTEXT],
	[I.NAV, L.NAV],
	[I.NOBR, L.NOBR],
	[I.NOFRAMES, L.NOFRAMES],
	[I.NOEMBED, L.NOEMBED],
	[I.NOSCRIPT, L.NOSCRIPT],
	[I.OBJECT, L.OBJECT],
	[I.OL, L.OL],
	[I.OPTGROUP, L.OPTGROUP],
	[I.OPTION, L.OPTION],
	[I.P, L.P],
	[I.PARAM, L.PARAM],
	[I.PLAINTEXT, L.PLAINTEXT],
	[I.PRE, L.PRE],
	[I.RB, L.RB],
	[I.RP, L.RP],
	[I.RT, L.RT],
	[I.RTC, L.RTC],
	[I.RUBY, L.RUBY],
	[I.S, L.S],
	[I.SCRIPT, L.SCRIPT],
	[I.SEARCH, L.SEARCH],
	[I.SECTION, L.SECTION],
	[I.SELECT, L.SELECT],
	[I.SOURCE, L.SOURCE],
	[I.SMALL, L.SMALL],
	[I.SPAN, L.SPAN],
	[I.STRIKE, L.STRIKE],
	[I.STRONG, L.STRONG],
	[I.STYLE, L.STYLE],
	[I.SUB, L.SUB],
	[I.SUMMARY, L.SUMMARY],
	[I.SUP, L.SUP],
	[I.TABLE, L.TABLE],
	[I.TBODY, L.TBODY],
	[I.TEMPLATE, L.TEMPLATE],
	[I.TEXTAREA, L.TEXTAREA],
	[I.TFOOT, L.TFOOT],
	[I.TD, L.TD],
	[I.TH, L.TH],
	[I.THEAD, L.THEAD],
	[I.TITLE, L.TITLE],
	[I.TR, L.TR],
	[I.TRACK, L.TRACK],
	[I.TT, L.TT],
	[I.U, L.U],
	[I.UL, L.UL],
	[I.SVG, L.SVG],
	[I.VAR, L.VAR],
	[I.WBR, L.WBR],
	[I.XMP, L.XMP]
]);
function Pn(e) {
	return Nn.get(e) ?? L.UNKNOWN;
}
var R = L, Fn = {
	[F.HTML]: /* @__PURE__ */ new Set([
		R.ADDRESS,
		R.APPLET,
		R.AREA,
		R.ARTICLE,
		R.ASIDE,
		R.BASE,
		R.BASEFONT,
		R.BGSOUND,
		R.BLOCKQUOTE,
		R.BODY,
		R.BR,
		R.BUTTON,
		R.CAPTION,
		R.CENTER,
		R.COL,
		R.COLGROUP,
		R.DD,
		R.DETAILS,
		R.DIR,
		R.DIV,
		R.DL,
		R.DT,
		R.EMBED,
		R.FIELDSET,
		R.FIGCAPTION,
		R.FIGURE,
		R.FOOTER,
		R.FORM,
		R.FRAME,
		R.FRAMESET,
		R.H1,
		R.H2,
		R.H3,
		R.H4,
		R.H5,
		R.H6,
		R.HEAD,
		R.HEADER,
		R.HGROUP,
		R.HR,
		R.HTML,
		R.IFRAME,
		R.IMG,
		R.INPUT,
		R.LI,
		R.LINK,
		R.LISTING,
		R.MAIN,
		R.MARQUEE,
		R.MENU,
		R.META,
		R.NAV,
		R.NOEMBED,
		R.NOFRAMES,
		R.NOSCRIPT,
		R.OBJECT,
		R.OL,
		R.P,
		R.PARAM,
		R.PLAINTEXT,
		R.PRE,
		R.SCRIPT,
		R.SECTION,
		R.SELECT,
		R.SOURCE,
		R.STYLE,
		R.SUMMARY,
		R.TABLE,
		R.TBODY,
		R.TD,
		R.TEMPLATE,
		R.TEXTAREA,
		R.TFOOT,
		R.TH,
		R.THEAD,
		R.TITLE,
		R.TR,
		R.TRACK,
		R.UL,
		R.WBR,
		R.XMP
	]),
	[F.MATHML]: /* @__PURE__ */ new Set([
		R.MI,
		R.MO,
		R.MN,
		R.MS,
		R.MTEXT,
		R.ANNOTATION_XML
	]),
	[F.SVG]: /* @__PURE__ */ new Set([
		R.TITLE,
		R.FOREIGN_OBJECT,
		R.DESC
	]),
	[F.XLINK]: /* @__PURE__ */ new Set(),
	[F.XML]: /* @__PURE__ */ new Set(),
	[F.XMLNS]: /* @__PURE__ */ new Set()
}, In = /* @__PURE__ */ new Set([
	R.H1,
	R.H2,
	R.H3,
	R.H4,
	R.H5,
	R.H6
]);
I.STYLE, I.SCRIPT, I.XMP, I.IFRAME, I.NOEMBED, I.NOFRAMES, I.PLAINTEXT;
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/parse5/8.1.0/38d3691d20fe3ab4780a57194f923f73d434662a4115031b2d4cf9cab8ac9fb9/node_modules/parse5/dist/tokenizer/index.js
var z;
(function(e) {
	e[e.DATA = 0] = "DATA", e[e.RCDATA = 1] = "RCDATA", e[e.RAWTEXT = 2] = "RAWTEXT", e[e.SCRIPT_DATA = 3] = "SCRIPT_DATA", e[e.PLAINTEXT = 4] = "PLAINTEXT", e[e.TAG_OPEN = 5] = "TAG_OPEN", e[e.END_TAG_OPEN = 6] = "END_TAG_OPEN", e[e.TAG_NAME = 7] = "TAG_NAME", e[e.RCDATA_LESS_THAN_SIGN = 8] = "RCDATA_LESS_THAN_SIGN", e[e.RCDATA_END_TAG_OPEN = 9] = "RCDATA_END_TAG_OPEN", e[e.RCDATA_END_TAG_NAME = 10] = "RCDATA_END_TAG_NAME", e[e.RAWTEXT_LESS_THAN_SIGN = 11] = "RAWTEXT_LESS_THAN_SIGN", e[e.RAWTEXT_END_TAG_OPEN = 12] = "RAWTEXT_END_TAG_OPEN", e[e.RAWTEXT_END_TAG_NAME = 13] = "RAWTEXT_END_TAG_NAME", e[e.SCRIPT_DATA_LESS_THAN_SIGN = 14] = "SCRIPT_DATA_LESS_THAN_SIGN", e[e.SCRIPT_DATA_END_TAG_OPEN = 15] = "SCRIPT_DATA_END_TAG_OPEN", e[e.SCRIPT_DATA_END_TAG_NAME = 16] = "SCRIPT_DATA_END_TAG_NAME", e[e.SCRIPT_DATA_ESCAPE_START = 17] = "SCRIPT_DATA_ESCAPE_START", e[e.SCRIPT_DATA_ESCAPE_START_DASH = 18] = "SCRIPT_DATA_ESCAPE_START_DASH", e[e.SCRIPT_DATA_ESCAPED = 19] = "SCRIPT_DATA_ESCAPED", e[e.SCRIPT_DATA_ESCAPED_DASH = 20] = "SCRIPT_DATA_ESCAPED_DASH", e[e.SCRIPT_DATA_ESCAPED_DASH_DASH = 21] = "SCRIPT_DATA_ESCAPED_DASH_DASH", e[e.SCRIPT_DATA_ESCAPED_LESS_THAN_SIGN = 22] = "SCRIPT_DATA_ESCAPED_LESS_THAN_SIGN", e[e.SCRIPT_DATA_ESCAPED_END_TAG_OPEN = 23] = "SCRIPT_DATA_ESCAPED_END_TAG_OPEN", e[e.SCRIPT_DATA_ESCAPED_END_TAG_NAME = 24] = "SCRIPT_DATA_ESCAPED_END_TAG_NAME", e[e.SCRIPT_DATA_DOUBLE_ESCAPE_START = 25] = "SCRIPT_DATA_DOUBLE_ESCAPE_START", e[e.SCRIPT_DATA_DOUBLE_ESCAPED = 26] = "SCRIPT_DATA_DOUBLE_ESCAPED", e[e.SCRIPT_DATA_DOUBLE_ESCAPED_DASH = 27] = "SCRIPT_DATA_DOUBLE_ESCAPED_DASH", e[e.SCRIPT_DATA_DOUBLE_ESCAPED_DASH_DASH = 28] = "SCRIPT_DATA_DOUBLE_ESCAPED_DASH_DASH", e[e.SCRIPT_DATA_DOUBLE_ESCAPED_LESS_THAN_SIGN = 29] = "SCRIPT_DATA_DOUBLE_ESCAPED_LESS_THAN_SIGN", e[e.SCRIPT_DATA_DOUBLE_ESCAPE_END = 30] = "SCRIPT_DATA_DOUBLE_ESCAPE_END", e[e.BEFORE_ATTRIBUTE_NAME = 31] = "BEFORE_ATTRIBUTE_NAME", e[e.ATTRIBUTE_NAME = 32] = "ATTRIBUTE_NAME", e[e.AFTER_ATTRIBUTE_NAME = 33] = "AFTER_ATTRIBUTE_NAME", e[e.BEFORE_ATTRIBUTE_VALUE = 34] = "BEFORE_ATTRIBUTE_VALUE", e[e.ATTRIBUTE_VALUE_DOUBLE_QUOTED = 35] = "ATTRIBUTE_VALUE_DOUBLE_QUOTED", e[e.ATTRIBUTE_VALUE_SINGLE_QUOTED = 36] = "ATTRIBUTE_VALUE_SINGLE_QUOTED", e[e.ATTRIBUTE_VALUE_UNQUOTED = 37] = "ATTRIBUTE_VALUE_UNQUOTED", e[e.AFTER_ATTRIBUTE_VALUE_QUOTED = 38] = "AFTER_ATTRIBUTE_VALUE_QUOTED", e[e.SELF_CLOSING_START_TAG = 39] = "SELF_CLOSING_START_TAG", e[e.BOGUS_COMMENT = 40] = "BOGUS_COMMENT", e[e.MARKUP_DECLARATION_OPEN = 41] = "MARKUP_DECLARATION_OPEN", e[e.COMMENT_START = 42] = "COMMENT_START", e[e.COMMENT_START_DASH = 43] = "COMMENT_START_DASH", e[e.COMMENT = 44] = "COMMENT", e[e.COMMENT_LESS_THAN_SIGN = 45] = "COMMENT_LESS_THAN_SIGN", e[e.COMMENT_LESS_THAN_SIGN_BANG = 46] = "COMMENT_LESS_THAN_SIGN_BANG", e[e.COMMENT_LESS_THAN_SIGN_BANG_DASH = 47] = "COMMENT_LESS_THAN_SIGN_BANG_DASH", e[e.COMMENT_LESS_THAN_SIGN_BANG_DASH_DASH = 48] = "COMMENT_LESS_THAN_SIGN_BANG_DASH_DASH", e[e.COMMENT_END_DASH = 49] = "COMMENT_END_DASH", e[e.COMMENT_END = 50] = "COMMENT_END", e[e.COMMENT_END_BANG = 51] = "COMMENT_END_BANG", e[e.DOCTYPE = 52] = "DOCTYPE", e[e.BEFORE_DOCTYPE_NAME = 53] = "BEFORE_DOCTYPE_NAME", e[e.DOCTYPE_NAME = 54] = "DOCTYPE_NAME", e[e.AFTER_DOCTYPE_NAME = 55] = "AFTER_DOCTYPE_NAME", e[e.AFTER_DOCTYPE_PUBLIC_KEYWORD = 56] = "AFTER_DOCTYPE_PUBLIC_KEYWORD", e[e.BEFORE_DOCTYPE_PUBLIC_IDENTIFIER = 57] = "BEFORE_DOCTYPE_PUBLIC_IDENTIFIER", e[e.DOCTYPE_PUBLIC_IDENTIFIER_DOUBLE_QUOTED = 58] = "DOCTYPE_PUBLIC_IDENTIFIER_DOUBLE_QUOTED", e[e.DOCTYPE_PUBLIC_IDENTIFIER_SINGLE_QUOTED = 59] = "DOCTYPE_PUBLIC_IDENTIFIER_SINGLE_QUOTED", e[e.AFTER_DOCTYPE_PUBLIC_IDENTIFIER = 60] = "AFTER_DOCTYPE_PUBLIC_IDENTIFIER", e[e.BETWEEN_DOCTYPE_PUBLIC_AND_SYSTEM_IDENTIFIERS = 61] = "BETWEEN_DOCTYPE_PUBLIC_AND_SYSTEM_IDENTIFIERS", e[e.AFTER_DOCTYPE_SYSTEM_KEYWORD = 62] = "AFTER_DOCTYPE_SYSTEM_KEYWORD", e[e.BEFORE_DOCTYPE_SYSTEM_IDENTIFIER = 63] = "BEFORE_DOCTYPE_SYSTEM_IDENTIFIER", e[e.DOCTYPE_SYSTEM_IDENTIFIER_DOUBLE_QUOTED = 64] = "DOCTYPE_SYSTEM_IDENTIFIER_DOUBLE_QUOTED", e[e.DOCTYPE_SYSTEM_IDENTIFIER_SINGLE_QUOTED = 65] = "DOCTYPE_SYSTEM_IDENTIFIER_SINGLE_QUOTED", e[e.AFTER_DOCTYPE_SYSTEM_IDENTIFIER = 66] = "AFTER_DOCTYPE_SYSTEM_IDENTIFIER", e[e.BOGUS_DOCTYPE = 67] = "BOGUS_DOCTYPE", e[e.CDATA_SECTION = 68] = "CDATA_SECTION", e[e.CDATA_SECTION_BRACKET = 69] = "CDATA_SECTION_BRACKET", e[e.CDATA_SECTION_END = 70] = "CDATA_SECTION_END", e[e.CHARACTER_REFERENCE = 71] = "CHARACTER_REFERENCE", e[e.AMBIGUOUS_AMPERSAND = 72] = "AMBIGUOUS_AMPERSAND";
})(z ||= {});
var B = {
	DATA: z.DATA,
	RCDATA: z.RCDATA,
	RAWTEXT: z.RAWTEXT,
	SCRIPT_DATA: z.SCRIPT_DATA,
	PLAINTEXT: z.PLAINTEXT,
	CDATA_SECTION: z.CDATA_SECTION
};
function Ln(e) {
	return e >= A.DIGIT_0 && e <= A.DIGIT_9;
}
function Rn(e) {
	return e >= A.LATIN_CAPITAL_A && e <= A.LATIN_CAPITAL_Z;
}
function zn(e) {
	return e >= A.LATIN_SMALL_A && e <= A.LATIN_SMALL_Z;
}
function Bn(e) {
	return zn(e) || Rn(e);
}
function Vn(e) {
	return Bn(e) || Ln(e);
}
function Hn(e) {
	return e + 32;
}
function Un(e) {
	return e === A.SPACE || e === A.LINE_FEED || e === A.TABULATION || e === A.FORM_FEED || e === A.CARRIAGE_RETURN;
}
function Wn(e) {
	return Un(e) || e === A.SOLIDUS || e === A.GREATER_THAN_SIGN;
}
function Gn(e) {
	return e === A.NULL ? M.nullCharacterReference : e > 1114111 ? M.characterReferenceOutsideUnicodeRange : kt(e) ? M.surrogateCharacterReference : Nt(e) ? M.noncharacterCharacterReference : Mt(e) || e === A.CARRIAGE_RETURN ? M.controlCharacterReference : null;
}
var Kn = class {
	constructor(e, t) {
		this.options = e, this.handler = t, this.paused = !1, this.inLoop = !1, this.inForeignNode = !1, this.lastStartTagName = "", this.active = !1, this.state = z.DATA, this.returnState = z.DATA, this.entityStartPos = 0, this.consumedAfterSnapshot = -1, this.currentCharacterToken = null, this.currentToken = null, this.currentAttr = {
			name: "",
			value: ""
		}, this.currentAttrLocation = null, this.runStart = 0, this.runEnd = 0, this.preprocessor = new Ft(t), this.currentLocation = this.getCurrentLocation(-1), this.entityDecoder = new Dn((e, t) => {
			this.preprocessor.pos = this.entityStartPos + t - 1, this._isCharacterReferenceInAttribute() || this._ensureCharacterToken(Un(e) ? N.WHITESPACE_CHARACTER : e === A.NULL ? N.NULL_CHARACTER : N.CHARACTER), this._extendRunWith(String.fromCodePoint(e), 0);
		}, t.onParseError ? {
			missingSemicolonAfterCharacterReference: () => {
				this._err(M.missingSemicolonAfterCharacterReference, 1);
			},
			absenceOfDigitsInNumericCharacterReference: (e) => {
				this._err(M.absenceOfDigitsInNumericCharacterReference, this.entityStartPos - this.preprocessor.pos + e);
			},
			validateNumericCharacterReference: (e) => {
				let t = Gn(e);
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
	flushCharacters() {
		if (this.inLoop) throw Error("Cannot flush characters from a tokenizer callback");
		let e = this.currentCharacterToken?.location, t = null;
		e && (t = this.getCharacterTokenEndLocation(e), this.currentLocation = t), this.inLoop = !0;
		try {
			this._emitCurrentCharacterToken(t);
		} finally {
			this.inLoop = !1;
		}
		this._runParsingLoop();
	}
	getCharacterTokenEndLocation(e) {
		let t = this.getCharacterTokenEndOffset(), n = this.preprocessor.html.slice(e.startOffset - this.preprocessor.droppedBufferSize, t - this.preprocessor.droppedBufferSize), r = e.startLine, i = -1;
		for (let e = n.indexOf("\n"); e !== -1; e = n.indexOf("\n", e + 1)) r++, i = e;
		for (let e = n.indexOf("\r"); e !== -1; e = n.indexOf("\r", e + 1)) n.charCodeAt(e + 1) !== A.LINE_FEED && (r++, i = Math.max(i, e));
		return {
			startLine: r,
			startCol: i === -1 ? e.startCol + n.length : n.length - i,
			startOffset: t,
			endLine: -1,
			endCol: -1,
			endOffset: -1
		};
	}
	getCharacterTokenEndOffset() {
		let e = this.preprocessor.droppedBufferSize + this.preprocessor.pos + 1, t = this.currentToken?.location;
		switch (this.state) {
			case z.TAG_OPEN:
			case z.RCDATA_LESS_THAN_SIGN:
			case z.RAWTEXT_LESS_THAN_SIGN:
			case z.SCRIPT_DATA_LESS_THAN_SIGN:
			case z.SCRIPT_DATA_ESCAPED_LESS_THAN_SIGN:
			case z.CDATA_SECTION_BRACKET: return e - 1;
			case z.END_TAG_OPEN:
			case z.MARKUP_DECLARATION_OPEN:
			case z.RCDATA_END_TAG_OPEN:
			case z.RCDATA_END_TAG_NAME:
			case z.RAWTEXT_END_TAG_OPEN:
			case z.RAWTEXT_END_TAG_NAME:
			case z.SCRIPT_DATA_END_TAG_OPEN:
			case z.SCRIPT_DATA_END_TAG_NAME:
			case z.SCRIPT_DATA_ESCAPED_END_TAG_OPEN:
			case z.SCRIPT_DATA_ESCAPED_END_TAG_NAME:
			case z.CDATA_SECTION_END: return e - 2;
			case z.CHARACTER_REFERENCE: return this._isCharacterReferenceInAttribute() && t ? t.startOffset : this.preprocessor.droppedBufferSize + this.entityStartPos;
			case z.DOCTYPE:
			case z.BEFORE_DOCTYPE_NAME: return this.currentLocation.startOffset;
			default: return t && (this.state === z.TAG_NAME || this.state >= z.BEFORE_ATTRIBUTE_NAME && this.state <= z.BOGUS_COMMENT || this.state >= z.COMMENT_START && this.state <= z.BOGUS_DOCTYPE) ? t.startOffset : e;
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
		this._flushRun(""), this.currentToken = {
			type: N.START_TAG,
			tagName: "",
			tagID: L.UNKNOWN,
			selfClosing: !1,
			ackSelfClosing: !1,
			attrs: [],
			location: this.getCurrentLocation(1)
		};
	}
	_createEndTagToken() {
		this._flushRun(""), this.currentToken = {
			type: N.END_TAG,
			tagName: "",
			tagID: L.UNKNOWN,
			selfClosing: !1,
			ackSelfClosing: !1,
			attrs: [],
			location: this.getCurrentLocation(2)
		};
	}
	_createCommentToken(e) {
		this._flushRun(""), this.currentToken = {
			type: N.COMMENT,
			data: "",
			location: this.getCurrentLocation(e)
		};
	}
	_createDoctypeToken(e) {
		this._flushRun(""), this.currentToken = {
			type: N.DOCTYPE,
			name: e,
			forceQuirks: !1,
			publicId: null,
			systemId: null,
			location: this.currentLocation
		};
	}
	_createCharacterToken(e) {
		if (this.currentLocation) {
			let e = this.currentLocation.startOffset - this.preprocessor.droppedBufferSize;
			this.preprocessor.html[e] === "\n" && this.preprocessor.html[e - 1] === "\r" && this.currentLocation.startOffset++;
		}
		this.currentCharacterToken = {
			type: e,
			chars: "",
			location: this.currentLocation
		};
	}
	_createAttr(e) {
		this._flushRun(""), this.currentAttr = {
			name: e,
			value: ""
		}, this.currentAttrLocation = this.getCurrentLocation(0);
	}
	_leaveAttrName() {
		var e;
		let t = this.currentToken;
		if (It(t, this.currentAttr.name) === null) {
			if (t.attrs.push(this.currentAttr), t.location && this.currentAttrLocation) {
				let n = (e = t.location).attrs ?? (e.attrs = Object.create(null));
				n[this.currentAttr.name] = this.currentAttrLocation, this._leaveAttrValue();
			}
		} else this._err(M.duplicateAttribute);
	}
	_leaveAttrValue() {
		this.currentAttrLocation && (this.currentAttrLocation.endLine = this.preprocessor.line, this.currentAttrLocation.endCol = this.preprocessor.col, this.currentAttrLocation.endOffset = this.preprocessor.offset);
	}
	_flushRun(e) {
		if (this.runEnd === this.runStart && e.length === 0) return;
		let t = this.preprocessor.html.slice(this.runStart, this.runEnd), n = (t.length < 13 ? t : ` ${t}`.slice(1)) + e, r = this.currentToken;
		this.runStart = this.runEnd, r === null ? this.currentCharacterToken.chars += n : r.type === N.COMMENT ? r.data += n : this.currentAttr.value += n;
	}
	_extendRun(e) {
		let { preprocessor: t } = this;
		if (e === A.LINE_FEED && t.html.charCodeAt(t.pos) !== A.LINE_FEED) this._flushRun("\n"), this.runStart = t.pos + 1;
		else {
			let n = t.pos - Number(e > 65535);
			n !== this.runEnd && (this._flushRun(""), this.runStart = n);
		}
		this.runEnd = t.pos + 1;
	}
	_extendRunWith(e, t) {
		let { preprocessor: n } = this, r = n.codePointStart + 1 + t - e.length;
		r >= 0 && n.html.startsWith(e, r) ? r !== this.runEnd && (this._flushRun(""), this.runStart = r) : (this._flushRun(e), this.runStart = r + e.length), this.runEnd = r + e.length;
	}
	prepareToken(e) {
		this._emitCurrentCharacterToken(e.location), this.currentToken = null, e.location && (e.location.endLine = this.preprocessor.line, e.location.endCol = this.preprocessor.col + 1, e.location.endOffset = this.preprocessor.offset + 1), this.currentLocation = this.getCurrentLocation(-1);
	}
	emitCurrentTagToken() {
		let e = this.currentToken;
		this.prepareToken(e), e.tagID = Pn(e.tagName), e.type === N.START_TAG ? (this.lastStartTagName = e.tagName, this.handler.onStartTag(e)) : (e.attrs.length > 0 && this._err(M.endTagWithAttributes), e.selfClosing && this._err(M.endTagWithTrailingSolidus), this.handler.onEndTag(e)), this.preprocessor.dropParsedChunk();
	}
	emitCurrentComment(e) {
		this.prepareToken(e), this.handler.onComment(e), this.preprocessor.dropParsedChunk();
	}
	emitCurrentDoctype(e) {
		this.prepareToken(e), this.handler.onDoctype(e), this.preprocessor.dropParsedChunk();
	}
	_emitCurrentCharacterToken(e) {
		if (this._flushRun(""), this.currentCharacterToken) {
			switch (e && this.currentCharacterToken.location && (this.currentCharacterToken.location.endLine = e.startLine, this.currentCharacterToken.location.endCol = e.startCol, this.currentCharacterToken.location.endOffset = e.startOffset), this.currentCharacterToken.type) {
				case N.CHARACTER:
					this.handler.onCharacter(this.currentCharacterToken);
					break;
				case N.NULL_CHARACTER:
					this.handler.onNullCharacter(this.currentCharacterToken);
					break;
				case N.WHITESPACE_CHARACTER: this.handler.onWhitespaceCharacter(this.currentCharacterToken);
			}
			this.currentCharacterToken = null;
		}
	}
	_emitEOFToken() {
		let e = this.getCurrentLocation(0);
		e && (e.endLine = e.startLine, e.endCol = e.startCol, e.endOffset = e.startOffset), this._emitCurrentCharacterToken(e), this.handler.onEof({
			type: N.EOF,
			location: e
		}), this.active = !1;
	}
	_ensureCharacterToken(e) {
		if (this.currentCharacterToken) {
			if (this.currentCharacterToken.type === e) return;
			this.currentLocation = this.getCurrentLocation(0), this._emitCurrentCharacterToken(this.currentLocation), this.preprocessor.dropParsedChunk();
		}
		this._createCharacterToken(e);
	}
	_emitCodePoint(e) {
		let t = Un(e) ? N.WHITESPACE_CHARACTER : e === A.NULL ? N.NULL_CHARACTER : N.CHARACTER, n = this.currentCharacterToken;
		(n === null || n.type !== t) && this._ensureCharacterToken(t), this._extendRun(e);
	}
	_emitChars(e, t = 0) {
		this.options.sourceCodeLocationInfo && t !== 0 && this.currentCharacterToken && this.currentCharacterToken.type !== N.CHARACTER ? (this.currentLocation = this.getCurrentLocation(e.length - 1 - t), this._emitCurrentCharacterToken(this.currentLocation), this._createCharacterToken(N.CHARACTER)) : this._ensureCharacterToken(N.CHARACTER), this._extendRunWith(e, t);
	}
	_startCharacterReference() {
		this.returnState = this.state, this.state = z.CHARACTER_REFERENCE, this.entityStartPos = this.preprocessor.pos, this.entityDecoder.startEntity(this._isCharacterReferenceInAttribute() ? P.Attribute : P.Legacy);
	}
	_isCharacterReferenceInAttribute() {
		return this.returnState === z.ATTRIBUTE_VALUE_DOUBLE_QUOTED || this.returnState === z.ATTRIBUTE_VALUE_SINGLE_QUOTED || this.returnState === z.ATTRIBUTE_VALUE_UNQUOTED;
	}
	_flushCodePointConsumedAsCharacterReference(e) {
		this._isCharacterReferenceInAttribute() ? this._extendRun(e) : this._emitCodePoint(e);
	}
	_callState(e) {
		switch (this.state) {
			case z.DATA:
				this._stateData(e);
				break;
			case z.RCDATA:
				this._stateRcdata(e);
				break;
			case z.RAWTEXT:
				this._stateRawtext(e);
				break;
			case z.SCRIPT_DATA:
				this._stateScriptData(e);
				break;
			case z.PLAINTEXT:
				this._statePlaintext(e);
				break;
			case z.TAG_OPEN:
				this._stateTagOpen(e);
				break;
			case z.END_TAG_OPEN:
				this._stateEndTagOpen(e);
				break;
			case z.TAG_NAME:
				this._stateTagName(e);
				break;
			case z.RCDATA_LESS_THAN_SIGN:
				this._stateRcdataLessThanSign(e);
				break;
			case z.RCDATA_END_TAG_OPEN:
				this._stateRcdataEndTagOpen(e);
				break;
			case z.RCDATA_END_TAG_NAME:
				this._stateRcdataEndTagName(e);
				break;
			case z.RAWTEXT_LESS_THAN_SIGN:
				this._stateRawtextLessThanSign(e);
				break;
			case z.RAWTEXT_END_TAG_OPEN:
				this._stateRawtextEndTagOpen(e);
				break;
			case z.RAWTEXT_END_TAG_NAME:
				this._stateRawtextEndTagName(e);
				break;
			case z.SCRIPT_DATA_LESS_THAN_SIGN:
				this._stateScriptDataLessThanSign(e);
				break;
			case z.SCRIPT_DATA_END_TAG_OPEN:
				this._stateScriptDataEndTagOpen(e);
				break;
			case z.SCRIPT_DATA_END_TAG_NAME:
				this._stateScriptDataEndTagName(e);
				break;
			case z.SCRIPT_DATA_ESCAPE_START:
				this._stateScriptDataEscapeStart(e);
				break;
			case z.SCRIPT_DATA_ESCAPE_START_DASH:
				this._stateScriptDataEscapeStartDash(e);
				break;
			case z.SCRIPT_DATA_ESCAPED:
				this._stateScriptDataEscaped(e);
				break;
			case z.SCRIPT_DATA_ESCAPED_DASH:
				this._stateScriptDataEscapedDash(e);
				break;
			case z.SCRIPT_DATA_ESCAPED_DASH_DASH:
				this._stateScriptDataEscapedDashDash(e);
				break;
			case z.SCRIPT_DATA_ESCAPED_LESS_THAN_SIGN:
				this._stateScriptDataEscapedLessThanSign(e);
				break;
			case z.SCRIPT_DATA_ESCAPED_END_TAG_OPEN:
				this._stateScriptDataEscapedEndTagOpen(e);
				break;
			case z.SCRIPT_DATA_ESCAPED_END_TAG_NAME:
				this._stateScriptDataEscapedEndTagName(e);
				break;
			case z.SCRIPT_DATA_DOUBLE_ESCAPE_START:
				this._stateScriptDataDoubleEscapeStart(e);
				break;
			case z.SCRIPT_DATA_DOUBLE_ESCAPED:
				this._stateScriptDataDoubleEscaped(e);
				break;
			case z.SCRIPT_DATA_DOUBLE_ESCAPED_DASH:
				this._stateScriptDataDoubleEscapedDash(e);
				break;
			case z.SCRIPT_DATA_DOUBLE_ESCAPED_DASH_DASH:
				this._stateScriptDataDoubleEscapedDashDash(e);
				break;
			case z.SCRIPT_DATA_DOUBLE_ESCAPED_LESS_THAN_SIGN:
				this._stateScriptDataDoubleEscapedLessThanSign(e);
				break;
			case z.SCRIPT_DATA_DOUBLE_ESCAPE_END:
				this._stateScriptDataDoubleEscapeEnd(e);
				break;
			case z.BEFORE_ATTRIBUTE_NAME:
				this._stateBeforeAttributeName(e);
				break;
			case z.ATTRIBUTE_NAME:
				this._stateAttributeName(e);
				break;
			case z.AFTER_ATTRIBUTE_NAME:
				this._stateAfterAttributeName(e);
				break;
			case z.BEFORE_ATTRIBUTE_VALUE:
				this._stateBeforeAttributeValue(e);
				break;
			case z.ATTRIBUTE_VALUE_DOUBLE_QUOTED:
				this._stateAttributeValueDoubleQuoted(e);
				break;
			case z.ATTRIBUTE_VALUE_SINGLE_QUOTED:
				this._stateAttributeValueSingleQuoted(e);
				break;
			case z.ATTRIBUTE_VALUE_UNQUOTED:
				this._stateAttributeValueUnquoted(e);
				break;
			case z.AFTER_ATTRIBUTE_VALUE_QUOTED:
				this._stateAfterAttributeValueQuoted(e);
				break;
			case z.SELF_CLOSING_START_TAG:
				this._stateSelfClosingStartTag(e);
				break;
			case z.BOGUS_COMMENT:
				this._stateBogusComment(e);
				break;
			case z.MARKUP_DECLARATION_OPEN:
				this._stateMarkupDeclarationOpen(e);
				break;
			case z.COMMENT_START:
				this._stateCommentStart(e);
				break;
			case z.COMMENT_START_DASH:
				this._stateCommentStartDash(e);
				break;
			case z.COMMENT:
				this._stateComment(e);
				break;
			case z.COMMENT_LESS_THAN_SIGN:
				this._stateCommentLessThanSign(e);
				break;
			case z.COMMENT_LESS_THAN_SIGN_BANG:
				this._stateCommentLessThanSignBang(e);
				break;
			case z.COMMENT_LESS_THAN_SIGN_BANG_DASH:
				this._stateCommentLessThanSignBangDash(e);
				break;
			case z.COMMENT_LESS_THAN_SIGN_BANG_DASH_DASH:
				this._stateCommentLessThanSignBangDashDash(e);
				break;
			case z.COMMENT_END_DASH:
				this._stateCommentEndDash(e);
				break;
			case z.COMMENT_END:
				this._stateCommentEnd(e);
				break;
			case z.COMMENT_END_BANG:
				this._stateCommentEndBang(e);
				break;
			case z.DOCTYPE:
				this._stateDoctype(e);
				break;
			case z.BEFORE_DOCTYPE_NAME:
				this._stateBeforeDoctypeName(e);
				break;
			case z.DOCTYPE_NAME:
				this._stateDoctypeName(e);
				break;
			case z.AFTER_DOCTYPE_NAME:
				this._stateAfterDoctypeName(e);
				break;
			case z.AFTER_DOCTYPE_PUBLIC_KEYWORD:
				this._stateAfterDoctypePublicKeyword(e);
				break;
			case z.BEFORE_DOCTYPE_PUBLIC_IDENTIFIER:
				this._stateBeforeDoctypePublicIdentifier(e);
				break;
			case z.DOCTYPE_PUBLIC_IDENTIFIER_DOUBLE_QUOTED:
				this._stateDoctypePublicIdentifierDoubleQuoted(e);
				break;
			case z.DOCTYPE_PUBLIC_IDENTIFIER_SINGLE_QUOTED:
				this._stateDoctypePublicIdentifierSingleQuoted(e);
				break;
			case z.AFTER_DOCTYPE_PUBLIC_IDENTIFIER:
				this._stateAfterDoctypePublicIdentifier(e);
				break;
			case z.BETWEEN_DOCTYPE_PUBLIC_AND_SYSTEM_IDENTIFIERS:
				this._stateBetweenDoctypePublicAndSystemIdentifiers(e);
				break;
			case z.AFTER_DOCTYPE_SYSTEM_KEYWORD:
				this._stateAfterDoctypeSystemKeyword(e);
				break;
			case z.BEFORE_DOCTYPE_SYSTEM_IDENTIFIER:
				this._stateBeforeDoctypeSystemIdentifier(e);
				break;
			case z.DOCTYPE_SYSTEM_IDENTIFIER_DOUBLE_QUOTED:
				this._stateDoctypeSystemIdentifierDoubleQuoted(e);
				break;
			case z.DOCTYPE_SYSTEM_IDENTIFIER_SINGLE_QUOTED:
				this._stateDoctypeSystemIdentifierSingleQuoted(e);
				break;
			case z.AFTER_DOCTYPE_SYSTEM_IDENTIFIER:
				this._stateAfterDoctypeSystemIdentifier(e);
				break;
			case z.BOGUS_DOCTYPE:
				this._stateBogusDoctype(e);
				break;
			case z.CDATA_SECTION:
				this._stateCdataSection(e);
				break;
			case z.CDATA_SECTION_BRACKET:
				this._stateCdataSectionBracket(e);
				break;
			case z.CDATA_SECTION_END:
				this._stateCdataSectionEnd(e);
				break;
			case z.CHARACTER_REFERENCE:
				this._stateCharacterReference();
				break;
			case z.AMBIGUOUS_AMPERSAND:
				this._stateAmbiguousAmpersand(e);
				break;
			default: throw Error("Unknown state");
		}
	}
	_stateData(e) {
		switch (e) {
			case A.LESS_THAN_SIGN:
				this.state = z.TAG_OPEN;
				break;
			case A.AMPERSAND:
				this._startCharacterReference();
				break;
			case A.NULL:
				this._err(M.unexpectedNullCharacter), this._emitCodePoint(e);
				break;
			case A.EOF:
				this._emitEOFToken();
				break;
			default: this._emitCodePoint(e);
		}
	}
	_stateRcdata(e) {
		switch (e) {
			case A.AMPERSAND:
				this._startCharacterReference();
				break;
			case A.LESS_THAN_SIGN:
				this.state = z.RCDATA_LESS_THAN_SIGN;
				break;
			case A.NULL:
				this._err(M.unexpectedNullCharacter), this._emitChars("�");
				break;
			case A.EOF:
				this._emitEOFToken();
				break;
			default: this._emitCodePoint(e);
		}
	}
	_stateRawtext(e) {
		switch (e) {
			case A.LESS_THAN_SIGN:
				this.state = z.RAWTEXT_LESS_THAN_SIGN;
				break;
			case A.NULL:
				this._err(M.unexpectedNullCharacter), this._emitChars("�");
				break;
			case A.EOF:
				this._emitEOFToken();
				break;
			default: this._emitCodePoint(e);
		}
	}
	_stateScriptData(e) {
		switch (e) {
			case A.LESS_THAN_SIGN:
				this.state = z.SCRIPT_DATA_LESS_THAN_SIGN;
				break;
			case A.NULL:
				this._err(M.unexpectedNullCharacter), this._emitChars("�");
				break;
			case A.EOF:
				this._emitEOFToken();
				break;
			default: this._emitCodePoint(e);
		}
	}
	_statePlaintext(e) {
		switch (e) {
			case A.NULL:
				this._err(M.unexpectedNullCharacter), this._emitChars("�");
				break;
			case A.EOF:
				this._emitEOFToken();
				break;
			default: this._emitCodePoint(e);
		}
	}
	_stateTagOpen(e) {
		if (Bn(e)) this._createStartTagToken(), this.state = z.TAG_NAME, this._stateTagName(e);
		else switch (e) {
			case A.EXCLAMATION_MARK:
				this.state = z.MARKUP_DECLARATION_OPEN;
				break;
			case A.SOLIDUS:
				this.state = z.END_TAG_OPEN;
				break;
			case A.QUESTION_MARK:
				this._err(M.unexpectedQuestionMarkInsteadOfTagName), this._createCommentToken(1), this.state = z.BOGUS_COMMENT, this._stateBogusComment(e);
				break;
			case A.EOF:
				this._err(M.eofBeforeTagName), this._emitChars("<", -1), this._emitEOFToken();
				break;
			default: this._err(M.invalidFirstCharacterOfTagName), this._emitChars("<", -1), this.state = z.DATA, this._stateData(e);
		}
	}
	_stateEndTagOpen(e) {
		if (Bn(e)) this._createEndTagToken(), this.state = z.TAG_NAME, this._stateTagName(e);
		else switch (e) {
			case A.GREATER_THAN_SIGN:
				this._err(M.missingEndTagName), this.state = z.DATA;
				break;
			case A.EOF:
				this._err(M.eofBeforeTagName), this._emitChars("</", -1), this._emitEOFToken();
				break;
			default: this._err(M.invalidFirstCharacterOfTagName), this._createCommentToken(2), this.state = z.BOGUS_COMMENT, this._stateBogusComment(e);
		}
	}
	_stateTagName(e) {
		let t = this.currentToken;
		switch (e) {
			case A.SPACE:
			case A.LINE_FEED:
			case A.TABULATION:
			case A.FORM_FEED:
				this.state = z.BEFORE_ATTRIBUTE_NAME;
				break;
			case A.SOLIDUS:
				this.state = z.SELF_CLOSING_START_TAG;
				break;
			case A.GREATER_THAN_SIGN:
				this.state = z.DATA, this.emitCurrentTagToken();
				break;
			case A.NULL:
				this._err(M.unexpectedNullCharacter), t.tagName += "�";
				break;
			case A.EOF:
				this._err(M.eofInTag), this._emitEOFToken();
				break;
			default: t.tagName += String.fromCodePoint(Rn(e) ? Hn(e) : e);
		}
	}
	_stateRcdataLessThanSign(e) {
		e === A.SOLIDUS ? this.state = z.RCDATA_END_TAG_OPEN : (this._emitChars("<", -1), this.state = z.RCDATA, this._stateRcdata(e));
	}
	_stateRcdataEndTagOpen(e) {
		Bn(e) ? (this.state = z.RCDATA_END_TAG_NAME, this._stateRcdataEndTagName(e)) : (this._emitChars("</", -1), this.state = z.RCDATA, this._stateRcdata(e));
	}
	handleSpecialEndTag(e) {
		if (!this.preprocessor.startsWith(this.lastStartTagName, !1)) return !this._ensureHibernation();
		this._createEndTagToken();
		let t = this.currentToken;
		switch (t.tagName = this.lastStartTagName, this.preprocessor.peek(this.lastStartTagName.length)) {
			case A.SPACE:
			case A.LINE_FEED:
			case A.TABULATION:
			case A.FORM_FEED: return this._advanceBy(this.lastStartTagName.length), this.state = z.BEFORE_ATTRIBUTE_NAME, !1;
			case A.SOLIDUS: return this._advanceBy(this.lastStartTagName.length), this.state = z.SELF_CLOSING_START_TAG, !1;
			case A.GREATER_THAN_SIGN: return this._advanceBy(this.lastStartTagName.length), this.emitCurrentTagToken(), this.state = z.DATA, !1;
			default: return this.currentToken = null, !this._ensureHibernation();
		}
	}
	_stateRcdataEndTagName(e) {
		this.handleSpecialEndTag(e) && (this._emitChars("</", -1), this.state = z.RCDATA, this._stateRcdata(e));
	}
	_stateRawtextLessThanSign(e) {
		e === A.SOLIDUS ? this.state = z.RAWTEXT_END_TAG_OPEN : (this._emitChars("<", -1), this.state = z.RAWTEXT, this._stateRawtext(e));
	}
	_stateRawtextEndTagOpen(e) {
		Bn(e) ? (this.state = z.RAWTEXT_END_TAG_NAME, this._stateRawtextEndTagName(e)) : (this._emitChars("</", -1), this.state = z.RAWTEXT, this._stateRawtext(e));
	}
	_stateRawtextEndTagName(e) {
		this.handleSpecialEndTag(e) && (this._emitChars("</", -1), this.state = z.RAWTEXT, this._stateRawtext(e));
	}
	_stateScriptDataLessThanSign(e) {
		switch (e) {
			case A.SOLIDUS:
				this.state = z.SCRIPT_DATA_END_TAG_OPEN;
				break;
			case A.EXCLAMATION_MARK:
				this.state = z.SCRIPT_DATA_ESCAPE_START, this._emitChars("<!");
				break;
			default: this._emitChars("<", -1), this.state = z.SCRIPT_DATA, this._stateScriptData(e);
		}
	}
	_stateScriptDataEndTagOpen(e) {
		Bn(e) ? (this.state = z.SCRIPT_DATA_END_TAG_NAME, this._stateScriptDataEndTagName(e)) : (this._emitChars("</", -1), this.state = z.SCRIPT_DATA, this._stateScriptData(e));
	}
	_stateScriptDataEndTagName(e) {
		this.handleSpecialEndTag(e) && (this._emitChars("</", -1), this.state = z.SCRIPT_DATA, this._stateScriptData(e));
	}
	_stateScriptDataEscapeStart(e) {
		e === A.HYPHEN_MINUS ? (this.state = z.SCRIPT_DATA_ESCAPE_START_DASH, this._emitChars("-")) : (this.state = z.SCRIPT_DATA, this._stateScriptData(e));
	}
	_stateScriptDataEscapeStartDash(e) {
		e === A.HYPHEN_MINUS ? (this.state = z.SCRIPT_DATA_ESCAPED_DASH_DASH, this._emitChars("-")) : (this.state = z.SCRIPT_DATA, this._stateScriptData(e));
	}
	_stateScriptDataEscaped(e) {
		switch (e) {
			case A.HYPHEN_MINUS:
				this.state = z.SCRIPT_DATA_ESCAPED_DASH, this._emitChars("-");
				break;
			case A.LESS_THAN_SIGN:
				this.state = z.SCRIPT_DATA_ESCAPED_LESS_THAN_SIGN;
				break;
			case A.NULL:
				this._err(M.unexpectedNullCharacter), this._emitChars("�");
				break;
			case A.EOF:
				this._err(M.eofInScriptHtmlCommentLikeText), this._emitEOFToken();
				break;
			default: this._emitCodePoint(e);
		}
	}
	_stateScriptDataEscapedDash(e) {
		switch (e) {
			case A.HYPHEN_MINUS:
				this.state = z.SCRIPT_DATA_ESCAPED_DASH_DASH, this._emitChars("-");
				break;
			case A.LESS_THAN_SIGN:
				this.state = z.SCRIPT_DATA_ESCAPED_LESS_THAN_SIGN;
				break;
			case A.NULL:
				this._err(M.unexpectedNullCharacter), this.state = z.SCRIPT_DATA_ESCAPED, this._emitChars("�");
				break;
			case A.EOF:
				this._err(M.eofInScriptHtmlCommentLikeText), this._emitEOFToken();
				break;
			default: this.state = z.SCRIPT_DATA_ESCAPED, this._emitCodePoint(e);
		}
	}
	_stateScriptDataEscapedDashDash(e) {
		switch (e) {
			case A.HYPHEN_MINUS:
				this._emitChars("-");
				break;
			case A.LESS_THAN_SIGN:
				this.state = z.SCRIPT_DATA_ESCAPED_LESS_THAN_SIGN;
				break;
			case A.GREATER_THAN_SIGN:
				this.state = z.SCRIPT_DATA, this._emitChars(">");
				break;
			case A.NULL:
				this._err(M.unexpectedNullCharacter), this.state = z.SCRIPT_DATA_ESCAPED, this._emitChars("�");
				break;
			case A.EOF:
				this._err(M.eofInScriptHtmlCommentLikeText), this._emitEOFToken();
				break;
			default: this.state = z.SCRIPT_DATA_ESCAPED, this._emitCodePoint(e);
		}
	}
	_stateScriptDataEscapedLessThanSign(e) {
		e === A.SOLIDUS ? this.state = z.SCRIPT_DATA_ESCAPED_END_TAG_OPEN : Bn(e) ? (this._emitChars("<", -1), this.state = z.SCRIPT_DATA_DOUBLE_ESCAPE_START, this._stateScriptDataDoubleEscapeStart(e)) : (this._emitChars("<", -1), this.state = z.SCRIPT_DATA_ESCAPED, this._stateScriptDataEscaped(e));
	}
	_stateScriptDataEscapedEndTagOpen(e) {
		Bn(e) ? (this.state = z.SCRIPT_DATA_ESCAPED_END_TAG_NAME, this._stateScriptDataEscapedEndTagName(e)) : (this._emitChars("</", -1), this.state = z.SCRIPT_DATA_ESCAPED, this._stateScriptDataEscaped(e));
	}
	_stateScriptDataEscapedEndTagName(e) {
		this.handleSpecialEndTag(e) && (this._emitChars("</", -1), this.state = z.SCRIPT_DATA_ESCAPED, this._stateScriptDataEscaped(e));
	}
	_stateScriptDataDoubleEscapeStart(e) {
		if (this.preprocessor.startsWith(j.SCRIPT, !1) && Wn(this.preprocessor.peek(j.SCRIPT.length))) {
			this._emitCodePoint(e);
			for (let e = 0; e < j.SCRIPT.length; e++) this._emitCodePoint(this._consume());
			this.state = z.SCRIPT_DATA_DOUBLE_ESCAPED;
		} else this._ensureHibernation() || (this.state = z.SCRIPT_DATA_ESCAPED, this._stateScriptDataEscaped(e));
	}
	_stateScriptDataDoubleEscaped(e) {
		switch (e) {
			case A.HYPHEN_MINUS:
				this.state = z.SCRIPT_DATA_DOUBLE_ESCAPED_DASH, this._emitChars("-");
				break;
			case A.LESS_THAN_SIGN:
				this.state = z.SCRIPT_DATA_DOUBLE_ESCAPED_LESS_THAN_SIGN, this._emitChars("<");
				break;
			case A.NULL:
				this._err(M.unexpectedNullCharacter), this._emitChars("�");
				break;
			case A.EOF:
				this._err(M.eofInScriptHtmlCommentLikeText), this._emitEOFToken();
				break;
			default: this._emitCodePoint(e);
		}
	}
	_stateScriptDataDoubleEscapedDash(e) {
		switch (e) {
			case A.HYPHEN_MINUS:
				this.state = z.SCRIPT_DATA_DOUBLE_ESCAPED_DASH_DASH, this._emitChars("-");
				break;
			case A.LESS_THAN_SIGN:
				this.state = z.SCRIPT_DATA_DOUBLE_ESCAPED_LESS_THAN_SIGN, this._emitChars("<");
				break;
			case A.NULL:
				this._err(M.unexpectedNullCharacter), this.state = z.SCRIPT_DATA_DOUBLE_ESCAPED, this._emitChars("�");
				break;
			case A.EOF:
				this._err(M.eofInScriptHtmlCommentLikeText), this._emitEOFToken();
				break;
			default: this.state = z.SCRIPT_DATA_DOUBLE_ESCAPED, this._emitCodePoint(e);
		}
	}
	_stateScriptDataDoubleEscapedDashDash(e) {
		switch (e) {
			case A.HYPHEN_MINUS:
				this._emitChars("-");
				break;
			case A.LESS_THAN_SIGN:
				this.state = z.SCRIPT_DATA_DOUBLE_ESCAPED_LESS_THAN_SIGN, this._emitChars("<");
				break;
			case A.GREATER_THAN_SIGN:
				this.state = z.SCRIPT_DATA, this._emitChars(">");
				break;
			case A.NULL:
				this._err(M.unexpectedNullCharacter), this.state = z.SCRIPT_DATA_DOUBLE_ESCAPED, this._emitChars("�");
				break;
			case A.EOF:
				this._err(M.eofInScriptHtmlCommentLikeText), this._emitEOFToken();
				break;
			default: this.state = z.SCRIPT_DATA_DOUBLE_ESCAPED, this._emitCodePoint(e);
		}
	}
	_stateScriptDataDoubleEscapedLessThanSign(e) {
		e === A.SOLIDUS ? (this.state = z.SCRIPT_DATA_DOUBLE_ESCAPE_END, this._emitChars("/")) : (this.state = z.SCRIPT_DATA_DOUBLE_ESCAPED, this._stateScriptDataDoubleEscaped(e));
	}
	_stateScriptDataDoubleEscapeEnd(e) {
		if (this.preprocessor.startsWith(j.SCRIPT, !1) && Wn(this.preprocessor.peek(j.SCRIPT.length))) {
			this._emitCodePoint(e);
			for (let e = 0; e < j.SCRIPT.length; e++) this._emitCodePoint(this._consume());
			this.state = z.SCRIPT_DATA_ESCAPED;
		} else this._ensureHibernation() || (this.state = z.SCRIPT_DATA_DOUBLE_ESCAPED, this._stateScriptDataDoubleEscaped(e));
	}
	_stateBeforeAttributeName(e) {
		switch (e) {
			case A.SPACE:
			case A.LINE_FEED:
			case A.TABULATION:
			case A.FORM_FEED: break;
			case A.SOLIDUS:
			case A.GREATER_THAN_SIGN:
			case A.EOF:
				this.state = z.AFTER_ATTRIBUTE_NAME, this._stateAfterAttributeName(e);
				break;
			case A.EQUALS_SIGN:
				this._err(M.unexpectedEqualsSignBeforeAttributeName), this._createAttr("="), this.state = z.ATTRIBUTE_NAME;
				break;
			default: this._createAttr(""), this.state = z.ATTRIBUTE_NAME, this._stateAttributeName(e);
		}
	}
	_stateAttributeName(e) {
		switch (e) {
			case A.SPACE:
			case A.LINE_FEED:
			case A.TABULATION:
			case A.FORM_FEED:
			case A.SOLIDUS:
			case A.GREATER_THAN_SIGN:
			case A.EOF:
				this._leaveAttrName(), this.state = z.AFTER_ATTRIBUTE_NAME, this._stateAfterAttributeName(e);
				break;
			case A.EQUALS_SIGN:
				this._leaveAttrName(), this.state = z.BEFORE_ATTRIBUTE_VALUE;
				break;
			case A.QUOTATION_MARK:
			case A.APOSTROPHE:
			case A.LESS_THAN_SIGN:
				this._err(M.unexpectedCharacterInAttributeName), this.currentAttr.name += String.fromCodePoint(e);
				break;
			case A.NULL:
				this._err(M.unexpectedNullCharacter), this.currentAttr.name += "�";
				break;
			default: this.currentAttr.name += String.fromCodePoint(Rn(e) ? Hn(e) : e);
		}
	}
	_stateAfterAttributeName(e) {
		switch (e) {
			case A.SPACE:
			case A.LINE_FEED:
			case A.TABULATION:
			case A.FORM_FEED: break;
			case A.SOLIDUS:
				this.state = z.SELF_CLOSING_START_TAG;
				break;
			case A.EQUALS_SIGN:
				this.state = z.BEFORE_ATTRIBUTE_VALUE;
				break;
			case A.GREATER_THAN_SIGN:
				this.state = z.DATA, this.emitCurrentTagToken();
				break;
			case A.EOF:
				this._err(M.eofInTag), this._emitEOFToken();
				break;
			default: this._createAttr(""), this.state = z.ATTRIBUTE_NAME, this._stateAttributeName(e);
		}
	}
	_stateBeforeAttributeValue(e) {
		switch (e) {
			case A.SPACE:
			case A.LINE_FEED:
			case A.TABULATION:
			case A.FORM_FEED: break;
			case A.QUOTATION_MARK:
				this.state = z.ATTRIBUTE_VALUE_DOUBLE_QUOTED;
				break;
			case A.APOSTROPHE:
				this.state = z.ATTRIBUTE_VALUE_SINGLE_QUOTED;
				break;
			case A.GREATER_THAN_SIGN:
				this._err(M.missingAttributeValue), this.state = z.DATA, this.emitCurrentTagToken();
				break;
			default: this.state = z.ATTRIBUTE_VALUE_UNQUOTED, this._stateAttributeValueUnquoted(e);
		}
	}
	_stateAttributeValueDoubleQuoted(e) {
		switch (e) {
			case A.QUOTATION_MARK:
				this.state = z.AFTER_ATTRIBUTE_VALUE_QUOTED;
				break;
			case A.AMPERSAND:
				this._startCharacterReference();
				break;
			case A.NULL:
				this._err(M.unexpectedNullCharacter), this._extendRunWith("�", 0);
				break;
			case A.EOF:
				this._err(M.eofInTag), this._emitEOFToken();
				break;
			default: this._extendRun(e);
		}
	}
	_stateAttributeValueSingleQuoted(e) {
		switch (e) {
			case A.APOSTROPHE:
				this.state = z.AFTER_ATTRIBUTE_VALUE_QUOTED;
				break;
			case A.AMPERSAND:
				this._startCharacterReference();
				break;
			case A.NULL:
				this._err(M.unexpectedNullCharacter), this._extendRunWith("�", 0);
				break;
			case A.EOF:
				this._err(M.eofInTag), this._emitEOFToken();
				break;
			default: this._extendRun(e);
		}
	}
	_stateAttributeValueUnquoted(e) {
		switch (e) {
			case A.SPACE:
			case A.LINE_FEED:
			case A.TABULATION:
			case A.FORM_FEED:
				this._leaveAttrValue(), this.state = z.BEFORE_ATTRIBUTE_NAME;
				break;
			case A.AMPERSAND:
				this._startCharacterReference();
				break;
			case A.GREATER_THAN_SIGN:
				this._leaveAttrValue(), this.state = z.DATA, this.emitCurrentTagToken();
				break;
			case A.NULL:
				this._err(M.unexpectedNullCharacter), this._extendRunWith("�", 0);
				break;
			case A.QUOTATION_MARK:
			case A.APOSTROPHE:
			case A.LESS_THAN_SIGN:
			case A.EQUALS_SIGN:
			case A.GRAVE_ACCENT:
				this._err(M.unexpectedCharacterInUnquotedAttributeValue), this._extendRun(e);
				break;
			case A.EOF:
				this._err(M.eofInTag), this._emitEOFToken();
				break;
			default: this._extendRun(e);
		}
	}
	_stateAfterAttributeValueQuoted(e) {
		switch (e) {
			case A.SPACE:
			case A.LINE_FEED:
			case A.TABULATION:
			case A.FORM_FEED:
				this._leaveAttrValue(), this.state = z.BEFORE_ATTRIBUTE_NAME;
				break;
			case A.SOLIDUS:
				this._leaveAttrValue(), this.state = z.SELF_CLOSING_START_TAG;
				break;
			case A.GREATER_THAN_SIGN:
				this._leaveAttrValue(), this.state = z.DATA, this.emitCurrentTagToken();
				break;
			case A.EOF:
				this._err(M.eofInTag), this._emitEOFToken();
				break;
			default: this._err(M.missingWhitespaceBetweenAttributes), this._leaveAttrValue(), this.state = z.BEFORE_ATTRIBUTE_NAME, this._stateBeforeAttributeName(e);
		}
	}
	_stateSelfClosingStartTag(e) {
		switch (e) {
			case A.GREATER_THAN_SIGN: {
				let e = this.currentToken;
				e.selfClosing = !0, this.state = z.DATA, this.emitCurrentTagToken();
				break;
			}
			case A.EOF:
				this._err(M.eofInTag), this._emitEOFToken();
				break;
			default: this._err(M.unexpectedSolidusInTag), this.state = z.BEFORE_ATTRIBUTE_NAME, this._stateBeforeAttributeName(e);
		}
	}
	_stateBogusComment(e) {
		let t = this.currentToken;
		switch (e) {
			case A.GREATER_THAN_SIGN:
				this.state = z.DATA, this.emitCurrentComment(t);
				break;
			case A.EOF:
				this.emitCurrentComment(t), this._emitEOFToken();
				break;
			case A.NULL:
				this._err(M.unexpectedNullCharacter), this._extendRunWith("�", 0);
				break;
			default: this._extendRun(e);
		}
	}
	_stateMarkupDeclarationOpen(e) {
		this._consumeSequenceIfMatch(j.DASH_DASH, !0) ? (this._createCommentToken(j.DASH_DASH.length + 1), this.state = z.COMMENT_START) : this._consumeSequenceIfMatch(j.DOCTYPE, !1) ? (this.currentLocation = this.getCurrentLocation(j.DOCTYPE.length + 1), this.state = z.DOCTYPE) : this._consumeSequenceIfMatch(j.CDATA_START, !0) ? this.inForeignNode ? this.state = z.CDATA_SECTION : (this._err(M.cdataInHtmlContent), this._createCommentToken(j.CDATA_START.length + 1), this.currentToken.data = "[CDATA[", this.state = z.BOGUS_COMMENT) : this._ensureHibernation() || (this._err(M.incorrectlyOpenedComment), this._createCommentToken(2), this.state = z.BOGUS_COMMENT, this._stateBogusComment(e));
	}
	_stateCommentStart(e) {
		switch (e) {
			case A.HYPHEN_MINUS:
				this.state = z.COMMENT_START_DASH;
				break;
			case A.GREATER_THAN_SIGN: {
				this._err(M.abruptClosingOfEmptyComment), this.state = z.DATA;
				let e = this.currentToken;
				this.emitCurrentComment(e);
				break;
			}
			default: this.state = z.COMMENT, this._stateComment(e);
		}
	}
	_stateCommentStartDash(e) {
		let t = this.currentToken;
		switch (e) {
			case A.HYPHEN_MINUS:
				this.state = z.COMMENT_END;
				break;
			case A.GREATER_THAN_SIGN:
				this._err(M.abruptClosingOfEmptyComment), this.state = z.DATA, this.emitCurrentComment(t);
				break;
			case A.EOF:
				this._err(M.eofInComment), this.emitCurrentComment(t), this._emitEOFToken();
				break;
			default: this._extendRunWith("-", -1), this.state = z.COMMENT, this._stateComment(e);
		}
	}
	_stateComment(e) {
		let t = this.currentToken;
		switch (e) {
			case A.HYPHEN_MINUS:
				this.state = z.COMMENT_END_DASH;
				break;
			case A.LESS_THAN_SIGN:
				this._extendRunWith("<", 0), this.state = z.COMMENT_LESS_THAN_SIGN;
				break;
			case A.NULL:
				this._err(M.unexpectedNullCharacter), this._extendRunWith("�", 0);
				break;
			case A.EOF:
				this._err(M.eofInComment), this.emitCurrentComment(t), this._emitEOFToken();
				break;
			default: this._extendRun(e);
		}
	}
	_stateCommentLessThanSign(e) {
		switch (e) {
			case A.EXCLAMATION_MARK:
				this._extendRunWith("!", 0), this.state = z.COMMENT_LESS_THAN_SIGN_BANG;
				break;
			case A.LESS_THAN_SIGN:
				this._extendRunWith("<", 0);
				break;
			default: this.state = z.COMMENT, this._stateComment(e);
		}
	}
	_stateCommentLessThanSignBang(e) {
		e === A.HYPHEN_MINUS ? this.state = z.COMMENT_LESS_THAN_SIGN_BANG_DASH : (this.state = z.COMMENT, this._stateComment(e));
	}
	_stateCommentLessThanSignBangDash(e) {
		e === A.HYPHEN_MINUS ? this.state = z.COMMENT_LESS_THAN_SIGN_BANG_DASH_DASH : (this.state = z.COMMENT_END_DASH, this._stateCommentEndDash(e));
	}
	_stateCommentLessThanSignBangDashDash(e) {
		e !== A.GREATER_THAN_SIGN && e !== A.EOF && this._err(M.nestedComment), this.state = z.COMMENT_END, this._stateCommentEnd(e);
	}
	_stateCommentEndDash(e) {
		let t = this.currentToken;
		switch (e) {
			case A.HYPHEN_MINUS:
				this.state = z.COMMENT_END;
				break;
			case A.EOF:
				this._err(M.eofInComment), this.emitCurrentComment(t), this._emitEOFToken();
				break;
			default: this._extendRunWith("-", -1), this.state = z.COMMENT, this._stateComment(e);
		}
	}
	_stateCommentEnd(e) {
		let t = this.currentToken;
		switch (e) {
			case A.GREATER_THAN_SIGN:
				this.state = z.DATA, this.emitCurrentComment(t);
				break;
			case A.EXCLAMATION_MARK:
				this.state = z.COMMENT_END_BANG;
				break;
			case A.HYPHEN_MINUS:
				this._extendRunWith("-", -2);
				break;
			case A.EOF:
				this._err(M.eofInComment), this.emitCurrentComment(t), this._emitEOFToken();
				break;
			default: this._extendRunWith("--", -1), this.state = z.COMMENT, this._stateComment(e);
		}
	}
	_stateCommentEndBang(e) {
		let t = this.currentToken;
		switch (e) {
			case A.HYPHEN_MINUS:
				this._extendRunWith("--!", -1), this.state = z.COMMENT_END_DASH;
				break;
			case A.GREATER_THAN_SIGN:
				this._err(M.incorrectlyClosedComment), this.state = z.DATA, this.emitCurrentComment(t);
				break;
			case A.EOF:
				this._err(M.eofInComment), this.emitCurrentComment(t), this._emitEOFToken();
				break;
			default: this._extendRunWith("--!", -1), this.state = z.COMMENT, this._stateComment(e);
		}
	}
	_stateDoctype(e) {
		switch (e) {
			case A.SPACE:
			case A.LINE_FEED:
			case A.TABULATION:
			case A.FORM_FEED:
				this.state = z.BEFORE_DOCTYPE_NAME;
				break;
			case A.GREATER_THAN_SIGN:
				this.state = z.BEFORE_DOCTYPE_NAME, this._stateBeforeDoctypeName(e);
				break;
			case A.EOF: {
				this._err(M.eofInDoctype), this._createDoctypeToken(null);
				let e = this.currentToken;
				e.forceQuirks = !0, this.emitCurrentDoctype(e), this._emitEOFToken();
				break;
			}
			default: this._err(M.missingWhitespaceBeforeDoctypeName), this.state = z.BEFORE_DOCTYPE_NAME, this._stateBeforeDoctypeName(e);
		}
	}
	_stateBeforeDoctypeName(e) {
		if (Rn(e)) this._createDoctypeToken(String.fromCharCode(Hn(e))), this.state = z.DOCTYPE_NAME;
		else switch (e) {
			case A.SPACE:
			case A.LINE_FEED:
			case A.TABULATION:
			case A.FORM_FEED: break;
			case A.NULL:
				this._err(M.unexpectedNullCharacter), this._createDoctypeToken("�"), this.state = z.DOCTYPE_NAME;
				break;
			case A.GREATER_THAN_SIGN: {
				this._err(M.missingDoctypeName), this._createDoctypeToken(null);
				let e = this.currentToken;
				e.forceQuirks = !0, this.emitCurrentDoctype(e), this.state = z.DATA;
				break;
			}
			case A.EOF: {
				this._err(M.eofInDoctype), this._createDoctypeToken(null);
				let e = this.currentToken;
				e.forceQuirks = !0, this.emitCurrentDoctype(e), this._emitEOFToken();
				break;
			}
			default: this._createDoctypeToken(String.fromCodePoint(e)), this.state = z.DOCTYPE_NAME;
		}
	}
	_stateDoctypeName(e) {
		let t = this.currentToken;
		switch (e) {
			case A.SPACE:
			case A.LINE_FEED:
			case A.TABULATION:
			case A.FORM_FEED:
				this.state = z.AFTER_DOCTYPE_NAME;
				break;
			case A.GREATER_THAN_SIGN:
				this.state = z.DATA, this.emitCurrentDoctype(t);
				break;
			case A.NULL:
				this._err(M.unexpectedNullCharacter), t.name += "�";
				break;
			case A.EOF:
				this._err(M.eofInDoctype), t.forceQuirks = !0, this.emitCurrentDoctype(t), this._emitEOFToken();
				break;
			default: t.name += String.fromCodePoint(Rn(e) ? Hn(e) : e);
		}
	}
	_stateAfterDoctypeName(e) {
		let t = this.currentToken;
		switch (e) {
			case A.SPACE:
			case A.LINE_FEED:
			case A.TABULATION:
			case A.FORM_FEED: break;
			case A.GREATER_THAN_SIGN:
				this.state = z.DATA, this.emitCurrentDoctype(t);
				break;
			case A.EOF:
				this._err(M.eofInDoctype), t.forceQuirks = !0, this.emitCurrentDoctype(t), this._emitEOFToken();
				break;
			default: this._consumeSequenceIfMatch(j.PUBLIC, !1) ? this.state = z.AFTER_DOCTYPE_PUBLIC_KEYWORD : this._consumeSequenceIfMatch(j.SYSTEM, !1) ? this.state = z.AFTER_DOCTYPE_SYSTEM_KEYWORD : this._ensureHibernation() || (this._err(M.invalidCharacterSequenceAfterDoctypeName), t.forceQuirks = !0, this.state = z.BOGUS_DOCTYPE, this._stateBogusDoctype(e));
		}
	}
	_stateAfterDoctypePublicKeyword(e) {
		let t = this.currentToken;
		switch (e) {
			case A.SPACE:
			case A.LINE_FEED:
			case A.TABULATION:
			case A.FORM_FEED:
				this.state = z.BEFORE_DOCTYPE_PUBLIC_IDENTIFIER;
				break;
			case A.QUOTATION_MARK:
				this._err(M.missingWhitespaceAfterDoctypePublicKeyword), t.publicId = "", this.state = z.DOCTYPE_PUBLIC_IDENTIFIER_DOUBLE_QUOTED;
				break;
			case A.APOSTROPHE:
				this._err(M.missingWhitespaceAfterDoctypePublicKeyword), t.publicId = "", this.state = z.DOCTYPE_PUBLIC_IDENTIFIER_SINGLE_QUOTED;
				break;
			case A.GREATER_THAN_SIGN:
				this._err(M.missingDoctypePublicIdentifier), t.forceQuirks = !0, this.state = z.DATA, this.emitCurrentDoctype(t);
				break;
			case A.EOF:
				this._err(M.eofInDoctype), t.forceQuirks = !0, this.emitCurrentDoctype(t), this._emitEOFToken();
				break;
			default: this._err(M.missingQuoteBeforeDoctypePublicIdentifier), t.forceQuirks = !0, this.state = z.BOGUS_DOCTYPE, this._stateBogusDoctype(e);
		}
	}
	_stateBeforeDoctypePublicIdentifier(e) {
		let t = this.currentToken;
		switch (e) {
			case A.SPACE:
			case A.LINE_FEED:
			case A.TABULATION:
			case A.FORM_FEED: break;
			case A.QUOTATION_MARK:
				t.publicId = "", this.state = z.DOCTYPE_PUBLIC_IDENTIFIER_DOUBLE_QUOTED;
				break;
			case A.APOSTROPHE:
				t.publicId = "", this.state = z.DOCTYPE_PUBLIC_IDENTIFIER_SINGLE_QUOTED;
				break;
			case A.GREATER_THAN_SIGN:
				this._err(M.missingDoctypePublicIdentifier), t.forceQuirks = !0, this.state = z.DATA, this.emitCurrentDoctype(t);
				break;
			case A.EOF:
				this._err(M.eofInDoctype), t.forceQuirks = !0, this.emitCurrentDoctype(t), this._emitEOFToken();
				break;
			default: this._err(M.missingQuoteBeforeDoctypePublicIdentifier), t.forceQuirks = !0, this.state = z.BOGUS_DOCTYPE, this._stateBogusDoctype(e);
		}
	}
	_stateDoctypePublicIdentifierDoubleQuoted(e) {
		let t = this.currentToken;
		switch (e) {
			case A.QUOTATION_MARK:
				this.state = z.AFTER_DOCTYPE_PUBLIC_IDENTIFIER;
				break;
			case A.NULL:
				this._err(M.unexpectedNullCharacter), t.publicId += "�";
				break;
			case A.GREATER_THAN_SIGN:
				this._err(M.abruptDoctypePublicIdentifier), t.forceQuirks = !0, this.emitCurrentDoctype(t), this.state = z.DATA;
				break;
			case A.EOF:
				this._err(M.eofInDoctype), t.forceQuirks = !0, this.emitCurrentDoctype(t), this._emitEOFToken();
				break;
			default: t.publicId += String.fromCodePoint(e);
		}
	}
	_stateDoctypePublicIdentifierSingleQuoted(e) {
		let t = this.currentToken;
		switch (e) {
			case A.APOSTROPHE:
				this.state = z.AFTER_DOCTYPE_PUBLIC_IDENTIFIER;
				break;
			case A.NULL:
				this._err(M.unexpectedNullCharacter), t.publicId += "�";
				break;
			case A.GREATER_THAN_SIGN:
				this._err(M.abruptDoctypePublicIdentifier), t.forceQuirks = !0, this.emitCurrentDoctype(t), this.state = z.DATA;
				break;
			case A.EOF:
				this._err(M.eofInDoctype), t.forceQuirks = !0, this.emitCurrentDoctype(t), this._emitEOFToken();
				break;
			default: t.publicId += String.fromCodePoint(e);
		}
	}
	_stateAfterDoctypePublicIdentifier(e) {
		let t = this.currentToken;
		switch (e) {
			case A.SPACE:
			case A.LINE_FEED:
			case A.TABULATION:
			case A.FORM_FEED:
				this.state = z.BETWEEN_DOCTYPE_PUBLIC_AND_SYSTEM_IDENTIFIERS;
				break;
			case A.GREATER_THAN_SIGN:
				this.state = z.DATA, this.emitCurrentDoctype(t);
				break;
			case A.QUOTATION_MARK:
				this._err(M.missingWhitespaceBetweenDoctypePublicAndSystemIdentifiers), t.systemId = "", this.state = z.DOCTYPE_SYSTEM_IDENTIFIER_DOUBLE_QUOTED;
				break;
			case A.APOSTROPHE:
				this._err(M.missingWhitespaceBetweenDoctypePublicAndSystemIdentifiers), t.systemId = "", this.state = z.DOCTYPE_SYSTEM_IDENTIFIER_SINGLE_QUOTED;
				break;
			case A.EOF:
				this._err(M.eofInDoctype), t.forceQuirks = !0, this.emitCurrentDoctype(t), this._emitEOFToken();
				break;
			default: this._err(M.missingQuoteBeforeDoctypeSystemIdentifier), t.forceQuirks = !0, this.state = z.BOGUS_DOCTYPE, this._stateBogusDoctype(e);
		}
	}
	_stateBetweenDoctypePublicAndSystemIdentifiers(e) {
		let t = this.currentToken;
		switch (e) {
			case A.SPACE:
			case A.LINE_FEED:
			case A.TABULATION:
			case A.FORM_FEED: break;
			case A.GREATER_THAN_SIGN:
				this.emitCurrentDoctype(t), this.state = z.DATA;
				break;
			case A.QUOTATION_MARK:
				t.systemId = "", this.state = z.DOCTYPE_SYSTEM_IDENTIFIER_DOUBLE_QUOTED;
				break;
			case A.APOSTROPHE:
				t.systemId = "", this.state = z.DOCTYPE_SYSTEM_IDENTIFIER_SINGLE_QUOTED;
				break;
			case A.EOF:
				this._err(M.eofInDoctype), t.forceQuirks = !0, this.emitCurrentDoctype(t), this._emitEOFToken();
				break;
			default: this._err(M.missingQuoteBeforeDoctypeSystemIdentifier), t.forceQuirks = !0, this.state = z.BOGUS_DOCTYPE, this._stateBogusDoctype(e);
		}
	}
	_stateAfterDoctypeSystemKeyword(e) {
		let t = this.currentToken;
		switch (e) {
			case A.SPACE:
			case A.LINE_FEED:
			case A.TABULATION:
			case A.FORM_FEED:
				this.state = z.BEFORE_DOCTYPE_SYSTEM_IDENTIFIER;
				break;
			case A.QUOTATION_MARK:
				this._err(M.missingWhitespaceAfterDoctypeSystemKeyword), t.systemId = "", this.state = z.DOCTYPE_SYSTEM_IDENTIFIER_DOUBLE_QUOTED;
				break;
			case A.APOSTROPHE:
				this._err(M.missingWhitespaceAfterDoctypeSystemKeyword), t.systemId = "", this.state = z.DOCTYPE_SYSTEM_IDENTIFIER_SINGLE_QUOTED;
				break;
			case A.GREATER_THAN_SIGN:
				this._err(M.missingDoctypeSystemIdentifier), t.forceQuirks = !0, this.state = z.DATA, this.emitCurrentDoctype(t);
				break;
			case A.EOF:
				this._err(M.eofInDoctype), t.forceQuirks = !0, this.emitCurrentDoctype(t), this._emitEOFToken();
				break;
			default: this._err(M.missingQuoteBeforeDoctypeSystemIdentifier), t.forceQuirks = !0, this.state = z.BOGUS_DOCTYPE, this._stateBogusDoctype(e);
		}
	}
	_stateBeforeDoctypeSystemIdentifier(e) {
		let t = this.currentToken;
		switch (e) {
			case A.SPACE:
			case A.LINE_FEED:
			case A.TABULATION:
			case A.FORM_FEED: break;
			case A.QUOTATION_MARK:
				t.systemId = "", this.state = z.DOCTYPE_SYSTEM_IDENTIFIER_DOUBLE_QUOTED;
				break;
			case A.APOSTROPHE:
				t.systemId = "", this.state = z.DOCTYPE_SYSTEM_IDENTIFIER_SINGLE_QUOTED;
				break;
			case A.GREATER_THAN_SIGN:
				this._err(M.missingDoctypeSystemIdentifier), t.forceQuirks = !0, this.state = z.DATA, this.emitCurrentDoctype(t);
				break;
			case A.EOF:
				this._err(M.eofInDoctype), t.forceQuirks = !0, this.emitCurrentDoctype(t), this._emitEOFToken();
				break;
			default: this._err(M.missingQuoteBeforeDoctypeSystemIdentifier), t.forceQuirks = !0, this.state = z.BOGUS_DOCTYPE, this._stateBogusDoctype(e);
		}
	}
	_stateDoctypeSystemIdentifierDoubleQuoted(e) {
		let t = this.currentToken;
		switch (e) {
			case A.QUOTATION_MARK:
				this.state = z.AFTER_DOCTYPE_SYSTEM_IDENTIFIER;
				break;
			case A.NULL:
				this._err(M.unexpectedNullCharacter), t.systemId += "�";
				break;
			case A.GREATER_THAN_SIGN:
				this._err(M.abruptDoctypeSystemIdentifier), t.forceQuirks = !0, this.emitCurrentDoctype(t), this.state = z.DATA;
				break;
			case A.EOF:
				this._err(M.eofInDoctype), t.forceQuirks = !0, this.emitCurrentDoctype(t), this._emitEOFToken();
				break;
			default: t.systemId += String.fromCodePoint(e);
		}
	}
	_stateDoctypeSystemIdentifierSingleQuoted(e) {
		let t = this.currentToken;
		switch (e) {
			case A.APOSTROPHE:
				this.state = z.AFTER_DOCTYPE_SYSTEM_IDENTIFIER;
				break;
			case A.NULL:
				this._err(M.unexpectedNullCharacter), t.systemId += "�";
				break;
			case A.GREATER_THAN_SIGN:
				this._err(M.abruptDoctypeSystemIdentifier), t.forceQuirks = !0, this.emitCurrentDoctype(t), this.state = z.DATA;
				break;
			case A.EOF:
				this._err(M.eofInDoctype), t.forceQuirks = !0, this.emitCurrentDoctype(t), this._emitEOFToken();
				break;
			default: t.systemId += String.fromCodePoint(e);
		}
	}
	_stateAfterDoctypeSystemIdentifier(e) {
		let t = this.currentToken;
		switch (e) {
			case A.SPACE:
			case A.LINE_FEED:
			case A.TABULATION:
			case A.FORM_FEED: break;
			case A.GREATER_THAN_SIGN:
				this.emitCurrentDoctype(t), this.state = z.DATA;
				break;
			case A.EOF:
				this._err(M.eofInDoctype), t.forceQuirks = !0, this.emitCurrentDoctype(t), this._emitEOFToken();
				break;
			default: this._err(M.unexpectedCharacterAfterDoctypeSystemIdentifier), this.state = z.BOGUS_DOCTYPE, this._stateBogusDoctype(e);
		}
	}
	_stateBogusDoctype(e) {
		let t = this.currentToken;
		switch (e) {
			case A.GREATER_THAN_SIGN:
				this.emitCurrentDoctype(t), this.state = z.DATA;
				break;
			case A.NULL:
				this._err(M.unexpectedNullCharacter);
				break;
			case A.EOF: this.emitCurrentDoctype(t), this._emitEOFToken();
		}
	}
	_stateCdataSection(e) {
		switch (e) {
			case A.RIGHT_SQUARE_BRACKET:
				this.state = z.CDATA_SECTION_BRACKET;
				break;
			case A.EOF:
				this._err(M.eofInCdata), this._emitEOFToken();
				break;
			default: this._emitCodePoint(e);
		}
	}
	_stateCdataSectionBracket(e) {
		e === A.RIGHT_SQUARE_BRACKET ? this.state = z.CDATA_SECTION_END : (this._emitChars("]", -1), this.state = z.CDATA_SECTION, this._stateCdataSection(e));
	}
	_stateCdataSectionEnd(e) {
		switch (e) {
			case A.GREATER_THAN_SIGN:
				this.state = z.DATA;
				break;
			case A.RIGHT_SQUARE_BRACKET:
				this._emitChars("]", -2);
				break;
			default: this._emitChars("]]", -1), this.state = z.CDATA_SECTION, this._stateCdataSection(e);
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
		e === 0 ? (this.preprocessor.pos = this.entityStartPos, this._flushCodePointConsumedAsCharacterReference(A.AMPERSAND), this.state = !this._isCharacterReferenceInAttribute() && Vn(this.preprocessor.peek(1)) ? z.AMBIGUOUS_AMPERSAND : this.returnState) : this.state = this.returnState;
	}
	_stateAmbiguousAmpersand(e) {
		Vn(e) ? this._flushCodePointConsumedAsCharacterReference(e) : (e === A.SEMICOLON && this._err(M.unknownNamedCharacterReference), this.state = this.returnState, this._callState(e));
	}
}, qn = /* @__PURE__ */ new Set([
	L.DD,
	L.DT,
	L.LI,
	L.OPTGROUP,
	L.OPTION,
	L.P,
	L.RB,
	L.RP,
	L.RT,
	L.RTC
]), Jn = /* @__PURE__ */ new Set([
	...qn,
	L.CAPTION,
	L.COLGROUP,
	L.TBODY,
	L.TD,
	L.TFOOT,
	L.TH,
	L.THEAD,
	L.TR
]), Yn = /* @__PURE__ */ new Set([
	L.APPLET,
	L.CAPTION,
	L.HTML,
	L.MARQUEE,
	L.OBJECT,
	L.TABLE,
	L.TD,
	L.TEMPLATE,
	L.TH
]), Xn = /* @__PURE__ */ new Set([
	...Yn,
	L.OL,
	L.UL
]), Zn = /* @__PURE__ */ new Set([...Yn, L.BUTTON]), Qn = /* @__PURE__ */ new Set([
	L.ANNOTATION_XML,
	L.MI,
	L.MN,
	L.MO,
	L.MS,
	L.MTEXT
]), $n = /* @__PURE__ */ new Set([
	L.DESC,
	L.FOREIGN_OBJECT,
	L.TITLE
]), er = /* @__PURE__ */ new Set([
	L.TR,
	L.TEMPLATE,
	L.HTML
]), tr = /* @__PURE__ */ new Set([
	L.TBODY,
	L.TFOOT,
	L.THEAD,
	L.TEMPLATE,
	L.HTML
]), nr = /* @__PURE__ */ new Set([
	L.TABLE,
	L.TEMPLATE,
	L.HTML
]), rr = /* @__PURE__ */ new Set([L.TD, L.TH]), ir = class {
	get currentTmplContentOrNode() {
		return this._isInTemplate() ? this.treeAdapter.getTemplateContent(this.current) : this.current;
	}
	constructor(e, t, n) {
		this.treeAdapter = t, this.handler = n, this.items = [], this.tagIDs = [], this.stackTop = -1, this.tmplCount = 0, this.currentTagId = L.UNKNOWN, this.current = e;
	}
	_indexOf(e) {
		return this.items.lastIndexOf(e, this.stackTop);
	}
	_isInTemplate() {
		return this.currentTagId === L.TEMPLATE && this.treeAdapter.getNamespaceURI(this.current) === F.HTML;
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
		while (t > 0 && this.treeAdapter.getNamespaceURI(this.items[t]) !== F.HTML);
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
		this.popUntilPopped(In, F.HTML);
	}
	popUntilTableCellPopped() {
		this.popUntilPopped(rr, F.HTML);
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
		this.clearBackTo(nr, F.HTML);
	}
	clearBackToTableBodyContext() {
		this.clearBackTo(tr, F.HTML);
	}
	clearBackToTableRowContext() {
		this.clearBackTo(er, F.HTML);
	}
	remove(e) {
		let t = this._indexOf(e);
		t >= 0 && (t === this.stackTop ? this.pop() : (this.items.splice(t, 1), this.tagIDs.splice(t, 1), this.stackTop--, this._updateCurrentElement(), this.handler.onItemPop(e, !1)));
	}
	tryPeekProperlyNestedBodyElement() {
		return this.stackTop >= 1 && this.tagIDs[1] === L.BODY ? this.items[1] : null;
	}
	contains(e) {
		return this._indexOf(e) > -1;
	}
	getCommonAncestor(e) {
		let t = this._indexOf(e) - 1;
		return t >= 0 ? this.items[t] : null;
	}
	isRootHtmlElementCurrent() {
		return this.stackTop === 0 && this.tagIDs[0] === L.HTML;
	}
	hasInDynamicScope(e, t) {
		for (let n = this.stackTop; n >= 0; n--) {
			let r = this.tagIDs[n];
			switch (this.treeAdapter.getNamespaceURI(this.items[n])) {
				case F.HTML:
					if (r === e) return !0;
					if (t.has(r)) return !1;
					break;
				case F.SVG:
					if ($n.has(r)) return !1;
					break;
				case F.MATHML: if (Qn.has(r)) return !1;
			}
		}
		return !0;
	}
	hasInScope(e) {
		return this.hasInDynamicScope(e, Yn);
	}
	hasInListItemScope(e) {
		return this.hasInDynamicScope(e, Xn);
	}
	hasInButtonScope(e) {
		return this.hasInDynamicScope(e, Zn);
	}
	hasNumberedHeaderInScope() {
		for (let e = this.stackTop; e >= 0; e--) {
			let t = this.tagIDs[e];
			switch (this.treeAdapter.getNamespaceURI(this.items[e])) {
				case F.HTML:
					if (In.has(t)) return !0;
					if (Yn.has(t)) return !1;
					break;
				case F.SVG:
					if ($n.has(t)) return !1;
					break;
				case F.MATHML: if (Qn.has(t)) return !1;
			}
		}
		return !0;
	}
	hasInTableScope(e) {
		for (let t = this.stackTop; t >= 0; t--) if (this.treeAdapter.getNamespaceURI(this.items[t]) === F.HTML) switch (this.tagIDs[t]) {
			case e: return !0;
			case L.TABLE:
			case L.HTML: return !1;
		}
		return !0;
	}
	hasTableBodyContextInTableScope() {
		for (let e = this.stackTop; e >= 0; e--) if (this.treeAdapter.getNamespaceURI(this.items[e]) === F.HTML) switch (this.tagIDs[e]) {
			case L.TBODY:
			case L.THEAD:
			case L.TFOOT: return !0;
			case L.TABLE:
			case L.HTML: return !1;
		}
		return !0;
	}
	hasInSelectScope(e) {
		for (let t = this.stackTop; t >= 0; t--) if (this.treeAdapter.getNamespaceURI(this.items[t]) === F.HTML) switch (this.tagIDs[t]) {
			case e: return !0;
			case L.OPTION:
			case L.OPTGROUP: break;
			default: return !1;
		}
		return !0;
	}
	generateImpliedEndTags() {
		for (; this.currentTagId !== void 0 && qn.has(this.currentTagId);) this.pop();
	}
	generateImpliedEndTagsThoroughly() {
		for (; this.currentTagId !== void 0 && Jn.has(this.currentTagId);) this.pop();
	}
	generateImpliedEndTagsWithExclusion(e) {
		for (; this.currentTagId !== void 0 && this.currentTagId !== e && Jn.has(this.currentTagId);) this.pop();
	}
}, ar = 3, or;
(function(e) {
	e[e.Marker = 0] = "Marker", e[e.Element = 1] = "Element";
})(or ||= {});
var sr = { type: or.Marker }, cr = class {
	constructor(e) {
		this.treeAdapter = e, this.entries = [], this.bookmark = null;
	}
	_getNoahArkConditionCandidates(e, t) {
		let n = [], r = t.length, i = this.treeAdapter.getTagName(e), a = this.treeAdapter.getNamespaceURI(e);
		for (let e = 0; e < this.entries.length; e++) {
			let t = this.entries[e];
			if (t.type === or.Marker) break;
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
		if (this.entries.length < ar) return;
		let t = this.treeAdapter.getAttrList(e), n = this._getNoahArkConditionCandidates(e, t);
		if (n.length < ar) return;
		let r = new Map(t.map((e) => [e.name, e.value])), i = 0;
		for (let e = 0; e < n.length; e++) {
			let t = n[e];
			t.attrs.every((e) => r.get(e.name) === e.value) && (i += 1, i >= ar && this.entries.splice(t.idx, 1));
		}
	}
	insertMarker() {
		this.entries.unshift(sr);
	}
	pushElement(e, t) {
		this._ensureNoahArkCondition(e), this.entries.unshift({
			type: or.Element,
			element: e,
			token: t
		});
	}
	insertElementAfterBookmark(e, t) {
		let n = this.entries.indexOf(this.bookmark);
		this.entries.splice(n, 0, {
			type: or.Element,
			element: e,
			token: t
		});
	}
	removeEntry(e) {
		let t = this.entries.indexOf(e);
		t !== -1 && this.entries.splice(t, 1);
	}
	clearToLastMarker() {
		let e = this.entries.indexOf(sr);
		e === -1 ? this.entries.length = 0 : this.entries.splice(0, e + 1);
	}
	getElementEntryInScopeWithTagName(e) {
		let t = this.entries.find((t) => t.type === or.Marker || this.treeAdapter.getTagName(t.element) === e);
		return t && t.type === or.Element ? t : null;
	}
	getElementEntry(e) {
		return this.entries.find((t) => t.type === or.Element && t.element === e);
	}
}, lr = {
	createDocument() {
		return {
			nodeName: "#document",
			mode: Mn.NO_QUIRKS,
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
			lr.appendChild(e, i);
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
			if (lr.isTextNode(n)) {
				n.value += t;
				return;
			}
		}
		lr.appendChild(e, lr.createTextNode(t));
	},
	insertTextBefore(e, t, n) {
		let r = e.childNodes[e.childNodes.indexOf(n) - 1];
		r && lr.isTextNode(r) ? r.value += t : lr.insertBefore(e, lr.createTextNode(t), n);
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
}, ur = "html", dr = "about:legacy-compat", fr = "http://www.ibm.com/data/dtd/v11/ibmxhtml1-transitional.dtd", pr = /* @__PURE__ */ "+//silmaril//dtd html pro v0r11 19970101//,-//as//dtd html 3.0 aswedit + extensions//,-//advasoft ltd//dtd html 3.0 aswedit + extensions//,-//ietf//dtd html 2.0 level 1//,-//ietf//dtd html 2.0 level 2//,-//ietf//dtd html 2.0 strict level 1//,-//ietf//dtd html 2.0 strict level 2//,-//ietf//dtd html 2.0 strict//,-//ietf//dtd html 2.0//,-//ietf//dtd html 2.1e//,-//ietf//dtd html 3.0//,-//ietf//dtd html 3.2 final//,-//ietf//dtd html 3.2//,-//ietf//dtd html 3//,-//ietf//dtd html level 0//,-//ietf//dtd html level 1//,-//ietf//dtd html level 2//,-//ietf//dtd html level 3//,-//ietf//dtd html strict level 0//,-//ietf//dtd html strict level 1//,-//ietf//dtd html strict level 2//,-//ietf//dtd html strict level 3//,-//ietf//dtd html strict//,-//ietf//dtd html//,-//metrius//dtd metrius presentational//,-//microsoft//dtd internet explorer 2.0 html strict//,-//microsoft//dtd internet explorer 2.0 html//,-//microsoft//dtd internet explorer 2.0 tables//,-//microsoft//dtd internet explorer 3.0 html strict//,-//microsoft//dtd internet explorer 3.0 html//,-//microsoft//dtd internet explorer 3.0 tables//,-//netscape comm. corp.//dtd html//,-//netscape comm. corp.//dtd strict html//,-//o'reilly and associates//dtd html 2.0//,-//o'reilly and associates//dtd html extended 1.0//,-//o'reilly and associates//dtd html extended relaxed 1.0//,-//sq//dtd html 2.0 hotmetal + extensions//,-//softquad software//dtd hotmetal pro 6.0::19990601::extensions to html 4.0//,-//softquad//dtd hotmetal pro 4.0::19971010::extensions to html 4.0//,-//spyglass//dtd html 2.0 extended//,-//sun microsystems corp.//dtd hotjava html//,-//sun microsystems corp.//dtd hotjava strict html//,-//w3c//dtd html 3 1995-03-24//,-//w3c//dtd html 3.2 draft//,-//w3c//dtd html 3.2 final//,-//w3c//dtd html 3.2//,-//w3c//dtd html 3.2s draft//,-//w3c//dtd html 4.0 frameset//,-//w3c//dtd html 4.0 transitional//,-//w3c//dtd html experimental 19960712//,-//w3c//dtd html experimental 970421//,-//w3c//dtd w3 html//,-//w3o//dtd w3 html 3.0//,-//webtechs//dtd mozilla html 2.0//,-//webtechs//dtd mozilla html//".split(","), mr = [
	...pr,
	"-//w3c//dtd html 4.01 frameset//",
	"-//w3c//dtd html 4.01 transitional//"
], hr = /* @__PURE__ */ new Set([
	"-//w3o//dtd w3 html strict 3.0//en//",
	"-/w3c/dtd html 4.0 transitional/en",
	"html"
]), gr = ["-//w3c//dtd xhtml 1.0 frameset//", "-//w3c//dtd xhtml 1.0 transitional//"], _r = [
	...gr,
	"-//w3c//dtd html 4.01 frameset//",
	"-//w3c//dtd html 4.01 transitional//"
];
function vr(e, t) {
	return t.some((t) => e.startsWith(t));
}
function yr(e) {
	return e.name === ur && e.publicId === null && (e.systemId === null || e.systemId === dr);
}
function br(e) {
	if (e.name !== ur) return Mn.QUIRKS;
	let { systemId: t } = e;
	if (t && t.toLowerCase() === fr) return Mn.QUIRKS;
	let { publicId: n } = e;
	if (n !== null) {
		if (n = n.toLowerCase(), hr.has(n)) return Mn.QUIRKS;
		let e = t === null ? mr : pr;
		if (vr(n, e)) return Mn.QUIRKS;
		if (e = t === null ? gr : _r, vr(n, e)) return Mn.LIMITED_QUIRKS;
	}
	return Mn.NO_QUIRKS;
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/parse5/8.1.0/38d3691d20fe3ab4780a57194f923f73d434662a4115031b2d4cf9cab8ac9fb9/node_modules/parse5/dist/common/foreign-content.js
var xr = {
	TEXT_HTML: "text/html",
	APPLICATION_XML: "application/xhtml+xml"
}, Sr = "definitionurl", Cr = "definitionURL", wr = new Map((/* @__PURE__ */ "attributeName.attributeType.baseFrequency.baseProfile.calcMode.clipPathUnits.diffuseConstant.edgeMode.filterUnits.glyphRef.gradientTransform.gradientUnits.kernelMatrix.kernelUnitLength.keyPoints.keySplines.keyTimes.lengthAdjust.limitingConeAngle.markerHeight.markerUnits.markerWidth.maskContentUnits.maskUnits.numOctaves.pathLength.patternContentUnits.patternTransform.patternUnits.pointsAtX.pointsAtY.pointsAtZ.preserveAlpha.preserveAspectRatio.primitiveUnits.refX.refY.repeatCount.repeatDur.requiredExtensions.requiredFeatures.specularConstant.specularExponent.spreadMethod.startOffset.stdDeviation.stitchTiles.surfaceScale.systemLanguage.tableValues.targetX.targetY.textLength.viewBox.viewTarget.xChannelSelector.yChannelSelector.zoomAndPan".split(".")).map((e) => [e.toLowerCase(), e])), Tr = /* @__PURE__ */ new Map([
	["xlink:actuate", {
		prefix: "xlink",
		name: "actuate",
		namespace: F.XLINK
	}],
	["xlink:arcrole", {
		prefix: "xlink",
		name: "arcrole",
		namespace: F.XLINK
	}],
	["xlink:href", {
		prefix: "xlink",
		name: "href",
		namespace: F.XLINK
	}],
	["xlink:role", {
		prefix: "xlink",
		name: "role",
		namespace: F.XLINK
	}],
	["xlink:show", {
		prefix: "xlink",
		name: "show",
		namespace: F.XLINK
	}],
	["xlink:title", {
		prefix: "xlink",
		name: "title",
		namespace: F.XLINK
	}],
	["xlink:type", {
		prefix: "xlink",
		name: "type",
		namespace: F.XLINK
	}],
	["xml:lang", {
		prefix: "xml",
		name: "lang",
		namespace: F.XML
	}],
	["xml:space", {
		prefix: "xml",
		name: "space",
		namespace: F.XML
	}],
	["xmlns", {
		prefix: "",
		name: "xmlns",
		namespace: F.XMLNS
	}],
	["xmlns:xlink", {
		prefix: "xmlns",
		name: "xlink",
		namespace: F.XMLNS
	}]
]), Er = new Map((/* @__PURE__ */ "altGlyph.altGlyphDef.altGlyphItem.animateColor.animateMotion.animateTransform.clipPath.feBlend.feColorMatrix.feComponentTransfer.feComposite.feConvolveMatrix.feDiffuseLighting.feDisplacementMap.feDistantLight.feFlood.feFuncA.feFuncB.feFuncG.feFuncR.feGaussianBlur.feImage.feMerge.feMergeNode.feMorphology.feOffset.fePointLight.feSpecularLighting.feSpotLight.feTile.feTurbulence.foreignObject.glyphRef.linearGradient.radialGradient.textPath".split(".")).map((e) => [e.toLowerCase(), e])), Dr = /* @__PURE__ */ new Set([
	L.B,
	L.BIG,
	L.BLOCKQUOTE,
	L.BODY,
	L.BR,
	L.CENTER,
	L.CODE,
	L.DD,
	L.DIV,
	L.DL,
	L.DT,
	L.EM,
	L.EMBED,
	L.H1,
	L.H2,
	L.H3,
	L.H4,
	L.H5,
	L.H6,
	L.HEAD,
	L.HR,
	L.I,
	L.IMG,
	L.LI,
	L.LISTING,
	L.MENU,
	L.META,
	L.NOBR,
	L.OL,
	L.P,
	L.PRE,
	L.RUBY,
	L.S,
	L.SMALL,
	L.SPAN,
	L.STRONG,
	L.STRIKE,
	L.SUB,
	L.SUP,
	L.TABLE,
	L.TT,
	L.U,
	L.UL,
	L.VAR
]);
function Or(e) {
	let t = e.tagID;
	return t === L.FONT && e.attrs.some(({ name: e }) => e === jn.COLOR || e === jn.SIZE || e === jn.FACE) || Dr.has(t);
}
function kr(e) {
	for (let t = 0; t < e.attrs.length; t++) if (e.attrs[t].name === Sr) {
		e.attrs[t].name = Cr;
		break;
	}
}
function Ar(e) {
	for (let t = 0; t < e.attrs.length; t++) {
		let n = wr.get(e.attrs[t].name);
		n != null && (e.attrs[t].name = n);
	}
}
function jr(e) {
	for (let t = 0; t < e.attrs.length; t++) {
		let n = Tr.get(e.attrs[t].name);
		n && (e.attrs[t].prefix = n.prefix, e.attrs[t].name = n.name, e.attrs[t].namespace = n.namespace);
	}
}
function Mr(e) {
	let t = Er.get(e.tagName);
	t != null && (e.tagName = t, e.tagID = Pn(e.tagName));
}
function Nr(e, t) {
	return t === F.MATHML && (e === L.MI || e === L.MO || e === L.MN || e === L.MS || e === L.MTEXT);
}
function Pr(e, t, n) {
	if (t === F.MATHML && e === L.ANNOTATION_XML) {
		for (let e = 0; e < n.length; e++) if (n[e].name === jn.ENCODING) {
			let t = n[e].value.toLowerCase();
			return t === xr.TEXT_HTML || t === xr.APPLICATION_XML;
		}
	}
	return t === F.SVG && (e === L.FOREIGN_OBJECT || e === L.DESC || e === L.TITLE);
}
function Fr(e, t, n, r) {
	return (!r || r === F.HTML) && Pr(e, t, n) || (!r || r === F.MATHML) && Nr(e, t);
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/parse5/8.1.0/38d3691d20fe3ab4780a57194f923f73d434662a4115031b2d4cf9cab8ac9fb9/node_modules/parse5/dist/parser/index.js
var Ir = "hidden", Lr = 8, Rr = 3, V;
(function(e) {
	e[e.INITIAL = 0] = "INITIAL", e[e.BEFORE_HTML = 1] = "BEFORE_HTML", e[e.BEFORE_HEAD = 2] = "BEFORE_HEAD", e[e.IN_HEAD = 3] = "IN_HEAD", e[e.IN_HEAD_NO_SCRIPT = 4] = "IN_HEAD_NO_SCRIPT", e[e.AFTER_HEAD = 5] = "AFTER_HEAD", e[e.IN_BODY = 6] = "IN_BODY", e[e.TEXT = 7] = "TEXT", e[e.IN_TABLE = 8] = "IN_TABLE", e[e.IN_TABLE_TEXT = 9] = "IN_TABLE_TEXT", e[e.IN_CAPTION = 10] = "IN_CAPTION", e[e.IN_COLUMN_GROUP = 11] = "IN_COLUMN_GROUP", e[e.IN_TABLE_BODY = 12] = "IN_TABLE_BODY", e[e.IN_ROW = 13] = "IN_ROW", e[e.IN_CELL = 14] = "IN_CELL", e[e.IN_SELECT = 15] = "IN_SELECT", e[e.IN_SELECT_IN_TABLE = 16] = "IN_SELECT_IN_TABLE", e[e.IN_TEMPLATE = 17] = "IN_TEMPLATE", e[e.AFTER_BODY = 18] = "AFTER_BODY", e[e.IN_FRAMESET = 19] = "IN_FRAMESET", e[e.AFTER_FRAMESET = 20] = "AFTER_FRAMESET", e[e.AFTER_AFTER_BODY = 21] = "AFTER_AFTER_BODY", e[e.AFTER_AFTER_FRAMESET = 22] = "AFTER_AFTER_FRAMESET";
})(V ||= {});
var zr = {
	startLine: -1,
	startCol: -1,
	startOffset: -1,
	endLine: -1,
	endCol: -1,
	endOffset: -1
}, Br = /* @__PURE__ */ new Set([
	L.TABLE,
	L.TBODY,
	L.TFOOT,
	L.THEAD,
	L.TR
]), Vr = {
	scriptingEnabled: !0,
	sourceCodeLocationInfo: !1,
	treeAdapter: lr,
	onParseError: null
}, Hr = class {
	constructor(e, t, n = null, r = null) {
		this.fragmentContext = n, this.scriptHandler = r, this.currentToken = null, this.stopped = !1, this.insertionMode = V.INITIAL, this.originalInsertionMode = V.INITIAL, this.headElement = null, this.formElement = null, this.currentNotInHTML = !1, this.tmplInsertionModeStack = [], this.pendingCharacterTokens = [], this.hasNonWhitespacePendingCharacterToken = !1, this.framesetOk = !0, this.skipNextNewLine = !1, this.fosterParentingEnabled = !1, this.options = {
			...Vr,
			...e
		}, this.treeAdapter = this.options.treeAdapter, this.onParseError = this.options.onParseError, this.onParseError && (this.options.sourceCodeLocationInfo = !0), this.document = t ?? this.treeAdapter.createDocument(), this.tokenizer = new Kn(this.options, this), this.activeFormattingElements = new cr(this.treeAdapter), this.fragmentContextID = n ? Pn(this.treeAdapter.getTagName(n)) : L.UNKNOWN, this._setContextModes(n ?? this.document, this.fragmentContextID), this.openElements = new ir(this.document, this.treeAdapter, this);
	}
	static parse(e, t) {
		let n = new this(t);
		return n.tokenizer.write(e, !0), n.document;
	}
	static getFragmentParser(e, t) {
		let n = {
			...Vr,
			...t
		};
		e ??= n.treeAdapter.createElement(I.TEMPLATE, F.HTML, []);
		let r = n.treeAdapter.createElement("documentmock", F.HTML, []), i = new this(n, r, e);
		return i.fragmentContextID === L.TEMPLATE && i.tmplInsertionModeStack.unshift(V.IN_TEMPLATE), i._initTokenizerForFragmentParsing(), i._insertFakeRootElement(), i._resetInsertionMode(), i._findFormInFragmentContext(), i;
	}
	getFragment() {
		let e = this.treeAdapter.getFirstChild(this.document), t = this.treeAdapter.createDocumentFragment();
		return this._adoptNodes(e, t), t;
	}
	_err(e, t, n) {
		if (!this.onParseError) return;
		let r = e.location ?? zr, i = {
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
		let n = e === this.document || e && this.treeAdapter.getNamespaceURI(e) === F.HTML;
		this.currentNotInHTML = !n, this.tokenizer.inForeignNode = !n && e !== void 0 && t !== void 0 && !this._isIntegrationPoint(t, e);
	}
	_switchToTextParsing(e, t) {
		this._insertElement(e, F.HTML), this.tokenizer.state = t, this.originalInsertionMode = this.insertionMode, this.insertionMode = V.TEXT;
	}
	switchToPlaintextParsing() {
		this.insertionMode = V.TEXT, this.originalInsertionMode = V.IN_BODY, this.tokenizer.state = B.PLAINTEXT;
	}
	_getAdjustedCurrentElement() {
		return this.openElements.stackTop === 0 && this.fragmentContext ? this.fragmentContext : this.openElements.current;
	}
	_findFormInFragmentContext() {
		let e = this.fragmentContext;
		for (; e;) {
			if (this.treeAdapter.getTagName(e) === I.FORM) {
				this.formElement = e;
				break;
			}
			e = this.treeAdapter.getParentNode(e);
		}
	}
	_initTokenizerForFragmentParsing() {
		if (this.fragmentContext && this.treeAdapter.getNamespaceURI(this.fragmentContext) === F.HTML) switch (this.fragmentContextID) {
			case L.TITLE:
			case L.TEXTAREA:
				this.tokenizer.state = B.RCDATA;
				break;
			case L.STYLE:
			case L.XMP:
			case L.IFRAME:
			case L.NOEMBED:
			case L.NOFRAMES:
			case L.NOSCRIPT:
				this.tokenizer.state = B.RAWTEXT;
				break;
			case L.SCRIPT:
				this.tokenizer.state = B.SCRIPT_DATA;
				break;
			case L.PLAINTEXT: this.tokenizer.state = B.PLAINTEXT;
		}
	}
	_setDocumentType(e) {
		let t = e.name || "", n = e.publicId || "", r = e.systemId || "";
		if (this.treeAdapter.setDocumentType(this.document, t, n, r), e.location) {
			let t = this.treeAdapter.getChildNodes(this.document).find((e) => this.treeAdapter.isDocumentTypeNode(e));
			t && this.treeAdapter.setNodeSourceCodeLocation(t, e.location);
		}
	}
	_attachElementToTree(e, t, n) {
		if (this.options.sourceCodeLocationInfo) {
			let n = t && {
				...t,
				startTag: t
			};
			this.treeAdapter.setNodeSourceCodeLocation(e, n);
		}
		n.beforeElement ? this.treeAdapter.insertBefore(n.parent, e, n.beforeElement) : this.treeAdapter.appendChild(n.parent, e);
	}
	_getInsertionLocation() {
		return this._shouldFosterParentOnInsertion() ? this._findFosterParentingLocation() : {
			parent: this.openElements.currentTmplContentOrNode ?? this.document,
			beforeElement: null
		};
	}
	_createElement(e, t, n, r) {
		var i, a;
		let o = this.treeAdapter.createElement(e, t, n);
		return (a = (i = this.treeAdapter).onElementCreated) == null || a.call(i, o, {
			intendedParent: r,
			formElement: this.formElement,
			isInTemplate: this.openElements.tmplCount > 0
		}), o;
	}
	_appendElement(e, t) {
		let n = this._getInsertionLocation(), r = this._createElement(e.tagName, t, e.attrs, n.parent);
		this._attachElementToTree(r, e.location, n);
	}
	_insertElement(e, t) {
		let n = this._getInsertionLocation(), r = this._createElement(e.tagName, t, e.attrs, n.parent);
		this._attachElementToTree(r, e.location, n), this.openElements.push(r, e.tagID);
	}
	_insertFakeElement(e, t) {
		let n = this._getInsertionLocation(), r = this._createElement(e, F.HTML, [], n.parent);
		this._attachElementToTree(r, null, n), this.openElements.push(r, t);
	}
	_insertTemplate(e) {
		let t = this._getInsertionLocation(), n = this._createElement(e.tagName, F.HTML, e.attrs, t.parent), r = this.treeAdapter.createDocumentFragment();
		this.treeAdapter.setTemplateContent(n, r), this._attachElementToTree(n, e.location, t), this.openElements.push(n, e.tagID), this.options.sourceCodeLocationInfo && this.treeAdapter.setNodeSourceCodeLocation(r, null);
	}
	_insertFakeRootElement() {
		let e = this._createElement(I.HTML, F.HTML, [], this.document);
		this.options.sourceCodeLocationInfo && this.treeAdapter.setNodeSourceCodeLocation(e, null), this.treeAdapter.appendChild(this.document, e), this.openElements.push(e, L.HTML);
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
			let n = t.location, r = this.treeAdapter.getTagName(e), i = t.type === N.END_TAG && r === t.tagName ? {
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
		return this.openElements.stackTop === 0 && this.fragmentContext ? (t = this.fragmentContext, n = this.fragmentContextID) : {current: t, currentTagId: n} = this.openElements, e.tagID === L.SVG && this.treeAdapter.getTagName(t) === I.ANNOTATION_XML && this.treeAdapter.getNamespaceURI(t) === F.MATHML ? !1 : this.tokenizer.inForeignNode || (e.tagID === L.MGLYPH || e.tagID === L.MALIGNMARK) && n !== void 0 && !this._isIntegrationPoint(n, t, F.HTML);
	}
	_processToken(e) {
		switch (e.type) {
			case N.CHARACTER:
				this.onCharacter(e);
				break;
			case N.NULL_CHARACTER:
				this.onNullCharacter(e);
				break;
			case N.COMMENT:
				this.onComment(e);
				break;
			case N.DOCTYPE:
				this.onDoctype(e);
				break;
			case N.START_TAG:
				this._processStartTag(e);
				break;
			case N.END_TAG:
				this.onEndTag(e);
				break;
			case N.EOF:
				this.onEof(e);
				break;
			case N.WHITESPACE_CHARACTER: this.onWhitespaceCharacter(e);
		}
	}
	_isIntegrationPoint(e, t, n) {
		return Fr(e, this.treeAdapter.getNamespaceURI(t), this.treeAdapter.getAttrList(t), n);
	}
	_reconstructActiveFormattingElements() {
		let e = this.activeFormattingElements.entries.length;
		if (e) {
			let t = this.activeFormattingElements.entries.findIndex((e) => e.type === or.Marker || this.openElements.contains(e.element)), n = t === -1 ? e - 1 : t - 1;
			for (let e = n; e >= 0; e--) {
				let t = this.activeFormattingElements.entries[e];
				this._insertElement(t.token, this.treeAdapter.getNamespaceURI(t.element)), t.element = this.openElements.current;
			}
		}
	}
	_closeTableCell() {
		this.openElements.generateImpliedEndTags(), this.openElements.popUntilTableCellPopped(), this.activeFormattingElements.clearToLastMarker(), this.insertionMode = V.IN_ROW;
	}
	_closePElement() {
		this.openElements.generateImpliedEndTagsWithExclusion(L.P), this.openElements.popUntilTagNamePopped(L.P);
	}
	_resetInsertionMode() {
		for (let e = this.openElements.stackTop; e >= 0; e--) switch (e === 0 && this.fragmentContext ? this.fragmentContextID : this.openElements.tagIDs[e]) {
			case L.TR:
				this.insertionMode = V.IN_ROW;
				return;
			case L.TBODY:
			case L.THEAD:
			case L.TFOOT:
				this.insertionMode = V.IN_TABLE_BODY;
				return;
			case L.CAPTION:
				this.insertionMode = V.IN_CAPTION;
				return;
			case L.COLGROUP:
				this.insertionMode = V.IN_COLUMN_GROUP;
				return;
			case L.TABLE:
				this.insertionMode = V.IN_TABLE;
				return;
			case L.BODY:
				this.insertionMode = V.IN_BODY;
				return;
			case L.FRAMESET:
				this.insertionMode = V.IN_FRAMESET;
				return;
			case L.SELECT:
				this._resetInsertionModeForSelect(e);
				return;
			case L.TEMPLATE:
				this.insertionMode = this.tmplInsertionModeStack[0];
				return;
			case L.HTML:
				this.insertionMode = this.headElement ? V.AFTER_HEAD : V.BEFORE_HEAD;
				return;
			case L.TD:
			case L.TH:
				if (e > 0) {
					this.insertionMode = V.IN_CELL;
					return;
				}
				break;
			case L.HEAD: if (e > 0) {
				this.insertionMode = V.IN_HEAD;
				return;
			}
		}
		this.insertionMode = V.IN_BODY;
	}
	_resetInsertionModeForSelect(e) {
		if (e > 0) for (let t = e - 1; t > 0; t--) {
			let e = this.openElements.tagIDs[t];
			if (e === L.TEMPLATE) break;
			if (e === L.TABLE) {
				this.insertionMode = V.IN_SELECT_IN_TABLE;
				return;
			}
		}
		this.insertionMode = V.IN_SELECT;
	}
	_isElementCausesFosterParenting(e) {
		return Br.has(e);
	}
	_shouldFosterParentOnInsertion() {
		return this.fosterParentingEnabled && this.openElements.currentTagId !== void 0 && this._isElementCausesFosterParenting(this.openElements.currentTagId);
	}
	_findFosterParentingLocation() {
		for (let e = this.openElements.stackTop; e >= 0; e--) {
			let t = this.openElements.items[e];
			switch (this.openElements.tagIDs[e]) {
				case L.TEMPLATE:
					if (this.treeAdapter.getNamespaceURI(t) === F.HTML) return {
						parent: this.treeAdapter.getTemplateContent(t),
						beforeElement: null
					};
					break;
				case L.TABLE: {
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
		return Fn[this.treeAdapter.getNamespaceURI(e)].has(t);
	}
	onCharacter(e) {
		if (this.skipNextNewLine = !1, this.tokenizer.inForeignNode) ao(this, e);
		else switch (this.insertionMode) {
			case V.INITIAL:
				ti(this, e);
				break;
			case V.BEFORE_HTML:
				ii(this, e);
				break;
			case V.BEFORE_HEAD:
				si(this, e);
				break;
			case V.IN_HEAD:
				di(this, e);
				break;
			case V.IN_HEAD_NO_SCRIPT:
				mi(this, e);
				break;
			case V.AFTER_HEAD:
				_i(this, e);
				break;
			case V.IN_BODY:
			case V.IN_CAPTION:
			case V.IN_CELL:
			case V.IN_TEMPLATE:
				bi(this, e);
				break;
			case V.TEXT:
			case V.IN_SELECT:
			case V.IN_SELECT_IN_TABLE:
				this._insertCharacters(e);
				break;
			case V.IN_TABLE:
			case V.IN_TABLE_BODY:
			case V.IN_ROW:
				ma(this, e);
				break;
			case V.IN_TABLE_TEXT:
				Da(this, e);
				break;
			case V.IN_COLUMN_GROUP:
				Pa(this, e);
				break;
			case V.AFTER_BODY:
				Xa(this, e);
				break;
			case V.AFTER_AFTER_BODY: no(this, e);
		}
	}
	onNullCharacter(e) {
		if (this.skipNextNewLine = !1, this.tokenizer.inForeignNode) io(this, e);
		else switch (this.insertionMode) {
			case V.INITIAL:
				ti(this, e);
				break;
			case V.BEFORE_HTML:
				ii(this, e);
				break;
			case V.BEFORE_HEAD:
				si(this, e);
				break;
			case V.IN_HEAD:
				di(this, e);
				break;
			case V.IN_HEAD_NO_SCRIPT:
				mi(this, e);
				break;
			case V.AFTER_HEAD:
				_i(this, e);
				break;
			case V.TEXT:
				this._insertCharacters(e);
				break;
			case V.IN_TABLE:
			case V.IN_TABLE_BODY:
			case V.IN_ROW:
				ma(this, e);
				break;
			case V.IN_COLUMN_GROUP:
				Pa(this, e);
				break;
			case V.AFTER_BODY:
				Xa(this, e);
				break;
			case V.AFTER_AFTER_BODY: no(this, e);
		}
	}
	onComment(e) {
		if (this.skipNextNewLine = !1, this.currentNotInHTML) Xr(this, e);
		else switch (this.insertionMode) {
			case V.INITIAL:
			case V.BEFORE_HTML:
			case V.BEFORE_HEAD:
			case V.IN_HEAD:
			case V.IN_HEAD_NO_SCRIPT:
			case V.AFTER_HEAD:
			case V.IN_BODY:
			case V.IN_TABLE:
			case V.IN_CAPTION:
			case V.IN_COLUMN_GROUP:
			case V.IN_TABLE_BODY:
			case V.IN_ROW:
			case V.IN_CELL:
			case V.IN_SELECT:
			case V.IN_SELECT_IN_TABLE:
			case V.IN_TEMPLATE:
			case V.IN_FRAMESET:
			case V.AFTER_FRAMESET:
				Xr(this, e);
				break;
			case V.IN_TABLE_TEXT:
				Oa(this, e);
				break;
			case V.AFTER_BODY:
				Zr(this, e);
				break;
			case V.AFTER_AFTER_BODY:
			case V.AFTER_AFTER_FRAMESET: Qr(this, e);
		}
	}
	onDoctype(e) {
		switch (this.skipNextNewLine = !1, this.insertionMode) {
			case V.INITIAL:
				ei(this, e);
				break;
			case V.BEFORE_HEAD:
			case V.IN_HEAD:
			case V.IN_HEAD_NO_SCRIPT:
			case V.AFTER_HEAD:
				this._err(e, M.misplacedDoctype);
				break;
			case V.IN_TABLE_TEXT: Oa(this, e);
		}
	}
	onStartTag(e) {
		this.skipNextNewLine = !1, this.currentToken = e, this._processStartTag(e), e.selfClosing && !e.ackSelfClosing && this._err(e, M.nonVoidHtmlElementStartTagWithTrailingSolidus);
	}
	_processStartTag(e) {
		this.shouldProcessStartTagTokenInForeignContent(e) ? so(this, e) : this._startTagOutsideForeignContent(e);
	}
	_startTagOutsideForeignContent(e) {
		switch (this.insertionMode) {
			case V.INITIAL:
				ti(this, e);
				break;
			case V.BEFORE_HTML:
				ni(this, e);
				break;
			case V.BEFORE_HEAD:
				ai(this, e);
				break;
			case V.IN_HEAD:
				ci(this, e);
				break;
			case V.IN_HEAD_NO_SCRIPT:
				fi(this, e);
				break;
			case V.AFTER_HEAD:
				hi(this, e);
				break;
			case V.IN_BODY:
				H(this, e);
				break;
			case V.IN_TABLE:
				Ca(this, e);
				break;
			case V.IN_TABLE_TEXT:
				Oa(this, e);
				break;
			case V.IN_CAPTION:
				Aa(this, e);
				break;
			case V.IN_COLUMN_GROUP:
				Ma(this, e);
				break;
			case V.IN_TABLE_BODY:
				Fa(this, e);
				break;
			case V.IN_ROW:
				La(this, e);
				break;
			case V.IN_CELL:
				za(this, e);
				break;
			case V.IN_SELECT:
				Va(this, e);
				break;
			case V.IN_SELECT_IN_TABLE:
				Ua(this, e);
				break;
			case V.IN_TEMPLATE:
				Ga(this, e);
				break;
			case V.AFTER_BODY:
				Ja(this, e);
				break;
			case V.IN_FRAMESET:
				Za(this, e);
				break;
			case V.AFTER_FRAMESET:
				$a(this, e);
				break;
			case V.AFTER_AFTER_BODY:
				to(this, e);
				break;
			case V.AFTER_AFTER_FRAMESET: ro(this, e);
		}
	}
	onEndTag(e) {
		this.skipNextNewLine = !1, this.currentToken = e, this.currentNotInHTML ? co(this, e) : this._endTagOutsideForeignContent(e);
	}
	_endTagOutsideForeignContent(e) {
		switch (this.insertionMode) {
			case V.INITIAL:
				ti(this, e);
				break;
			case V.BEFORE_HTML:
				ri(this, e);
				break;
			case V.BEFORE_HEAD:
				oi(this, e);
				break;
			case V.IN_HEAD:
				li(this, e);
				break;
			case V.IN_HEAD_NO_SCRIPT:
				pi(this, e);
				break;
			case V.AFTER_HEAD:
				gi(this, e);
				break;
			case V.IN_BODY:
				ua(this, e);
				break;
			case V.TEXT:
				fa(this, e);
				break;
			case V.IN_TABLE:
				wa(this, e);
				break;
			case V.IN_TABLE_TEXT:
				Oa(this, e);
				break;
			case V.IN_CAPTION:
				ja(this, e);
				break;
			case V.IN_COLUMN_GROUP:
				Na(this, e);
				break;
			case V.IN_TABLE_BODY:
				Ia(this, e);
				break;
			case V.IN_ROW:
				Ra(this, e);
				break;
			case V.IN_CELL:
				Ba(this, e);
				break;
			case V.IN_SELECT:
				Ha(this, e);
				break;
			case V.IN_SELECT_IN_TABLE:
				Wa(this, e);
				break;
			case V.IN_TEMPLATE:
				Ka(this, e);
				break;
			case V.AFTER_BODY:
				Ya(this, e);
				break;
			case V.IN_FRAMESET:
				Qa(this, e);
				break;
			case V.AFTER_FRAMESET:
				eo(this, e);
				break;
			case V.AFTER_AFTER_BODY: no(this, e);
		}
	}
	onEof(e) {
		switch (this.insertionMode) {
			case V.INITIAL:
				ti(this, e);
				break;
			case V.BEFORE_HTML:
				ii(this, e);
				break;
			case V.BEFORE_HEAD:
				si(this, e);
				break;
			case V.IN_HEAD:
				di(this, e);
				break;
			case V.IN_HEAD_NO_SCRIPT:
				mi(this, e);
				break;
			case V.AFTER_HEAD:
				_i(this, e);
				break;
			case V.IN_BODY:
			case V.IN_TABLE:
			case V.IN_CAPTION:
			case V.IN_COLUMN_GROUP:
			case V.IN_TABLE_BODY:
			case V.IN_ROW:
			case V.IN_CELL:
			case V.IN_SELECT:
			case V.IN_SELECT_IN_TABLE:
				da(this, e);
				break;
			case V.TEXT:
				pa(this, e);
				break;
			case V.IN_TABLE_TEXT:
				Oa(this, e);
				break;
			case V.IN_TEMPLATE:
				qa(this, e);
				break;
			case V.AFTER_BODY:
			case V.IN_FRAMESET:
			case V.AFTER_FRAMESET:
			case V.AFTER_AFTER_BODY:
			case V.AFTER_AFTER_FRAMESET: $r(this, e);
		}
	}
	onWhitespaceCharacter(e) {
		if (this.skipNextNewLine && (this.skipNextNewLine = !1, e.chars.charCodeAt(0) === A.LINE_FEED)) {
			if (e.chars.length === 1) return;
			e.chars = e.chars.substr(1);
		}
		if (this.tokenizer.inForeignNode) this._insertCharacters(e);
		else switch (this.insertionMode) {
			case V.IN_HEAD:
			case V.IN_HEAD_NO_SCRIPT:
			case V.AFTER_HEAD:
			case V.TEXT:
			case V.IN_COLUMN_GROUP:
			case V.IN_SELECT:
			case V.IN_SELECT_IN_TABLE:
			case V.IN_FRAMESET:
			case V.AFTER_FRAMESET:
				this._insertCharacters(e);
				break;
			case V.IN_BODY:
			case V.IN_CAPTION:
			case V.IN_CELL:
			case V.IN_TEMPLATE:
			case V.AFTER_BODY:
			case V.AFTER_AFTER_BODY:
			case V.AFTER_AFTER_FRAMESET:
				yi(this, e);
				break;
			case V.IN_TABLE:
			case V.IN_TABLE_BODY:
			case V.IN_ROW:
				ma(this, e);
				break;
			case V.IN_TABLE_TEXT: Ea(this, e);
		}
	}
};
function Ur(e, t) {
	let n = e.activeFormattingElements.getElementEntryInScopeWithTagName(t.tagName);
	return n ? e.openElements.contains(n.element) ? e.openElements.hasInScope(t.tagID) || (n = null) : (e.activeFormattingElements.removeEntry(n), n = null) : la(e, t), n;
}
function Wr(e, t) {
	let n = null, r = e.openElements.stackTop;
	for (; r >= 0; r--) {
		let i = e.openElements.items[r];
		if (i === t.element) break;
		e._isSpecialElement(i, e.openElements.tagIDs[r]) && (n = i);
	}
	return n || (e.openElements.shortenToLength(Math.max(r, 0)), e.activeFormattingElements.removeEntry(t)), n;
}
function Gr(e, t, n) {
	let r = t, i = e.openElements.getCommonAncestor(t), a = e.openElements.getCommonAncestor(n);
	for (let o = 0, s = i; s !== n; o++, s = i) {
		i = e.openElements.getCommonAncestor(s);
		let n = e.activeFormattingElements.getElementEntry(s), c = n && o >= Rr;
		!n || c ? (c && e.activeFormattingElements.removeEntry(n), e.openElements.remove(s)) : (s = Kr(e, n, a), r === t && (e.activeFormattingElements.bookmark = n), e.treeAdapter.detachNode(r), e.treeAdapter.appendChild(s, r), r = s);
	}
	return r;
}
function Kr(e, t, n) {
	let r = e.treeAdapter.getNamespaceURI(t.element), i = e._createElement(t.token.tagName, r, t.token.attrs, n);
	return e.openElements.replace(t.element, i), t.element = i, i;
}
function qr(e, t, n) {
	let r = Pn(e.treeAdapter.getTagName(t));
	if (e._isElementCausesFosterParenting(r)) e._fosterParentElement(n);
	else {
		let i = e.treeAdapter.getNamespaceURI(t);
		r === L.TEMPLATE && i === F.HTML && (t = e.treeAdapter.getTemplateContent(t)), e.treeAdapter.appendChild(t, n);
	}
}
function Jr(e, t, n) {
	let r = e.treeAdapter.getNamespaceURI(n.element), { token: i } = n, a = e._createElement(i.tagName, r, i.attrs, t);
	e._adoptNodes(t, a), e.treeAdapter.appendChild(t, a), e.activeFormattingElements.insertElementAfterBookmark(a, i), e.activeFormattingElements.removeEntry(n), e.openElements.remove(n.element), e.openElements.insertAfter(t, a, i.tagID);
}
function Yr(e, t) {
	for (let n = 0; n < Lr; n++) {
		let n = Ur(e, t);
		if (!n) break;
		let r = Wr(e, n);
		if (!r) break;
		e.activeFormattingElements.bookmark = n;
		let i = Gr(e, r, n.element), a = e.openElements.getCommonAncestor(n.element);
		e.treeAdapter.detachNode(i), a && qr(e, a, i), Jr(e, r, n);
	}
}
function Xr(e, t) {
	e._appendCommentNode(t, e.openElements.currentTmplContentOrNode);
}
function Zr(e, t) {
	e._appendCommentNode(t, e.openElements.items[0]);
}
function Qr(e, t) {
	e._appendCommentNode(t, e.document);
}
function $r(e, t) {
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
function ei(e, t) {
	e._setDocumentType(t);
	let n = t.forceQuirks ? Mn.QUIRKS : br(t);
	yr(t) || e._err(t, M.nonConformingDoctype), e.treeAdapter.setDocumentMode(e.document, n), e.insertionMode = V.BEFORE_HTML;
}
function ti(e, t) {
	e._err(t, M.missingDoctype, !0), e.treeAdapter.setDocumentMode(e.document, Mn.QUIRKS), e.insertionMode = V.BEFORE_HTML, e._processToken(t);
}
function ni(e, t) {
	t.tagID === L.HTML ? (e._insertElement(t, F.HTML), e.insertionMode = V.BEFORE_HEAD) : ii(e, t);
}
function ri(e, t) {
	let n = t.tagID;
	(n === L.HTML || n === L.HEAD || n === L.BODY || n === L.BR) && ii(e, t);
}
function ii(e, t) {
	e._insertFakeRootElement(), e.insertionMode = V.BEFORE_HEAD, e._processToken(t);
}
function ai(e, t) {
	switch (t.tagID) {
		case L.HTML:
			H(e, t);
			break;
		case L.HEAD:
			e._insertElement(t, F.HTML), e.headElement = e.openElements.current, e.insertionMode = V.IN_HEAD;
			break;
		default: si(e, t);
	}
}
function oi(e, t) {
	let n = t.tagID;
	n === L.HEAD || n === L.BODY || n === L.HTML || n === L.BR ? si(e, t) : e._err(t, M.endTagWithoutMatchingOpenElement);
}
function si(e, t) {
	e._insertFakeElement(I.HEAD, L.HEAD), e.headElement = e.openElements.current, e.insertionMode = V.IN_HEAD, e._processToken(t);
}
function ci(e, t) {
	switch (t.tagID) {
		case L.HTML:
			H(e, t);
			break;
		case L.BASE:
		case L.BASEFONT:
		case L.BGSOUND:
		case L.LINK:
		case L.META:
			e._appendElement(t, F.HTML), t.ackSelfClosing = !0;
			break;
		case L.TITLE:
			e._switchToTextParsing(t, B.RCDATA);
			break;
		case L.NOSCRIPT:
			e.options.scriptingEnabled ? e._switchToTextParsing(t, B.RAWTEXT) : (e._insertElement(t, F.HTML), e.insertionMode = V.IN_HEAD_NO_SCRIPT);
			break;
		case L.NOFRAMES:
		case L.STYLE:
			e._switchToTextParsing(t, B.RAWTEXT);
			break;
		case L.SCRIPT:
			e._switchToTextParsing(t, B.SCRIPT_DATA);
			break;
		case L.TEMPLATE:
			e._insertTemplate(t), e.activeFormattingElements.insertMarker(), e.framesetOk = !1, e.insertionMode = V.IN_TEMPLATE, e.tmplInsertionModeStack.unshift(V.IN_TEMPLATE);
			break;
		case L.HEAD:
			e._err(t, M.misplacedStartTagForHeadElement);
			break;
		default: di(e, t);
	}
}
function li(e, t) {
	switch (t.tagID) {
		case L.HEAD:
			e.openElements.pop(), e.insertionMode = V.AFTER_HEAD;
			break;
		case L.BODY:
		case L.BR:
		case L.HTML:
			di(e, t);
			break;
		case L.TEMPLATE:
			ui(e, t);
			break;
		default: e._err(t, M.endTagWithoutMatchingOpenElement);
	}
}
function ui(e, t) {
	e.openElements.tmplCount > 0 ? (e.openElements.generateImpliedEndTagsThoroughly(), e.openElements.currentTagId !== L.TEMPLATE && e._err(t, M.closingOfElementWithOpenChildElements), e.openElements.popUntilTagNamePopped(L.TEMPLATE), e.activeFormattingElements.clearToLastMarker(), e.tmplInsertionModeStack.shift(), e._resetInsertionMode()) : e._err(t, M.endTagWithoutMatchingOpenElement);
}
function di(e, t) {
	e.openElements.pop(), e.insertionMode = V.AFTER_HEAD, e._processToken(t);
}
function fi(e, t) {
	switch (t.tagID) {
		case L.HTML:
			H(e, t);
			break;
		case L.BASEFONT:
		case L.BGSOUND:
		case L.HEAD:
		case L.LINK:
		case L.META:
		case L.NOFRAMES:
		case L.STYLE:
			ci(e, t);
			break;
		case L.NOSCRIPT:
			e._err(t, M.nestedNoscriptInHead);
			break;
		default: mi(e, t);
	}
}
function pi(e, t) {
	switch (t.tagID) {
		case L.NOSCRIPT:
			e.openElements.pop(), e.insertionMode = V.IN_HEAD;
			break;
		case L.BR:
			mi(e, t);
			break;
		default: e._err(t, M.endTagWithoutMatchingOpenElement);
	}
}
function mi(e, t) {
	let n = t.type === N.EOF ? M.openElementsLeftAfterEof : M.disallowedContentInNoscriptInHead;
	e._err(t, n), e.openElements.pop(), e.insertionMode = V.IN_HEAD, e._processToken(t);
}
function hi(e, t) {
	switch (t.tagID) {
		case L.HTML:
			H(e, t);
			break;
		case L.BODY:
			e._insertElement(t, F.HTML), e.framesetOk = !1, e.insertionMode = V.IN_BODY;
			break;
		case L.FRAMESET:
			e._insertElement(t, F.HTML), e.insertionMode = V.IN_FRAMESET;
			break;
		case L.BASE:
		case L.BASEFONT:
		case L.BGSOUND:
		case L.LINK:
		case L.META:
		case L.NOFRAMES:
		case L.SCRIPT:
		case L.STYLE:
		case L.TEMPLATE:
		case L.TITLE:
			e._err(t, M.abandonedHeadElementChild), e.openElements.push(e.headElement, L.HEAD), ci(e, t), e.openElements.remove(e.headElement);
			break;
		case L.HEAD:
			e._err(t, M.misplacedStartTagForHeadElement);
			break;
		default: _i(e, t);
	}
}
function gi(e, t) {
	switch (t.tagID) {
		case L.BODY:
		case L.HTML:
		case L.BR:
			_i(e, t);
			break;
		case L.TEMPLATE:
			ui(e, t);
			break;
		default: e._err(t, M.endTagWithoutMatchingOpenElement);
	}
}
function _i(e, t) {
	e._insertFakeElement(I.BODY, L.BODY), e.insertionMode = V.IN_BODY, vi(e, t);
}
function vi(e, t) {
	switch (t.type) {
		case N.CHARACTER:
			bi(e, t);
			break;
		case N.WHITESPACE_CHARACTER:
			yi(e, t);
			break;
		case N.COMMENT:
			Xr(e, t);
			break;
		case N.START_TAG:
			H(e, t);
			break;
		case N.END_TAG:
			ua(e, t);
			break;
		case N.EOF: da(e, t);
	}
}
function yi(e, t) {
	e._reconstructActiveFormattingElements(), e._insertCharacters(t);
}
function bi(e, t) {
	e._reconstructActiveFormattingElements(), e._insertCharacters(t), e.framesetOk = !1;
}
function xi(e, t) {
	e.openElements.tmplCount === 0 && e.treeAdapter.adoptAttributes(e.openElements.items[0], t.attrs);
}
function Si(e, t) {
	let n = e.openElements.tryPeekProperlyNestedBodyElement();
	n && e.openElements.tmplCount === 0 && (e.framesetOk = !1, e.treeAdapter.adoptAttributes(n, t.attrs));
}
function Ci(e, t) {
	let n = e.openElements.tryPeekProperlyNestedBodyElement();
	e.framesetOk && n && (e.treeAdapter.detachNode(n), e.openElements.popAllUpToHtmlElement(), e._insertElement(t, F.HTML), e.insertionMode = V.IN_FRAMESET);
}
function wi(e, t) {
	e.openElements.hasInButtonScope(L.P) && e._closePElement(), e._insertElement(t, F.HTML);
}
function Ti(e, t) {
	e.openElements.hasInButtonScope(L.P) && e._closePElement(), e.openElements.currentTagId !== void 0 && In.has(e.openElements.currentTagId) && e.openElements.pop(), e._insertElement(t, F.HTML);
}
function Ei(e, t) {
	e.openElements.hasInButtonScope(L.P) && e._closePElement(), e._insertElement(t, F.HTML), e.skipNextNewLine = !0, e.framesetOk = !1;
}
function Di(e, t) {
	let n = e.openElements.tmplCount > 0;
	(!e.formElement || n) && (e.openElements.hasInButtonScope(L.P) && e._closePElement(), e._insertElement(t, F.HTML), n || (e.formElement = e.openElements.current));
}
function Oi(e, t) {
	e.framesetOk = !1;
	let n = t.tagID;
	for (let t = e.openElements.stackTop; t >= 0; t--) {
		let r = e.openElements.tagIDs[t];
		if (n === L.LI && r === L.LI || (n === L.DD || n === L.DT) && (r === L.DD || r === L.DT)) {
			e.openElements.generateImpliedEndTagsWithExclusion(r), e.openElements.popUntilTagNamePopped(r);
			break;
		}
		if (r !== L.ADDRESS && r !== L.DIV && r !== L.P && e._isSpecialElement(e.openElements.items[t], r)) break;
	}
	e.openElements.hasInButtonScope(L.P) && e._closePElement(), e._insertElement(t, F.HTML);
}
function ki(e, t) {
	e.openElements.hasInButtonScope(L.P) && e._closePElement(), e._insertElement(t, F.HTML), e.tokenizer.state = B.PLAINTEXT;
}
function Ai(e, t) {
	e.openElements.hasInScope(L.BUTTON) && (e.openElements.generateImpliedEndTags(), e.openElements.popUntilTagNamePopped(L.BUTTON)), e._reconstructActiveFormattingElements(), e._insertElement(t, F.HTML), e.framesetOk = !1;
}
function ji(e, t) {
	let n = e.activeFormattingElements.getElementEntryInScopeWithTagName(I.A);
	n && (Yr(e, t), e.openElements.remove(n.element), e.activeFormattingElements.removeEntry(n)), e._reconstructActiveFormattingElements(), e._insertElement(t, F.HTML), e.activeFormattingElements.pushElement(e.openElements.current, t);
}
function Mi(e, t) {
	e._reconstructActiveFormattingElements(), e._insertElement(t, F.HTML), e.activeFormattingElements.pushElement(e.openElements.current, t);
}
function Ni(e, t) {
	e._reconstructActiveFormattingElements(), e.openElements.hasInScope(L.NOBR) && (Yr(e, t), e._reconstructActiveFormattingElements()), e._insertElement(t, F.HTML), e.activeFormattingElements.pushElement(e.openElements.current, t);
}
function Pi(e, t) {
	e._reconstructActiveFormattingElements(), e._insertElement(t, F.HTML), e.activeFormattingElements.insertMarker(), e.framesetOk = !1;
}
function Fi(e, t) {
	e.treeAdapter.getDocumentMode(e.document) !== Mn.QUIRKS && e.openElements.hasInButtonScope(L.P) && e._closePElement(), e._insertElement(t, F.HTML), e.framesetOk = !1, e.insertionMode = V.IN_TABLE;
}
function Ii(e, t) {
	e._reconstructActiveFormattingElements(), e._appendElement(t, F.HTML), e.framesetOk = !1, t.ackSelfClosing = !0;
}
function Li(e) {
	let t = It(e, jn.TYPE);
	return t != null && t.toLowerCase() === Ir;
}
function Ri(e, t) {
	e._reconstructActiveFormattingElements(), e._appendElement(t, F.HTML), Li(t) || (e.framesetOk = !1), t.ackSelfClosing = !0;
}
function zi(e, t) {
	e._appendElement(t, F.HTML), t.ackSelfClosing = !0;
}
function Bi(e, t) {
	e.openElements.hasInButtonScope(L.P) && e._closePElement(), e._appendElement(t, F.HTML), e.framesetOk = !1, t.ackSelfClosing = !0;
}
function Vi(e, t) {
	t.tagName = I.IMG, t.tagID = L.IMG, Ii(e, t);
}
function Hi(e, t) {
	e._insertElement(t, F.HTML), e.skipNextNewLine = !0, e.tokenizer.state = B.RCDATA, e.originalInsertionMode = e.insertionMode, e.framesetOk = !1, e.insertionMode = V.TEXT;
}
function Ui(e, t) {
	e.openElements.hasInButtonScope(L.P) && e._closePElement(), e._reconstructActiveFormattingElements(), e.framesetOk = !1, e._switchToTextParsing(t, B.RAWTEXT);
}
function Wi(e, t) {
	e.framesetOk = !1, e._switchToTextParsing(t, B.RAWTEXT);
}
function Gi(e, t) {
	e._switchToTextParsing(t, B.RAWTEXT);
}
function Ki(e, t) {
	e._reconstructActiveFormattingElements(), e._insertElement(t, F.HTML), e.framesetOk = !1, e.insertionMode = e.insertionMode === V.IN_TABLE || e.insertionMode === V.IN_CAPTION || e.insertionMode === V.IN_TABLE_BODY || e.insertionMode === V.IN_ROW || e.insertionMode === V.IN_CELL ? V.IN_SELECT_IN_TABLE : V.IN_SELECT;
}
function qi(e, t) {
	e.openElements.currentTagId === L.OPTION && e.openElements.pop(), e._reconstructActiveFormattingElements(), e._insertElement(t, F.HTML);
}
function Ji(e, t) {
	e.openElements.hasInScope(L.RUBY) && e.openElements.generateImpliedEndTags(), e._insertElement(t, F.HTML);
}
function Yi(e, t) {
	e.openElements.hasInScope(L.RUBY) && e.openElements.generateImpliedEndTagsWithExclusion(L.RTC), e._insertElement(t, F.HTML);
}
function Xi(e, t) {
	e._reconstructActiveFormattingElements(), kr(t), jr(t), t.selfClosing ? e._appendElement(t, F.MATHML) : e._insertElement(t, F.MATHML), t.ackSelfClosing = !0;
}
function Zi(e, t) {
	e._reconstructActiveFormattingElements(), Ar(t), jr(t), t.selfClosing ? e._appendElement(t, F.SVG) : e._insertElement(t, F.SVG), t.ackSelfClosing = !0;
}
function Qi(e, t) {
	e._reconstructActiveFormattingElements(), e._insertElement(t, F.HTML);
}
function H(e, t) {
	switch (t.tagID) {
		case L.I:
		case L.S:
		case L.B:
		case L.U:
		case L.EM:
		case L.TT:
		case L.BIG:
		case L.CODE:
		case L.FONT:
		case L.SMALL:
		case L.STRIKE:
		case L.STRONG:
			Mi(e, t);
			break;
		case L.A:
			ji(e, t);
			break;
		case L.H1:
		case L.H2:
		case L.H3:
		case L.H4:
		case L.H5:
		case L.H6:
			Ti(e, t);
			break;
		case L.P:
		case L.DL:
		case L.OL:
		case L.UL:
		case L.DIV:
		case L.DIR:
		case L.NAV:
		case L.MAIN:
		case L.MENU:
		case L.ASIDE:
		case L.CENTER:
		case L.FIGURE:
		case L.FOOTER:
		case L.HEADER:
		case L.HGROUP:
		case L.DIALOG:
		case L.DETAILS:
		case L.ADDRESS:
		case L.ARTICLE:
		case L.SEARCH:
		case L.SECTION:
		case L.SUMMARY:
		case L.FIELDSET:
		case L.BLOCKQUOTE:
		case L.FIGCAPTION:
			wi(e, t);
			break;
		case L.LI:
		case L.DD:
		case L.DT:
			Oi(e, t);
			break;
		case L.BR:
		case L.IMG:
		case L.WBR:
		case L.AREA:
		case L.EMBED:
		case L.KEYGEN:
			Ii(e, t);
			break;
		case L.HR:
			Bi(e, t);
			break;
		case L.RB:
		case L.RTC:
			Ji(e, t);
			break;
		case L.RT:
		case L.RP:
			Yi(e, t);
			break;
		case L.PRE:
		case L.LISTING:
			Ei(e, t);
			break;
		case L.XMP:
			Ui(e, t);
			break;
		case L.SVG:
			Zi(e, t);
			break;
		case L.HTML:
			xi(e, t);
			break;
		case L.BASE:
		case L.LINK:
		case L.META:
		case L.STYLE:
		case L.TITLE:
		case L.SCRIPT:
		case L.BGSOUND:
		case L.BASEFONT:
		case L.TEMPLATE:
			ci(e, t);
			break;
		case L.BODY:
			Si(e, t);
			break;
		case L.FORM:
			Di(e, t);
			break;
		case L.NOBR:
			Ni(e, t);
			break;
		case L.MATH:
			Xi(e, t);
			break;
		case L.TABLE:
			Fi(e, t);
			break;
		case L.INPUT:
			Ri(e, t);
			break;
		case L.PARAM:
		case L.TRACK:
		case L.SOURCE:
			zi(e, t);
			break;
		case L.IMAGE:
			Vi(e, t);
			break;
		case L.BUTTON:
			Ai(e, t);
			break;
		case L.APPLET:
		case L.OBJECT:
		case L.MARQUEE:
			Pi(e, t);
			break;
		case L.IFRAME:
			Wi(e, t);
			break;
		case L.SELECT:
			Ki(e, t);
			break;
		case L.OPTION:
		case L.OPTGROUP:
			qi(e, t);
			break;
		case L.NOEMBED:
		case L.NOFRAMES:
			Gi(e, t);
			break;
		case L.FRAMESET:
			Ci(e, t);
			break;
		case L.TEXTAREA:
			Hi(e, t);
			break;
		case L.NOSCRIPT:
			e.options.scriptingEnabled ? Gi(e, t) : Qi(e, t);
			break;
		case L.PLAINTEXT:
			ki(e, t);
			break;
		case L.COL:
		case L.TH:
		case L.TD:
		case L.TR:
		case L.HEAD:
		case L.FRAME:
		case L.TBODY:
		case L.TFOOT:
		case L.THEAD:
		case L.CAPTION:
		case L.COLGROUP: break;
		default: Qi(e, t);
	}
}
function $i(e, t) {
	if (e.openElements.hasInScope(L.BODY) && (e.insertionMode = V.AFTER_BODY, e.options.sourceCodeLocationInfo)) {
		let n = e.openElements.tryPeekProperlyNestedBodyElement();
		n && e._setEndLocation(n, t);
	}
}
function ea(e, t) {
	e.openElements.hasInScope(L.BODY) && (e.insertionMode = V.AFTER_BODY, Ya(e, t));
}
function ta(e, t) {
	let n = t.tagID;
	e.openElements.hasInScope(n) && (e.openElements.generateImpliedEndTags(), e.openElements.popUntilTagNamePopped(n));
}
function na(e) {
	let t = e.openElements.tmplCount > 0, { formElement: n } = e;
	t || (e.formElement = null), (n || t) && e.openElements.hasInScope(L.FORM) && (e.openElements.generateImpliedEndTags(), t ? e.openElements.popUntilTagNamePopped(L.FORM) : n && e.openElements.remove(n));
}
function ra(e) {
	e.openElements.hasInButtonScope(L.P) || e._insertFakeElement(I.P, L.P), e._closePElement();
}
function ia(e) {
	e.openElements.hasInListItemScope(L.LI) && (e.openElements.generateImpliedEndTagsWithExclusion(L.LI), e.openElements.popUntilTagNamePopped(L.LI));
}
function aa(e, t) {
	let n = t.tagID;
	e.openElements.hasInScope(n) && (e.openElements.generateImpliedEndTagsWithExclusion(n), e.openElements.popUntilTagNamePopped(n));
}
function oa(e) {
	e.openElements.hasNumberedHeaderInScope() && (e.openElements.generateImpliedEndTags(), e.openElements.popUntilNumberedHeaderPopped());
}
function sa(e, t) {
	let n = t.tagID;
	e.openElements.hasInScope(n) && (e.openElements.generateImpliedEndTags(), e.openElements.popUntilTagNamePopped(n), e.activeFormattingElements.clearToLastMarker());
}
function ca(e) {
	e._reconstructActiveFormattingElements(), e._insertFakeElement(I.BR, L.BR), e.openElements.pop(), e.framesetOk = !1;
}
function la(e, t) {
	let n = t.tagName, r = t.tagID;
	for (let t = e.openElements.stackTop; t > 0; t--) {
		let i = e.openElements.items[t], a = e.openElements.tagIDs[t];
		if (r === a && e.treeAdapter.getNamespaceURI(i) === F.HTML && (r !== L.UNKNOWN || e.treeAdapter.getTagName(i) === n)) {
			e.openElements.generateImpliedEndTagsWithExclusion(r), e.openElements.stackTop >= t && e.openElements.shortenToLength(t);
			break;
		}
		if (e._isSpecialElement(i, a)) break;
	}
}
function ua(e, t) {
	switch (t.tagID) {
		case L.A:
		case L.B:
		case L.I:
		case L.S:
		case L.U:
		case L.EM:
		case L.TT:
		case L.BIG:
		case L.CODE:
		case L.FONT:
		case L.NOBR:
		case L.SMALL:
		case L.STRIKE:
		case L.STRONG:
			Yr(e, t);
			break;
		case L.P:
			ra(e);
			break;
		case L.DL:
		case L.UL:
		case L.OL:
		case L.DIR:
		case L.DIV:
		case L.NAV:
		case L.PRE:
		case L.MAIN:
		case L.MENU:
		case L.ASIDE:
		case L.BUTTON:
		case L.CENTER:
		case L.FIGURE:
		case L.FOOTER:
		case L.HEADER:
		case L.HGROUP:
		case L.DIALOG:
		case L.ADDRESS:
		case L.ARTICLE:
		case L.DETAILS:
		case L.SEARCH:
		case L.SECTION:
		case L.SUMMARY:
		case L.LISTING:
		case L.FIELDSET:
		case L.BLOCKQUOTE:
		case L.FIGCAPTION:
			ta(e, t);
			break;
		case L.LI:
			ia(e);
			break;
		case L.DD:
		case L.DT:
			aa(e, t);
			break;
		case L.H1:
		case L.H2:
		case L.H3:
		case L.H4:
		case L.H5:
		case L.H6:
			oa(e);
			break;
		case L.BR:
			ca(e);
			break;
		case L.BODY:
			$i(e, t);
			break;
		case L.HTML:
			ea(e, t);
			break;
		case L.FORM:
			na(e);
			break;
		case L.APPLET:
		case L.OBJECT:
		case L.MARQUEE:
			sa(e, t);
			break;
		case L.TEMPLATE:
			ui(e, t);
			break;
		default: la(e, t);
	}
}
function da(e, t) {
	e.tmplInsertionModeStack.length > 0 ? qa(e, t) : $r(e, t);
}
function fa(e, t) {
	var n;
	t.tagID === L.SCRIPT && ((n = e.scriptHandler) == null || n.call(e, e.openElements.current)), e.openElements.pop(), e.insertionMode = e.originalInsertionMode;
}
function pa(e, t) {
	e._err(t, M.eofInElementThatCanContainOnlyText), e.openElements.pop(), e.insertionMode = e.originalInsertionMode, e.onEof(t);
}
function ma(e, t) {
	if (e.openElements.currentTagId !== void 0 && Br.has(e.openElements.currentTagId)) switch (e.pendingCharacterTokens.length = 0, e.hasNonWhitespacePendingCharacterToken = !1, e.originalInsertionMode = e.insertionMode, e.insertionMode = V.IN_TABLE_TEXT, t.type) {
		case N.CHARACTER:
			Da(e, t);
			break;
		case N.WHITESPACE_CHARACTER: Ea(e, t);
	}
	else Ta(e, t);
}
function ha(e, t) {
	e.openElements.clearBackToTableContext(), e.activeFormattingElements.insertMarker(), e._insertElement(t, F.HTML), e.insertionMode = V.IN_CAPTION;
}
function ga(e, t) {
	e.openElements.clearBackToTableContext(), e._insertElement(t, F.HTML), e.insertionMode = V.IN_COLUMN_GROUP;
}
function _a(e, t) {
	e.openElements.clearBackToTableContext(), e._insertFakeElement(I.COLGROUP, L.COLGROUP), e.insertionMode = V.IN_COLUMN_GROUP, Ma(e, t);
}
function va(e, t) {
	e.openElements.clearBackToTableContext(), e._insertElement(t, F.HTML), e.insertionMode = V.IN_TABLE_BODY;
}
function ya(e, t) {
	e.openElements.clearBackToTableContext(), e._insertFakeElement(I.TBODY, L.TBODY), e.insertionMode = V.IN_TABLE_BODY, Fa(e, t);
}
function ba(e, t) {
	e.openElements.hasInTableScope(L.TABLE) && (e.openElements.popUntilTagNamePopped(L.TABLE), e._resetInsertionMode(), e._processStartTag(t));
}
function xa(e, t) {
	Li(t) ? e._appendElement(t, F.HTML) : Ta(e, t), t.ackSelfClosing = !0;
}
function Sa(e, t) {
	!e.formElement && e.openElements.tmplCount === 0 && (e._insertElement(t, F.HTML), e.formElement = e.openElements.current, e.openElements.pop());
}
function Ca(e, t) {
	switch (t.tagID) {
		case L.TD:
		case L.TH:
		case L.TR:
			ya(e, t);
			break;
		case L.STYLE:
		case L.SCRIPT:
		case L.TEMPLATE:
			ci(e, t);
			break;
		case L.COL:
			_a(e, t);
			break;
		case L.FORM:
			Sa(e, t);
			break;
		case L.TABLE:
			ba(e, t);
			break;
		case L.TBODY:
		case L.TFOOT:
		case L.THEAD:
			va(e, t);
			break;
		case L.INPUT:
			xa(e, t);
			break;
		case L.CAPTION:
			ha(e, t);
			break;
		case L.COLGROUP:
			ga(e, t);
			break;
		default: Ta(e, t);
	}
}
function wa(e, t) {
	switch (t.tagID) {
		case L.TABLE:
			e.openElements.hasInTableScope(L.TABLE) && (e.openElements.popUntilTagNamePopped(L.TABLE), e._resetInsertionMode());
			break;
		case L.TEMPLATE:
			ui(e, t);
			break;
		case L.BODY:
		case L.CAPTION:
		case L.COL:
		case L.COLGROUP:
		case L.HTML:
		case L.TBODY:
		case L.TD:
		case L.TFOOT:
		case L.TH:
		case L.THEAD:
		case L.TR: break;
		default: Ta(e, t);
	}
}
function Ta(e, t) {
	let n = e.fosterParentingEnabled;
	e.fosterParentingEnabled = !0, vi(e, t), e.fosterParentingEnabled = n;
}
function Ea(e, t) {
	e.pendingCharacterTokens.push(t);
}
function Da(e, t) {
	e.pendingCharacterTokens.push(t), e.hasNonWhitespacePendingCharacterToken = !0;
}
function Oa(e, t) {
	let n = 0;
	if (e.hasNonWhitespacePendingCharacterToken) for (; n < e.pendingCharacterTokens.length; n++) Ta(e, e.pendingCharacterTokens[n]);
	else for (; n < e.pendingCharacterTokens.length; n++) e._insertCharacters(e.pendingCharacterTokens[n]);
	e.insertionMode = e.originalInsertionMode, e._processToken(t);
}
var ka = /* @__PURE__ */ new Set([
	L.CAPTION,
	L.COL,
	L.COLGROUP,
	L.TBODY,
	L.TD,
	L.TFOOT,
	L.TH,
	L.THEAD,
	L.TR
]);
function Aa(e, t) {
	let n = t.tagID;
	ka.has(n) ? e.openElements.hasInTableScope(L.CAPTION) && (e.openElements.generateImpliedEndTags(), e.openElements.popUntilTagNamePopped(L.CAPTION), e.activeFormattingElements.clearToLastMarker(), e.insertionMode = V.IN_TABLE, Ca(e, t)) : H(e, t);
}
function ja(e, t) {
	let n = t.tagID;
	switch (n) {
		case L.CAPTION:
		case L.TABLE:
			e.openElements.hasInTableScope(L.CAPTION) && (e.openElements.generateImpliedEndTags(), e.openElements.popUntilTagNamePopped(L.CAPTION), e.activeFormattingElements.clearToLastMarker(), e.insertionMode = V.IN_TABLE, n === L.TABLE && wa(e, t));
			break;
		case L.BODY:
		case L.COL:
		case L.COLGROUP:
		case L.HTML:
		case L.TBODY:
		case L.TD:
		case L.TFOOT:
		case L.TH:
		case L.THEAD:
		case L.TR: break;
		default: ua(e, t);
	}
}
function Ma(e, t) {
	switch (t.tagID) {
		case L.HTML:
			H(e, t);
			break;
		case L.COL:
			e._appendElement(t, F.HTML), t.ackSelfClosing = !0;
			break;
		case L.TEMPLATE:
			ci(e, t);
			break;
		default: Pa(e, t);
	}
}
function Na(e, t) {
	switch (t.tagID) {
		case L.COLGROUP:
			e.openElements.currentTagId === L.COLGROUP && (e.openElements.pop(), e.insertionMode = V.IN_TABLE);
			break;
		case L.TEMPLATE:
			ui(e, t);
			break;
		case L.COL: break;
		default: Pa(e, t);
	}
}
function Pa(e, t) {
	e.openElements.currentTagId === L.COLGROUP && (e.openElements.pop(), e.insertionMode = V.IN_TABLE, e._processToken(t));
}
function Fa(e, t) {
	switch (t.tagID) {
		case L.TR:
			e.openElements.clearBackToTableBodyContext(), e._insertElement(t, F.HTML), e.insertionMode = V.IN_ROW;
			break;
		case L.TH:
		case L.TD:
			e.openElements.clearBackToTableBodyContext(), e._insertFakeElement(I.TR, L.TR), e.insertionMode = V.IN_ROW, La(e, t);
			break;
		case L.CAPTION:
		case L.COL:
		case L.COLGROUP:
		case L.TBODY:
		case L.TFOOT:
		case L.THEAD:
			e.openElements.hasTableBodyContextInTableScope() && (e.openElements.clearBackToTableBodyContext(), e.openElements.pop(), e.insertionMode = V.IN_TABLE, Ca(e, t));
			break;
		default: Ca(e, t);
	}
}
function Ia(e, t) {
	let n = t.tagID;
	switch (t.tagID) {
		case L.TBODY:
		case L.TFOOT:
		case L.THEAD:
			e.openElements.hasInTableScope(n) && (e.openElements.clearBackToTableBodyContext(), e.openElements.pop(), e.insertionMode = V.IN_TABLE);
			break;
		case L.TABLE:
			e.openElements.hasTableBodyContextInTableScope() && (e.openElements.clearBackToTableBodyContext(), e.openElements.pop(), e.insertionMode = V.IN_TABLE, wa(e, t));
			break;
		case L.BODY:
		case L.CAPTION:
		case L.COL:
		case L.COLGROUP:
		case L.HTML:
		case L.TD:
		case L.TH:
		case L.TR: break;
		default: wa(e, t);
	}
}
function La(e, t) {
	switch (t.tagID) {
		case L.TH:
		case L.TD:
			e.openElements.clearBackToTableRowContext(), e._insertElement(t, F.HTML), e.insertionMode = V.IN_CELL, e.activeFormattingElements.insertMarker();
			break;
		case L.CAPTION:
		case L.COL:
		case L.COLGROUP:
		case L.TBODY:
		case L.TFOOT:
		case L.THEAD:
		case L.TR:
			e.openElements.hasInTableScope(L.TR) && (e.openElements.clearBackToTableRowContext(), e.openElements.pop(), e.insertionMode = V.IN_TABLE_BODY, Fa(e, t));
			break;
		default: Ca(e, t);
	}
}
function Ra(e, t) {
	switch (t.tagID) {
		case L.TR:
			e.openElements.hasInTableScope(L.TR) && (e.openElements.clearBackToTableRowContext(), e.openElements.pop(), e.insertionMode = V.IN_TABLE_BODY);
			break;
		case L.TABLE:
			e.openElements.hasInTableScope(L.TR) && (e.openElements.clearBackToTableRowContext(), e.openElements.pop(), e.insertionMode = V.IN_TABLE_BODY, Ia(e, t));
			break;
		case L.TBODY:
		case L.TFOOT:
		case L.THEAD:
			(e.openElements.hasInTableScope(t.tagID) || e.openElements.hasInTableScope(L.TR)) && (e.openElements.clearBackToTableRowContext(), e.openElements.pop(), e.insertionMode = V.IN_TABLE_BODY, Ia(e, t));
			break;
		case L.BODY:
		case L.CAPTION:
		case L.COL:
		case L.COLGROUP:
		case L.HTML:
		case L.TD:
		case L.TH: break;
		default: wa(e, t);
	}
}
function za(e, t) {
	let n = t.tagID;
	ka.has(n) ? (e.openElements.hasInTableScope(L.TD) || e.openElements.hasInTableScope(L.TH)) && (e._closeTableCell(), La(e, t)) : H(e, t);
}
function Ba(e, t) {
	let n = t.tagID;
	switch (n) {
		case L.TD:
		case L.TH:
			e.openElements.hasInTableScope(n) && (e.openElements.generateImpliedEndTags(), e.openElements.popUntilTagNamePopped(n), e.activeFormattingElements.clearToLastMarker(), e.insertionMode = V.IN_ROW);
			break;
		case L.TABLE:
		case L.TBODY:
		case L.TFOOT:
		case L.THEAD:
		case L.TR:
			e.openElements.hasInTableScope(n) && (e._closeTableCell(), Ra(e, t));
			break;
		case L.BODY:
		case L.CAPTION:
		case L.COL:
		case L.COLGROUP:
		case L.HTML: break;
		default: ua(e, t);
	}
}
function Va(e, t) {
	switch (t.tagID) {
		case L.HTML:
			H(e, t);
			break;
		case L.OPTION:
			e.openElements.currentTagId === L.OPTION && e.openElements.pop(), e._insertElement(t, F.HTML);
			break;
		case L.OPTGROUP:
			e.openElements.currentTagId === L.OPTION && e.openElements.pop(), e.openElements.currentTagId === L.OPTGROUP && e.openElements.pop(), e._insertElement(t, F.HTML);
			break;
		case L.HR:
			e.openElements.currentTagId === L.OPTION && e.openElements.pop(), e.openElements.currentTagId === L.OPTGROUP && e.openElements.pop(), e._appendElement(t, F.HTML), t.ackSelfClosing = !0;
			break;
		case L.INPUT:
		case L.KEYGEN:
		case L.TEXTAREA:
		case L.SELECT:
			e.openElements.hasInSelectScope(L.SELECT) && (e.openElements.popUntilTagNamePopped(L.SELECT), e._resetInsertionMode(), t.tagID !== L.SELECT && e._processStartTag(t));
			break;
		case L.SCRIPT:
		case L.TEMPLATE: ci(e, t);
	}
}
function Ha(e, t) {
	switch (t.tagID) {
		case L.OPTGROUP:
			e.openElements.stackTop > 0 && e.openElements.currentTagId === L.OPTION && e.openElements.tagIDs[e.openElements.stackTop - 1] === L.OPTGROUP && e.openElements.pop(), e.openElements.currentTagId === L.OPTGROUP && e.openElements.pop();
			break;
		case L.OPTION:
			e.openElements.currentTagId === L.OPTION && e.openElements.pop();
			break;
		case L.SELECT:
			e.openElements.hasInSelectScope(L.SELECT) && (e.openElements.popUntilTagNamePopped(L.SELECT), e._resetInsertionMode());
			break;
		case L.TEMPLATE: ui(e, t);
	}
}
function Ua(e, t) {
	let n = t.tagID;
	n === L.CAPTION || n === L.TABLE || n === L.TBODY || n === L.TFOOT || n === L.THEAD || n === L.TR || n === L.TD || n === L.TH ? (e.openElements.popUntilTagNamePopped(L.SELECT), e._resetInsertionMode(), e._processStartTag(t)) : Va(e, t);
}
function Wa(e, t) {
	let n = t.tagID;
	n === L.CAPTION || n === L.TABLE || n === L.TBODY || n === L.TFOOT || n === L.THEAD || n === L.TR || n === L.TD || n === L.TH ? e.openElements.hasInTableScope(n) && (e.openElements.popUntilTagNamePopped(L.SELECT), e._resetInsertionMode(), e.onEndTag(t)) : Ha(e, t);
}
function Ga(e, t) {
	switch (t.tagID) {
		case L.BASE:
		case L.BASEFONT:
		case L.BGSOUND:
		case L.LINK:
		case L.META:
		case L.NOFRAMES:
		case L.SCRIPT:
		case L.STYLE:
		case L.TEMPLATE:
		case L.TITLE:
			ci(e, t);
			break;
		case L.CAPTION:
		case L.COLGROUP:
		case L.TBODY:
		case L.TFOOT:
		case L.THEAD:
			e.tmplInsertionModeStack[0] = V.IN_TABLE, e.insertionMode = V.IN_TABLE, Ca(e, t);
			break;
		case L.COL:
			e.tmplInsertionModeStack[0] = V.IN_COLUMN_GROUP, e.insertionMode = V.IN_COLUMN_GROUP, Ma(e, t);
			break;
		case L.TR:
			e.tmplInsertionModeStack[0] = V.IN_TABLE_BODY, e.insertionMode = V.IN_TABLE_BODY, Fa(e, t);
			break;
		case L.TD:
		case L.TH:
			e.tmplInsertionModeStack[0] = V.IN_ROW, e.insertionMode = V.IN_ROW, La(e, t);
			break;
		default: e.tmplInsertionModeStack[0] = V.IN_BODY, e.insertionMode = V.IN_BODY, H(e, t);
	}
}
function Ka(e, t) {
	t.tagID === L.TEMPLATE && ui(e, t);
}
function qa(e, t) {
	e.openElements.tmplCount > 0 ? (e.openElements.popUntilTagNamePopped(L.TEMPLATE), e.activeFormattingElements.clearToLastMarker(), e.tmplInsertionModeStack.shift(), e._resetInsertionMode(), e.onEof(t)) : $r(e, t);
}
function Ja(e, t) {
	t.tagID === L.HTML ? H(e, t) : Xa(e, t);
}
function Ya(e, t) {
	if (t.tagID === L.HTML) {
		if (e.fragmentContext || (e.insertionMode = V.AFTER_AFTER_BODY), e.options.sourceCodeLocationInfo && e.openElements.tagIDs[0] === L.HTML) {
			e._setEndLocation(e.openElements.items[0], t);
			let n = e.openElements.items[1];
			n && !e.treeAdapter.getNodeSourceCodeLocation(n)?.endTag && e._setEndLocation(n, t);
		}
	} else Xa(e, t);
}
function Xa(e, t) {
	e.insertionMode = V.IN_BODY, vi(e, t);
}
function Za(e, t) {
	switch (t.tagID) {
		case L.HTML:
			H(e, t);
			break;
		case L.FRAMESET:
			e._insertElement(t, F.HTML);
			break;
		case L.FRAME:
			e._appendElement(t, F.HTML), t.ackSelfClosing = !0;
			break;
		case L.NOFRAMES: ci(e, t);
	}
}
function Qa(e, t) {
	t.tagID === L.FRAMESET && !e.openElements.isRootHtmlElementCurrent() && (e.openElements.pop(), !e.fragmentContext && e.openElements.currentTagId !== L.FRAMESET && (e.insertionMode = V.AFTER_FRAMESET));
}
function $a(e, t) {
	switch (t.tagID) {
		case L.HTML:
			H(e, t);
			break;
		case L.NOFRAMES: ci(e, t);
	}
}
function eo(e, t) {
	t.tagID === L.HTML && (e.insertionMode = V.AFTER_AFTER_FRAMESET);
}
function to(e, t) {
	t.tagID === L.HTML ? H(e, t) : no(e, t);
}
function no(e, t) {
	e.insertionMode = V.IN_BODY, vi(e, t);
}
function ro(e, t) {
	switch (t.tagID) {
		case L.HTML:
			H(e, t);
			break;
		case L.NOFRAMES: ci(e, t);
	}
}
function io(e, t) {
	t.chars = "�", e._insertCharacters(t);
}
function ao(e, t) {
	e._insertCharacters(t), e.framesetOk = !1;
}
function oo(e) {
	for (; e.treeAdapter.getNamespaceURI(e.openElements.current) !== F.HTML && e.openElements.currentTagId !== void 0 && !e._isIntegrationPoint(e.openElements.currentTagId, e.openElements.current);) e.openElements.pop();
}
function so(e, t) {
	if (Or(t)) oo(e), e._startTagOutsideForeignContent(t);
	else {
		let n = e._getAdjustedCurrentElement(), r = e.treeAdapter.getNamespaceURI(n);
		r === F.MATHML ? kr(t) : r === F.SVG && (Mr(t), Ar(t)), jr(t), t.selfClosing ? e._appendElement(t, r) : e._insertElement(t, r), t.ackSelfClosing = !0;
	}
}
function co(e, t) {
	if (t.tagID === L.P || t.tagID === L.BR) oo(e), e._endTagOutsideForeignContent(t);
	else for (let n = e.openElements.stackTop; n > 0; n--) {
		let r = e.openElements.items[n];
		if (e.treeAdapter.getNamespaceURI(r) === F.HTML) {
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
I.AREA, I.BASE, I.BASEFONT, I.BGSOUND, I.BR, I.COL, I.EMBED, I.FRAME, I.HR, I.IMG, I.INPUT, I.KEYGEN, I.LINK, I.META, I.PARAM, I.SOURCE, I.TRACK, I.WBR;
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/unist-util-position/5.0.0/1d79e0203c9b065483fc8ef1df203f43199c23ab75fff695b368215f52aea4a0/node_modules/unist-util-position/lib/index.js
var lo = fo("end"), uo = fo("start");
function fo(e) {
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
function po(e) {
	let t = uo(e), n = lo(e);
	if (t && n) return {
		start: t,
		end: n
	};
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/hast-util-raw/9.2.0/1063f4c60c2357ac12ae00e730ba1824e643d0e748f42b622d364123328453ed/node_modules/hast-util-raw/lib/index.js
var mo = /<(\/?)(iframe|noembed|noframes|plaintext|script|style|textarea|title|xmp)(?=[\t\n\f\r />])/gi, ho = /* @__PURE__ */ new Set([
	"mdxFlowExpression",
	"mdxJsxFlowElement",
	"mdxJsxTextElement",
	"mdxTextExpression",
	"mdxjsEsm"
]), go = {
	sourceCodeLocationInfo: !0,
	scriptingEnabled: !1
};
function _o(e, t) {
	let n = jo(e), r = _("type", {
		handlers: {
			root: yo,
			element: bo,
			text: xo,
			comment: wo,
			doctype: So,
			raw: To
		},
		unknown: Eo
	}), i = {
		parser: n ? new Hr(go) : Hr.getFragmentParser(void 0, go),
		handle(e) {
			r(e, i);
		},
		stitches: !1,
		options: t || {}
	};
	r(e, i), Do(i, uo());
	let a = st(n ? i.parser.document : i.parser.getFragment(), { file: i.options.file });
	return i.stitches && x(a, "comment", function(e, t, n) {
		let r = e;
		if (r.value.stitch && n && t !== void 0) {
			let e = n.children;
			return e[t] = r.value.stitch, t;
		}
	}), a.type === "root" && a.children.length === 1 && a.children[0].type === e.type ? a.children[0] : a;
}
function vo(e, t) {
	let n = -1;
	/* istanbul ignore else - invalid nodes, see rehypejs/rehype-raw#7. */
	if (e) for (; ++n < e.length;) t.handle(e[n]);
}
function yo(e, t) {
	vo(e.children, t);
}
function bo(e, t) {
	ko(e, t), vo(e.children, t), Ao(e, t);
}
function xo(e, t) {
	t.parser.tokenizer.state > 4 && (t.parser.tokenizer.state = 0), Do(t, uo(e));
	let n = Mo(e), { value: r } = e, i = 0;
	for (; i < r.length;) {
		let e = Po(r.charCodeAt(i)), a = i + 1;
		for (; a < r.length && Po(r.charCodeAt(a)) === e;) a++;
		let o = {
			type: e,
			chars: r.slice(i, a),
			location: n
		};
		t.parser.currentToken = o, t.parser._processToken(t.parser.currentToken), i = a;
	}
}
function So(e, t) {
	let n = {
		type: N.DOCTYPE,
		name: "html",
		forceQuirks: !1,
		publicId: "",
		systemId: "",
		location: Mo(e)
	};
	Do(t, uo(e)), t.parser.currentToken = n, t.parser._processToken(t.parser.currentToken);
}
function Co(e, t) {
	t.stitches = !0;
	let n = No(e);
	"children" in e && "children" in n && (n.children = _o({
		type: "root",
		children: e.children
	}, t.options).children), wo({
		type: "comment",
		value: { stitch: n }
	}, t);
}
function wo(e, t) {
	let n = e.value, r = {
		type: N.COMMENT,
		data: n,
		location: Mo(e)
	};
	Do(t, uo(e)), t.parser.currentToken = r, t.parser._processToken(t.parser.currentToken);
}
function To(e, t) {
	t.parser.tokenizer.flushCharacters(), t.parser.tokenizer.preprocessor.html = "", t.parser.tokenizer.preprocessor.pos = -1, t.parser.tokenizer.preprocessor.lastGapPos = -2, t.parser.tokenizer.preprocessor.gapStack = [], t.parser.tokenizer.preprocessor.skipNextNewLine = !1, t.parser.tokenizer.preprocessor.lastChunkWritten = !1, t.parser.tokenizer.preprocessor.endOfChunkHit = !1, t.parser.tokenizer.preprocessor.isEol = !1, Oo(t, uo(e)), t.parser.tokenizer.write(t.options.tagfilter ? e.value.replace(mo, "&lt;$1$2") : e.value, !1), t.parser.tokenizer._runParsingLoop(), t.parser.tokenizer.state === 71 && (t.parser.tokenizer.preprocessor.pos = t.parser.tokenizer.preprocessor.html.length, t.parser.tokenizer.preprocessor.lastChunkWritten = !0, t.parser.tokenizer._stateCharacterReference(), t.parser.tokenizer.preprocessor.lastChunkWritten = !1, t.parser.tokenizer.preprocessor.endOfChunkHit = !1, t.parser.tokenizer.active = !0, t.parser.tokenizer._runParsingLoop());
}
function Eo(e, t) {
	let n = e;
	if (t.options.passThrough && t.options.passThrough.includes(n.type)) Co(n, t);
	else {
		let e = "";
		throw ho.has(n.type) && (e = ". It looks like you are using MDX nodes with `hast-util-raw` (or `rehype-raw`). If you use this because you are using remark or rehype plugins that inject `'html'` nodes, then please raise an issue with that plugin, as its a bad and slow idea. If you use this because you are using markdown syntax, then you have to configure this utility (or plugin) to pass through these nodes (see `passThrough` in docs), but you can also migrate to use the MDX syntax"), Error("Cannot compile `" + n.type + "` node" + e);
	}
}
function Do(e, t) {
	e.parser.tokenizer.flushCharacters(), Oo(e, t), e.parser.tokenizer.paused = !1, e.parser.tokenizer.inLoop = !1, e.parser.tokenizer.active = !1, e.parser.tokenizer.returnState = B.DATA, e.parser.tokenizer.charRefCode = -1, e.parser.tokenizer.consumedAfterSnapshot = -1, e.parser.tokenizer.currentLocation = null, e.parser.tokenizer.currentCharacterToken = null, e.parser.tokenizer.currentToken = null, e.parser.tokenizer.currentAttr = {
		name: "",
		value: ""
	};
}
function Oo(e, t) {
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
function ko(e, t) {
	if (t.parser.tokenizer.state === B.PLAINTEXT) return;
	let n = e.tagName.toLowerCase();
	Do(t, uo(e));
	let { current: r } = t.parser.openElements, i = r && "namespaceURI" in r ? r.namespaceURI : it.html;
	i === it.html && n === "svg" && (i = it.svg);
	let a = vt({
		...e,
		children: []
	}, { space: i === it.svg ? "svg" : "html" }), o = {
		type: N.START_TAG,
		tagName: n,
		tagID: Pn(n),
		selfClosing: !1,
		ackSelfClosing: !1,
		/* c8 ignore next */
		attrs: "attrs" in a ? a.attrs : [],
		location: Mo(e)
	};
	t.parser.currentToken = o, t.parser._processToken(t.parser.currentToken), t.parser.tokenizer.lastStartTagName = n;
}
function Ao(e, t) {
	let n = e.tagName.toLowerCase();
	if (!t.parser.tokenizer.inForeignNode && g.includes(n) || t.parser.tokenizer.state === B.PLAINTEXT) return;
	Do(t, lo(e));
	let r = {
		type: N.END_TAG,
		tagName: n,
		tagID: Pn(n),
		selfClosing: !1,
		ackSelfClosing: !1,
		attrs: [],
		location: Mo(e)
	};
	t.parser.currentToken = r, t.parser._processToken(t.parser.currentToken), n === t.parser.tokenizer.lastStartTagName && (t.parser.tokenizer.state === B.RCDATA || t.parser.tokenizer.state === B.RAWTEXT || t.parser.tokenizer.state === B.SCRIPT_DATA) && (t.parser.tokenizer.state = B.DATA);
}
function jo(e) {
	let t = e.type === "root" ? e.children[0] : e;
	return !!(t && (t.type === "doctype" || t.type === "element" && t.tagName.toLowerCase() === "html"));
}
function Mo(e) {
	let t = uo(e) || {
		line: void 0,
		column: void 0,
		offset: void 0
	}, n = lo(e) || {
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
function No(e) {
	return We("children" in e ? {
		...e,
		children: []
	} : e);
}
function Po(e) {
	return e === 9 || e === 10 || e === 12 || e === 13 || e === 32 ? N.WHITESPACE_CHARACTER : N.CHARACTER;
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/rehype-raw/7.0.0/734078c2f0a20ce62f58b63d79780f200745ac42ad9e1bcb734272bd58080def/node_modules/rehype-raw/lib/index.js
function Fo(e) {
	return function(t, n) {
		return _o(t, {
			...e,
			file: n
		});
	};
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/hast-util-sanitize/5.0.2/c3dc153ddecab75ab59b84200c9aac478b2239feca9ac0313d02705f62f81336/node_modules/hast-util-sanitize/lib/schema.js
var Io = [
	"ariaDescribedBy",
	"ariaLabel",
	"ariaLabelledBy"
], Lo = {
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
			...Io,
			"dataFootnoteBackref",
			"dataFootnoteRef",
			["className", "data-footnote-backref"],
			"href"
		],
		blockquote: ["cite"],
		code: [["className", /^language-./]],
		del: ["cite"],
		div: ["itemScope", "itemType"],
		dl: [...Io],
		h2: [["className", "sr-only"]],
		img: [
			...Io,
			"longDesc",
			"src"
		],
		input: [["disabled", !0], ["type", "checkbox"]],
		ins: ["cite"],
		li: [["className", "task-list-item"]],
		ol: [...Io, ["className", "contains-task-list"]],
		q: ["cite"],
		section: ["dataFootnotes", ["className", "footnotes"]],
		source: ["srcSet"],
		summary: [...Io],
		table: [...Io],
		ul: [...Io, ["className", "contains-task-list"]],
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
}, Ro = {}.hasOwnProperty;
function zo(e, t) {
	let n = {
		type: "root",
		children: []
	}, r = Bo({
		schema: t ? {
			...Lo,
			...t
		} : Lo,
		stack: []
	}, e);
	return r && (Array.isArray(r) ? r.length === 1 ? n = r[0] : n.children = r : n = r), n;
}
function Bo(e, t) {
	if (t && typeof t == "object") {
		let n = t;
		switch (typeof n.type == "string" ? n.type : "") {
			case "comment": return Vo(e, n);
			case "doctype": return Ho(e, n);
			case "element": return Uo(e, n);
			case "root": return Wo(e, n);
			case "text": return Go(e, n);
		}
	}
}
function Vo(e, t) {
	if (e.schema.allowComments) {
		let e = typeof t.value == "string" ? t.value : "", n = e.indexOf("-->"), r = {
			type: "comment",
			value: n < 0 ? e : e.slice(0, n)
		};
		return Qo(r, t), r;
	}
}
function Ho(e, t) {
	if (e.schema.allowDoctypes) {
		let e = { type: "doctype" };
		return Qo(e, t), e;
	}
}
function Uo(e, t) {
	let n = typeof t.tagName == "string" ? t.tagName : "";
	e.stack.push(n);
	let r = Ko(e, t.children), i = qo(e, t.properties);
	e.stack.pop();
	let a = !1;
	if (n && n !== "*" && (!e.schema.tagNames || e.schema.tagNames.includes(n)) && (a = !0, e.schema.ancestors && Ro.call(e.schema.ancestors, n))) {
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
	return Qo(o, t), o;
}
function Wo(e, t) {
	let n = {
		type: "root",
		children: Ko(e, t.children)
	};
	return Qo(n, t), n;
}
function Go(e, t) {
	let n = {
		type: "text",
		value: typeof t.value == "string" ? t.value : ""
	};
	return Qo(n, t), n;
}
function Ko(e, t) {
	let n = [];
	if (Array.isArray(t)) {
		let r = t, i = -1;
		for (; ++i < r.length;) {
			let t = Bo(e, r[i]);
			t && (Array.isArray(t) ? n.push(...t) : n.push(t));
		}
	}
	return n;
}
function qo(e, t) {
	let n = e.stack[e.stack.length - 1], r = e.schema.attributes, i = e.schema.required, a = r && Ro.call(r, n) ? r[n] : void 0, o = r && Ro.call(r, "*") ? r["*"] : void 0, s = t && typeof t == "object" ? t : {}, c = {}, l;
	for (l in s) if (Ro.call(s, l)) {
		let t = s[l], n = Jo(e, $o(a, l), l, t);
		n ??= Jo(e, $o(o, l), l, t), n != null && (c[l] = n);
	}
	if (i && Ro.call(i, n)) {
		let e = i[n];
		for (l in e) Ro.call(e, l) && !Ro.call(c, l) && (c[l] = e[l]);
	}
	return c;
}
function Jo(e, t, n, r) {
	return t ? Array.isArray(r) ? Yo(e, t, n, r) : Xo(e, t, n, r) : void 0;
}
function Yo(e, t, n, r) {
	let i = -1, a = [];
	for (; ++i < r.length;) {
		let o = Xo(e, t, n, r[i]);
		(typeof o == "number" || typeof o == "string") && a.push(o);
	}
	return a;
}
function Xo(e, t, n, r) {
	if ((typeof r == "boolean" || typeof r == "number" || typeof r == "string") && Zo(e, n, r)) {
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
function Zo(e, t, n) {
	let r = e.schema.protocols && Ro.call(e.schema.protocols, t) ? e.schema.protocols[t] : void 0;
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
function Qo(e, t) {
	let n = po(t);
	t.data && (e.data = We(t.data)), n && (e.position = n);
}
function $o(e, t) {
	let n, r = -1;
	if (e) for (; ++r < e.length;) {
		let i = e[r], a = typeof i == "string" ? i : i[0];
		if (a === t) return i;
		a === "data*" && (n = i);
	}
	if (t.length > 4 && t.slice(0, 4).toLowerCase() === "data") return n;
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/rehype-sanitize/6.0.0/e86312c10b3ea6fac0e127dc05f18d2f26c5f19ba836eeb4d37116fbf9354cf2/node_modules/rehype-sanitize/lib/index.js
function es(e) {
	return function(t) {
		return zo(t, e);
	};
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/mdast-util-find-and-replace/3.0.4/2d74b726887af636126d19425862fd9a149e8887d7a20ba91d09b446a6769a34/node_modules/mdast-util-find-and-replace/lib/index.js
function ts(e, t, n) {
	let r = C((n || {}).ignore || []), i = ns(t), a = -1;
	for (; ++a < i.length;) pe(e, o);
	function o(e, t, n) {
		if (e.type === "text" && typeof n == "number") return s(e, t, n);
		if (r(e, n, t[t.length - 1])) return w;
	}
	function s(e, t, n) {
		let r = t[t.length - 1], o = i[a][0], s = i[a][1], c = 0, l = !1, u = [];
		o.lastIndex = 0;
		let d = o.exec(e.value);
		for (; d;) {
			let n = d.index, r = {
				index: d.index,
				input: d.input,
				stack: [...t, e]
			}, i = s(...d, r);
			if (typeof i == "string" && (i = i.length > 0 ? {
				type: "text",
				value: i
			} : void 0), i === !1) o.lastIndex = n + 1;
			else {
				if (c !== n && f({
					type: "text",
					value: e.value.slice(c, n)
				}), Array.isArray(i)) for (let e of i) f(e);
				else i && f(i);
				c = n + d[0].length, l = !0;
			}
			if (!o.global) break;
			d = o.exec(e.value);
		}
		return l ? (c < e.value.length && f({
			type: "text",
			value: e.value.slice(c)
		}), r.children.splice(n, 1, ...u)) : u = [e], n + u.length;
		function f(e) {
			let t = u[u.length - 1];
			t && t.type === "text" && e.type === "text" ? t.value += e.value : u.push(e);
		}
	}
}
function ns(e) {
	if (!Array.isArray(e)) throw TypeError("Expected find and replace tuple or list of tuples");
	let t = [], n = !e[0] || Array.isArray(e[0]) ? e : [e], r = -1;
	for (; ++r < n.length;) {
		let e = n[r];
		t.push([rs(e[0]), is(e[1])]);
	}
	return t;
}
function rs(e) {
	return typeof e == "string" ? new RegExp(b(e), "g") : e;
}
function is(e) {
	return typeof e == "function" ? e : function() {
		return e;
	};
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/mdast-util-gfm-autolink-literal/2.1.0/eee5d74e9b7f9580f7bcbffd6fa81b3d13e30bfd110b0d9427b1e8c66fd02766/node_modules/mdast-util-gfm-autolink-literal/lib/index.js
var as = "phrasing", os = [
	"autolink",
	"link",
	"image",
	"label"
];
function ss() {
	return {
		transforms: [hs],
		enter: {
			literalAutolink: ls,
			literalAutolinkEmail: us,
			literalAutolinkHttp: us,
			literalAutolinkWww: us
		},
		exit: {
			literalAutolink: ms,
			literalAutolinkEmail: ps,
			literalAutolinkHttp: ds,
			literalAutolinkWww: fs
		}
	};
}
function cs() {
	return { unsafe: [
		{
			character: "@",
			before: "[+\\-.\\w]",
			after: "[\\-.\\w]",
			inConstruct: as,
			notInConstruct: os
		},
		{
			character: ".",
			before: "[Ww]",
			after: "[\\-.\\w]",
			inConstruct: as,
			notInConstruct: os
		},
		{
			character: ":",
			before: "[ps]",
			after: "\\/",
			inConstruct: as,
			notInConstruct: os
		}
	] };
}
function ls(e) {
	this.enter({
		type: "link",
		title: null,
		url: "",
		children: [],
		position: void 0
	}, e);
}
function us(e) {
	this.config.enter.autolinkProtocol.call(this, e);
}
function ds(e) {
	this.config.exit.autolinkProtocol.call(this, e);
}
function fs(e) {
	this.config.exit.data.call(this, e);
	let t = this.stack[this.stack.length - 1];
	t.type, t.url = "http://" + this.sliceSerialize(e);
}
function ps(e) {
	this.config.exit.autolinkEmail.call(this, e);
}
function ms(e) {
	this.exit(e);
}
function hs(e) {
	ts(e, [
		[/(?<![\da-z])(https?:\/\/|www(?=\.))([\w-]*\.[\w\-.]*)([^\t\n\r ]*)/gi, gs],
		[/(?<![\w+\-.@])xmpp:([\w+\-.]+)@([\w-]+(?:\.[\w-]+)+)(\/(?:[\w\-/]|\.(?=[\dA-Za-z]))*)?/g, ys],
		[/(?<![\w+\-.@])mailto:([\w+\-.]+)@([\w-]+(?:\.[\w-]+)+)/g, vs],
		[/(?<![\w+\-.]|(?:^|[^\d@A-Za-z])(?:mailto|xmpp):)([\w+\-.]+)@([\w-]+(?:\.[\w-]+)+)/g, _s]
	], { ignore: ["link", "linkReference"] });
}
function gs(e, t, n, r, i) {
	if (!Ss(i)) return !1;
	let a = "";
	if (/^w/i.test(t) && (n = t + n, t = "", a = "http://"), ee(n.charCodeAt(0)) || !bs(n)) return !1;
	let o = xs(n + r), s = {
		type: "link",
		title: null,
		url: a + t + o[0],
		children: [{
			type: "text",
			value: t + o[0],
			position: void 0
		}],
		position: void 0
	};
	return o[1] ? [s, {
		type: "text",
		value: o[1],
		position: void 0
	}] : s;
}
function _s(e, t, n) {
	return !/[\d\-_]$/.test(n) && {
		type: "link",
		title: null,
		url: "mailto:" + t + "@" + n,
		children: [{
			type: "text",
			value: t + "@" + n,
			position: void 0
		}],
		position: void 0
	};
}
function vs(e, t, n) {
	if (/[\d\-_]$/.test(n)) return !1;
	let r = "mailto:" + t + "@" + n;
	return {
		type: "link",
		title: null,
		url: r,
		children: [{
			type: "text",
			value: r,
			position: void 0
		}],
		position: void 0
	};
}
function ys(e, t, n, r, i) {
	if (/[\d\-_]$/.test(n) || r && (!/[A-Za-z]$/.test(r) || i.input.charAt(i.index + e.length) === "@")) return !1;
	let a = "xmpp:" + t + "@" + n + (r || "");
	return {
		type: "link",
		title: null,
		url: a,
		children: [{
			type: "text",
			value: a,
			position: void 0
		}],
		position: void 0
	};
}
function bs(e) {
	let t = e.split(".");
	return !(t.length < 2 || t[t.length - 1] && (/_/.test(t[t.length - 1]) || !/[\dA-Za-z]/.test(t[t.length - 1])) || t[t.length - 2] && (/_/.test(t[t.length - 2]) || !/[\dA-Za-z]/.test(t[t.length - 2])));
}
function xs(e) {
	let t = e.length;
	for (; t > 0;) {
		let n = e.charCodeAt(t - 1);
		if (n === 33 || n === 34 || n === 38 || n === 39 || n === 41 || n === 44 || n === 46 || n === 58 || n === 59 || n === 60 || n === 62 || n === 63 || n === 93 || n === 125) t--;
		else break;
	}
	if (t === e.length) return [e, void 0];
	let n = e.slice(t);
	e = e.slice(0, t);
	let r = n.indexOf(")"), i = v(e, "("), a = v(e, ")");
	for (; r !== -1 && i > a;) e += n.slice(0, r + 1), n = n.slice(r + 1), r = n.indexOf(")"), a++;
	return [e, n];
}
function Ss(e) {
	let t = e.input.charCodeAt(e.index - 1);
	return e.index === 0 || re(t) || ee(t);
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/mdast-util-gfm-footnote/2.1.0/893a3a48d2e931f331c34c8889b67c836c2ce32fa83fda728325811ff43b845c/node_modules/mdast-util-gfm-footnote/lib/index.js
Ms.peek = js;
function Cs() {
	this.buffer();
}
function ws(e) {
	this.enter({
		type: "footnoteReference",
		identifier: "",
		label: ""
	}, e);
}
function Ts() {
	this.buffer();
}
function Es(e) {
	this.enter({
		type: "footnoteDefinition",
		identifier: "",
		label: "",
		children: []
	}, e);
}
function Ds(e) {
	let t = this.resume(), n = this.stack[this.stack.length - 1];
	n.type, n.identifier = D(this.sliceSerialize(e)).toLowerCase(), n.label = t;
}
function Os(e) {
	this.exit(e);
}
function ks(e) {
	let t = this.resume(), n = this.stack[this.stack.length - 1];
	n.type, n.identifier = D(this.sliceSerialize(e)).toLowerCase(), n.label = t;
}
function As(e) {
	this.exit(e);
}
function js() {
	return "[";
}
function Ms(e, t, n, r) {
	let i = n.createTracker(r), a = i.move("[^"), o = n.enter("footnoteReference"), s = n.enter("reference");
	return a += i.move(n.safe(n.associationId(e), {
		after: "]",
		before: a
	})), s(), o(), a += i.move("]"), a;
}
function Ns() {
	return {
		enter: {
			gfmFootnoteCallString: Cs,
			gfmFootnoteCall: ws,
			gfmFootnoteDefinitionLabelString: Ts,
			gfmFootnoteDefinition: Es
		},
		exit: {
			gfmFootnoteCallString: Ds,
			gfmFootnoteCall: Os,
			gfmFootnoteDefinitionLabelString: ks,
			gfmFootnoteDefinition: As
		}
	};
}
function Ps(e) {
	let t = !1;
	return e && e.firstLineBlank && (t = !0), {
		handlers: {
			footnoteDefinition: n,
			footnoteReference: Ms
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
		})), c(), o += a.move("]:"), e.children && e.children.length > 0 && (a.shift(4), o += a.move((t ? "\n" : " ") + r.indentLines(r.containerFlow(e, a.current()), t ? Is : Fs))), s(), o;
	}
}
function Fs(e, t, n) {
	return t === 0 ? e : Is(e, t, n);
}
function Is(e, t, n) {
	return (n ? "" : "    ") + e;
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/mdast-util-gfm/3.1.0/e3b03b67b30e49bdfa91b364d75193ae8e66c48db8733de5c92b82495bcb2f21/node_modules/mdast-util-gfm/lib/index.js
function Ls() {
	return [
		ss(),
		Ns(),
		E(),
		he(),
		le()
	];
}
function Rs(e) {
	return { extensions: [
		cs(),
		Ps(e),
		te(),
		ge(e),
		ie()
	] };
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/micromark-extension-gfm-autolink-literal/2.1.0/4c2b6e13d31e3105fc7a7f9819380fba5f6026708372ae1c76f8713ff168f466/node_modules/micromark-extension-gfm-autolink-literal/lib/syntax.js
var zs = {
	tokenize: $s,
	partial: !0
}, Bs = {
	tokenize: ec,
	partial: !0
}, Vs = {
	tokenize: tc,
	partial: !0
}, Hs = {
	tokenize: nc,
	partial: !0
}, Us = {
	tokenize: rc,
	partial: !0
}, Ws = {
	name: "wwwAutolink",
	tokenize: Zs,
	previous: ic
}, Gs = {
	name: "protocolAutolink",
	tokenize: Qs,
	previous: ac
}, Ks = {
	name: "emailAutolink",
	tokenize: Xs,
	previous: oc
}, qs = {};
function Js() {
	return { text: qs };
}
for (var Ys = 48; Ys < 123;) qs[Ys] = Ks, Ys++, Ys === 58 ? Ys = 65 : Ys === 91 && (Ys = 97);
qs[43] = Ks, qs[45] = Ks, qs[46] = Ks, qs[95] = Ks, qs[72] = [Ks, Gs], qs[104] = [Ks, Gs], qs[87] = [Ks, Ws], qs[119] = [Ks, Ws];
function Xs(e, t, n) {
	let r = this, i, a;
	return o;
	function o(t) {
		return !sc(t) || !oc.call(r, r.previous) || cc(r.events) ? n(t) : (e.enter("literalAutolink"), e.enter("literalAutolinkEmail"), s(t));
	}
	function s(t) {
		return sc(t) ? (e.consume(t), s) : t === 64 ? (e.consume(t), c) : n(t);
	}
	function c(t) {
		return t === 46 ? e.check(Us, u, l)(t) : t === 45 || t === 95 || T(t) ? (a = !0, e.consume(t), c) : u(t);
	}
	function l(t) {
		return e.consume(t), i = !0, c;
	}
	function u(o) {
		return a && i && S(r.previous) ? (e.exit("literalAutolinkEmail"), e.exit("literalAutolink"), t(o)) : n(o);
	}
}
function Zs(e, t, n) {
	let r = this;
	return i;
	function i(t) {
		return t !== 87 && t !== 119 || !ic.call(r, r.previous) || cc(r.events) ? n(t) : (e.enter("literalAutolink"), e.enter("literalAutolinkWww"), e.check(zs, e.attempt(Bs, e.attempt(Vs, a), n), n)(t));
	}
	function a(n) {
		return e.exit("literalAutolinkWww"), e.exit("literalAutolink"), t(n);
	}
}
function Qs(e, t, n) {
	let r = this, i = "", a = !1;
	return o;
	function o(t) {
		return (t === 72 || t === 104) && ac.call(r, r.previous) && !cc(r.events) ? (e.enter("literalAutolink"), e.enter("literalAutolinkHttp"), i += String.fromCodePoint(t), e.consume(t), s) : n(t);
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
		return t === null || ne(t) || y(t) || re(t) || ee(t) ? n(t) : e.attempt(Bs, e.attempt(Vs, u), n)(t);
	}
	function u(n) {
		return e.exit("literalAutolinkHttp"), e.exit("literalAutolink"), t(n);
	}
}
function $s(e, t, n) {
	let r = 0;
	return i;
	function i(t) {
		return (t === 87 || t === 119) && r < 3 ? (r++, e.consume(t), i) : t === 46 && r === 3 ? (e.consume(t), a) : n(t);
	}
	function a(e) {
		return e === null ? n(e) : t(e);
	}
}
function ec(e, t, n) {
	let r, i, a;
	return o;
	function o(t) {
		return t === 46 || t === 95 ? e.check(Hs, c, s)(t) : t === null || y(t) || re(t) || t !== 45 && ee(t) ? c(t) : (a = !0, e.consume(t), o);
	}
	function s(t) {
		return t === 95 ? r = !0 : (i = r, r = void 0), e.consume(t), o;
	}
	function c(e) {
		return i || r || !a ? n(e) : t(e);
	}
}
function tc(e, t) {
	let n = 0, r = 0;
	return i;
	function i(o) {
		return o === 40 ? (n++, e.consume(o), i) : o === 41 && r < n ? a(o) : o === 33 || o === 34 || o === 38 || o === 39 || o === 41 || o === 42 || o === 44 || o === 46 || o === 58 || o === 59 || o === 60 || o === 63 || o === 93 || o === 95 || o === 126 ? e.check(Hs, t, a)(o) : o === null || y(o) || re(o) ? t(o) : (e.consume(o), i);
	}
	function a(t) {
		return t === 41 && r++, e.consume(t), i;
	}
}
function nc(e, t, n) {
	return r;
	function r(o) {
		return o === 33 || o === 34 || o === 39 || o === 41 || o === 42 || o === 44 || o === 46 || o === 58 || o === 59 || o === 63 || o === 95 || o === 126 ? (e.consume(o), r) : o === 38 ? (e.consume(o), a) : o === 93 ? (e.consume(o), i) : o === 60 || o === null || y(o) || re(o) ? t(o) : n(o);
	}
	function i(e) {
		return e === null || e === 40 || e === 91 || y(e) || re(e) ? t(e) : r(e);
	}
	function a(e) {
		return S(e) ? o(e) : n(e);
	}
	function o(t) {
		return t === 59 ? (e.consume(t), r) : S(t) ? (e.consume(t), o) : n(t);
	}
}
function rc(e, t, n) {
	return r;
	function r(t) {
		return e.consume(t), i;
	}
	function i(e) {
		return T(e) ? n(e) : t(e);
	}
}
function ic(e) {
	return e === null || e === 40 || e === 42 || e === 95 || e === 91 || e === 93 || e === 126 || y(e);
}
function ac(e) {
	return !S(e);
}
function oc(e) {
	return !(e === 47 || sc(e));
}
function sc(e) {
	return e === 43 || e === 45 || e === 46 || e === 95 || T(e);
}
function cc(e) {
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
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/micromark-util-sanitize-uri/2.0.1/4299cd011039987e8843760b0580e66f7b0e8e5d4c6c4b135ad03382581f7392/node_modules/micromark-util-sanitize-uri/index.js
function lc(e) {
	let t = [], n = -1, r = 0, i = 0;
	for (; ++n < e.length;) {
		let a = e.charCodeAt(n), o = "";
		if (a === 37 && T(e.charCodeAt(n + 1)) && T(e.charCodeAt(n + 2))) i = 2;
		else if (a < 128) /[!#$&-;=?-Z_a-z~]/.test(String.fromCharCode(a)) || (o = String.fromCharCode(a));
		else if (a > 55295 && a < 57344) {
			let t = e.charCodeAt(n + 1);
			a < 56320 && t > 56319 && t < 57344 ? (o = String.fromCharCode(a, t), i = 1) : o = "�";
		} else o = String.fromCharCode(a);
		o &&= (t.push(e.slice(r, n), encodeURIComponent(o)), r = n + i + 1, ""), i &&= (n += i, 0);
	}
	return t.join("") + e.slice(r);
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/micromark-extension-gfm-footnote/2.1.0/29df0e4d296d9cc279ad183e48321c6d7c1893aedec9f237a6d2c9c9884388f5/node_modules/micromark-extension-gfm-footnote/lib/syntax.js
var uc = {
	tokenize: vc,
	partial: !0
};
function dc() {
	return {
		document: { 91: {
			name: "gfmFootnoteDefinition",
			tokenize: hc,
			continuation: { tokenize: gc },
			exit: _c
		} },
		text: {
			91: {
				name: "gfmFootnoteCall",
				tokenize: mc
			},
			93: {
				name: "gfmPotentialFootnoteCall",
				add: "after",
				tokenize: fc,
				resolveTo: pc
			}
		}
	};
}
function fc(e, t, n) {
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
		let s = D(r.sliceSerialize({
			start: o.end,
			end: r.now()
		}));
		return s.codePointAt(0) !== 94 || !a.includes(s.slice(1)) ? n(i) : (e.enter("gfmFootnoteCallLabelMarker"), e.consume(i), e.exit("gfmFootnoteCallLabelMarker"), t(i));
	}
}
function pc(e, t) {
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
function mc(e, t, n) {
	let r = this, i = r.parser.gfmFootnotes || (r.parser.gfmFootnotes = []), a = 0, o;
	return s;
	function s(t) {
		return e.enter("gfmFootnoteCall"), e.enter("gfmFootnoteCallLabelMarker"), e.consume(t), e.exit("gfmFootnoteCallLabelMarker"), c;
	}
	function c(t) {
		return t === 94 ? (e.enter("gfmFootnoteCallMarker"), e.consume(t), e.exit("gfmFootnoteCallMarker"), e.enter("gfmFootnoteCallString"), e.enter("chunkString").contentType = "string", l) : n(t);
	}
	function l(s) {
		if (a > 999 || s === 93 && !o || s === null || s === 91 || y(s)) return n(s);
		if (s === 93) {
			e.exit("chunkString");
			let a = e.exit("gfmFootnoteCallString");
			return i.includes(D(r.sliceSerialize(a))) ? (e.enter("gfmFootnoteCallLabelMarker"), e.consume(s), e.exit("gfmFootnoteCallLabelMarker"), e.exit("gfmFootnoteCall"), t) : n(s);
		}
		return y(s) || (o = !0), a++, e.consume(s), s === 92 ? u : l;
	}
	function u(t) {
		return t === 91 || t === 92 || t === 93 ? (e.consume(t), a++, l) : l(t);
	}
}
function hc(e, t, n) {
	let r = this, i = r.parser.gfmFootnotes || (r.parser.gfmFootnotes = []), a, o = 0, s;
	return c;
	function c(t) {
		return e.enter("gfmFootnoteDefinition")._container = !0, e.enter("gfmFootnoteDefinitionLabel"), e.enter("gfmFootnoteDefinitionLabelMarker"), e.consume(t), e.exit("gfmFootnoteDefinitionLabelMarker"), l;
	}
	function l(t) {
		return t === 94 ? (e.enter("gfmFootnoteDefinitionMarker"), e.consume(t), e.exit("gfmFootnoteDefinitionMarker"), e.enter("gfmFootnoteDefinitionLabelString"), e.enter("chunkString").contentType = "string", u) : n(t);
	}
	function u(t) {
		if (o > 999 || t === 93 && !s || t === null || t === 91 || y(t)) return n(t);
		if (t === 93) {
			e.exit("chunkString");
			let n = e.exit("gfmFootnoteDefinitionLabelString");
			return a = D(r.sliceSerialize(n)), e.enter("gfmFootnoteDefinitionLabelMarker"), e.consume(t), e.exit("gfmFootnoteDefinitionLabelMarker"), e.exit("gfmFootnoteDefinitionLabel"), f;
		}
		return y(t) || (s = !0), o++, e.consume(t), t === 92 ? d : u;
	}
	function d(t) {
		return t === 91 || t === 92 || t === 93 ? (e.consume(t), o++, u) : u(t);
	}
	function f(t) {
		return t === 58 ? (e.enter("definitionMarker"), e.consume(t), e.exit("definitionMarker"), i.includes(a) || i.push(a), ce(e, p, "gfmFootnoteDefinitionWhitespace")) : n(t);
	}
	function p(e) {
		return t(e);
	}
}
function gc(e, t, n) {
	return e.check(se, t, e.attempt(uc, t, n));
}
function _c(e) {
	e.exit("gfmFootnoteDefinition");
}
function vc(e, t, n) {
	let r = this;
	return ce(e, i, "gfmFootnoteDefinitionIndent", 5);
	function i(e) {
		let i = r.events[r.events.length - 1];
		return i && i[1].type === "gfmFootnoteDefinitionIndent" && i[2].sliceSerialize(i[1], !0).length === 4 ? t(e) : n(e);
	}
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/micromark-extension-gfm/3.0.0/ddabeeb1737ccbdf8eca4dd4ef4acedaff18c7ef058f356e5983cc268c0a7f34/node_modules/micromark-extension-gfm/index.js
function yc(e) {
	return de([
		Js(),
		dc(),
		ue(e),
		oe(),
		O()
	]);
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/remark-gfm/4.0.1/58e4915f4ed43a77132433a5d6c1e67eda733cd52c04e8623f9e43f8fa695984/node_modules/remark-gfm/lib/index.js
var bc = {};
function xc(e) {
	let t = this, n = e || bc, r = t.data(), i = r.micromarkExtensions ||= [], a = r.fromMarkdownExtensions ||= [], o = r.toMarkdownExtensions ||= [];
	i.push(yc(n)), a.push(Ls()), o.push(Rs(n));
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/remend/1.4.0/11d16e7bb37b15e8341645aa788cb6bc7fce595f1fe869c5ba9269e10eb5e227/node_modules/remend/dist/index.js
var Sc = Object.defineProperty, Cc = Object.defineProperties, wc = Object.getOwnPropertyDescriptors, Tc = Object.getOwnPropertySymbols, Ec = Object.prototype.hasOwnProperty, Dc = Object.prototype.propertyIsEnumerable, Oc = (e, t, n) => t in e ? Sc(e, t, {
	enumerable: !0,
	configurable: !0,
	writable: !0,
	value: n
}) : e[t] = n, kc = (e, t) => {
	for (var n in t ||= {}) Ec.call(t, n) && Oc(e, n, t[n]);
	if (Tc) for (var n of Tc(t)) Dc.call(t, n) && Oc(e, n, t[n]);
	return e;
}, Ac = (e, t) => Cc(e, wc(t)), U = {
	PROSE: 0,
	FENCE_MARKER: 1,
	FENCE_INFO: 2,
	FENCE_BODY: 3,
	CODE_SPAN: 4,
	CODE_SPAN_OPEN: 5
}, jc = /^(`{3,}|~{3,})(.*)$/, Mc = (e) => e !== void 0 && e >= "0" && e <= "9", Nc = (e, t) => {
	let n = t;
	if (e[n] === "-" || e[n] === "*" || e[n] === "+") n += 1;
	else {
		for (; Mc(e[n]) && n - t < 9;) n += 1;
		if (n === t || e[n] !== "." && e[n] !== ")") return -1;
		n += 1;
	}
	return e[n] === " " ? n + 1 : -1;
}, Pc = (e, t, n) => {
	let r = 0, i = t, a = !1, o = t;
	for (;;) {
		for (; o < n && e[o] === " ";) o += 1;
		if (e[o] === ">") {
			r += 1, o += e[o + 1] === " " ? 2 : 1, i = o, a = !1;
			continue;
		}
		let t = o < n ? Nc(e, o) : -1;
		if (t === -1) return {
			contentStart: o,
			listIndent: a ? o - i : 0,
			quoteDepth: r
		};
		o = t, a = !0;
	}
}, Fc = (e, t, n, r) => {
	let i = t;
	for (let t = 0; t < r; t += 1) {
		for (; i < n && e[i] === " ";) i += 1;
		if (e[i] !== ">") return -1;
		i += 1, e[i] === " " && (i += 1);
	}
	return i;
}, Ic = (e, t, n, r, i) => {
	let a = t + r;
	e.fill(U.FENCE_MARKER, a, a + i), e.fill(U.FENCE_INFO, a + i, Math.min(n + 1, e.length));
}, Lc = (e, t, n, r) => {
	let i = t;
	for (; i < n && e[i] === " ";) i += 1;
	let a = 0;
	for (; i < n && e[i] === r.char;) i += 1, a += 1;
	if (a < r.length) return !1;
	for (; i < n;) {
		if (e[i] !== " " && e[i] !== "	" && e[i] !== "\r") return !1;
		i += 1;
	}
	return !0;
}, Rc = (e, t, n, r) => {
	let i = r > n && e[r - 1] === "\r" ? r - 1 : r, a = Pc(e, n, i), o = e.slice(a.contentStart, i).match(jc);
	if (!o) return null;
	let [, s, c] = o, l = s[0];
	return l === "`" && c.includes("`") ? null : (Ic(t, n, r, a.contentStart - n, s.length), {
		char: l,
		length: s.length,
		listIndent: a.listIndent,
		quoteDepth: a.quoteDepth
	});
}, zc = (e, t, n, r) => {
	let i = Fc(e, t, n, r.quoteDepth);
	if (i === -1 || r.listIndent === 0) return i;
	let a = i;
	for (; a < n && e[a] === " ";) a += 1;
	return a === n || e[a] === "\r" && a + 1 === n || a - i >= r.listIndent ? i : -1;
}, Bc = (e, t) => {
	let n = e.length, r = null, i = 0;
	for (; i < n;) {
		let a = e.indexOf("\n", i);
		a === -1 && (a = n);
		let o = r ? zc(e, i, a, r) : -1;
		r && o === -1 && (r = null), r ? Lc(e, o, a, r) ? (t.fill(U.FENCE_MARKER, i, a), r = null) : t.fill(U.FENCE_BODY, i, Math.min(a + 1, n)) : r = Rc(e, t, i, a), i = a + 1;
	}
	return r;
}, Vc = (e, t) => {
	let n = t + 1;
	for (; n < e.length && e[n] === "`";) n += 1;
	return n;
}, Hc = (e, t) => {
	let n = t + 1;
	for (; n < e.length && (e[n] === " " || e[n] === "	" || e[n] === "\r");) n += 1;
	return n < e.length && e[n] === "\n";
}, Uc = (e, t, n) => {
	let r = n;
	for (;;) {
		let n = e.indexOf("`", r);
		if (n === -1) return -1;
		let i = n > r && e[n - 1] === "\\" && t[n - 1] === U.PROSE;
		if (t[n] === U.PROSE && !i) return n;
		r = n + 1;
	}
}, Wc = (e, t) => {
	let n = e.length, r = -1, i = 0, a = 0;
	for (; a < n;) {
		if (r < 0) {
			if (a = Uc(e, t, a), a === -1) break;
			let n = Vc(e, a);
			r = a, i = n - a, a = n;
			continue;
		}
		if (t[a] !== U.PROSE) {
			t.fill(U.CODE_SPAN_OPEN, r, a), r = -1, a += 1;
			continue;
		}
		if (e[a] === "\n" && Hc(e, a)) {
			r = -1, a += 1;
			continue;
		}
		if (e[a] !== "`") {
			a += 1;
			continue;
		}
		let n = Vc(e, a);
		n - a === i && (t.fill(U.CODE_SPAN, r, n), r = -1), a = n;
	}
	return r >= 0 ? (t.fill(U.CODE_SPAN_OPEN, r, n), {
		start: r,
		runLength: i
	}) : null;
}, Gc = (e) => {
	let t = new Uint8Array(e.length);
	return {
		text: e,
		regions: t,
		openFence: Bc(e, t),
		openSpan: Wc(e, t),
		mathMask: null,
		linkUrlMask: null,
		htmlTagMask: null
	};
}, Kc = null, qc = null, Jc = (e) => {
	if (qc !== null && e === Kc) return qc;
	let t = Gc(e);
	return Kc = e, qc = t, t;
}, Yc = () => {
	Kc = null, qc = null;
}, Xc = (e, t) => t >= e.regions.length ? e.openFence !== null || e.openSpan !== null : t < 0 ? !1 : e.regions[t] !== U.PROSE, Zc = (e, t) => e.regions[t] === U.CODE_SPAN, Qc = (e, t) => {
	let n = Jc(e), r = t + t, i = 0;
	for (let t = e.indexOf(r); t !== -1; t = e.indexOf(r, t)) n.regions[t] === U.PROSE ? (i += 1, t += 2) : t += 1;
	return i;
}, $c = /* @__PURE__ */ new Uint8Array(), el = (e) => e === "inlineLatex" || e === "blockLatex", tl = (e, t) => t === "[" && e === "none" ? "blockLatex" : t === "]" && e === "blockLatex" ? "none" : t === "(" && e === "none" ? "inlineLatex" : t === ")" && e === "inlineLatex" ? "none" : null, nl = (e, t) => t ? e === "blockDollar" ? "none" : "blockDollar" : e === "blockDollar" ? e : e === "inlineDollar" ? "none" : "inlineDollar", rl = (e) => e.includes("$") || e.includes("\\(") || e.includes("\\["), il = (e, t, n) => {
	let r = e[t + 1];
	if (e[t] === "\\") {
		if (r === "$") return {
			context: n,
			length: 2
		};
		let e = tl(n, r);
		return e === null ? null : {
			context: e,
			length: 2
		};
	}
	if (e[t] === "$" && !el(n)) {
		let e = r === "$";
		return {
			context: nl(n, e),
			length: e ? 2 : 1
		};
	}
	return null;
}, al = (e) => {
	let { text: t, regions: n } = e, r = t.length, i = new Uint8Array(r), a = "none", o = 0;
	for (; o < r;) {
		i[o] = a === "none" ? 0 : 1;
		let e = n[o] === U.PROSE ? il(t, o, a) : null;
		e === null ? o += 1 : (a = e.context, e.length === 2 && (i[o + 1] = a === "none" ? 0 : 1), o += e.length);
	}
	return i;
}, ol = (e, t) => t < 0 || t >= e.text.length ? !1 : (e.mathMask === null && (e.mathMask = rl(e.text) ? al(e) : $c), e.mathMask[t] === 1), sl = (e, t, n, r) => {
	let { text: i, regions: a } = e, o = new Uint8Array(n - t), s = 0;
	for (let e = n - 1; e >= t; --e) i[e] === ")" && a[e] === U.PROSE && (s = 1), o[e - t] = s;
	let c = !1;
	for (let e = t; e < n; e += 1) c && o[e - t] === 1 && (r[e] = 1), a[e] === U.PROSE && (i[e] === ")" ? c = !1 : i[e] === "(" && (c = e > 0 && i[e - 1] === "]"));
}, cl = (e) => {
	let { text: t } = e, n = t.length, r = new Uint8Array(n), i = 0;
	for (; i < n;) {
		let a = t.indexOf("\n", i);
		a === -1 && (a = n), sl(e, i, a, r), i = a + 1;
	}
	return r;
}, ll = (e, t) => t < 0 || t >= e.text.length ? !1 : (e.linkUrlMask === null && (e.linkUrlMask = e.text.includes("](") ? cl(e) : $c), e.linkUrlMask[t] === 1), ul = (e) => {
	let { text: t, regions: n } = e, r = t.length, i = new Uint8Array(r), a = !1;
	for (let e = 0; e < r; e += 1) if (t[e] === "\n") a = !1;
	else if (i[e] = +!!a, n[e] === U.PROSE) {
		if (t[e] === ">") a = !1;
		else if (t[e] === "<") {
			let n = t[e + 1];
			a = n !== void 0 && (n >= "a" && n <= "z" || n >= "A" && n <= "Z" || n === "/");
		}
	}
	return i;
}, dl = (e, t) => t < 0 || t >= e.text.length ? !1 : (e.htmlTagMask === null && (e.htmlTagMask = e.text.includes("<") ? ul(e) : $c), e.htmlTagMask[t] === 1), W = (e, t) => Xc(Jc(e), t), fl = (e, t) => Zc(Jc(e), t), pl = /^(\s*(?:[-*+]|\d+[.)]) +)>(=?\s*[$]?\d)/gm, ml = (e) => !e || typeof e != "string" || !e.includes(">") ? e : e.replace(pl, (t, n, r, i) => W(e, i) ? t : `${n}\\>${r}`), hl = (e, { delimiter: t, markerLength: n, regex: r }) => {
	let i = e.lastIndexOf(t, e.length - 2), a = i === -1 ? e.length - 1 : i;
	return e.slice(Math.max(0, a - n + 1)).match(r);
}, gl = {
	delimiter: "*",
	markerLength: 2,
	regex: /(\*\*)([^*]*\*?)$/
}, _l = {
	delimiter: "_",
	markerLength: 2,
	regex: /(__)([^_]*?)$/
}, vl = {
	delimiter: "*",
	markerLength: 3,
	regex: /(\*\*\*)([^*]*?)$/
}, yl = {
	delimiter: "*",
	markerLength: 1,
	regex: /(\*)([^*]*?)$/
}, bl = {
	delimiter: "_",
	markerLength: 1,
	regex: /(_)([^_]*?)$/
}, xl = {
	delimiter: "~",
	markerLength: 2,
	regex: /(~~)([^~]*?)$/
}, Sl = {
	delimiter: "_",
	markerLength: 2,
	regex: /(__)([^_]+)_$/
}, Cl = {
	delimiter: "~",
	markerLength: 2,
	regex: /(~~)([^~]+)~$/
}, wl = /^[\s_~*`]*$/, Tl = /^[\s]*[-*+][\s]+$/, El = /[\p{L}\p{N}_]/u, Dl = /^\*{4,}$/, Ol = (e) => {
	if (!e) return !1;
	let t = e.charCodeAt(0);
	return t >= 48 && t <= 57 || t >= 65 && t <= 90 || t >= 97 && t <= 122 || t === 95 || El.test(e);
}, kl = (e, t) => {
	let n = 1;
	for (let r = t - 1; r >= 0; --r) if (e[r] === "]") n += 1;
	else if (e[r] === "[" && (--n, n === 0)) return r;
	return -1;
}, Al = (e, t) => {
	let n = 1;
	for (let r = t + 1; r < e.length; r += 1) if (e[r] === "[") n += 1;
	else if (e[r] === "]" && (--n, n === 0)) return r;
	return -1;
}, jl = (e, t) => ol(Jc(e), t), Ml = (e, t, n) => {
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
}, Nl = (e, t, n, r) => n === "\\" || ol(e, t) ? !0 : n !== "*" && r === "*" ? (t < e.text.length - 2 ? e.text[t + 2] : "") !== "*" : n === "*" || (!n || n === " " || n === "	" || n === "\n") && (!r || r === " " || r === "	" || r === "\n"), Pl = (e) => e === " " || e === "	" || e === "\n", Fl = (e, t) => !!(e && t && Ol(e) && Ol(t)), Il = (e, t, n, r) => {
	let i = Fl(e, t), a = !!t && !Pl(t), o = !!e && !Pl(e);
	return i && n % 2 == 0 && !r ? { count: !1 } : o && n % 2 == 1 || a ? {
		count: !0,
		inWordAsteriskChain: i
	} : { count: !1 };
}, Ll = (e) => {
	let t = Jc(e), n = 0, r = !1, i = e.length;
	for (let a = 0; a < i; a += 1) {
		if (e[a] !== "*" || t.regions[a] !== U.PROSE) {
			Ol(e[a]) || (r = !1);
			continue;
		}
		let o = a > 0 ? e[a - 1] : "", s = a < i - 1 ? e[a + 1] : "";
		if (Nl(t, a, o, s)) continue;
		let c = Il(o, s, n, r);
		c.count && (n += 1, r = c.inWordAsteriskChain);
	}
	return n;
}, Rl = (e, t, n, r) => !!(n === "\\" || ol(e, t) || ll(e, t) || dl(e, t) || n === "_" || r === "_" || n && r && Ol(n) && Ol(r)), zl = (e) => {
	let t = Jc(e), n = 0, r = e.length;
	for (let i = 0; i < r; i += 1) {
		if (e[i] !== "_" || t.regions[i] !== U.PROSE) continue;
		let a = i > 0 ? e[i - 1] : "", o = i < r - 1 ? e[i + 1] : "";
		Rl(t, i, a, o) || (n += 1);
	}
	return n;
}, Bl = (e) => {
	let t = Jc(e), n = 0, r = 0;
	for (let i = 0; i < e.length; i += 1) e[i] === "*" && t.regions[i] === U.PROSE ? r += 1 : (r >= 3 && (n += Math.floor(r / 3)), r = 0);
	return r >= 3 && (n += Math.floor(r / 3)), n;
}, Vl = (e) => Qc(e, "*"), Hl = (e) => e === "" || e === " " || e === "	" || e === "\n", Ul = (e, t, n, r, i) => {
	if (!(Hl(n) && Hl(r))) return !1;
	if (t > i.lineEnd) {
		let n = e.indexOf("\n", t);
		i.lineEnd = n === -1 ? e.length : n, i.result = Ml(e, t, "_");
	}
	return i.result;
}, Wl = (e, t, n, r) => {
	let i = e.text, a = t, o = !1;
	a > 0 && i[a - 1] === "\\" && (a += 1, o = !0);
	let s = n - a;
	if (s < 2) return !1;
	let c = a > 0 ? i[a - 1] : "", l = o ? "\\" : c, u = n < i.length ? i[n] : "";
	return Ol(l) && Ol(u) || Ul(i, a, l, u, r) || ol(e, a) || ll(e, a) || dl(e, a) ? !1 : Math.floor(s / 2) % 2 == 1;
}, Gl = (e) => {
	let t = Jc(e), n = e.length, r = {
		lineEnd: -1,
		result: !1
	}, i = !1, a = 0;
	for (; a < n;) {
		if (e[a] !== "_" || t.regions[a] !== U.PROSE) {
			a += 1;
			continue;
		}
		let o = a, s = a + 1;
		for (; s < n && e[s] === "_" && t.regions[s] === U.PROSE;) s += 1;
		a = s, Wl(t, o, s, r) && (i = !i);
	}
	return i;
}, Kl = (e, t, n) => {
	if (!t || wl.test(t)) return !0;
	let r = e.substring(0, n).lastIndexOf("\n"), i = r === -1 ? 0 : r + 1, a = e.substring(i, n);
	return Tl.test(a) && t.includes("\n") ? !0 : Ml(e, n, "*");
}, ql = (e) => {
	let t = hl(e, gl);
	if (!t) return e;
	let n = t[2], r = e.lastIndexOf(t[1]);
	return W(e, r) || fl(e, r) || Kl(e, n, r) ? e : Vl(e) % 2 == 1 ? n.endsWith("*") ? `${e}*` : `${e}**` : e;
}, Jl = (e, t, n) => {
	if (!t || wl.test(t)) return !0;
	let r = e.substring(0, n).lastIndexOf("\n"), i = r === -1 ? 0 : r + 1, a = e.substring(i, n);
	return Tl.test(a) && t.includes("\n") ? !0 : Ml(e, n, "_");
}, Yl = (e) => {
	let t = hl(e, _l);
	if (!t) {
		let t = hl(e, Sl);
		if (t) {
			let n = e.lastIndexOf(t[1]);
			if (!(W(e, n) || fl(e, n)) && Gl(e)) return `${e}_`;
		}
		return e;
	}
	let n = t[2], r = e.lastIndexOf(t[1]);
	return W(e, r) || fl(e, r) || Jl(e, n, r) ? e : Gl(e) ? `${e}__` : e;
}, Xl = (e, t) => {
	let { text: n } = e;
	return n[t] === "*" && e.regions[t] === U.PROSE && n[t - 1] !== "*" && n[t + 1] !== "*" && n[t - 1] !== "\\" && !ol(e, t);
}, Zl = (e) => {
	let t = Jc(e);
	for (let n = e.indexOf("*"); n !== -1; n = e.indexOf("*", n + 1)) {
		if (!Xl(t, n)) continue;
		let r = n > 0 ? e[n - 1] : "", i = n < e.length - 1 ? e[n + 1] : "", a = !r || Pl(r), o = !i || Pl(i);
		if (!(a && o) && !(r && i && Ol(r) && Ol(i)) && !o) return n;
	}
	return -1;
}, Ql = (e) => {
	if (!hl(e, yl)) return e;
	let t = Zl(e);
	if (t === -1 || W(e, t) || fl(e, t)) return e;
	let n = e.substring(t + 1);
	return !n || wl.test(n) ? e : Ll(e) % 2 == 1 ? `${e}*` : e;
}, $l = (e) => {
	let t = Jc(e);
	for (let n = e.indexOf("_"); n !== -1; n = e.indexOf("_", n + 1)) if (t.regions[n] === U.PROSE && e[n - 1] !== "_" && e[n + 1] !== "_" && e[n - 1] !== "\\" && !ol(t, n) && !ll(t, n)) {
		let t = n > 0 ? e[n - 1] : "", r = n < e.length - 1 ? e[n + 1] : "";
		if (t && r && Ol(t) && Ol(r)) continue;
		return n;
	}
	return -1;
}, eu = (e) => {
	let t = e.length;
	for (; t > 0 && e[t - 1] === "\n";) --t;
	return t < e.length ? `${e.slice(0, t)}_${e.slice(t)}` : `${e}_`;
}, tu = (e) => {
	if (!e.endsWith("**")) return null;
	let t = e.slice(0, -2);
	if (Vl(t) % 2 != 1) return null;
	let n = t.indexOf("**"), r = $l(t);
	return n !== -1 && r !== -1 && n < r ? `${t}_**` : null;
}, nu = (e) => {
	if (!hl(e, bl)) return e;
	let t = $l(e);
	if (t === -1) return e;
	let n = e.substring(t + 1);
	if (!n || wl.test(n) || W(e, t) || fl(e, t)) return e;
	if (zl(e) % 2 == 1) {
		let t = tu(e);
		return t === null ? eu(e) : t;
	}
	return e;
}, ru = (e) => {
	let t = Vl(e), n = Ll(e);
	return t % 2 == 0 && n % 2 == 0;
}, iu = (e, t, n) => !t || wl.test(t) || W(e, n) || fl(e, n) ? !0 : Ml(e, n, "*"), au = (e) => {
	if (Dl.test(e)) return e;
	let t = hl(e, vl);
	if (!t) return e;
	let n = t[2];
	return iu(e, n, e.lastIndexOf(t[1])) ? e : Bl(e) % 2 == 1 ? ru(e) ? e : `${e}***` : e;
}, ou = /<[a-zA-Z/][^>]*$/, su = /[a-zA-Z/]/, cu = (e) => e.includes("$") || e.includes("\\(") || e.includes("\\["), lu = (e, t) => {
	let n = e[t + 1];
	return n !== void 0 && su.test(n);
}, uu = (e) => {
	let t = e.match(ou);
	if (!t || t.index === void 0) return e;
	let n = cu(e);
	for (let r = t.index; r < e.length; r += 1) if (e[r] === "<" && lu(e, r) && !W(e, r) && !(n && jl(e, r))) return e.substring(0, r).trimEnd();
	return e;
}, du = (e) => {
	let t = Jc(e);
	if (t.openFence) return e;
	let n = t.openSpan;
	if (!n) return e;
	let r = e.slice(n.start + n.runLength);
	if (!r || wl.test(r)) return e;
	let i = 0, a = e.length - 1;
	for (; a >= 0 && e[a] === "`";) i += 1, --a;
	return i >= n.runLength ? e : e + "`".repeat(n.runLength - i);
}, fu = (e) => Qc(e, "$"), pu = (e) => {
	let t = Jc(e), n = 0;
	for (let r = 0; r < e.length; r += 1) e[r] === "\\" ? r += 1 : t.regions[r] === U.PROSE && e[r] === "$" && (r + 1 < e.length && e[r + 1] === "$" ? r += 1 : n += 1);
	return n;
}, mu = (e) => {
	if (e.endsWith("$") && !e.endsWith("$$")) return `${e}$`;
	let t = e.indexOf("$$");
	return t !== -1 && e.indexOf("\n", t) !== -1 && !e.endsWith("\n") ? `${e}
$$` : `${e}$$`;
}, hu = (e) => fu(e) % 2 == 0 ? e : mu(e), gu = (e) => pu(e) % 2 == 1 ? `${e}$` : e, _u = "streamdown:incomplete-image", vu = (e, t, n) => {
	if (e.substring(t + 2).includes(")")) return null;
	let r = kl(e, t);
	if (r === -1 || W(e, r)) return null;
	let i = r > 0 && e[r - 1] === "!", a = i ? r - 1 : r, o = e.substring(0, a), s = e.substring(r + 1, t);
	return i ? `${o}![${s}](${_u})` : n === "text-only" ? `${o}${s}` : `${o}[${s}](streamdown:incomplete-link)`;
}, yu = (e, t) => {
	for (let n = 0; n < t; n++) if (e[n] === "[" && !W(e, n)) {
		if (n > 0 && e[n - 1] === "!") continue;
		let t = Al(e, n);
		if (t === -1) return n;
		if (t + 1 < e.length && e[t + 1] === "(") {
			let r = e.indexOf(")", t + 2);
			r !== -1 && (n = r);
		}
	}
	return t;
}, bu = (e, t, n) => {
	let r = t > 0 && e[t - 1] === "!", i = r ? t - 1 : t;
	if (!e.substring(t + 1).includes("]") || Al(e, t) === -1) {
		let a = e.substring(0, i);
		if (r) return `${a}![${e.substring(t + 1)}](${_u})`;
		if (n === "text-only") {
			let n = yu(e, t);
			return e.substring(0, n) + e.substring(n + 1);
		}
		return `${e}](streamdown:incomplete-link)`;
	}
	return null;
}, xu = (e, t) => {
	let n = e.lastIndexOf("](");
	if (n !== -1 && !W(e, n)) {
		let r = vu(e, n, t);
		if (r !== null) return r;
	}
	for (let n = e.lastIndexOf("["); n !== -1; n = n === 0 ? -1 : e.lastIndexOf("[", n - 1)) if (!W(e, n)) {
		let r = bu(e, n, t);
		if (r !== null) return r;
	}
	return e;
}, Su = 32, Cu = (e, t = "protocol") => {
	let n = e;
	for (let e = 0; e < Su; e += 1) {
		let e = xu(n, t);
		if (e.length >= n.length) return e;
		n = e;
	}
	return n;
}, wu = /^-{1,2}$/, Tu = /^[\s]*-{1,2}[\s]+$/, Eu = /^={1,2}$/, Du = /^[\s]*={1,2}[\s]+$/, Ou = (e) => {
	if (!e || typeof e != "string") return e;
	let t = e.lastIndexOf("\n");
	if (t === -1) return e;
	let n = e.substring(t + 1), r = e.substring(0, t), i = n.trim();
	if (wu.test(i) && !n.match(Tu)) {
		let t = r.split("\n").at(-1);
		if (t && t.trim().length > 0) return `${e}\u200B`;
	}
	if (Eu.test(i) && !n.match(Du)) {
		let t = r.split("\n").at(-1);
		if (t && t.trim().length > 0) return `${e}\u200B`;
	}
	return e;
}, ku = /([\p{L}\p{N}_])~(?!~)(?=[\p{L}\p{N}_])/gu, Au = (e) => !e || typeof e != "string" || !e.includes("~") ? e : e.replace(ku, (t, n, r) => W(e, r + n.length) ? t : `${n}\\~`), ju = (e) => Qc(e, "~"), Mu = (e) => {
	let t = hl(e, xl);
	if (t) {
		let n = t[2];
		if (!n || wl.test(n)) return e;
		let r = e.lastIndexOf(t[1]);
		if (W(e, r) || fl(e, r)) return e;
		if (ju(e) % 2 == 1) return `${e}~~`;
	} else {
		let t = hl(e, Cl);
		if (t) {
			let n = e.lastIndexOf(t[0].slice(0, 2));
			if (W(e, n) || fl(e, n)) return e;
			if (ju(e) % 2 == 1) return `${e}~`;
		}
	}
	return e;
}, Nu = (e) => e !== !1, Pu = (e) => e === !0, G = {
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
}, Fu = [
	{
		handler: {
			name: "singleTilde",
			handle: Au,
			priority: G.SINGLE_TILDE
		},
		optionKey: "singleTilde"
	},
	{
		handler: {
			name: "comparisonOperators",
			handle: ml,
			priority: G.COMPARISON_OPERATORS
		},
		optionKey: "comparisonOperators"
	},
	{
		handler: {
			name: "htmlTags",
			handle: uu,
			priority: G.HTML_TAGS
		},
		optionKey: "htmlTags"
	},
	{
		handler: {
			name: "setextHeadings",
			handle: Ou,
			priority: G.SETEXT_HEADINGS
		},
		optionKey: "setextHeadings"
	},
	{
		handler: {
			name: "links",
			handle: Cu,
			priority: G.LINKS
		},
		optionKey: "links",
		earlyReturn: (e) => e.endsWith("](streamdown:incomplete-link)") || e.endsWith("](streamdown:incomplete-image)")
	},
	{
		handler: {
			name: "boldItalic",
			handle: au,
			priority: G.BOLD_ITALIC
		},
		optionKey: "boldItalic"
	},
	{
		handler: {
			name: "bold",
			handle: ql,
			priority: G.BOLD
		},
		optionKey: "bold"
	},
	{
		handler: {
			name: "italicDoubleUnderscore",
			handle: Yl,
			priority: G.ITALIC_DOUBLE_UNDERSCORE
		},
		optionKey: "italic"
	},
	{
		handler: {
			name: "italicSingleAsterisk",
			handle: Ql,
			priority: G.ITALIC_SINGLE_ASTERISK
		},
		optionKey: "italic"
	},
	{
		handler: {
			name: "italicSingleUnderscore",
			handle: nu,
			priority: G.ITALIC_SINGLE_UNDERSCORE
		},
		optionKey: "italic"
	},
	{
		handler: {
			name: "inlineCode",
			handle: du,
			priority: G.INLINE_CODE
		},
		optionKey: "inlineCode"
	},
	{
		handler: {
			name: "strikethrough",
			handle: Mu,
			priority: G.STRIKETHROUGH
		},
		optionKey: "strikethrough"
	},
	{
		handler: {
			name: "katex",
			handle: hu,
			priority: G.KATEX
		},
		optionKey: "katex"
	},
	{
		handler: {
			name: "inlineKatex",
			handle: gu,
			priority: G.INLINE_KATEX
		},
		optionKey: "inlineKatex"
	}
], Iu = (e) => {
	let t = e?.linkMode ?? "protocol";
	return Fu.filter(({ handler: t, optionKey: n }) => t.name === "links" ? Nu(e?.links) || Nu(e?.images) : t.name === "inlineKatex" ? Pu(e?.inlineKatex) : Nu(e?.[n])).map(({ handler: e, earlyReturn: n }) => e.name === "links" ? {
		handler: Ac(kc({}, e), { handle: (e) => Cu(e, t) }),
		earlyReturn: t === "protocol" ? n : void 0
	} : {
		handler: e,
		earlyReturn: n
	});
}, Lu = (e, t) => {
	if (!e || typeof e != "string") return e;
	let n = e.endsWith(" ") && !e.endsWith("  ") ? e.slice(0, -1) : e, r = Iu(t), i = (t?.handlers ?? []).map((e) => ({
		handler: Ac(kc({}, e), { priority: e.priority ?? G.DEFAULT }),
		earlyReturn: void 0
	})), a = [...r, ...i].sort((e, t) => (e.handler.priority ?? 0) - (t.handler.priority ?? 0));
	try {
		for (let { handler: e, earlyReturn: t } of a) if (n = e.handle(n), t != null && t(n)) return n;
		return n.endsWith(" ") && !n.endsWith("  ") ? n.slice(0, -1) : n;
	} finally {
		Yc();
	}
}, Ru = /\/\*(?:[^*]|\*(?!\/))*\*\//g, zu = /\n/g, Bu = /^\s*/, Vu = /^(\*?[-#/*\\\w]+(\[[0-9a-z_-]+\])?)\s*/, Hu = /^:\s*/, Uu = /^((?:'(?:\\.|[^'\\])*'|"(?:\\.|[^"\\])*"|url\((?:'(?:\\.|[^'\\])*'|"(?:\\.|[^"\\])*"|[^)]*)\)|[^};])+)/, Wu = /^[;\s]*/, Gu = "\n", Ku = "/", qu = "*", Ju = "", Yu = "comment", Xu = "declaration";
function Zu(e, t) {
	if (typeof e != "string") throw TypeError("First argument must be a string");
	if (!e) return [];
	t ||= {};
	var n = 1, r = 1;
	function i(e) {
		var t = e.match(zu);
		t && (n += t.length);
		var i = e.lastIndexOf(Gu);
		r = ~i ? e.length - i : r + e.length;
	}
	function a() {
		var e = {
			line: n,
			column: r
		};
		return function(t) {
			return t.position = new o(e), l(), t;
		};
	}
	function o(e) {
		this.start = e, this.end = {
			line: n,
			column: r
		}, this.source = t.source;
	}
	o.prototype.content = e;
	function s(i) {
		var a = /* @__PURE__ */ Error(t.source + ":" + n + ":" + r + ": " + i);
		if (a.reason = i, a.filename = t.source, a.line = n, a.column = r, a.source = e, !t.silent) throw a;
	}
	function c(t) {
		var n = t.exec(e);
		if (n) {
			var r = n[0];
			return i(r), e = e.slice(r.length), n;
		}
	}
	function l() {
		c(Bu);
	}
	function u(e) {
		for (var t; t = d();) e.push(t);
		return e;
	}
	function d() {
		var t = a();
		if (Ku == e.charAt(0) && qu == e.charAt(1)) {
			for (var n = 2; Ju != e.charAt(n) && (qu != e.charAt(n) || Ku != e.charAt(n + 1));) ++n;
			if (n += 2, Ju === e.charAt(n - 1)) return s("End of comment missing");
			var o = e.slice(2, n - 2);
			return r += 2, i(o), e = e.slice(n), r += 2, t({
				type: Yu,
				comment: o
			});
		}
	}
	function f() {
		var e = a(), t = c(Vu);
		if (t) {
			if (d(), !c(Hu)) return s("property missing ':'");
			var n = c(Uu), r = e({
				type: Xu,
				property: t[0].replace(Ru, Ju).trim(),
				value: n ? n[0].replace(Ru, Ju).trim() : Ju
			});
			return c(Wu), r;
		}
	}
	function p() {
		var e = [];
		u(e);
		for (var t; t = f();) e.push(t), u(e);
		return e;
	}
	return l(), p();
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/style-to-object/2.0.2/a30f1425598a361f1af6c50c118e29da138b4f9b135fdd30fdc4edfb07db7b56/node_modules/style-to-object/esm/index.mjs
function Qu(e, t) {
	let n = null;
	if (!e || typeof e != "string") return n;
	let r = Zu(e), i = typeof t == "function";
	return r.forEach((e) => {
		if (e.type !== "declaration") return;
		let { property: r, value: a } = e;
		i ? t(r, a, e) : a && (n ??= {}, n[r] = a);
	}), n;
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/hast-util-to-jsx-runtime/2.3.7/f0c3536a723b2f0a0e7cdd958750dc849414307a09ad48cc1e76157aa270b133/node_modules/hast-util-to-jsx-runtime/lib/index.js
var $u = {}.hasOwnProperty, ed = /* @__PURE__ */ new Map(), td = /[A-Z]/g, nd = /-([a-z])/g, rd = /* @__PURE__ */ new Set([
	"table",
	"tbody",
	"thead",
	"tfoot",
	"tr"
]), id = /* @__PURE__ */ new Set(["td", "th"]), ad = "https://github.com/syntax-tree/hast-util-to-jsx-runtime";
function od(e, t) {
	if (!t || t.Fragment === void 0) throw TypeError("Expected `Fragment` in options");
	let n = t.filePath || void 0, r;
	if (t.development) {
		if (typeof t.jsxDEV != "function") throw TypeError("Expected `jsxDEV` in options when `development: true`");
		r = _d(n, t.jsxDEV);
	} else {
		if (typeof t.jsx != "function") throw TypeError("Expected `jsx` in production options");
		if (typeof t.jsxs != "function") throw TypeError("Expected `jsxs` in production options");
		r = gd(n, t.jsx, t.jsxs);
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
	}, a = sd(i, e, void 0);
	return a && typeof a != "string" ? a : i.create(e, i.Fragment, { children: a || void 0 }, void 0);
}
function sd(e, t, n) {
	if (t.type === "element") return cd(e, t, n);
	if (t.type === "mdxFlowExpression" || t.type === "mdxTextExpression") return ld(e, t);
	if (t.type === "mdxJsxFlowElement" || t.type === "mdxJsxTextElement") return dd(e, t, n);
	if (t.type === "mdxjsEsm") return ud(e, t);
	if (t.type === "root") return fd(e, t, n);
	if (t.type === "text") return pd(e, t);
}
function cd(e, t, n) {
	let r = e.schema, i = r;
	t.tagName.toLowerCase() === "svg" && r.space === "html" && (i = m, e.schema = i), e.ancestors.push(t);
	let a = Cd(e, t.tagName, !1), o = vd(e, t), s = bd(e, t);
	return rd.has(t.tagName) && (s = s.filter(function(e) {
		return typeof e != "string" || !h(e);
	})), md(e, o, a, t), hd(o, s), e.ancestors.pop(), e.schema = r, e.create(t, a, o, n);
}
function ld(e, t) {
	if (t.data && t.data.estree && e.evaluater) {
		let n = t.data.estree.body[0];
		return n.type, e.evaluater.evaluateExpression(n.expression);
	}
	wd(e, t.position);
}
function ud(e, t) {
	if (t.data && t.data.estree && e.evaluater) return e.evaluater.evaluateProgram(t.data.estree);
	wd(e, t.position);
}
function dd(e, t, n) {
	let r = e.schema, i = r;
	t.name === "svg" && r.space === "html" && (i = m, e.schema = i), e.ancestors.push(t);
	let a = t.name === null ? e.Fragment : Cd(e, t.name, !0), o = yd(e, t), s = bd(e, t);
	return md(e, o, a, t), hd(o, s), e.ancestors.pop(), e.schema = r, e.create(t, a, o, n);
}
function fd(e, t, n) {
	let r = {};
	return hd(r, bd(e, t)), e.create(t, e.Fragment, r, n);
}
function pd(e, t) {
	return t.value;
}
function md(e, t, n, r) {
	typeof n != "string" && n !== e.Fragment && e.passNode && (t.node = r);
}
function hd(e, t) {
	if (t.length > 0) {
		let n = t.length > 1 ? t : t[0];
		n && (e.children = n);
	}
}
function gd(e, t, n) {
	return r;
	function r(e, r, i, a) {
		let o = Array.isArray(i.children) ? n : t;
		return a ? o(r, i, a) : o(r, i);
	}
}
function _d(e, t) {
	return n;
	function n(n, r, i, a) {
		let o = Array.isArray(i.children), s = uo(n);
		return t(r, i, a, o, {
			columnNumber: s ? s.column - 1 : void 0,
			fileName: e,
			lineNumber: s ? s.line : void 0
		}, void 0);
	}
}
function vd(e, t) {
	let n = {}, r, i;
	for (i in t.properties) if (i !== "children" && $u.call(t.properties, i)) {
		let a = xd(e, i, t.properties[i]);
		if (a) {
			let [i, o] = a;
			e.tableCellAlignToStyle && i === "align" && typeof o == "string" && id.has(t.tagName) ? r = o : n[i] = o;
		}
	}
	if (r) {
		let t = n.style ||= {};
		t[e.stylePropertyNameCase === "css" ? "text-align" : "textAlign"] = r;
	}
	return n;
}
function yd(e, t) {
	let n = {};
	for (let r of t.attributes) if (r.type === "mdxJsxExpressionAttribute") {
		if (r.data && r.data.estree && e.evaluater) {
			let t = r.data.estree.body[0];
			t.type;
			let i = t.expression;
			i.type;
			let a = i.properties[0];
			a.type, Object.assign(n, e.evaluater.evaluateExpression(a.argument));
		} else wd(e, t.position);
	} else {
		let { name: i } = r, a;
		if (r.value && typeof r.value == "object") {
			if (r.value.data && r.value.data.estree && e.evaluater) {
				let t = r.value.data.estree.body[0];
				t.type, a = e.evaluater.evaluateExpression(t.expression);
			} else wd(e, t.position);
		} else a = r.value === null || r.value;
		n[i] = a;
	}
	return n;
}
function bd(e, t) {
	let n = [], r = -1, i = e.passKeys ? /* @__PURE__ */ new Map() : ed;
	for (; ++r < t.children.length;) {
		let a = t.children[r], o;
		if (e.passKeys) {
			let e = a.type === "element" ? a.tagName : a.type === "mdxJsxFlowElement" || a.type === "mdxJsxTextElement" ? a.name : void 0;
			if (e) {
				let t = i.get(e) || 0;
				o = e + "-" + t, i.set(e, t + 1);
			}
		}
		let s = sd(e, a, o);
		s !== void 0 && n.push(s);
	}
	return n;
}
function xd(e, t, n) {
	if (n == null || typeof n == "number" && Number.isNaN(n)) return;
	let r = c(e.schema, t);
	return Array.isArray(n) && (n = r.commaSeparated ? s(n) : p(n)), r.property === "style" ? ["style", typeof n == "object" ? e.stylePropertyNameCase === "dom" ? n : Td(n) : Sd(e, String(n))] : [e.elementAttributeNameCase === "react" && r.space ? _e[r.property] || r.property : r.attribute, n];
}
function Sd(e, t) {
	let n = {};
	try {
		Qu(t, function(t, r) {
			if (t && r) {
				let i = t;
				i.slice(0, 2) !== "--" && (i = i.toLowerCase(), e.stylePropertyNameCase === "dom" && (i.slice(0, 4) === "-ms-" && (i = i.slice(1)), i = i.replace(nd, Od))), n[i] = r;
			}
		});
	} catch (t) {
		if (e.ignoreInvalidStyle) return {};
		let n = t, r = new fe("Cannot parse `style` attribute", {
			ancestors: e.ancestors,
			cause: n,
			ruleId: "style",
			source: "hast-util-to-jsx-runtime"
		});
		throw r.file = e.filePath || void 0, r.url = ad + "#cannot-parse-style-attribute", r;
	}
	return n;
}
function Cd(e, t, n) {
	let r;
	if (!n) r = {
		type: "Literal",
		value: t
	};
	else if (t.includes(".")) {
		let e = t.split("."), n = -1, i;
		for (; ++n < e.length;) {
			let t = ae(e[n]) ? {
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
	} else r = ae(t) && !/^[a-z]/.test(t) ? {
		type: "Identifier",
		name: t
	} : {
		type: "Literal",
		value: t
	};
	if (r.type === "Literal") {
		let t = r.value;
		return $u.call(e.components, t) ? e.components[t] : t;
	}
	if (e.evaluater) return e.evaluater.evaluateExpression(r);
	wd(e);
}
function wd(e, t) {
	let n = new fe("Cannot handle MDX estrees without `createEvaluater`", {
		ancestors: e.ancestors,
		place: t,
		ruleId: "mdx-estree",
		source: "hast-util-to-jsx-runtime"
	});
	throw n.file = e.filePath || void 0, n.url = ad + "#cannot-handle-mdx-estrees-without-createevaluater", n;
}
function Td(e) {
	let t = {}, n;
	for (n in e) $u.call(e, n) && (t[Ed(n)] = e[n]);
	return t;
}
function Ed(e) {
	if (e.slice(0, 2) === "--") return e;
	let t = e.replace(td, Dd);
	return t.slice(0, 3) === "ms-" && (t = "-" + t), t;
}
function Dd(e) {
	return "-" + e.toLowerCase();
}
function Od(e, t) {
	return t.toUpperCase();
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/html-url-attributes/3.0.1/4ff74860c28cab7a302474ce2fd1dfeaa22ae8a9d7cc188f8eeff706084f44a8/node_modules/html-url-attributes/lib/index.js
var kd = {
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
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/remark-parse/11.0.0/9115b9734ec428a36c24d2c31ba82a6a98bcfe73e6232b695aa5862ff5654bfb/node_modules/remark-parse/lib/index.js
function Ad(e) {
	let t = this;
	t.parser = n;
	function n(n) {
		return me(n, {
			...t.data("settings"),
			...e,
			extensions: t.data("micromarkExtensions") || [],
			mdastExtensions: t.data("fromMarkdownExtensions") || []
		});
	}
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/mdast-util-to-hast/13.2.1/0601b1701843aa9b863db3511935f9f2758af7297665ae75d5d7ed3bcb3cad64/node_modules/mdast-util-to-hast/lib/handlers/blockquote.js
function jd(e, t) {
	let n = {
		type: "element",
		tagName: "blockquote",
		properties: {},
		children: e.wrap(e.all(t), !0)
	};
	return e.patch(t, n), e.applyData(t, n);
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/mdast-util-to-hast/13.2.1/0601b1701843aa9b863db3511935f9f2758af7297665ae75d5d7ed3bcb3cad64/node_modules/mdast-util-to-hast/lib/handlers/break.js
function Md(e, t) {
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
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/mdast-util-to-hast/13.2.1/0601b1701843aa9b863db3511935f9f2758af7297665ae75d5d7ed3bcb3cad64/node_modules/mdast-util-to-hast/lib/handlers/code.js
function Nd(e, t) {
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
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/mdast-util-to-hast/13.2.1/0601b1701843aa9b863db3511935f9f2758af7297665ae75d5d7ed3bcb3cad64/node_modules/mdast-util-to-hast/lib/handlers/delete.js
function Pd(e, t) {
	let n = {
		type: "element",
		tagName: "del",
		properties: {},
		children: e.all(t)
	};
	return e.patch(t, n), e.applyData(t, n);
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/mdast-util-to-hast/13.2.1/0601b1701843aa9b863db3511935f9f2758af7297665ae75d5d7ed3bcb3cad64/node_modules/mdast-util-to-hast/lib/handlers/emphasis.js
function Fd(e, t) {
	let n = {
		type: "element",
		tagName: "em",
		properties: {},
		children: e.all(t)
	};
	return e.patch(t, n), e.applyData(t, n);
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/mdast-util-to-hast/13.2.1/0601b1701843aa9b863db3511935f9f2758af7297665ae75d5d7ed3bcb3cad64/node_modules/mdast-util-to-hast/lib/handlers/footnote-reference.js
function Id(e, t) {
	let n = typeof e.options.clobberPrefix == "string" ? e.options.clobberPrefix : "user-content-", r = String(t.identifier).toUpperCase(), i = lc(r.toLowerCase()), a = e.footnoteOrder.indexOf(r), o, s = e.footnoteCounts.get(r);
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
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/mdast-util-to-hast/13.2.1/0601b1701843aa9b863db3511935f9f2758af7297665ae75d5d7ed3bcb3cad64/node_modules/mdast-util-to-hast/lib/handlers/heading.js
function Ld(e, t) {
	let n = {
		type: "element",
		tagName: "h" + t.depth,
		properties: {},
		children: e.all(t)
	};
	return e.patch(t, n), e.applyData(t, n);
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/mdast-util-to-hast/13.2.1/0601b1701843aa9b863db3511935f9f2758af7297665ae75d5d7ed3bcb3cad64/node_modules/mdast-util-to-hast/lib/handlers/html.js
function Rd(e, t) {
	if (e.options.allowDangerousHtml) {
		let n = {
			type: "raw",
			value: t.value
		};
		return e.patch(t, n), e.applyData(t, n);
	}
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/mdast-util-to-hast/13.2.1/0601b1701843aa9b863db3511935f9f2758af7297665ae75d5d7ed3bcb3cad64/node_modules/mdast-util-to-hast/lib/revert.js
function zd(e, t) {
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
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/mdast-util-to-hast/13.2.1/0601b1701843aa9b863db3511935f9f2758af7297665ae75d5d7ed3bcb3cad64/node_modules/mdast-util-to-hast/lib/handlers/image-reference.js
function Bd(e, t) {
	let n = String(t.identifier).toUpperCase(), r = e.definitionById.get(n);
	if (!r) return zd(e, t);
	let i = {
		src: lc(r.url || ""),
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
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/mdast-util-to-hast/13.2.1/0601b1701843aa9b863db3511935f9f2758af7297665ae75d5d7ed3bcb3cad64/node_modules/mdast-util-to-hast/lib/handlers/image.js
function Vd(e, t) {
	let n = { src: lc(t.url) };
	t.alt !== null && t.alt !== void 0 && (n.alt = t.alt), t.title !== null && t.title !== void 0 && (n.title = t.title);
	let r = {
		type: "element",
		tagName: "img",
		properties: n,
		children: []
	};
	return e.patch(t, r), e.applyData(t, r);
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/mdast-util-to-hast/13.2.1/0601b1701843aa9b863db3511935f9f2758af7297665ae75d5d7ed3bcb3cad64/node_modules/mdast-util-to-hast/lib/handlers/inline-code.js
function Hd(e, t) {
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
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/mdast-util-to-hast/13.2.1/0601b1701843aa9b863db3511935f9f2758af7297665ae75d5d7ed3bcb3cad64/node_modules/mdast-util-to-hast/lib/handlers/link-reference.js
function Ud(e, t) {
	let n = String(t.identifier).toUpperCase(), r = e.definitionById.get(n);
	if (!r) return zd(e, t);
	let i = { href: lc(r.url || "") };
	r.title !== null && r.title !== void 0 && (i.title = r.title);
	let a = {
		type: "element",
		tagName: "a",
		properties: i,
		children: e.all(t)
	};
	return e.patch(t, a), e.applyData(t, a);
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/mdast-util-to-hast/13.2.1/0601b1701843aa9b863db3511935f9f2758af7297665ae75d5d7ed3bcb3cad64/node_modules/mdast-util-to-hast/lib/handlers/link.js
function Wd(e, t) {
	let n = { href: lc(t.url) };
	t.title !== null && t.title !== void 0 && (n.title = t.title);
	let r = {
		type: "element",
		tagName: "a",
		properties: n,
		children: e.all(t)
	};
	return e.patch(t, r), e.applyData(t, r);
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/mdast-util-to-hast/13.2.1/0601b1701843aa9b863db3511935f9f2758af7297665ae75d5d7ed3bcb3cad64/node_modules/mdast-util-to-hast/lib/handlers/list-item.js
function Gd(e, t, n) {
	let r = e.all(t), i = n ? Kd(n) : qd(t), a = {}, o = [];
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
function Kd(e) {
	let t = !1;
	if (e.type === "list") {
		t = e.spread || !1;
		let n = e.children, r = -1;
		for (; !t && ++r < n.length;) t = qd(n[r]);
	}
	return t;
}
function qd(e) {
	return e.spread ?? e.children.length > 1;
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/mdast-util-to-hast/13.2.1/0601b1701843aa9b863db3511935f9f2758af7297665ae75d5d7ed3bcb3cad64/node_modules/mdast-util-to-hast/lib/handlers/list.js
function Jd(e, t) {
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
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/mdast-util-to-hast/13.2.1/0601b1701843aa9b863db3511935f9f2758af7297665ae75d5d7ed3bcb3cad64/node_modules/mdast-util-to-hast/lib/handlers/paragraph.js
function Yd(e, t) {
	let n = {
		type: "element",
		tagName: "p",
		properties: {},
		children: e.all(t)
	};
	return e.patch(t, n), e.applyData(t, n);
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/mdast-util-to-hast/13.2.1/0601b1701843aa9b863db3511935f9f2758af7297665ae75d5d7ed3bcb3cad64/node_modules/mdast-util-to-hast/lib/handlers/root.js
function Xd(e, t) {
	let n = {
		type: "root",
		children: e.wrap(e.all(t))
	};
	return e.patch(t, n), e.applyData(t, n);
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/mdast-util-to-hast/13.2.1/0601b1701843aa9b863db3511935f9f2758af7297665ae75d5d7ed3bcb3cad64/node_modules/mdast-util-to-hast/lib/handlers/strong.js
function Zd(e, t) {
	let n = {
		type: "element",
		tagName: "strong",
		properties: {},
		children: e.all(t)
	};
	return e.patch(t, n), e.applyData(t, n);
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/mdast-util-to-hast/13.2.1/0601b1701843aa9b863db3511935f9f2758af7297665ae75d5d7ed3bcb3cad64/node_modules/mdast-util-to-hast/lib/handlers/table.js
function Qd(e, t) {
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
		}, a = uo(t.children[1]), o = lo(t.children[t.children.length - 1]);
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
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/mdast-util-to-hast/13.2.1/0601b1701843aa9b863db3511935f9f2758af7297665ae75d5d7ed3bcb3cad64/node_modules/mdast-util-to-hast/lib/handlers/table-row.js
function $d(e, t, n) {
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
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/mdast-util-to-hast/13.2.1/0601b1701843aa9b863db3511935f9f2758af7297665ae75d5d7ed3bcb3cad64/node_modules/mdast-util-to-hast/lib/handlers/table-cell.js
function ef(e, t) {
	let n = {
		type: "element",
		tagName: "td",
		properties: {},
		children: e.all(t)
	};
	return e.patch(t, n), e.applyData(t, n);
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/trim-lines/3.0.1/3e05e1ebee3bd4d32745ba97f18d7b40732dc36fb1129ef08059a5ff4e875642/node_modules/trim-lines/index.js
var tf = 9, nf = 32;
function rf(e) {
	let t = String(e), n = /\r?\n|\r/g, r = n.exec(t), i = 0, a = [];
	for (; r;) a.push(af(t.slice(i, r.index), i > 0, !0), r[0]), i = r.index + r[0].length, r = n.exec(t);
	return a.push(af(t.slice(i), i > 0, !1)), a.join("");
}
function af(e, t, n) {
	let r = 0, i = e.length;
	if (t) {
		let t = e.codePointAt(r);
		for (; t === tf || t === nf;) r++, t = e.codePointAt(r);
	}
	if (n) {
		let t = e.codePointAt(i - 1);
		for (; t === tf || t === nf;) i--, t = e.codePointAt(i - 1);
	}
	return i > r ? e.slice(r, i) : "";
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/mdast-util-to-hast/13.2.1/0601b1701843aa9b863db3511935f9f2758af7297665ae75d5d7ed3bcb3cad64/node_modules/mdast-util-to-hast/lib/handlers/text.js
function of(e, t) {
	let n = {
		type: "text",
		value: rf(String(t.value))
	};
	return e.patch(t, n), e.applyData(t, n);
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/mdast-util-to-hast/13.2.1/0601b1701843aa9b863db3511935f9f2758af7297665ae75d5d7ed3bcb3cad64/node_modules/mdast-util-to-hast/lib/handlers/thematic-break.js
function sf(e, t) {
	let n = {
		type: "element",
		tagName: "hr",
		properties: {},
		children: []
	};
	return e.patch(t, n), e.applyData(t, n);
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/mdast-util-to-hast/13.2.1/0601b1701843aa9b863db3511935f9f2758af7297665ae75d5d7ed3bcb3cad64/node_modules/mdast-util-to-hast/lib/handlers/index.js
var cf = {
	blockquote: jd,
	break: Md,
	code: Nd,
	delete: Pd,
	emphasis: Fd,
	footnoteReference: Id,
	heading: Ld,
	html: Rd,
	imageReference: Bd,
	image: Vd,
	inlineCode: Hd,
	linkReference: Ud,
	link: Wd,
	listItem: Gd,
	list: Jd,
	paragraph: Yd,
	root: Xd,
	strong: Zd,
	table: Qd,
	tableCell: ef,
	tableRow: $d,
	text: of,
	thematicBreak: sf,
	toml: lf,
	yaml: lf,
	definition: lf,
	footnoteDefinition: lf
};
function lf() {}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/mdast-util-to-hast/13.2.1/0601b1701843aa9b863db3511935f9f2758af7297665ae75d5d7ed3bcb3cad64/node_modules/mdast-util-to-hast/lib/footer.js
function uf(e, t) {
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
function df(e, t) {
	return "Back to reference " + (e + 1) + (t > 1 ? "-" + t : "");
}
function ff(e) {
	let t = typeof e.options.clobberPrefix == "string" ? e.options.clobberPrefix : "user-content-", n = e.options.footnoteBackContent || uf, r = e.options.footnoteBackLabel || df, i = e.options.footnoteLabel || "Footnotes", a = e.options.footnoteLabelTagName || "h2", o = e.options.footnoteLabelProperties || { className: ["sr-only"] }, s = [], c = -1;
	for (; ++c < e.footnoteOrder.length;) {
		let i = e.footnoteById.get(e.footnoteOrder[c]);
		if (!i) continue;
		let a = e.all(i), o = String(i.identifier).toUpperCase(), l = lc(o.toLowerCase()), u = 0, d = [], f = e.footnoteCounts.get(o);
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
					...We(o),
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
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/mdast-util-to-hast/13.2.1/0601b1701843aa9b863db3511935f9f2758af7297665ae75d5d7ed3bcb3cad64/node_modules/mdast-util-to-hast/lib/state.js
var pf = {}.hasOwnProperty, mf = {};
function hf(e, t) {
	let n = t || mf, r = /* @__PURE__ */ new Map(), i = /* @__PURE__ */ new Map(), a = {
		all: s,
		applyData: _f,
		definitionById: r,
		footnoteById: i,
		footnoteCounts: /* @__PURE__ */ new Map(),
		footnoteOrder: [],
		handlers: {
			...cf,
			...n.handlers
		},
		one: o,
		options: n,
		patch: gf,
		wrap: yf
	};
	return x(e, function(e) {
		if (e.type === "definition" || e.type === "footnoteDefinition") {
			let t = e.type === "definition" ? r : i, n = String(e.identifier).toUpperCase();
			t.has(n) || t.set(n, e);
		}
	}), a;
	function o(e, t) {
		let n = e.type, r = a.handlers[n];
		if (pf.call(a.handlers, n) && r) return r(a, e, t);
		if (a.options.passThrough && a.options.passThrough.includes(n)) {
			if ("children" in e) {
				let { children: t, ...n } = e, r = We(n);
				return r.children = a.all(e), r;
			}
			return We(e);
		}
		return (a.options.unknownHandler || vf)(a, e, t);
	}
	function s(e) {
		let t = [];
		if ("children" in e) {
			let n = e.children, r = -1;
			for (; ++r < n.length;) {
				let i = a.one(n[r], e);
				if (i) {
					if (r && n[r - 1].type === "break" && (!Array.isArray(i) && i.type === "text" && (i.value = bf(i.value)), !Array.isArray(i) && i.type === "element")) {
						let e = i.children[0];
						e && e.type === "text" && (e.value = bf(e.value));
					}
					Array.isArray(i) ? t.push(...i) : t.push(i);
				}
			}
		}
		return t;
	}
}
function gf(e, t) {
	e.position && (t.position = po(e));
}
function _f(e, t) {
	let n = t;
	if (e && e.data) {
		let t = e.data.hName, r = e.data.hChildren, i = e.data.hProperties;
		typeof t == "string" && (n.type === "element" ? n.tagName = t : n = {
			type: "element",
			tagName: t,
			properties: {},
			children: "children" in n ? n.children : [n]
		}), n.type === "element" && i && Object.assign(n.properties, We(i)), "children" in n && n.children && r != null && (n.children = r);
	}
	return n;
}
function vf(e, t) {
	let n = t.data || {}, r = "value" in t && !(pf.call(n, "hProperties") || pf.call(n, "hChildren")) ? {
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
function yf(e, t) {
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
function bf(e) {
	let t = 0, n = e.charCodeAt(t);
	for (; n === 9 || n === 32;) t++, n = e.charCodeAt(t);
	return e.slice(t);
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/mdast-util-to-hast/13.2.1/0601b1701843aa9b863db3511935f9f2758af7297665ae75d5d7ed3bcb3cad64/node_modules/mdast-util-to-hast/lib/index.js
function xf(e, t) {
	let n = hf(e, t), r = n.one(e, void 0), i = ff(n), a = Array.isArray(r) ? {
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
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/remark-rehype/11.1.2/413cca51fa6919c2b8d1f13be7e3dcd4b61ffe5e743b854a370ca50359ab2f05/node_modules/remark-rehype/lib/index.js
function Sf(e, t) {
	return e && "run" in e ? async function(n, r) {
		let i = xf(n, {
			file: r,
			...t
		});
		await e.run(i, r);
	} : function(n, r) {
		return xf(n, {
			file: r,
			...e || t
		});
	};
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/bail/2.0.2/838d1c0654e3975d06f7e82c73b61b5a08f8fc8de6d14d08bf3bb0164fb24956/node_modules/bail/index.js
function Cf(e) {
	if (e) throw e;
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/extend/3.0.2/5eb62444e2a85cf73742fe2e1f08618c58e4d2cc6aae35a1cde37dd81613cc59/node_modules/extend/index.js
var wf = /* @__PURE__ */ t(((e, t) => {
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
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/is-plain-obj/4.1.0/d7fab263ffdd134b8e9e6130a60458baf7017c81e96947d5a2e6383c3d47b281/node_modules/is-plain-obj/index.js
function Tf(e) {
	if (typeof e != "object" || !e) return !1;
	let t = Object.getPrototypeOf(e);
	return (t === null || t === Object.prototype || Object.getPrototypeOf(t) === null) && !(Symbol.toStringTag in e) && !(Symbol.iterator in e);
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/trough/2.2.0/d3fda8c9b431ad5b37c76eca267b819119248e1cd71c2d1cb0490532cebf1f16/node_modules/trough/lib/index.js
function Ef() {
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
				t = o, s ? Df(s, i)(...o) : r(null, ...o);
			}
		}
	}
	function r(n) {
		if (typeof n != "function") throw TypeError("Expected `middelware` to be a function, not " + n);
		return e.push(n), t;
	}
}
function Df(e, t) {
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
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/vfile/6.0.3/c95f08038d7738c3693dc5b3e6949fc88e9e20afbbd4556e89a70570a86279d5/node_modules/vfile/lib/minpath.browser.js
var Of = {
	basename: kf,
	dirname: Af,
	extname: jf,
	join: Mf,
	sep: "/"
};
function kf(e, t) {
	if (t !== void 0 && typeof t != "string") throw TypeError("\"ext\" argument must be a string");
	Ff(e);
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
function Af(e) {
	if (Ff(e), e.length === 0) return ".";
	let t = -1, n = e.length, r;
	for (; --n;) if (e.codePointAt(n) === 47) {
		if (r) {
			t = n;
			break;
		}
	} else r ||= !0;
	return t < 0 ? e.codePointAt(0) === 47 ? "/" : "." : t === 1 && e.codePointAt(0) === 47 ? "//" : e.slice(0, t);
}
function jf(e) {
	Ff(e);
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
function Mf(...e) {
	let t = -1, n;
	for (; ++t < e.length;) Ff(e[t]), e[t] && (n = n === void 0 ? e[t] : n + "/" + e[t]);
	return n === void 0 ? "." : Nf(n);
}
function Nf(e) {
	Ff(e);
	let t = e.codePointAt(0) === 47, n = Pf(e, !t);
	return n.length === 0 && !t && (n = "."), n.length > 0 && e.codePointAt(e.length - 1) === 47 && (n += "/"), t ? "/" + n : n;
}
function Pf(e, t) {
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
function Ff(e) {
	if (typeof e != "string") throw TypeError("Path must be a string. Received " + JSON.stringify(e));
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/vfile/6.0.3/c95f08038d7738c3693dc5b3e6949fc88e9e20afbbd4556e89a70570a86279d5/node_modules/vfile/lib/minproc.browser.js
var If = { cwd: Lf };
function Lf() {
	return "/";
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/vfile/6.0.3/c95f08038d7738c3693dc5b3e6949fc88e9e20afbbd4556e89a70570a86279d5/node_modules/vfile/lib/minurl.shared.js
function Rf(e) {
	return !!(typeof e == "object" && e && "href" in e && e.href && "protocol" in e && e.protocol && e.auth === void 0);
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/vfile/6.0.3/c95f08038d7738c3693dc5b3e6949fc88e9e20afbbd4556e89a70570a86279d5/node_modules/vfile/lib/minurl.browser.js
function zf(e) {
	if (typeof e == "string") e = new URL(e);
	else if (!Rf(e)) {
		let t = /* @__PURE__ */ TypeError("The \"path\" argument must be of type string or an instance of URL. Received `" + e + "`");
		throw t.code = "ERR_INVALID_ARG_TYPE", t;
	}
	if (e.protocol !== "file:") {
		let e = /* @__PURE__ */ TypeError("The URL must be of scheme file");
		throw e.code = "ERR_INVALID_URL_SCHEME", e;
	}
	return Bf(e);
}
function Bf(e) {
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
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/vfile/6.0.3/c95f08038d7738c3693dc5b3e6949fc88e9e20afbbd4556e89a70570a86279d5/node_modules/vfile/lib/index.js
var Vf = [
	"history",
	"path",
	"basename",
	"stem",
	"extname",
	"dirname"
], Hf = class {
	constructor(e) {
		let t;
		t = e ? Rf(e) ? { path: e } : typeof e == "string" || Kf(e) ? { value: e } : e : {}, this.cwd = "cwd" in t ? "" : If.cwd(), this.data = {}, this.history = [], this.messages = [], this.value, this.map, this.result, this.stored;
		let n = -1;
		for (; ++n < Vf.length;) {
			let e = Vf[n];
			e in t && t[e] !== void 0 && t[e] !== null && (this[e] = e === "history" ? [...t[e]] : t[e]);
		}
		let r;
		for (r in t) Vf.includes(r) || (this[r] = t[r]);
	}
	get basename() {
		return typeof this.path == "string" ? Of.basename(this.path) : void 0;
	}
	set basename(e) {
		Wf(e, "basename"), Uf(e, "basename"), this.path = Of.join(this.dirname || "", e);
	}
	get dirname() {
		return typeof this.path == "string" ? Of.dirname(this.path) : void 0;
	}
	set dirname(e) {
		Gf(this.basename, "dirname"), this.path = Of.join(e || "", this.basename);
	}
	get extname() {
		return typeof this.path == "string" ? Of.extname(this.path) : void 0;
	}
	set extname(e) {
		if (Uf(e, "extname"), Gf(this.dirname, "extname"), e) {
			if (e.codePointAt(0) !== 46) throw Error("`extname` must start with `.`");
			if (e.includes(".", 1)) throw Error("`extname` cannot contain multiple dots");
		}
		this.path = Of.join(this.dirname, this.stem + (e || ""));
	}
	get path() {
		return this.history[this.history.length - 1];
	}
	set path(e) {
		Rf(e) && (e = zf(e)), Wf(e, "path"), this.path !== e && this.history.push(e);
	}
	get stem() {
		return typeof this.path == "string" ? Of.basename(this.path, this.extname) : void 0;
	}
	set stem(e) {
		Wf(e, "stem"), Uf(e, "stem"), this.path = Of.join(this.dirname || "", e + (this.extname || ""));
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
		let r = new fe(e, t, n);
		return this.path && (r.name = this.path + ":" + r.name, r.file = this.path), r.fatal = !1, this.messages.push(r), r;
	}
	toString(e) {
		return this.value === void 0 ? "" : typeof this.value == "string" ? this.value : new TextDecoder(e || void 0).decode(this.value);
	}
};
function Uf(e, t) {
	if (e && e.includes(Of.sep)) throw Error("`" + t + "` cannot be a path: did not expect `" + Of.sep + "`");
}
function Wf(e, t) {
	if (!e) throw Error("`" + t + "` cannot be empty");
}
function Gf(e, t) {
	if (!e) throw Error("Setting `" + t + "` requires `path` to be set too");
}
function Kf(e) {
	return !!(e && typeof e == "object" && "byteLength" in e && "byteOffset" in e);
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/unified/11.0.5/bc453fd1ce38b454018aa30afab98a43140ddb9d705be13a6e19f903ecb01c2d/node_modules/unified/lib/callable-instance.js
var qf = (function(e) {
	let t = this.constructor.prototype, n = t[e], r = function() {
		return n.apply(r, arguments);
	};
	return Object.setPrototypeOf(r, t), r;
}), Jf = /* @__PURE__ */ e(wf(), 1), Yf = {}.hasOwnProperty, Xf = new class e extends qf {
	constructor() {
		super("copy"), this.Compiler = void 0, this.Parser = void 0, this.attachers = [], this.compiler = void 0, this.freezeIndex = -1, this.frozen = void 0, this.namespace = {}, this.parser = void 0, this.transformers = Ef();
	}
	copy() {
		let t = new e(), n = -1;
		for (; ++n < this.attachers.length;) {
			let e = this.attachers[n];
			t.use(...e);
		}
		return t.data((0, Jf.default)(!0, {}, this.namespace)), t;
	}
	data(e, t) {
		return typeof e == "string" ? arguments.length === 2 ? ($f("data", this.frozen), this.namespace[e] = t, this) : Yf.call(this.namespace, e) && this.namespace[e] || void 0 : e ? ($f("data", this.frozen), this.namespace = e, this) : this.namespace;
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
		let t = np(e), n = this.parser || this.Parser;
		return Zf("parse", n), n(String(t), t);
	}
	process(e, t) {
		let n = this;
		return this.freeze(), Zf("process", this.parser || this.Parser), Qf("process", this.compiler || this.Compiler), t ? r(void 0, t) : new Promise(r);
		function r(r, i) {
			let a = np(e), o = n.parse(a);
			n.run(o, a, function(e, t, r) {
				if (e || !t || !r) return s(e);
				let i = t, a = n.stringify(i, r);
				ip(a) ? r.value = a : r.result = a, s(e, r);
			});
			function s(e, n) {
				e || !n ? i(e) : r ? r(n) : t(void 0, n);
			}
		}
	}
	processSync(e) {
		let t = !1, n;
		return this.freeze(), Zf("processSync", this.parser || this.Parser), Qf("processSync", this.compiler || this.Compiler), this.process(e, r), tp("processSync", "process", t), n;
		function r(e, r) {
			t = !0, Cf(e), n = r;
		}
	}
	run(e, t, n) {
		ep(e), this.freeze();
		let r = this.transformers;
		return !n && typeof t == "function" && (n = t, t = void 0), n ? i(void 0, n) : new Promise(i);
		function i(i, a) {
			let o = np(t);
			r.run(e, o, s);
			function s(t, r, o) {
				let s = r || e;
				t ? a(t) : i ? i(s) : n(void 0, s, o);
			}
		}
	}
	runSync(e, t) {
		let n = !1, r;
		return this.run(e, t, i), tp("runSync", "run", n), r;
		function i(e, t) {
			Cf(e), r = t, n = !0;
		}
	}
	stringify(e, t) {
		this.freeze();
		let n = np(t), r = this.compiler || this.Compiler;
		return Qf("stringify", r), ep(e), r(e, n);
	}
	use(e, ...t) {
		let n = this.attachers, r = this.namespace;
		if ($f("use", this.frozen), e != null) {
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
			o(e.plugins), e.settings && (r.settings = (0, Jf.default)(!0, r.settings, e.settings));
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
				Tf(o) && Tf(r) && (r = (0, Jf.default)(!0, o, r)), n[i] = [
					e,
					r,
					...a
				];
			}
		}
	}
}().freeze();
function Zf(e, t) {
	if (typeof t != "function") throw TypeError("Cannot `" + e + "` without `parser`");
}
function Qf(e, t) {
	if (typeof t != "function") throw TypeError("Cannot `" + e + "` without `compiler`");
}
function $f(e, t) {
	if (t) throw Error("Cannot call `" + e + "` on a frozen processor.\nCreate a new processor first, by calling it: use `processor()` instead of `processor`.");
}
function ep(e) {
	if (!Tf(e) || typeof e.type != "string") throw TypeError("Expected node, got `" + e + "`");
}
function tp(e, t, n) {
	if (!n) throw Error("`" + e + "` finished async. Use `" + t + "` instead");
}
function np(e) {
	return rp(e) ? e : new Hf(e);
}
function rp(e) {
	return !!(e && typeof e == "object" && "message" in e && "messages" in e);
}
function ip(e) {
	return typeof e == "string" || ap(e);
}
function ap(e) {
	return !!(e && typeof e == "object" && "byteLength" in e && "byteOffset" in e);
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/marked/18.1.3/d8ff126ff7ac2fca41a9ce7d860be7e920fb88e13835eb7e75330ba4a6584397/node_modules/marked/lib/marked.esm.js
var op = /* @__PURE__ */ e(i(), 1), K = r();
function sp() {
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
var cp = sp();
function lp(e) {
	cp = e;
}
var up = { exec: () => null };
function dp(e) {
	let t = [];
	return (n) => {
		let r = Math.max(0, Math.min(3, n - 1)), i = t[r];
		return i || (i = e(r), t[r] = i), i;
	};
}
function q(e, t = "") {
	let n = typeof e == "string" ? e : e.source, r = {
		replace: (e, t) => {
			let i = typeof t == "string" ? t : t.source;
			return i = i.replace(J.caret, "$1"), n = n.replace(e, i), r;
		},
		getRegex: () => new RegExp(n, t)
	};
	return r;
}
var fp = ((e = "") => {
	try {
		return !!RegExp("(?<=1)(?<!1)" + e);
	} catch {
		return !1;
	}
})(), J = {
	codeRemoveIndent: /^(?: {0,3}\t| {1,4})/gm,
	outputLinkReplace: /\\([\[\]])/g,
	indentCodeCompensation: /^(\s+)(?:```)/,
	beginningSpace: /^\s+/,
	endingHash: /#$/,
	startingSpaceChar: /^ /,
	endingSpaceChar: / $/,
	endingSpaceTabChar: /[ \t]$/,
	nonSpaceChar: /[^ ]/,
	newLineCharGlobal: /\n/g,
	tabCharGlobal: /\t/g,
	leadingSpaceTab: /^[ \t]+/,
	multipleSpaceGlobal: /\s+/g,
	blankLine: /^[ \t]*$/,
	doubleBlankLine: /\n[ \t]*\n[ \t]*$/,
	blockquoteStart: /^ {0,3}>/,
	blockquoteSetextReplace: /\n {0,3}((?:=+|-+)[ \t]*)(?=\n|$)/g,
	blockquoteSetextReplace2: /^ {0,3}>[ \t]?/gm,
	blockquoteMarkerWhitespace: /^( {0,3}>)([ \t]+)/gm,
	listReplaceNesting: /^ {1,4}(?=( {4})*[^ ])/g,
	listIsTask: /^\[[ xX]\][ \t]+\S/,
	listReplaceTask: /^\[[ xX]\][ \t]+/,
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
	numericCharacterReference: /&#(?:(\d{1,7})|[Xx]([A-Fa-f0-9]{1,6}));/,
	numericCharacterReferenceOrAmpersand: /&#(?:(\d{1,7})|[Xx]([A-Fa-f0-9]{1,6}));|&(?!(#\d{1,7}|#[Xx][a-fA-F0-9]{1,6}|\w+);)/g,
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
	nextBulletRegex: dp((e) => RegExp(`^ {0,${e}}(?:[*+-]|\\d{1,9}[.)])((?:[ 	][^\\n]*)?(?:\\n|$))`)),
	hrRegex: dp((e) => RegExp(`^ {0,${e}}((?:-[ 	]*){3,}|(?:_[ 	]*){3,}|(?:\\*[ 	]*){3,})(?:\\n+|$)`)),
	fencesBeginRegex: dp((e) => RegExp(`^ {0,${e}}(?:\`\`\`|~~~)`)),
	headingBeginRegex: dp((e) => RegExp(`^ {0,${e}}#{1,6}(?:\\s|$)`)),
	pedanticHeadingBeginRegex: dp((e) => RegExp(`^ {0,${e}}#`)),
	htmlBeginRegex: dp((e) => RegExp(`^ {0,${e}}(?:</?(?:${Dp})(?: +|$|/?>)|<(?:(?:script|pre|style|textarea)(?:[\\s>]|$)|!--)|<\\?|<![A-Za-z]|<!\\[CDATA\\[)`, "i")),
	blockquoteBeginRegex: dp((e) => RegExp(`^ {0,${e}}>`))
}, pp = /^(?:[ \t]*(?:\n|$))+/, mp = /^((?: {4}| {0,3}\t)[^\n]+(?:\n(?:[ \t]*(?:\n|$))*)?)+/, hp = /^ {0,3}(`{3,}(?=[^`\n]*(?:\n|$))|~{3,})([^\n]*)(?:\n|$)(?:|([\s\S]*?)(?:\n|$))(?: {0,3}(?:(?=`)\1`*|(?=~)\1~*) *(?=\n|$)|$)/, gp = /^ {0,3}((?:-[\t ]*){3,}|(?:_[ \t]*){3,}|(?:\*[ \t]*){3,})(?:\n+|$)/, _p = /^ {0,3}(#{1,6})(?=\s|$)(.*)(?:\n+|$)/, vp = / {0,3}(?:[*+-]|\d{1,9}[.)])/, yp = /^(?!bull |blockCode|fences|blockquote|heading|html|table)((?:.|\n(?!\s*?\n|bull |fences|blockquote|heading|hr|html|table))+?)\n {0,3}(=+|-+)[ \t]*(?:\n+|$)/, bp = q(yp).replace(/bull/g, vp).replace(/blockCode/g, /(?: {4}| {0,3}\t)/).replace(/fences/g, / {0,3}(?:`{3,}|~{3,})/).replace(/blockquote/g, / {0,3}>/).replace(/heading/g, / {0,3}#{1,6}(?:\s|$)/).replace(/hr/g, / {0,3}(?:(?:-[\t ]*){3,}|(?:_[ \t]*){3,}|(?:\*[ \t]*){3,})(?:\n+|$)/).replace(/html/g, / {0,3}<[^\n>]+>\n/).replace(/\|table/g, "").getRegex(), xp = q(yp).replace(/bull/g, vp).replace(/blockCode/g, /(?: {4}| {0,3}\t)/).replace(/fences/g, / {0,3}(?:`{3,}|~{3,})/).replace(/blockquote/g, / {0,3}>/).replace(/heading/g, / {0,3}#{1,6}(?:\s|$)/).replace(/hr/g, / {0,3}(?:(?:-[\t ]*){3,}|(?:_[ \t]*){3,}|(?:\*[ \t]*){3,})(?:\n+|$)/).replace(/html/g, / {0,3}<[^\n>]+>\n/).replace(/table/g, / {0,3}\|?(?:[:\- ]*\|)+[\:\- ]*\n/).getRegex(), Sp = /^([^\n]+(?:\n(?!hr|heading|lheading|blockquote|fences|list|html|table|[ \t]+\n)[^\n]+)*)/, Cp = /^[^\n]+/, wp = /(?!\s*\])(?:\\[\s\S]|[^\[\]\\\n]|\n(?![ \t]*\n))+/, Tp = q(/^ {0,3}\[(label)\]: *(?:\n[ \t]*)?([^<\s][^\s]*|<.*?>)(?:(?: +(?:\n[ \t]*)?| *\n[ \t]*)(title))? *(?:\n+|$)/).replace("label", wp).replace("title", /(?:"(?:\\[\s\S]|[^"\\])*"|'(?:\\[\s\S]|[^'\\\n]|\n(?!\n))*'|\((?:\\[\s\S]|[^()\\])*\))/).getRegex(), Ep = q(/^(bull)([ \t][^\n]*?)?(?:\n|$)/).replace(/bull/g, vp).getRegex(), Dp = "address|article|aside|base|basefont|blockquote|body|caption|center|col|colgroup|dd|details|dialog|dir|div|dl|dt|fieldset|figcaption|figure|footer|form|frame|frameset|h[1-6]|head|header|hr|html|iframe|legend|li|link|main|menu|menuitem|meta|nav|noframes|ol|optgroup|option|p|param|search|section|summary|table|tbody|td|tfoot|th|thead|title|tr|track|ul", Op = /<!--(?:-?>|[\s\S]*?(?:-->|$))/, kp = q("^ {0,3}(?:<(script|pre|style|textarea)[\\s>][\\s\\S]*?(?:</\\1>[^\\n]*\\n*|$)|comment[^\\n]*(\\n+|$)|<\\?[\\s\\S]*?(?:\\?>[^\\n]*\\n*|$)|<![A-Z][\\s\\S]*?(?:>[^\\n]*\\n*|$)|<!\\[CDATA\\[[\\s\\S]*?(?:\\]\\]>[^\\n]*\\n*|$)|</?(tag)(?:[ \\t]+|\\n|/?>|$)[\\s\\S]*?(?:(?:\\n[ 	]*)+\\n|$)|<(?!(?:script|pre|style|textarea)(?![a-z0-9-]))([a-z][a-z0-9-]*)(?:attribute)*? */?>(?=[ \\t]*(?:\\n|$))[\\s\\S]*?(?:(?:\\n[ 	]*)+\\n|$)|</[a-z][a-z0-9-]*\\s*>(?=[ \\t]*(?:\\n|$))[\\s\\S]*?(?:(?:\\n[ 	]*)+\\n|$))", "i").replace("comment", Op).replace("tag", Dp).replace("attribute", / +[a-zA-Z:_][\w.:-]*(?: *= *"[^"\n]*"| *= *'[^'\n]*'| *= *[^\s"'=<>`]+)?/).getRegex(), Ap = (e) => q(Sp).replace("hr", gp).replace("heading", " {0,3}#{1,6}(?:\\s|$)").replace("|lheading", "").replace("|table", "").replace("blockquote", " {0,3}>").replace("fences", " {0,3}(?:`{3,}(?=[^`\\n]*(?:\\n|$))|~~~)[^\\n]*(?:\\n|$)").replace("list", e).replace("html", " {0,3}(?:</?(?:tag)(?: +|\\n|/?>)|<(?:(?:script|pre|style|textarea)(?:[\\s>]|$)|!--)|<\\?|<![A-Za-z]|<!\\[CDATA\\[)").replace("tag", Dp).getRegex(), jp = Ap(/ {0,3}(?:[*+-]|1[.)])[ \t]+[^ \t\n]/), Mp = Ap(/ {0,3}(?:[*+-]|\d{1,9}[.)])(?:[ \t]|\n|$)/), Np = {
	blockquote: q(/^( {0,3}> ?(?:(?![ \t]*(?:\n|$))(paragraph)|[^\n]*)(?:\n|$))+/).replace("paragraph", Mp).getRegex(),
	code: mp,
	def: Tp,
	fences: hp,
	heading: _p,
	hr: gp,
	html: kp,
	lheading: bp,
	list: Ep,
	newline: pp,
	paragraph: jp,
	table: up,
	text: Cp
}, Pp = q("^ *([^\\n ].*)\\n {0,3}((?:\\| *)?:?-+:? *(?:\\| *:?-+:? *)*(?:\\| *)?)(?:\\n((?:(?! *\\n|hr|heading|blockquote|code|fences|list|html).*(?:\\n|$))*)\\n*|$)").replace("hr", gp).replace("heading", " {0,3}#{1,6}(?:\\s|$)").replace("blockquote", " {0,3}>").replace("code", "(?: {4}| {0,3}	)[^\\n]").replace("fences", " {0,3}(?:`{3,}(?=[^`\\n]*(?:\\n|$))|~~~)[^\\n]*(?:\\n|$)").replace("list", " {0,3}(?:[*+-]|1[.)])[ \\t]").replace("html", " {0,3}(?:</?(?:tag)(?: +|\\n|/?>)|<(?:(?:script|pre|style|textarea)(?:[\\s>]|$)|!--)|<\\?|<![A-Za-z]|<!\\[CDATA\\[)").replace("tag", Dp).getRegex(), Fp = {
	...Np,
	lheading: xp,
	table: Pp,
	paragraph: q(Sp).replace("hr", gp).replace("heading", " {0,3}#{1,6}(?:\\s|$)").replace("|lheading", "").replace("table", Pp).replace("blockquote", " {0,3}>").replace("fences", " {0,3}(?:`{3,}(?=[^`\\n]*(?:\\n|$))|~~~)[^\\n]*(?:\\n|$)").replace("list", " {0,3}(?:[*+-]|1[.)])[ \\t]+[^ \\t\\n]").replace("html", " {0,3}(?:</?(?:tag)(?: +|\\n|/?>)|<(?:(?:script|pre|style|textarea)(?:[\\s>]|$)|!--)|<\\?|<![A-Za-z]|<!\\[CDATA\\[)").replace("tag", Dp).getRegex()
}, Ip = {
	...Np,
	html: q("^ *(?:comment(?: *\\n|\\s*$)|<(tag)[\\s\\S]+?</\\1>(?: *\\n{2,}|\\s*$)|<tag(?:\"[^\"]*\"|'[^']*'|\\s[^'\"/>\\s]*)*?/?>(?: *\\n{2,}|\\s*$))").replace("comment", Op).replace(/tag/g, "(?!(?:a|em|strong|small|s|cite|q|dfn|abbr|data|time|code|var|samp|kbd|sub|sup|i|b|u|mark|ruby|rt|rp|bdi|bdo|span|br|wbr|ins|del|img)\\b)\\w+(?!:|[^\\w\\s@]*@)\\b").getRegex(),
	def: /^ *\[([^\]]+)\]: *<?([^\s>]+)>?(?: +(["(][^\n]+[")]))? *(?:\n+|$)/,
	heading: /^(#{1,6})(.*)(?:\n+|$)/,
	fences: up,
	lheading: /^(.+?)\n {0,3}(=+|-+) *(?:\n+|$)/,
	paragraph: q(Sp).replace("hr", gp).replace("heading", " *#{1,6} *[^\n]").replace("lheading", bp).replace("|table", "").replace("blockquote", " {0,3}>").replace("|fences", "").replace("|list", "").replace("|html", "").replace("|tag", "").getRegex()
}, Lp = /^\\([!"#$%&'()*+,\-./:;<=>?@\[\]\\^_`{|}~])/, Rp = /^(`+)([^`]|[^`][\s\S]*?[^`])\1(?!`)/, zp = /^( {2,}|\\)\n(?!\s*$)[ \t]*/, Bp = /^(`+|[^`])(?:(?= {2,}\n)|[\s\S]*?(?:(?=[\\<!\[`*_]|\b_|$)|[^ ](?= {2,}\n)))/, Vp = /[\p{P}\p{S}]/u, Hp = /[\s\p{P}\p{S}]/u, Up = /[^\s\p{P}\p{S}]/u, Wp = q(/^((?![*_])punctSpace)/, "u").replace(/punctSpace/g, Hp).getRegex(), Gp = /[\p{Pi}\p{Ps}"']/u, Kp = /(?!~)[\p{P}\p{S}]/u, qp = /(?!~)[\s\p{P}\p{S}]/u, Jp = /(?:[^\s\p{P}\p{S}]|~)/u, Yp = q(/link|precode-code|html/, "g").replace("link", /\[(?:[^\[\]`]|(?<a>`+)[^`]+\k<a>(?!`))*?\]\((?:\\[\s\S]|[^\\\(\)]|\((?:\\[\s\S]|[^\\\(\)])*\))*\)/).replace("precode-", fp ? "(?<!`)()" : "(^^|[^`])").replace("code", /(?<b>`+)[^`]+\k<b>(?!`)/).replace("html", /<(?! )[^<>]*?>/).getRegex(), Xp = /^(?:\*+(?:((?!\*)punct)|([^\s*]))?)|^_+(?:((?!_)punct)|([^\s_]))?/, Zp = q(Xp, "u").replace(/punct/g, Vp).getRegex(), Qp = q(Xp, "u").replace(/punct/g, Kp).getRegex(), $p = q(/^(?:\*+(?:((?!\*)(?!openQuote)punct)|([^\s*]))?)|^_+(?:((?!_)(?!openQuote)punct)|([^\s_]))?/, "u").replace(/openQuote/g, Gp).replace(/punct/g, Vp).getRegex(), em = "^[^_*]*?__[^_*]*?\\*[^_*]*?(?=__)|[^*]+(?=[^*])|(?!\\*)punct(\\*+)(?=[\\s]|$)|notPunctSpace(\\*+)(?!\\*)(?=punctSpace|$)|(?!\\*)punctSpace(\\*+)(?=notPunctSpace)|[\\s](\\*+)(?!\\*)(?=punct)|(?!\\*)punct(\\*+)(?!\\*)(?=punct)|notPunctSpace(\\*+)(?=notPunctSpace)", tm = q(em, "gu").replace(/notPunctSpace/g, Up).replace(/punctSpace/g, Hp).replace(/punct/g, Vp).getRegex(), nm = q(em, "gu").replace(/notPunctSpace/g, Jp).replace(/punctSpace/g, qp).replace(/punct/g, Kp).getRegex(), rm = q("^[^_*]*?__[^_*]*?\\*[^_*]*?(?=__)|[^*]+(?=[^*])|(?!\\*)punct(\\*+)(?=[\\s]|$)|notPunctSpace(\\*+)(?!\\*)(?=punctSpace|$)|(?!\\*)[\\s](\\*+)(?=notPunctSpace)|[\\s](\\*+)(?!\\*)(?=punct)|(?!\\*)punct(\\*+)(?!\\*)(?=punct)|(?:(?!\\*)punct|notPunctSpace)(\\*+)(?!\\*)(?=notPunctSpace)", "gu").replace(/notPunctSpace/g, Up).replace(/punctSpace/g, Hp).replace(/punct/g, Vp).getRegex(), im = q("^[^_*]*?\\*\\*[^_*]*?_[^_*]*?(?=\\*\\*)|[^_]+(?=[^_])|(?!_)punct(_+)(?=[\\s]|$)|notPunctSpace(_+)(?!_)(?=punctSpace|$)|(?!_)punctSpace(_+)(?=notPunctSpace)|[\\s](_+)(?!_)(?=punct)|(?!_)punct(_+)(?!_)(?=punct)", "gu").replace(/notPunctSpace/g, Up).replace(/punctSpace/g, Hp).replace(/punct/g, Vp).getRegex(), am = q("^[^_*]*?\\*\\*[^_*]*?_[^_*]*?(?=\\*\\*)|[^_]+(?=[^_])|(?!_)punct(_+)(?=[\\s]|$)|notPunctSpace(_+)(?!_)(?=punctSpace|$)|(?!_)[\\s](_+)(?=notPunctSpace)|[\\s](_+)(?!_)(?=punct)|(?!_)punct(_+)(?!_)(?=punct)|(?:(?!_)punct|notPunctSpace)(_+)(?!_)(?=notPunctSpace)", "gu").replace(/notPunctSpace/g, Up).replace(/punctSpace/g, Hp).replace(/punct/g, Vp).getRegex(), om = q(/^~~?(?:((?!~)punct)|[^\s~])/, "u").replace(/punct/g, Vp).getRegex(), sm = q("^[^~]+(?=[^~])|(?!~)punct(~~?)(?=[\\s]|$)|notPunctSpace(~~?)(?!~)(?=punctSpace|$)|(?!~)punctSpace(~~?)(?=notPunctSpace)|[\\s](~~?)(?!~)(?=punct)|(?!~)punct(~~?)(?!~)(?=punct)|notPunctSpace(~~?)(?=notPunctSpace)", "gu").replace(/notPunctSpace/g, Up).replace(/punctSpace/g, Hp).replace(/punct/g, Vp).getRegex(), cm = q(/\\(punct)/, "gu").replace(/punct/g, Vp).getRegex(), lm = q(/(`+)(?:[^`]|[^`][\s\S]*?[^`])\1(?!`)|`+|\\(punct)/, "gu").replace(/punct/g, Vp).getRegex(), um = q(/^<(scheme:[^\s\x00-\x1f<>]*|email)>/).replace("scheme", /[a-zA-Z][a-zA-Z0-9+.-]{1,31}/).replace("email", /[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+(@)[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)+(?![-_])/).getRegex(), dm = q(Op).replace("(?:-->|$)", "-->").getRegex(), fm = q("^comment|^</[a-zA-Z][a-zA-Z0-9-]*\\s*>|^<[a-zA-Z][a-zA-Z0-9-]*(?:attribute)*?\\s*/?>|^<\\?[\\s\\S]*?\\?>|^<![a-zA-Z]+\\s[\\s\\S]*?>|^<!\\[CDATA\\[[\\s\\S]*?\\]\\]>").replace("comment", dm).replace("attribute", /\s+[a-zA-Z:_][\w.:-]*(?:\s*=\s*"[^"]*"|\s*=\s*'[^']*'|\s*=\s*[^\s"'=<>`]+)?/).getRegex(), pm = /\[(?:\\[\s\S]|[^\[\]\\])*\]/, mm = q(/(?:\[(?:brackets|\\[\s\S]|[^\[\]\\])*\]|\\[\s\S]|`+(?!`)[^`]*?`+(?!`)|``+(?=\])|[^\[\]\\`])*?/).replace("brackets", pm).getRegex(), hm = q(/^!?\[(label)\]\([ \t\n]*(href)(?:(?:[ \t]+(?:\n[ \t]*)?|\n[ \t]*)(title))?[ \t\n]*\)/).replace("label", mm).replace("href", /<(?:\\.|[^\n<>\\])+>|[^ \t\n\x00-\x1f]+|(?=\))/).replace("title", /"(?:\\[\s\S]|[^"\\])*"|'(?:\\[\s\S]|[^'\\])*'|\((?:\\[\s\S]|[^)\\])*\)/).getRegex(), gm = q(/^!?\[(label)\]\[(ref)\]/).replace("label", mm).replace("ref", wp).getRegex(), _m = q(/^!?\[(ref)\](?:\[\])?/).replace("ref", wp).getRegex(), vm = /(?!\s*\])(?:\\[\s\S]|[^\[\]\\]){1,999}/, ym = q(/(?:[^\[\]\\`]*(?:\[(?:brackets|\\[\s\S]|[^\[\]\\])*\]|\\[\s\S]|`+(?!`)[^`]*?`+(?!`)|``+(?=\]))){0,999}?[^\[\]\\`]*?/).replace("brackets", pm).getRegex(), bm = q("reflink|nolink(?!\\()", "g").replace("reflink", q(/^!?\[(label)\]\[(ref)\]/).replace("label", ym).replace("ref", vm).getRegex()).replace("nolink", q(/^!?\[(ref)\](?:\[\])?/).replace("ref", vm).getRegex()).getRegex(), xm = /[hH][tT][tT][pP][sS]?|[fF][tT][pP]/, Sm = q(/(?:mailto:email|xmpp:email(?:\/[A-Za-z0-9@.]+)?)/).replace(/email/g, /[A-Za-z0-9._+-]+@[a-zA-Z0-9-_]+(?:\.[a-zA-Z0-9-_]*[a-zA-Z0-9])+(?![\w-])/).getRegex(), Cm = {
	_backpedal: up,
	anyPunctuation: cm,
	anyPunctuationOutsideCode: lm,
	autolink: um,
	blockSkip: Yp,
	br: zp,
	code: Rp,
	del: up,
	delLDelim: up,
	delRDelim: up,
	emStrongLDelim: Zp,
	emStrongRDelimAst: tm,
	emStrongRDelimUnd: im,
	escape: Lp,
	link: hm,
	nolink: _m,
	punctuation: Wp,
	reflink: gm,
	reflinkSearch: bm,
	tag: fm,
	text: Bp,
	url: up
}, wm = {
	...Cm,
	emStrongLDelim: $p,
	emStrongRDelimAst: rm,
	emStrongRDelimUnd: am,
	link: q(/^!?\[(label)\]\((.*?)\)/).replace("label", mm).getRegex(),
	reflink: q(/^!?\[(label)\]\s*\[([^\]]*)\]/).replace("label", mm).getRegex()
}, Tm = {
	...Cm,
	emStrongRDelimAst: nm,
	emStrongLDelim: Qp,
	delLDelim: om,
	delRDelim: sm,
	url: q(/^emailProtocol|^((?:protocol):\/\/|www\.)(?:[a-zA-Z0-9\-]+\.?)+[^\s<]*|^email/).replace("emailProtocol", Sm).replace("protocol", xm).replace("email", /[A-Za-z0-9._+-]+(@)[a-zA-Z0-9-_]+(?:\.[a-zA-Z0-9-_]*[a-zA-Z0-9])+(?![\w-])/).getRegex(),
	_backpedal: /(?:[^?!.,:;*_'"~()&]+|\([^)]*\)|&(?![a-zA-Z0-9]+;$)|[?!.,:;*_'"~)]+(?!$))+/,
	del: /^(~~?)(?=[^\s~])((?:\\[\s\S]|[^\\])*?(?:\\[\s\S]|[^\s~\\]))\1(?=[^~]|$)/,
	text: q(/^(?:[^a-zA-Z0-9](?=emailProtocol)|(`+|~+|[^`~])(?:(?=[`~])|(?= {2,}\n)|(?=[a-zA-Z0-9.!#$%&'*+\/=?_`{\|}~-]+@)|[\s\S]*?(?:(?=[\\<!\[`*~_]|\b_|protocol:\/\/|wwwStart|$)|[^ ](?= {2,}\n)|[^a-zA-Z0-9](?=emailProtocol)|[^a-zA-Z0-9.!#$%&'*+\/=?_`{\|}~-](?=[a-zA-Z0-9.!#$%&'*+\/=?_`{\|}~-]+@))))/).replace("protocol", xm).replace(/emailProtocol/g, /(?:mailto|xmpp):/).replace("wwwStart", fp ? "(?<![^\\s*_~(])www\\." : "www\\.").getRegex()
}, Em = {
	...Tm,
	br: q(zp).replace("{2,}", "*").getRegex(),
	text: q(Tm.text).replace("\\b_", "\\b_| {2,}\\n").replace(/\{2,\}/g, "*").getRegex()
}, Dm = {
	normal: Np,
	gfm: Fp,
	pedantic: Ip
}, Om = {
	normal: Cm,
	gfm: Tm,
	breaks: Em,
	pedantic: wm
}, km = {
	"&": "&amp;",
	"<": "&lt;",
	">": "&gt;",
	"\"": "&quot;",
	"'": "&#39;"
}, Am = (e) => km[e];
function Y(e, t) {
	if (t) {
		if (J.escapeTest.test(e)) return e.replace(J.escapeReplace, Am);
	} else if (J.escapeTestNoEncode.test(e)) return e.replace(J.escapeReplaceNoEncode, Am);
	return e;
}
function jm(e) {
	return J.numericCharacterReference.test(e) ? e.replace(J.numericCharacterReferenceOrAmpersand, (e, t, n) => {
		if (t === void 0 && n === void 0) return "&amp;";
		let r = t === void 0 ? Number.parseInt(n, 16) : Number.parseInt(t, 10);
		return r === 0 || r > 1114111 || r >= 55296 && r <= 57343 ? "�" : r === 38 ? "&amp;" : String.fromCodePoint(r);
	}) : e;
}
function Mm(e) {
	try {
		e = encodeURI(e).replace(J.percentDecode, "%");
	} catch {
		return null;
	}
	return e;
}
function Nm(e, t) {
	let n = e.replace(J.findPipe, (e, t, n) => {
		let r = !1, i = t;
		for (; --i >= 0 && n[i] === "\\";) r = !r;
		return r ? "|" : " |";
	}).split(J.splitPipe), r = 0;
	if (n[0].trim() || n.shift(), n.length > 0 && !n.at(-1)?.trim() && n.pop(), t) {
		if (n.length > t) n.splice(t);
		else for (; n.length < t;) n.push("");
	}
	for (; r < n.length; r++) n[r] = n[r].trim().replace(J.slashPipe, "|");
	return n;
}
function Pm(e, t, n) {
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
function Fm(e) {
	let t = e.split("\n"), n = t.length - 1;
	for (; n >= 0 && J.blankLine.test(t[n]);) n--;
	return t.length - n <= 2 ? e : t.slice(0, n + 1).join("\n");
}
function Im(e) {
	return e.trim().replace(J.multipleSpaceGlobal, " ").toLowerCase().toUpperCase().toLowerCase();
}
function Lm(e, t) {
	if (e.indexOf(t[0]) === -1 && e.indexOf(t[1]) === -1) return -1;
	let n = 0;
	for (let r = 0; r < e.length; r++) if (e[r] === "\\") r++;
	else if (e[r] === t[0]) n++;
	else if (e[r] === t[1] && (n--, n < 0)) return r;
	return n > 0 ? -2 : -1;
}
var Rm = /* @__PURE__ */ new WeakMap();
function zm(e) {
	let t = {
		"*": -1,
		_: -1
	}, n = -1, r = e.length;
	for (let i = 0; i < r; i++) {
		let a = e[i];
		if (a !== "*" && a !== "_" && a !== "~") continue;
		let o = i === 0 ? "\n" : e[i - 1], s = i + 1;
		for (; s < r && e[s] === a;) s++;
		o !== " " && o !== "	" && o !== "\n" && o !== "\r" && o !== "\f" && (a === "~" ? n = i : t[a] = i), i = s - 1;
	}
	return {
		maskedSrc: e,
		lastPossibleEmStrongCloserIndex: t,
		lastPossibleDelCloserIndex: n
	};
}
function Bm(e, t) {
	let n = zm(t), r = Rm.get(e);
	r ? r.push(n) : Rm.set(e, [n]);
}
function Vm(e) {
	let t = Rm.get(e);
	t && (t.pop(), t.length === 0 && Rm.delete(e));
}
function Hm(e, t, n) {
	let r = Rm.get(e)?.at(-1);
	return r && r.maskedSrc === t ? r.lastPossibleEmStrongCloserIndex[n] : zm(t).lastPossibleEmStrongCloserIndex[n];
}
function Um(e, t) {
	let n = Rm.get(e)?.at(-1);
	return n && n.maskedSrc === t ? n.lastPossibleDelCloserIndex : zm(t).lastPossibleDelCloserIndex;
}
function Wm(e, t = 0) {
	let n = t, r = "";
	for (let t of e) if (t === "	") {
		let e = 4 - n % 4;
		r += " ".repeat(e), n += e;
	} else r += t, n++;
	return r;
}
function Gm(e, t, n, r, i) {
	let a = t.href, o = t.title || null, s = e[1].replace(i.other.outputLinkReplace, "$1"), c = e[0].charAt(0) === "!", l = r.state.inLink, u = r.state.linkEmitted, d = r.state.inRawBlock;
	r.state.inLink = !0, r.state.linkEmitted = !1;
	let f = r.inlineTokens(s), p = r.state.linkEmitted;
	if (r.state.linkEmitted = u, r.state.inLink = l, !c) {
		if (p) {
			r.state.inRawBlock = d;
			return;
		}
		r.state.linkEmitted = !0;
	}
	return {
		type: c ? "image" : "link",
		raw: n,
		href: a,
		title: o,
		text: s,
		tokens: f
	};
}
function Km(e, t, n) {
	let r = e.match(n.other.indentCodeCompensation);
	if (r === null) return t;
	let i = r[1];
	return t.split("\n").map((e) => {
		let t = e.match(n.other.beginningSpace);
		if (t === null) return e;
		let [r] = t;
		return e.slice(Math.min(r.length, i.length));
	}).join("\n");
}
function qm(e, t, n, r) {
	if (!t.includes("<")) return !1;
	for (let i = 0; i < t.length; i++) {
		if (t[i] === "\\") {
			i++;
			continue;
		}
		if (t[i] === "`") {
			let e = r.inline.code.exec(t.slice(i));
			if (e) {
				i += e[0].length - 1;
				continue;
			}
		}
		if (t[i] !== "<") continue;
		let a = e.slice(n + i), o = r.inline.tag.exec(a) || r.inline.autolink.exec(a);
		if (o) {
			if (o[0].length > t.length - i) return !0;
			i += o[0].length - 1;
		}
	}
	return !1;
}
var Jm = class {
	options;
	rules;
	lexer;
	constructor(e) {
		this.options = e || cp;
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
			let e = this.options.pedantic ? t[0] : Fm(t[0]);
			return {
				type: "code",
				raw: e,
				codeBlockStyle: "indented",
				text: e.replace(this.rules.other.codeRemoveIndent, "")
			};
		}
	}
	fences(e) {
		let t = this.rules.block.fences.exec(e);
		if (t) {
			let e = t[0], n = Km(e, t[3] || "", this.rules);
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
				let t = Pm(e, "#");
				(this.options.pedantic || !t || this.rules.other.endingSpaceTabChar.test(t)) && (e = t.trim());
			}
			return {
				type: "heading",
				raw: Pm(t[0], "\n"),
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
			raw: Pm(t[0], "\n")
		};
	}
	blockquote(e) {
		let t = this.rules.block.blockquote.exec(e);
		if (t) {
			let e = Pm(t[0], "\n").split("\n"), n = "", r = "", i = [];
			for (; e.length > 0;) {
				let t = !1, a = [], o = 0;
				for (; o < e.length; o++) if (this.rules.other.blockquoteStart.test(e[o])) a.push(e[o]), t = !0;
				else if (!t) a.push(e[o]);
				else break;
				e = e.slice(o);
				let s = a.join("\n"), c = s.replace(this.rules.other.blockquoteMarkerWhitespace, (e, t, n) => t + Wm(n, t.length)).replace(this.rules.other.blockquoteSetextReplace, "\n    $1").replace(this.rules.other.blockquoteSetextReplace2, "");
				n = n ? `${n}
${s}` : s, r = r ? `${r}
${c}` : c;
				let l = this.lexer.state.top;
				if (this.lexer.state.top = !0, this.lexer.blockTokens(c, i, !0), this.lexer.state.top = l, e.length === 0) break;
				let u = i.at(-1);
				if (u?.type === "code" || u?.type === "heading" || u?.type === "hr" || u?.type === "html" || u?.type === "table") break;
				if (u?.type === "blockquote") {
					let t = u, a = e.join("\n"), o = t.raw + "\n" + a.replace(this.rules.other.blockquoteSetextReplace2, ""), s = this.blockquote(o);
					i[i.length - 1] = s;
					let c = o.substring(s.raw.length).replace(/^\n/, ""), l = c ? c.split("\n").length : 0, d = l ? e.slice(0, -l) : e;
					if (d.length > 0 && (n = `${n}
${d.join("\n")}`), r = r.substring(0, r.length - t.text.length) + s.text, e = e.slice(d.length), e.length > 0 && this.rules.other.blockquoteStart.test(e[0])) continue;
					break;
				}
				if (u?.type === "list") {
					let t = u, a = t.raw + "\n" + e.join("\n"), o = this.list(a);
					i[i.length - 1] = o;
					let s = o.raw.length > t.raw.length ? a.slice(t.raw.length + 1, o.raw.length).split("\n").length : 0;
					if (s > 0) {
						let t = e.slice(0, s).join("\n");
						n = `${n}
${t}`, r = `${r}
${t}`;
					}
					e = e.slice(s);
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
			let n = e, r = () => n.length - e.length, i = t[1].trim(), a = i.length > 1, o = {
				type: "list",
				raw: "",
				ordered: a,
				start: a ? +i.slice(0, -1) : "",
				loose: !1,
				items: []
			};
			i = a ? `\\d{1,9}\\${i.slice(-1)}` : `\\${i}`, this.options.pedantic && (i = a ? i : "[*+-]");
			let s = this.rules.other.listItemRegex(i), c = !1;
			for (; e;) {
				let i = !1, a = "";
				if (!(t = s.exec(e)) || this.rules.block.hr.test(e)) break;
				let l = r();
				e = e.substring(t[0].length);
				let u = t[2].split("\n", 1)[0], d = t[1].length, f = this.options.pedantic ? Wm(u, d) : u.replace(this.rules.other.leadingSpaceTab, (e) => Wm(e, d)), p = e.split("\n", 1)[0], m = !f.trim(), h = 0;
				if (this.options.pedantic ? (h = 2, a = f.trimStart()) : m ? h = d + 1 : (h = f.search(this.rules.other.nonSpaceChar), h = h > 4 ? 1 : h, a = f.slice(h), h += d), m && this.rules.other.blankLine.test(p) && (e = e.substring(p.length + 1), i = !0), !i) {
					let t = this.rules.other.nextBulletRegex(h), n = this.rules.other.hrRegex(h), r = this.rules.other.fencesBeginRegex(h), i = this.options.pedantic ? this.rules.other.pedanticHeadingBeginRegex(h) : this.rules.other.headingBeginRegex(h), o = this.rules.other.htmlBeginRegex(h), s = this.rules.other.blockquoteBeginRegex(h);
					for (; e;) {
						let c = e.split("\n", 1)[0], l;
						if (p = c, this.options.pedantic ? (p = p.replace(this.rules.other.listReplaceNesting, "  "), l = p) : l = p.replace(this.rules.other.leadingSpaceTab, (e) => Wm(e)), r.test(p) || i.test(p) || o.test(p) || s.test(p) || t.test(p) || n.test(p)) break;
						if (l.search(this.rules.other.nonSpaceChar) >= h || !p.trim()) a += "\n" + l.slice(h);
						else {
							if (m || f.replace(this.rules.other.tabCharGlobal, "    ").search(this.rules.other.nonSpaceChar) >= 4 || r.test(f) || i.test(f) || o.test(f) || n.test(f)) break;
							let e = !this.options.pedantic && /^ {0,3}(?:=+|-+) *$/.test(p) && this.rules.block.lheading.test(`${f}
${p}`) ? `    ${l.replace(/^ {0,3}/, "")}` : p;
							a += "\n" + e;
						}
						m = !p.trim(), e = e.substring(c.length + 1), f = l.slice(h);
					}
				}
				let g = n.slice(l, r());
				o.loose || (c ? o.loose = !0 : this.rules.other.doubleBlankLine.test(g) && (c = !0)), o.items.push({
					type: "list_item",
					raw: g,
					task: !!this.options.gfm && this.rules.other.listIsTask.test(a),
					loose: !1,
					text: a,
					tokens: []
				});
			}
			let l = o.items.at(-1);
			if (l) l.raw = l.raw.trimEnd(), l.text = l.text.trimEnd();
			else return;
			o.raw = n.slice(0, r()).trimEnd();
			for (let e of o.items) if (this.lexer.state.top = !1, e.tokens = this.lexer.blockTokens(e.text, []), !o.loose) {
				let t = e.tokens.filter((e) => e.type === "space");
				o.loose = t.length > 0 && t.some((e) => this.rules.other.anyLine.test(e.raw));
			}
			for (let e of o.items) {
				let t = e.tokens[0];
				if (e.task && (t?.type === "text" || t?.type === "paragraph")) {
					e.text = e.text.replace(this.rules.other.listReplaceTask, ""), t.raw = t.raw.replace(this.rules.other.listReplaceTask, ""), t.text = t.text.replace(this.rules.other.listReplaceTask, "");
					for (let e = this.lexer.inlineQueue.length - 1; e >= 0; e--) if (this.rules.other.listIsTask.test(this.lexer.inlineQueue[e].src)) {
						this.lexer.inlineQueue[e].src = this.lexer.inlineQueue[e].src.replace(this.rules.other.listReplaceTask, "");
						break;
					}
					let n = this.rules.other.listTaskCheckbox.exec(e.raw);
					if (n) {
						let t = {
							type: "checkbox",
							raw: n[0] + " ",
							checked: n[0] !== "[ ]"
						};
						e.checked = t.checked, o.loose ? e.tokens[0] && ["paragraph", "text"].includes(e.tokens[0].type) && "tokens" in e.tokens[0] && e.tokens[0].tokens ? (e.tokens[0].raw = t.raw + e.tokens[0].raw, e.tokens[0].text = t.raw + e.tokens[0].text, e.tokens[0].tokens.unshift(t)) : e.tokens.unshift({
							type: "paragraph",
							raw: t.raw,
							text: t.raw,
							tokens: [t]
						}) : e.tokens.unshift(t);
					}
				} else e.task &&= !1;
			}
			if (o.loose) for (let e of o.items) {
				e.loose = !0;
				for (let t of e.tokens) t.type === "text" && (t.type = "paragraph");
			}
			return o;
		}
	}
	html(e) {
		let t = this.rules.block.html.exec(e);
		if (t) {
			let e = Fm(t[0]);
			return {
				type: "html",
				block: !0,
				raw: e,
				pre: t[1] === "pre" || t[1] === "script" || t[1] === "style",
				text: e
			};
		}
	}
	def(e) {
		let t = this.rules.block.def.exec(e);
		if (t) {
			if (!this.rules.other.startAngleBracket.test(t[2]) && Lm(t[2], "()") !== -1) return;
			let e = Im(t[1]), n = t[2] ? t[2].replace(this.rules.other.hrefBrackets, "$1").replace(this.rules.inline.anyPunctuation, "$1") : "", r = t[3] ? t[3].substring(1, t[3].length - 1).replace(this.rules.inline.anyPunctuation, "$1") : t[3];
			return {
				type: "def",
				tag: e,
				raw: Pm(t[0], "\n"),
				href: n,
				title: r
			};
		}
	}
	table(e) {
		let t = this.rules.block.table.exec(e);
		if (!t || !this.rules.other.tableDelimiter.test(t[2])) return;
		let n = Nm(t[1]), r = t[2].replace(this.rules.other.tableAlignChars, "").split("|"), i = t[3]?.trim() ? t[3].replace(this.rules.other.tableRowBlankLine, "").split("\n") : [], a = {
			type: "table",
			raw: Pm(t[0], "\n"),
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
			for (let e of i) a.rows.push(Nm(e, a.header.length).map((e, t) => ({
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
				raw: Pm(t[0], "\n"),
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
		if (this.lexer.state.linkParenPossible === !1) return;
		let t = this.rules.inline.link.exec(e);
		if (t) {
			let n = t[0].charAt(0) === "!" ? 2 : 1;
			if (!this.options.pedantic && qm(e, t[1], n, this.rules)) return;
			let r = t[2].trim();
			if (!this.options.pedantic && this.rules.other.startAngleBracket.test(r)) {
				if (!this.rules.other.endAngleBracket.test(r)) return;
				let e = Pm(r.slice(0, -1), "\\");
				if ((r.length - e.length) % 2 == 0) return;
			} else {
				let e = Lm(t[2], "()");
				if (e === -2) return;
				if (e > -1) {
					let n = (t[0].indexOf("!") === 0 ? 5 : 4) + t[1].length + e;
					t[2] = t[2].substring(0, e), t[0] = t[0].substring(0, n).trim(), t[3] = "";
				}
			}
			let i = Pm(t[0].slice(0, -1), "\\");
			if ((t[0].length - i.length) % 2 == 0) return;
			let a = t[2], o = "";
			if (this.options.pedantic) {
				let e = this.rules.other.pedanticHrefTitle.exec(a);
				e && (a = e[1], o = e[3]);
			} else o = t[3] ? t[3].slice(1, -1) : "";
			return a = a.trim(), this.rules.other.startAngleBracket.test(a) && (a = this.options.pedantic && !this.rules.other.endAngleBracket.test(r) ? a.slice(1) : a.slice(1, -1)), Gm(t, {
				href: a && a.replace(this.rules.inline.anyPunctuation, "$1"),
				title: o && o.replace(this.rules.inline.anyPunctuation, "$1")
			}, t[0], this.lexer, this.rules);
		}
	}
	reflink(e, t) {
		let n;
		if ((n = this.rules.inline.reflink.exec(e)) || (n = this.rules.inline.nolink.exec(e))) {
			let r = n[0].charAt(0) === "!" ? 2 : 1;
			if (!this.options.pedantic && qm(e, n[1], r, this.rules)) return;
			let i = t[Im(n[2] || n[1])];
			if (!i) {
				let e = n[0].charAt(0);
				return {
					type: "text",
					raw: e,
					text: e
				};
			}
			return Gm(n, i, n[0], this.lexer, this.rules);
		}
	}
	emStrong(e, t, n = "") {
		let r = this.rules.inline.emStrongLDelim.exec(e);
		if (!r || !r[1] && !r[2] && !r[3] && !r[4]) return;
		let i = r[0][0], a = n === i, o = t.length - e.length, s = o;
		for (; !this.options.pedantic && s > 0 && t[s - 1] === i;) s--;
		let c = [...t.slice(Math.max(0, s - 2), s)].at(-1) ?? "", l = c === "*" || c === "_" || !!this.rules.inline.punctuation.exec(c);
		s < o && (n = c);
		let u = s < o ? l : !!this.rules.inline.punctuation.exec(n);
		if (r[4] && n.match(this.rules.other.unicodeAlphaNumeric)) return;
		let d = r[1] || r[3] || "";
		if (!d || !n || u) {
			let n = [...r[0]].length - 1, u, f, p = n, m = 0, h = o - s + n, g = !this.options.pedantic && !!c.trim() && (!!d || i === "*" && !l);
			if (i !== "*" && i !== "_") return;
			let _ = i === "*" ? this.rules.inline.emStrongRDelimAst : this.rules.inline.emStrongRDelimUnd, v = Hm(this, t, i), y = t.length - e.length + n;
			if (v < y) return;
			t = t.slice(-1 * e.length + n);
			let b = v - y;
			for (_.lastIndex = 0; (r = _.exec(t)) !== null && !(r.index > b);) {
				if (u = r[1] || r[2] || r[3] || r[4] || r[5] || r[6], !u) continue;
				if (f = [...u].length, r[3] || r[4]) {
					p += f;
					continue;
				}
				if (r[5] || r[6] || g) {
					if (a && (r[5] || r[6])) break;
					if (h % 3 && !((h + f) % 3)) {
						m += f;
						continue;
					}
				}
				if (p -= f, p > 0) continue;
				f = Math.min(f, f + p + m);
				let t = [...r[0]][0].length, i = e.slice(0, n + r.index + t + f);
				if (Math.min(n, f) % 2) {
					let e = i.slice(1, -1);
					return {
						type: "em",
						raw: i,
						text: e,
						tokens: this.lexer.inlineTokens(e)
					};
				}
				let o = i.slice(2, -2);
				return {
					type: "strong",
					raw: i,
					text: o,
					tokens: this.lexer.inlineTokens(o)
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
			let n = [...r[0]].length - 1, i, a, o = n, s = this.rules.inline.delRDelim, c = Um(this, t), l = t.length - e.length + n;
			if (c < l) return;
			t = t.slice(-1 * e.length + n);
			let u = c - l;
			for (s.lastIndex = 0; (r = s.exec(t)) !== null && !(r.index > u);) {
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
				autolink: !0,
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
				autolink: !0,
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
				text: e ? t[0] : jm(t[0]),
				escaped: e
			};
		}
	}
}, Ym = class e {
	tokens;
	options;
	state;
	inlineQueue;
	tokenizer;
	constructor(e) {
		this.tokens = [], this.tokens.links = Object.create(null), this.options = e || cp, this.options.tokenizer = this.options.tokenizer || new Jm(), this.tokenizer = this.options.tokenizer, this.tokenizer.options = this.options, this.tokenizer.lexer = this, this.inlineQueue = [], this.state = {
			inLink: !1,
			inRawBlock: !1,
			linkEmitted: !1,
			linkParenPossible: !0,
			top: !0
		};
		let t = {
			other: J,
			block: Dm.normal,
			inline: Om.normal
		};
		this.options.pedantic ? (t.block = Dm.pedantic, t.inline = Om.pedantic) : this.options.gfm && (t.block = Dm.gfm, t.inline = this.options.breaks ? Om.breaks : Om.gfm), this.tokenizer.rules = t;
	}
	static get rules() {
		return {
			block: Dm,
			inline: Om
		};
	}
	static lex(t, n) {
		return new e(n).lex(t);
	}
	static lexInline(t, n) {
		return new e(n).inlineTokens(t);
	}
	lex(e) {
		e = e.replace(J.carriageReturn, "\n"), this.blockTokens(e, this.tokens);
		for (let e = 0; e < this.inlineQueue.length; e++) {
			let t = this.inlineQueue[e];
			this.state.inRawBlock = !1, this.inlineTokens(t.src, t.tokens);
		}
		return this.inlineQueue = [], this.tokens;
	}
	blockTokens(e, t = [], n = !1) {
		this.tokenizer.lexer = this, this.options.pedantic && (e = e.replace(J.tabCharGlobal, "    ").replace(J.spaceLine, ""));
		let r = 1 / 0;
		for (; e;) {
			if (e.length < r) r = e.length;
			else {
				this.infiniteLoopError(e.charCodeAt(0));
				break;
			}
			let i;
			if (this.options.extensions?.block?.some((n) => (i = n.call({ lexer: this }, e, t)) ? (e = e.substring(i.raw.length), t.push(i), !0) : !1)) continue;
			if (i = this.tokenizer.space(e)) {
				e = e.substring(i.raw.length);
				let n = t.at(-1);
				i.raw.length === 1 && n !== void 0 ? n.raw += "\n" : t.push(i);
				continue;
			}
			if (i = this.tokenizer.code(e)) {
				e = e.substring(i.raw.length);
				let n = t.at(-1);
				n?.type === "paragraph" || n?.type === "text" ? (n.raw += (n.raw.endsWith("\n") ? "" : "\n") + i.raw, n.text += "\n" + i.text, this.inlineQueue.at(-1).src = n.text) : t.push(i);
				continue;
			}
			if ((i = this.tokenizer.fences(e)) || (i = this.tokenizer.heading(e)) || (i = this.tokenizer.hr(e)) || (i = this.tokenizer.blockquote(e)) || (i = this.tokenizer.list(e)) || (i = this.tokenizer.html(e))) {
				e = e.substring(i.raw.length), t.push(i);
				continue;
			}
			if (i = this.tokenizer.def(e)) {
				e = e.substring(i.raw.length);
				let n = t.at(-1);
				n?.type === "paragraph" || n?.type === "text" ? (n.raw += (n.raw.endsWith("\n") ? "" : "\n") + i.raw, n.text += "\n" + i.raw, this.inlineQueue.at(-1).src = n.text) : (this.tokens.links[i.tag] || (this.tokens.links[i.tag] = {
					href: i.href,
					title: i.title
				}), t.push(i));
				continue;
			}
			if ((i = this.tokenizer.table(e)) || (i = this.tokenizer.lheading(e))) {
				e = e.substring(i.raw.length), t.push(i);
				continue;
			}
			let a = e;
			if (this.options.extensions?.startBlock) {
				let t = 1 / 0, n = e.slice(1), r;
				this.options.extensions.startBlock.forEach((e) => {
					r = e.call({ lexer: this }, n), typeof r == "number" && r >= 0 && (t = Math.min(t, r));
				}), t < 1 / 0 && t >= 0 && (a = e.substring(0, t + 1));
			}
			if (this.state.top && (i = this.tokenizer.paragraph(a))) {
				let r = t.at(-1);
				n && r?.type === "paragraph" ? (r.raw += (r.raw.endsWith("\n") ? "" : "\n") + i.raw, r.text += "\n" + i.text, this.inlineQueue.pop(), this.inlineQueue.at(-1).src = r.text) : t.push(i), n = a.length !== e.length, e = e.substring(i.raw.length);
			} else if (i = this.tokenizer.text(e)) {
				e = e.substring(i.raw.length);
				let n = t.at(-1);
				n?.type === "text" ? (n.raw += (n.raw.endsWith("\n") ? "" : "\n") + i.raw, n.text += "\n" + i.text, this.inlineQueue.pop(), this.inlineQueue.at(-1).src = n.text) : t.push(i);
			} else if (e) {
				this.infiniteLoopError(e.charCodeAt(0));
				break;
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
	linkInText(e) {
		if (!e.includes("[")) return !1;
		let t = this.tokenizer.rules.inline.link;
		for (let n of e.matchAll(this.tokenizer.rules.inline.blockSkip)) if (t.test(n[0]) && e.charAt(n.index - 1) !== "!") return !0;
		for (let t of e.matchAll(this.tokenizer.rules.inline.reflinkSearch)) {
			let e = t[0], n = e.lastIndexOf("[");
			if (e.charAt(0) !== "!" && Object.hasOwn(this.tokens.links, Im(e.slice(n + 1, -1))) && !(n > 1 && this.linkInText(e.slice(1, n - 1)))) return !0;
		}
		return !1;
	}
	inlineTokens(e, t = []) {
		this.tokenizer.lexer = this;
		let n = this.state.linkParenPossible;
		this.state.linkParenPossible = n && e.includes(")");
		try {
			return this.#e(e, t);
		} finally {
			this.state.linkParenPossible = n;
		}
	}
	#e(e, t) {
		let n = e.includes("*") || e.includes("_") || e.includes("~") || this.options.hooks?.emStrongMask || this.tokenizer.emStrong !== Jm.prototype.emStrong || this.tokenizer.del !== Jm.prototype.del ? this.#n(e) : e;
		Bm(this.tokenizer, n);
		try {
			return this.#t(e, n, t);
		} finally {
			Vm(this.tokenizer);
		}
	}
	#t(e, t, n) {
		let r = !1, i = "", a = 1 / 0;
		for (; e;) {
			if (e.length < a) a = e.length;
			else {
				this.infiniteLoopError(e.charCodeAt(0));
				break;
			}
			r || (i = ""), r = !1;
			let o;
			if (this.options.extensions?.inline?.some((t) => (o = t.call({ lexer: this }, e, n)) ? (e = e.substring(o.raw.length), n.push(o), !0) : !1)) continue;
			if ((o = this.tokenizer.escape(e)) || (o = this.tokenizer.tag(e)) || (o = this.tokenizer.link(e))) {
				e = e.substring(o.raw.length), n.push(o);
				continue;
			}
			if (o = this.tokenizer.reflink(e, this.tokens.links)) {
				e = e.substring(o.raw.length);
				let t = n.at(-1);
				o.type === "text" && t?.type === "text" ? (t.raw += o.raw, t.text += o.text) : n.push(o);
				continue;
			}
			if ((o = this.tokenizer.emStrong(e, t, i)) || (o = this.tokenizer.codespan(e)) || (o = this.tokenizer.br(e)) || (o = this.tokenizer.del(e, t, i)) || (o = this.tokenizer.autolink(e)) || !this.state.inLink && (o = this.tokenizer.url(e))) {
				e = e.substring(o.raw.length), n.push(o);
				continue;
			}
			let s = e;
			if (this.options.extensions?.startInline) {
				let t = 1 / 0, n = e.slice(1), r;
				this.options.extensions.startInline.forEach((e) => {
					r = e.call({ lexer: this }, n), typeof r == "number" && r >= 0 && (t = Math.min(t, r));
				}), t < 1 / 0 && t >= 0 && (s = e.substring(0, t + 1));
			}
			if (o = this.tokenizer.inlineText(s)) {
				e = e.substring(o.raw.length), o.raw.slice(-1) !== "_" && (i = o.raw.slice(-1)), r = !0;
				let t = n.at(-1);
				t?.type === "text" ? (t.raw += o.raw, t.text += o.text) : n.push(o);
			} else if (e) {
				this.infiniteLoopError(e.charCodeAt(0));
				break;
			}
		}
		return n;
	}
	#n(e) {
		let t = e;
		if (this.tokens.links && e.includes("[")) {
			let e = this.tokenizer.rules.inline.reflinkSearch, n = (t) => {
				let r = t.lastIndexOf("[");
				if (!Object.hasOwn(this.tokens.links, Im(t.slice(r + 1, -1)))) return t;
				if (r > 1 && t.charAt(0) !== "!") {
					let i = t.slice(1, r - 1);
					if (this.linkInText(i)) return "[" + i.replace(e, n) + "][" + "a".repeat(t.length - r - 2) + "]";
				}
				return "[" + "a".repeat(t.length - 2) + "]";
			};
			t = t.replace(e, n);
		}
		return t = t.replace(this.tokenizer.rules.inline.anyPunctuationOutsideCode, (e, t, n) => n ? "+".repeat(e.length) : t ? t + e.slice(t.length, -t.length).replace(/`/g, "+") + t : e), t = t.replace(this.tokenizer.rules.inline.blockSkip, (e, t, n) => {
			let r = n ? n.length : 0;
			return e.slice(0, r) + "[" + "a".repeat(e.length - r - 2) + "]";
		}), t = this.options.hooks?.emStrongMask?.call({ lexer: this }, t) ?? t, t;
	}
	infiniteLoopError(e) {
		let t = "Infinite loop on byte: " + e;
		if (this.options.silent) console.error(t);
		else throw Error(t);
	}
}, Xm = class {
	options;
	parser;
	constructor(e) {
		this.options = e || cp;
	}
	space(e) {
		return "";
	}
	code({ text: e, lang: t, escaped: n }) {
		let r = (t || "").match(J.notSpaceStart)?.[0], i = e ? e.replace(J.endingNewline, "") + "\n" : "";
		return r ? "<pre><code class=\"language-" + Y(r) + "\">" + (n ? i : Y(i, !0)) + "</code></pre>\n" : "<pre><code>" + (n ? i : Y(i, !0)) + "</code></pre>\n";
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
		return `<code>${Y(e, !0)}</code>`;
	}
	br(e) {
		return "<br>";
	}
	del({ tokens: e }) {
		return `<del>${this.parser.parseInline(e)}</del>`;
	}
	link({ href: e, title: t, text: n, tokens: r, autolink: i }) {
		let a = i ? Y(n, !0) : this.parser.parseInline(r), o = Mm(e);
		if (o === null) return a;
		e = Y(o, i);
		let s = "<a href=\"" + e + "\"";
		return t && (s += " title=\"" + Y(t) + "\""), s += ">" + a + "</a>", s;
	}
	image({ href: e, title: t, text: n, tokens: r }) {
		r && (n = this.parser.parseInline(r, this.parser.textRenderer));
		let i = Mm(e);
		if (i === null) return Y(n);
		e = i;
		let a = `<img src="${Y(e)}" alt="${Y(n)}"`;
		return t && (a += ` title="${Y(t)}"`), a += ">", a;
	}
	text(e) {
		return "tokens" in e && e.tokens ? this.parser.parseInline(e.tokens) : "escaped" in e && e.escaped ? e.text : Y(e.text);
	}
}, Zm = class {
	parser;
	strong({ text: e, tokens: t }) {
		return t ? this.parser.parseInline(t, this) : e;
	}
	em({ text: e, tokens: t }) {
		return t ? this.parser.parseInline(t, this) : e;
	}
	codespan({ text: e }) {
		return e;
	}
	del({ text: e, tokens: t }) {
		return t ? this.parser.parseInline(t, this) : e;
	}
	html({ text: e }) {
		return e;
	}
	text({ text: e }) {
		return e;
	}
	link({ text: e, tokens: t }) {
		return t ? this.parser.parseInline(t, this) : e;
	}
	image({ text: e, tokens: t }) {
		return t ? this.parser.parseInline(t, this) : e;
	}
	br() {
		return "";
	}
	checkbox({ raw: e }) {
		return e;
	}
}, Qm = class e {
	options;
	renderer;
	textRenderer;
	constructor(e) {
		this.options = e || cp, this.options.renderer = this.options.renderer || new Xm(), this.renderer = this.options.renderer, this.renderer.options = this.options, this.renderer.parser = this, this.textRenderer = new Zm(), this.textRenderer.parser = this;
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
					"checkbox",
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
					"checkbox",
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
}, $m = class {
	options;
	block;
	constructor(e) {
		this.options = e || cp;
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
		return e ? Ym.lex : Ym.lexInline;
	}
	provideParser(e = this.block) {
		return e ? Qm.parse : Qm.parseInline;
	}
};
function eh(e, t) {
	if (Array.isArray(t)) for (let n = 0; n < t.length; n++) e.push(t[n]);
	else e.push(t);
}
var th = new class {
	defaults = sp();
	options = this.setOptions;
	parse = this.parseMarkdown(!0);
	parseInline = this.parseMarkdown(!1);
	Parser = Qm;
	Renderer = Xm;
	TextRenderer = Zm;
	Lexer = Ym;
	Tokenizer = Jm;
	Hooks = $m;
	constructor(...e) {
		this.use(...e);
	}
	walkTokens(e, t) {
		let n = [];
		for (let r of e) switch (eh(n, t.call(this, r)), r.type) {
			case "table": {
				let e = r;
				for (let r of e.header) eh(n, this.walkTokens(r.tokens, t));
				for (let r of e.rows) for (let e of r) eh(n, this.walkTokens(e.tokens, t));
				break;
			}
			case "list": {
				let e = r;
				eh(n, this.walkTokens(e.items, t));
				break;
			}
			default: {
				let e = r;
				this.defaults.extensions?.childTokens?.[e.type] ? this.defaults.extensions.childTokens[e.type].forEach((r) => {
					let i = e[r].flat(1 / 0);
					eh(n, this.walkTokens(i, t));
				}) : e.tokens && eh(n, this.walkTokens(e.tokens, t));
			}
		}
		return n;
	}
	use(...e) {
		let t = this.defaults.extensions || {
			renderers: Object.create(null),
			childTokens: Object.create(null)
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
				let t = this.defaults.renderer || new Xm(this.defaults);
				for (let n in e.renderer) {
					if (!(n in t)) throw Error(`renderer '${n}' does not exist`);
					if (["options", "parser"].includes(n)) continue;
					let r = n, i = e.renderer[r], a = t[r];
					t[r] = (...e) => {
						let n = i.apply(t, e);
						return n === !1 && (n = a.apply(t, e)), n === !1 ? "" : n ?? "";
					};
				}
				n.renderer = t;
			}
			if (e.tokenizer) {
				let t = this.defaults.tokenizer || new Jm(this.defaults);
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
				let t = this.defaults.hooks || new $m();
				for (let n in e.hooks) {
					if (!(n in t)) throw Error(`hook '${n}' does not exist`);
					if (["options", "block"].includes(n)) continue;
					let r = n, i = e.hooks[r], a = t[r];
					t[r] = $m.passThroughHooks.has(n) ? (e) => {
						if (t.options.async && $m.passThroughHooksRespectAsync.has(n)) return (async () => {
							let n = await i.call(t, e);
							return a.call(t, n);
						})();
						let r = i.call(t, e);
						return a.call(t, r);
					} : (...e) => {
						if (t.options.async) return (async () => {
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
					return eh(n, r.call(this, e)), t && eh(n, t.call(this, e)), n;
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
		return Ym.lex(e, t ?? this.defaults);
	}
	parser(e, t) {
		return Qm.parse(e, t ?? this.defaults);
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
				let n = i.hooks ? await i.hooks.preprocess(t) : t, r = await (i.hooks ? await i.hooks.provideLexer(e) : e ? Ym.lex : Ym.lexInline)(n, i), a = i.hooks ? await i.hooks.processAllTokens(r) : r;
				i.walkTokens && await Promise.all(this.walkTokens(a, i.walkTokens));
				let o = await (i.hooks ? await i.hooks.provideParser(e) : e ? Qm.parse : Qm.parseInline)(a, i);
				return i.hooks ? await i.hooks.postprocess(o) : o;
			})().catch(a);
			try {
				i.hooks && (t = i.hooks.preprocess(t));
				let n = (i.hooks ? i.hooks.provideLexer(e) : e ? Ym.lex : Ym.lexInline)(t, i);
				i.hooks && (n = i.hooks.processAllTokens(n)), i.walkTokens && this.walkTokens(n, i.walkTokens);
				let r = (i.hooks ? i.hooks.provideParser(e) : e ? Qm.parse : Qm.parseInline)(n, i);
				return i.hooks && (r = i.hooks.postprocess(r)), r;
			} catch (e) {
				return a(e);
			}
		};
	}
	onError(e, t) {
		return (n) => {
			if (n.message += "\nPlease report this to https://github.com/markedjs/marked.", e) {
				let e = "<p>An error occurred:</p><pre>" + Y(n.message + "", !0) + "</pre>";
				return t ? Promise.resolve(e) : e;
			}
			if (t) return Promise.reject(n);
			throw n;
		};
	}
}();
function X(e, t) {
	return th.parse(e, t);
}
X.options = X.setOptions = function(e) {
	return th.setOptions(e), X.defaults = th.defaults, lp(X.defaults), X;
}, X.getDefaults = sp, X.defaults = cp;
function nh(...e) {
	return th.use(...e), X.defaults = th.defaults, lp(X.defaults), X;
}
X.use = nh, X.walkTokens = function(e, t) {
	return th.walkTokens(e, t);
}, X.parseInline = th.parseInline, X.Parser = Qm, X.parser = Qm.parse, X.Renderer = Xm, X.TextRenderer = Zm, X.Lexer = Ym, X.lexer = Ym.lex, X.Tokenizer = Jm, X.Hooks = $m, X.parse = X, X.options, X.setOptions, X.walkTokens, X.parseInline, Qm.parse, Ym.lex;
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/streamdown/2.7.1/cd86399a25cc60a263bb5bc5019500d5ee112acb0fe8c90f08576a062b1a2e36/node_modules/streamdown/dist/chunk-7TTQLYRW.js
var rh = 300, ih = "300px", ah = 500;
function oh(e = {}) {
	let { immediate: t = !1, debounceDelay: n = rh, rootMargin: r = ih, idleTimeout: i = ah } = e, [a, o] = (0, k.useState)(!1), s = (0, k.useRef)(null), c = (0, k.useRef)(null), l = (0, k.useRef)(null), u = (0, k.useMemo)(() => (e) => {
		let t = Date.now();
		return window.setTimeout(() => {
			e({
				didTimeout: !1,
				timeRemaining: () => Math.max(0, 50 - (Date.now() - t))
			});
		}, 1);
	}, []), d = (0, k.useMemo)(() => typeof window < "u" && window.requestIdleCallback ? (e, t) => window.requestIdleCallback(e, t) : u, [u]), f = (0, k.useMemo)(() => typeof window < "u" && window.cancelIdleCallback ? (e) => window.cancelIdleCallback(e) : (e) => {
		clearTimeout(e);
	}, []);
	return (0, k.useEffect)(() => {
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
var sh = 320, ch = 4, lh = () => typeof performance > "u" ? Date.now() : performance.now();
function uh(e) {
	let t = e?.now ?? lh, n = e?.maxBacklogMs ?? sh, r = 0, i = 0;
	return {
		now: t,
		beginPass(e) {
			i = Math.min(Math.max(r, e), e + n);
		},
		mark() {
			return i;
		},
		rewind(e) {
			i = e;
		},
		take(e, t, r) {
			if (e <= 0) return {
				baseDelay: 0,
				step: Math.max(0, t)
			};
			let a = Math.max(0, t), o = a === 0 ? 0 : Math.min(a, ch), s = Math.max(i, r), c = r + n, l = s + Math.max(0, e - 1) * a, u = a;
			if (l > c && e > 1) {
				if (s < c) {
					let t = c - s;
					u = Math.max(o, t / (e - 1));
				} else u = o;
			}
			let d = Math.max(0, Math.round(s - r));
			return i = s + e * u, {
				baseDelay: d,
				step: u
			};
		},
		commitPass() {
			r = i;
		}
	};
}
var dh = /\s/, fh = /[\p{Script=Han}\p{Script=Hiragana}\p{Script=Katakana}\u3000-\u303F\uFF00-\uFFEF]/u, ph = /^\s+$/, mh = /* @__PURE__ */ new Set([
	"pre",
	"svg",
	"math",
	"annotation"
]), hh = /* @__PURE__ */ new Set(["img", "hr"]), gh = (e) => typeof e == "object" && !!e && "type" in e && e.type === "element", _h = (e) => e.some((e) => gh(e) && mh.has(e.tagName)), vh = (e) => {
	for (let t = e.length - 1; t >= 0; t--) {
		let n = e[t];
		if (gh(n) && n.tagName === "li") return n;
	}
}, yh = (e, t, n, r) => {
	e.properties ??= {}, e.properties["data-sd-animate-marker"] = !0;
	let i = typeof e.properties.style == "string" ? `${e.properties.style};` : "";
	e.properties.style = `${i}--sd-marker-duration:${t}ms;--sd-marker-delay:${Math.round(n)}ms;--sd-marker-easing:${r}`;
}, bh = /* @__PURE__ */ new Set([
	"ul",
	"ol",
	"li"
]), xh = (e) => {
	for (let t of e.children) if (gh(t)) {
		if (t.tagName === "input") return t;
		if (bh.has(t.tagName)) continue;
		let e = xh(t);
		if (e) return e;
	}
}, Sh = (e, t, n, r) => {
	e.properties ??= {}, e.properties["data-sd-animate"] = !0;
	let i = typeof e.properties.style == "string" ? `${e.properties.style};` : "";
	e.properties.style = `${i}--sd-animation:sd-${t.animation};--sd-duration:${n}ms;--sd-easing:${t.easing};--sd-delay:${Math.round(r)}ms`;
}, Ch = (e, t, n, r) => {
	let i = xh(e);
	i && Sh(i, t, n, r);
}, wh = (e, t, n, r, i, a) => {
	if (_h(t)) return;
	let o = r.prevContentLength, s = i.count;
	i.count += 1;
	let c = o > 0 && s < o, l = c ? 0 : a.baseDelay + i.newIndex++ * a.step;
	Sh(e, n, c ? 0 : n.duration, l);
}, Th = (e) => {
	var t;
	for (let n = e.length - 1; n >= 0; --n) {
		let r = e[n];
		if (gh(r)) {
			if ((t = r.properties) != null && t["data-sd-animated"]) break;
			r.properties = {
				...r.properties,
				"data-sd-animated": !0
			};
		}
	}
}, Eh = (e) => {
	let t = [], n = "", r = !1, i = !1;
	for (let a of e) {
		let e = dh.test(a), o = !e && fh.test(a);
		if (!e && n !== "" && (o || i)) t.push(n), n = "", r = !1;
		else if (e !== r && n) {
			if (e) {
				n += a, r = !0;
				continue;
			}
			t.push(n), n = "";
		}
		n += a, r = e, i = o;
	}
	return n && t.push(n), t;
}, Dh = (e) => {
	let t = [], n = "";
	for (let r of e) dh.test(r) ? t.length > 0 && !ph.test(t.at(-1)) ? t[t.length - 1] += r : n += r : (n &&= (t.push(n), ""), t.push(r));
	return n && (t.length > 0 ? t[t.length - 1] += n : t.push(n)), t;
}, Oh = /^[^\s;:'"\\/](?:[^;:'"\\/\r\n]*[^\s;:'"\\/])?$/, kh = (e, t, n, r) => {
	let i = `sd-${t.animation}`, a = `${n ? 0 : t.duration}ms`, o;
	return t.styleObjects ? (o = {
		"--sd-animation": i,
		"--sd-duration": a,
		"--sd-easing": t.easing
	}, r && (o["--sd-delay"] = `${Math.round(r)}ms`)) : (o = `--sd-animation:${i};--sd-duration:${a};--sd-easing:${t.easing}`, r && (o += `;--sd-delay:${Math.round(r)}ms`)), {
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
}, Ah = (e, t) => !(e > 0 && t < e), jh = (e) => gh(e) && hh.has(e.tagName), Mh = (e, t, n) => {
	let r = 0, i = 0;
	return pe(e, (e) => e.type === "text" || jh(e), (e, a) => {
		if (_h(a)) return w;
		if (jh(e)) {
			Ah(n, i) && (r += 1), i += 1;
			return;
		}
		let o = e.value;
		if (!o.trim()) {
			i += o.length;
			return;
		}
		let s = t.sep === "char" ? Dh(o) : Eh(o);
		for (let e of s) {
			let t = i;
			i += e.length, !ph.test(e) && Ah(n, t) && (r += 1);
		}
	}), r;
}, Nh = (e, t, n, r, i, a) => {
	var o;
	let s = t.at(-1);
	if (!(s && "children" in s)) return;
	if (_h(t)) return w;
	let c = s, l = c.children.indexOf(e);
	if (l === -1) return;
	let u = e.value;
	if (!u.trim()) {
		i.count += u.length;
		return;
	}
	let d = n.sep === "char" ? Dh(u) : Eh(u), f = r.prevContentLength, p = !1, m = vh(t), h = !(!m || (o = m.properties) != null && o["data-sd-animate-marker"]), g = !1, _ = d.map((e) => {
		let t = i.count;
		if (i.count += e.length, ph.test(e)) return {
			type: "text",
			value: e
		};
		let r = f > 0 && t < f, o = r ? 0 : a.baseDelay + i.newIndex++ * a.step;
		if (p = !0, m && h && !g) {
			let e = r ? 0 : n.duration;
			yh(m, e, o, n.easing), Ch(m, n, e, o), g = !0;
		}
		return kh(e, n, r, o);
	});
	return p && Th(t), c.children.splice(l, 1, ..._), l + _.length;
}, Ph = 0;
function Fh(e) {
	return Lh(e, !1);
}
var Ih = (e) => Lh(e, !0);
function Lh(e, t) {
	let n = e?.animation ?? "fadeIn", r = e?.easing ?? "ease", i = {
		animation: n,
		duration: e?.duration ?? 150,
		easing: r,
		sep: e?.sep ?? "word",
		stagger: e?.stagger ?? 40,
		styleObjects: t && Oh.test(n) && Oh.test(r),
		timeline: e?.timeline
	}, a = {
		committedCharCount: 0,
		prevContentLength: 0,
		lastRenderCharCount: 0,
		pendingMark: null
	}, o = Ph++, s = () => (e) => {
		let t = {
			count: 0,
			newIndex: 0
		};
		a.prevContentLength = a.committedCharCount;
		let n = i.timeline, r = n?.now() ?? lh();
		n && (a.pendingMark === null ? a.pendingMark = n.mark() : n.rewind(a.pendingMark));
		let o = n ? n.take(Mh(e, i, a.prevContentLength), i.stagger, r) : {
			baseDelay: 0,
			step: i.stagger
		};
		pe(e, (e) => e.type === "text" || jh(e), (e, n) => {
			if (e.type === "text") return Nh(e, n, i, a, t, o);
			wh(e, n, i, a, t, o);
		}), a.lastRenderCharCount = t.count, a.prevContentLength = 0;
	};
	return Object.defineProperty(s, "name", { value: `rehypeAnimate$${o}` }), {
		name: "animate",
		type: "animate",
		rehypePlugin: s,
		setPrevContentLength(e) {
			a.committedCharCount = e, a.prevContentLength = e;
		},
		getLastRenderCharCount() {
			return a.lastRenderCharCount;
		},
		commit() {
			a.committedCharCount = a.lastRenderCharCount, a.pendingMark = null;
		}
	};
}
Fh();
var Rh = (0, k.createContext)(!1), zh = () => (0, k.useContext)(Rh), Bh = (...e) => a(o(e)), Vh = (e, t) => {
	if (!e || !t) return t;
	let n = `${e}:`;
	return t.split(/\s+/).filter(Boolean).map((t) => t.startsWith(n) ? t : `${e}:${t}`).join(" ");
}, Hh = (e) => e ? (...t) => Vh(e, a(o(t))) : Bh, Uh = (e, t, n) => {
	let r = typeof t == "string" && n.startsWith("text/csv") ? "﻿" : "", i = typeof t == "string" ? new Blob([r + t], { type: n }) : t, a = URL.createObjectURL(i), o = document.createElement("a");
	o.href = a, o.download = e, document.body.appendChild(o), o.click(), document.body.removeChild(o), URL.revokeObjectURL(a);
}, Wh = (0, k.createContext)(Bh), Z = () => (0, k.useContext)(Wh), Gh = 8, Kh = (e) => {
	if (e !== void 0 && e !== 0 && e !== Infinity) {
		if (typeof e == "number") return Number.isFinite(e) && e > 0 ? `${e}px` : void 0;
		if (e !== "0" && e !== "none" && e !== "Infinity") return e;
	}
}, qh = (e, t, n) => {
	let r = (0, k.useRef)(null), i = (0, k.useRef)(!0), a = (0, k.useRef)(!1), o = (0, k.useRef)(null);
	return (0, k.useEffect)(() => {
		let e = r.current;
		if (!(e && t)) return;
		let n = () => {
			let t = e.scrollHeight - e.scrollTop - e.clientHeight < Gh;
			i.current = t;
		};
		return e.addEventListener("scroll", n, { passive: !0 }), () => e.removeEventListener("scroll", n);
	}, [t]), (0, k.useEffect)(() => {
		e && !a.current && (i.current = !0), e || (i.current = !0), a.current = e;
	}, [e]), (0, k.useEffect)(() => () => {
		o.current !== null && cancelAnimationFrame(o.current);
	}, []), (0, k.useEffect)(() => {
		let n = r.current;
		!(n && t && e && i.current) || o.current !== null || (o.current = requestAnimationFrame(() => {
			o.current = null, i.current && n.scrollTo({
				top: n.scrollHeight,
				behavior: "instant"
			});
		}));
	}, [
		e,
		t,
		n
	]), r;
}, Jh = Bh("block"), Yh = Bh("block", "before:content-[counter(line)]", "before:inline-block", "before:[counter-increment:line]", "before:w-6", "before:mr-4", "before:text-[13px]", "before:text-right", "before:text-muted-foreground/50", "before:font-mono", "before:select-none"), Xh = (e) => {
	let t = {};
	for (let n of e.split(";")) {
		let e = n.indexOf(":");
		if (e > 0) {
			let r = n.slice(0, e).trim(), i = n.slice(e + 1).trim();
			r && i && (t[r] = i);
		}
	}
	return t;
}, Zh = (0, k.memo)(({ children: e, result: t, language: n, className: r, maxHeight: i, startLine: a, lineNumbers: o = !0, ...s }) => {
	let c = Z(), { isAnimating: l } = (0, k.useContext)($), u = Kh(i), d = qh(l, !!u, t), f = (0, k.useMemo)(() => c(Yh), [c]), p = (0, k.useMemo)(() => c(Jh), [c]), m = (0, k.useMemo)(() => {
		let e = {};
		return t.bg && (e["--sdm-bg"] = t.bg), t.fg && (e["--sdm-fg"] = t.fg), t.rootStyle && Object.assign(e, Xh(t.rootStyle)), e;
	}, [
		t.bg,
		t.fg,
		t.rootStyle
	]);
	return (0, K.jsx)("div", {
		className: c(r, u ? "overflow-y-auto" : null, "overflow-x-auto rounded-md border border-border bg-background p-4 text-sm"),
		"data-language": n,
		"data-streamdown": "code-block-body",
		ref: d,
		style: u ? { maxHeight: u } : void 0,
		...s,
		children: (0, K.jsx)("pre", {
			className: c(r, "bg-[var(--sdm-bg,inherit)]", "dark:bg-[var(--shiki-dark-bg,var(--sdm-bg,inherit))]"),
			style: m,
			children: (0, K.jsx)("code", {
				className: o ? c("[counter-increment:line_0] [counter-reset:line]") : void 0,
				style: o && a && a > 1 ? { counterReset: `line ${a - 1}` } : void 0,
				children: t.tokens.map((e, t) => (0, K.jsx)("span", {
					className: o ? f : p,
					children: e.length === 0 || e.length === 1 && e[0].content === "" ? "\n" : e.map((e, t) => {
						let n = {}, r = !!e.bgColor;
						if (e.color && (n["--sdm-c"] = e.color), e.bgColor && (n["--sdm-tbg"] = e.bgColor), e.htmlStyle) for (let [t, i] of Object.entries(e.htmlStyle)) t === "color" ? n["--sdm-c"] = i : t === "background-color" ? (n["--sdm-tbg"] = i, r = !0) : n[t] = i;
						return (0, K.jsx)("span", {
							className: c("text-[var(--sdm-c,inherit)]", "dark:text-[var(--shiki-dark,var(--sdm-c,inherit))]", r && "bg-[var(--sdm-tbg)]", r && "dark:bg-[var(--shiki-dark-bg,var(--sdm-tbg))]"),
							style: n,
							...e.htmlAttrs,
							children: e.content
						}, t);
					})
				}, t))
			})
		})
	});
}), Qh = ({ className: e, language: t, style: n, isIncomplete: r, ...i }) => {
	let a = Z();
	return (0, K.jsx)("div", {
		className: a("relative my-4 flex w-full flex-col gap-2 rounded-xl border border-border bg-sidebar p-2", e),
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
}, $h = (0, k.createContext)({ code: "" }), eg = () => (0, k.useContext)($h), tg = ({ language: e }) => {
	let t = Z();
	return (0, K.jsx)("div", {
		className: t("flex h-8 items-center text-muted-foreground text-xs"),
		"data-language": e,
		"data-streamdown": "code-block-header",
		children: (0, K.jsx)("span", {
			className: t("ml-1 font-mono lowercase"),
			children: e
		})
	});
}, ng = (e) => {
	let t = e.length;
	for (; t > 0 && e[t - 1] === "\n";) t--;
	return e.slice(0, t);
}, rg = (0, k.lazy)(() => import("./highlighted-body-JLMTFACH-C2BuPZPw.js").then((e) => ({ default: e.HighlightedCodeBlockBody }))), ig = ({ code: e, language: t, className: n, children: r, isIncomplete: i = !1, startLine: a, lineNumbers: o, ...s }) => {
	let c = Z(), { codeBlockMaxHeight: l } = (0, k.useContext)($), u = (0, k.useMemo)(() => ng(e), [e]), d = (0, k.useMemo)(() => ({
		bg: "transparent",
		fg: "inherit",
		tokens: u.split("\n").map((e) => [{
			content: e,
			color: "inherit",
			bgColor: "transparent",
			htmlStyle: {},
			offset: 0
		}])
	}), [u]);
	return (0, K.jsx)($h.Provider, {
		value: { code: e },
		children: (0, K.jsxs)(Qh, {
			dir: "ltr",
			isIncomplete: i,
			language: t,
			children: [
				(0, K.jsx)(tg, { language: t }),
				r ? (0, K.jsx)("div", {
					className: c("pointer-events-none absolute top-2 right-2 z-10 flex items-center"),
					children: (0, K.jsx)("div", {
						className: c("pointer-events-auto flex shrink-0 items-center gap-2 rounded-md border border-sidebar bg-sidebar/80 px-1.5 py-1 supports-[backdrop-filter]:bg-sidebar/70 supports-[backdrop-filter]:backdrop-blur"),
						"data-streamdown": "code-block-actions",
						children: r
					})
				}) : null,
				(0, K.jsx)(k.Suspense, {
					fallback: (0, K.jsx)(Zh, {
						className: n,
						language: t,
						lineNumbers: o,
						maxHeight: l,
						result: d,
						startLine: a,
						...s
					}),
					children: (0, K.jsx)(rg, {
						className: n,
						code: u,
						isIncomplete: i,
						language: t,
						lineNumbers: o,
						maxHeight: l,
						raw: d,
						startLine: a,
						...s
					})
				})
			]
		})
	});
}, ag = {
	CheckIcon: (e) => (0, K.jsx)("svg", {
		"aria-hidden": "true",
		color: "currentColor",
		height: 16,
		strokeLinejoin: "round",
		viewBox: "0 0 16 16",
		width: 16,
		...e,
		children: (0, K.jsx)("path", {
			clipRule: "evenodd",
			d: "M15.5607 3.99999L15.0303 4.53032L6.23744 13.3232C5.55403 14.0066 4.44599 14.0066 3.76257 13.3232L4.2929 12.7929L3.76257 13.3232L0.969676 10.5303L0.439346 9.99999L1.50001 8.93933L2.03034 9.46966L4.82323 12.2626C4.92086 12.3602 5.07915 12.3602 5.17678 12.2626L13.9697 3.46966L14.5 2.93933L15.5607 3.99999Z",
			fill: "currentColor",
			fillRule: "evenodd"
		})
	}),
	CopyIcon: (e) => (0, K.jsx)("svg", {
		"aria-hidden": "true",
		color: "currentColor",
		height: 16,
		strokeLinejoin: "round",
		viewBox: "0 0 16 16",
		width: 16,
		...e,
		children: (0, K.jsx)("path", {
			clipRule: "evenodd",
			d: "M2.75 0.5C1.7835 0.5 1 1.2835 1 2.25V9.75C1 10.7165 1.7835 11.5 2.75 11.5H3.75H4.5V10H3.75H2.75C2.61193 10 2.5 9.88807 2.5 9.75V2.25C2.5 2.11193 2.61193 2 2.75 2H8.25C8.38807 2 8.5 2.11193 8.5 2.25V3H10V2.25C10 1.2835 9.2165 0.5 8.25 0.5H2.75ZM7.75 4.5C6.7835 4.5 6 5.2835 6 6.25V13.75C6 14.7165 6.7835 15.5 7.75 15.5H13.25C14.2165 15.5 15 14.7165 15 13.75V6.25C15 5.2835 14.2165 4.5 13.25 4.5H7.75ZM7.5 6.25C7.5 6.11193 7.61193 6 7.75 6H13.25C13.3881 6 13.5 6.11193 13.5 6.25V13.75C13.5 13.8881 13.3881 14 13.25 14H7.75C7.61193 14 7.5 13.8881 7.5 13.75V6.25Z",
			fill: "currentColor",
			fillRule: "evenodd"
		})
	}),
	DownloadIcon: (e) => (0, K.jsx)("svg", {
		"aria-hidden": "true",
		color: "currentColor",
		height: 16,
		strokeLinejoin: "round",
		viewBox: "0 0 16 16",
		width: 16,
		...e,
		children: (0, K.jsx)("path", {
			clipRule: "evenodd",
			d: "M8.75 1V1.75V8.68934L10.7197 6.71967L11.25 6.18934L12.3107 7.25L11.7803 7.78033L8.70711 10.8536C8.31658 11.2441 7.68342 11.2441 7.29289 10.8536L4.21967 7.78033L3.68934 7.25L4.75 6.18934L5.28033 6.71967L7.25 8.68934V1.75V1H8.75ZM13.5 9.25V13.5H2.5V9.25V8.5H1V9.25V14C1 14.5523 1.44771 15 2 15H14C14.5523 15 15 14.5523 15 14V9.25V8.5H13.5V9.25Z",
			fill: "currentColor",
			fillRule: "evenodd"
		})
	}),
	ExternalLinkIcon: (e) => (0, K.jsx)("svg", {
		"aria-hidden": "true",
		color: "currentColor",
		height: 16,
		strokeLinejoin: "round",
		viewBox: "0 0 16 16",
		width: 16,
		...e,
		children: (0, K.jsx)("path", {
			clipRule: "evenodd",
			d: "M13.5 10.25V13.25C13.5 13.3881 13.3881 13.5 13.25 13.5H2.75C2.61193 13.5 2.5 13.3881 2.5 13.25L2.5 2.75C2.5 2.61193 2.61193 2.5 2.75 2.5H5.75H6.5V1H5.75H2.75C1.7835 1 1 1.7835 1 2.75V13.25C1 14.2165 1.7835 15 2.75 15H13.25C14.2165 15 15 14.2165 15 13.25V10.25V9.5H13.5V10.25ZM9 1H9.75H14.2495C14.6637 1 14.9995 1.33579 14.9995 1.75V6.25V7H13.4995V6.25V3.56066L8.53033 8.52978L8 9.06011L6.93934 7.99945L7.46967 7.46912L12.4388 2.5H9.75H9V1Z",
			fill: "currentColor",
			fillRule: "evenodd"
		})
	}),
	Loader2Icon: (e) => (0, K.jsxs)("svg", {
		"aria-hidden": "true",
		color: "currentColor",
		height: 16,
		strokeLinejoin: "round",
		viewBox: "0 0 16 16",
		width: 16,
		...e,
		children: [
			(0, K.jsx)("path", {
				d: "M8 0V4",
				stroke: "currentColor",
				strokeWidth: "1.5"
			}),
			(0, K.jsx)("path", {
				d: "M8 16V12",
				opacity: "0.5",
				stroke: "currentColor",
				strokeWidth: "1.5"
			}),
			(0, K.jsx)("path", {
				d: "M3.29773 1.52783L5.64887 4.7639",
				opacity: "0.9",
				stroke: "currentColor",
				strokeWidth: "1.5"
			}),
			(0, K.jsx)("path", {
				d: "M12.7023 1.52783L10.3511 4.7639",
				opacity: "0.1",
				stroke: "currentColor",
				strokeWidth: "1.5"
			}),
			(0, K.jsx)("path", {
				d: "M12.7023 14.472L10.3511 11.236",
				opacity: "0.4",
				stroke: "currentColor",
				strokeWidth: "1.5"
			}),
			(0, K.jsx)("path", {
				d: "M3.29773 14.472L5.64887 11.236",
				opacity: "0.6",
				stroke: "currentColor",
				strokeWidth: "1.5"
			}),
			(0, K.jsx)("path", {
				d: "M15.6085 5.52783L11.8043 6.7639",
				opacity: "0.2",
				stroke: "currentColor",
				strokeWidth: "1.5"
			}),
			(0, K.jsx)("path", {
				d: "M0.391602 10.472L4.19583 9.23598",
				opacity: "0.7",
				stroke: "currentColor",
				strokeWidth: "1.5"
			}),
			(0, K.jsx)("path", {
				d: "M15.6085 10.4722L11.8043 9.2361",
				opacity: "0.3",
				stroke: "currentColor",
				strokeWidth: "1.5"
			}),
			(0, K.jsx)("path", {
				d: "M0.391602 5.52783L4.19583 6.7639",
				opacity: "0.8",
				stroke: "currentColor",
				strokeWidth: "1.5"
			})
		]
	}),
	Maximize2Icon: (e) => (0, K.jsx)("svg", {
		"aria-hidden": "true",
		color: "currentColor",
		height: 16,
		strokeLinejoin: "round",
		viewBox: "0 0 16 16",
		width: 16,
		...e,
		children: (0, K.jsx)("path", {
			clipRule: "evenodd",
			d: "M1 5.25V6H2.5V5.25V2.5H5.25H6V1H5.25H2C1.44772 1 1 1.44772 1 2V5.25ZM5.25 14.9994H6V13.4994H5.25H2.5V10.7494V9.99939H1V10.7494V13.9994C1 14.5517 1.44772 14.9994 2 14.9994H5.25ZM15 10V10.75V14C15 14.5523 14.5523 15 14 15H10.75H10V13.5H10.75H13.5V10.75V10H15ZM10.75 1H10V2.5H10.75H13.5V5.25V6H15V5.25V2C15 1.44772 14.5523 1 14 1H10.75Z",
			fill: "currentColor",
			fillRule: "evenodd"
		})
	}),
	RotateCcwIcon: (e) => (0, K.jsx)("svg", {
		"aria-hidden": "true",
		color: "currentColor",
		height: 16,
		strokeLinejoin: "round",
		viewBox: "0 0 16 16",
		width: 16,
		...e,
		children: (0, K.jsx)("path", {
			clipRule: "evenodd",
			d: "M13.5 8C13.5 4.96643 11.0257 2.5 7.96452 2.5C5.42843 2.5 3.29365 4.19393 2.63724 6.5H5.25H6V8H5.25H0.75C0.335787 8 0 7.66421 0 7.25V2.75V2H1.5V2.75V5.23347C2.57851 2.74164 5.06835 1 7.96452 1C11.8461 1 15 4.13001 15 8C15 11.87 11.8461 15 7.96452 15C5.62368 15 3.54872 13.8617 2.27046 12.1122L1.828 11.5066L3.03915 10.6217L3.48161 11.2273C4.48831 12.6051 6.12055 13.5 7.96452 13.5C11.0257 13.5 13.5 11.0336 13.5 8Z",
			fill: "currentColor",
			fillRule: "evenodd"
		})
	}),
	XIcon: (e) => (0, K.jsx)("svg", {
		"aria-hidden": "true",
		color: "currentColor",
		height: 16,
		strokeLinejoin: "round",
		viewBox: "0 0 16 16",
		width: 16,
		...e,
		children: (0, K.jsx)("path", {
			clipRule: "evenodd",
			d: "M12.4697 13.5303L13 14.0607L14.0607 13L13.5303 12.4697L9.06065 7.99999L13.5303 3.53032L14.0607 2.99999L13 1.93933L12.4697 2.46966L7.99999 6.93933L3.53032 2.46966L2.99999 1.93933L1.93933 2.99999L2.46966 3.53032L6.93933 7.99999L2.46966 12.4697L1.93933 13L2.99999 14.0607L3.53032 13.5303L7.99999 9.06065L12.4697 13.5303Z",
			fill: "currentColor",
			fillRule: "evenodd"
		})
	}),
	ZoomInIcon: (e) => (0, K.jsx)("svg", {
		"aria-hidden": "true",
		color: "currentColor",
		height: 16,
		strokeLinejoin: "round",
		viewBox: "0 0 16 16",
		width: 16,
		...e,
		children: (0, K.jsx)("path", {
			clipRule: "evenodd",
			d: "M1.5 6.5C1.5 3.73858 3.73858 1.5 6.5 1.5C9.26142 1.5 11.5 3.73858 11.5 6.5C11.5 9.26142 9.26142 11.5 6.5 11.5C3.73858 11.5 1.5 9.26142 1.5 6.5ZM6.5 0C2.91015 0 0 2.91015 0 6.5C0 10.0899 2.91015 13 6.5 13C8.02469 13 9.42677 12.475 10.5353 11.596L13.9697 15.0303L14.5 15.5607L15.5607 14.5L15.0303 13.9697L11.596 10.5353C12.475 9.42677 13 8.02469 13 6.5C13 2.91015 10.0899 0 6.5 0ZM4.125 5.875H4.75H5.875V4.75V4.125H7.125V4.75V5.875H8.25H8.875V7.125H8.25H7.125V8.25V8.875H5.875V8.25V7.125H4.75H4.125V5.875Z",
			fill: "currentColor",
			fillRule: "evenodd"
		})
	}),
	ZoomOutIcon: (e) => (0, K.jsx)("svg", {
		"aria-hidden": "true",
		color: "currentColor",
		height: 16,
		strokeLinejoin: "round",
		viewBox: "0 0 16 16",
		width: 16,
		...e,
		children: (0, K.jsx)("path", {
			clipRule: "evenodd",
			d: "M1.5 6.5C1.5 3.73858 3.73858 1.5 6.5 1.5C9.26142 1.5 11.5 3.73858 11.5 6.5C11.5 9.26142 9.26142 11.5 6.5 11.5C3.73858 11.5 1.5 9.26142 1.5 6.5ZM6.5 0C2.91015 0 0 2.91015 0 6.5C0 10.0899 2.91015 13 6.5 13C8.02469 13 9.42677 12.475 10.5353 11.596L13.9697 15.0303L14.5 15.5607L15.5607 14.5L15.0303 13.9697L11.596 10.5353C12.475 9.42677 13 8.02469 13 6.5C13 2.91015 10.0899 0 6.5 0ZM4.125 5.875H4.75H8.25H8.875V7.125H8.25H4.75H4.125V5.875Z",
			fill: "currentColor",
			fillRule: "evenodd"
		})
	})
}, og = (0, k.createContext)(ag), sg = (e, t) => {
	if (e === t) return !0;
	if (!(e && t)) return e === t;
	let n = Object.keys(e), r = Object.keys(t);
	return n.length === r.length && n.every((n) => e[n] === t[n]);
}, cg = ({ icons: e, children: t }) => {
	let n = (0, k.useRef)(e), r = (0, k.useRef)(e ? {
		...ag,
		...e
	} : ag);
	sg(n.current, e) || (n.current = e, r.current = e ? {
		...ag,
		...e
	} : ag);
	let i = r.current;
	return (0, K.jsx)(og.Provider, {
		value: i,
		children: t
	});
}, lg = () => (0, k.useContext)(og), ug = {
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
	zoomIn: "Zoom in",
	zoomOut: "Zoom out",
	resetView: "Reset zoom and pan",
	mermaidAriaLabel: "Mermaid chart",
	errorMermaid: "Failed to render Mermaid chart",
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
}, dg = (0, k.createContext)(ug), fg = () => (0, k.useContext)(dg), pg = ({ onCopy: e, onError: t, timeout: n = 2e3, children: r, className: i, code: a, ...o }) => {
	let s = Z(), [c, l] = (0, k.useState)(!1), u = (0, k.useRef)(0), { code: d } = eg(), { isAnimating: f } = (0, k.useContext)($), p = fg(), m = a ?? d, h = async () => {
		var r;
		if (typeof window > "u" || !((r = navigator == null ? void 0 : navigator.clipboard) != null && r.writeText)) t?.(/* @__PURE__ */ Error("Clipboard API not available"));
		else try {
			c || (await navigator.clipboard.writeText(m), l(!0), e?.(), u.current = window.setTimeout(() => l(!1), n));
		} catch (e) {
			t?.(e);
		}
	};
	(0, k.useEffect)(() => () => {
		window.clearTimeout(u.current);
	}, []);
	let g = lg(), _ = c ? g.CheckIcon : g.CopyIcon;
	return (0, K.jsxs)(K.Fragment, { children: [(0, K.jsx)("button", {
		"aria-label": p.copyCode,
		className: s("cursor-pointer p-1 text-muted-foreground transition-all hover:text-foreground disabled:cursor-not-allowed disabled:opacity-50", i),
		"data-streamdown": "code-block-copy-button",
		disabled: f,
		onClick: h,
		title: p.copyCode,
		type: "button",
		...o,
		children: r ?? (0, K.jsx)(_, {
			"aria-hidden": "true",
			size: 14
		})
	}), c && (0, K.jsx)("output", {
		"aria-live": "polite",
		className: "sr-only",
		children: p.copied
	})] });
}, mg = (e, t, n) => {
	if (typeof e == "boolean") return n;
	let r = e[t];
	if (typeof r != "object") return n;
	let i = r.download;
	return typeof i == "object" && i.filename || n;
}, hg = (e, t) => {
	if (typeof e == "boolean") return {};
	let n = e[t];
	if (typeof n != "object") return {};
	let r = n.copy;
	return typeof r != "object" || !r ? {} : {
		onCopy: r.onCopy,
		onError: r.onError
	};
}, gg = {
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
}, _g = ({ onDownload: e, onError: t, language: n, children: r, className: i, code: a, ...o }) => {
	let s = Z(), { code: c } = eg(), { isAnimating: l, controls: u } = (0, k.useContext)($), d = fg(), f = lg(), p = a ?? c, m = n && n in gg ? gg[n] : "txt", h = `${mg(u, "code", "file")}.${m}`;
	return (0, K.jsx)("button", {
		"aria-label": d.downloadFile,
		className: s("cursor-pointer p-1 text-muted-foreground transition-all hover:text-foreground disabled:cursor-not-allowed disabled:opacity-50", i),
		"data-streamdown": "code-block-download-button",
		disabled: l,
		onClick: () => {
			try {
				Uh(h, p, "text/plain"), e?.();
			} catch (e) {
				t?.(e);
			}
		},
		title: d.downloadFile,
		type: "button",
		...o,
		children: r ?? (0, K.jsx)(f.DownloadIcon, { size: 14 })
	});
}, vg = () => {
	let { Loader2Icon: e } = lg(), t = Z();
	return (0, K.jsxs)("div", {
		className: t("w-full divide-y divide-border overflow-hidden rounded-xl border border-border"),
		children: [(0, K.jsx)("div", { className: t("h-[46px] w-full bg-muted/80") }), (0, K.jsx)("div", {
			className: t("flex w-full items-center justify-center p-4"),
			children: (0, K.jsx)(e, { className: t("size-4 animate-spin") })
		})]
	});
}, yg = /\.[^/.]+$/, bg = ({ node: e, className: t, src: n, alt: r, onLoad: i, onError: a, showControls: o = !0, showDownloadControl: s = !0, ...c }) => {
	let { DownloadIcon: l } = lg(), u = Z(), d = (0, k.useRef)(null), [f, p] = (0, k.useState)(!1), [m, h] = (0, k.useState)(!1), g = fg(), _ = c.width != null || c.height != null, v = (f || _) && !m && o && s, y = m && !_;
	(0, k.useEffect)(() => {
		let e = d.current;
		if (e != null && e.complete) {
			let t = e.naturalWidth > 0;
			p(t), h(!t);
		}
	}, []);
	let b = (0, k.useCallback)((e) => {
		p(!0), h(!1), i?.(e);
	}, [i]), x = (0, k.useCallback)((e) => {
		p(!1), h(!0), a?.(e);
	}, [a]), ee = async () => {
		if (n) try {
			let e = await (await fetch(n)).blob(), t = new URL(n, window.location.origin).pathname.split("/").pop() || "", i = t.split(".").pop(), a = t.includes(".") && i !== void 0 && i.length <= 4, o = "";
			if (a) o = t;
			else {
				let n = e.type, i = "png";
				n.includes("jpeg") || n.includes("jpg") ? i = "jpg" : n.includes("png") ? i = "png" : n.includes("svg") ? i = "svg" : n.includes("gif") ? i = "gif" : n.includes("webp") && (i = "webp"), o = `${(r || t || "image").replace(yg, "")}.${i}`;
			}
			Uh(o, e, e.type);
		} catch {
			window.open(n, "_blank");
		}
	};
	return n ? n === "streamdown:incomplete-image" ? (0, K.jsx)("div", {
		className: u("group relative my-4 inline-block"),
		"data-incomplete": "true",
		"data-streamdown": "image-wrapper",
		children: (0, K.jsx)("div", {
			className: u("h-24 w-48 animate-pulse rounded-lg bg-muted"),
			"data-streamdown": "image-placeholder"
		})
	}) : (0, K.jsxs)("div", {
		className: u("group relative my-4 inline-block"),
		"data-streamdown": "image-wrapper",
		children: [
			(0, K.jsx)("img", {
				alt: r,
				className: u("max-w-full rounded-lg", y && "hidden", t),
				"data-streamdown": "image",
				onError: x,
				onLoad: b,
				ref: d,
				src: n,
				...c
			}),
			y && (0, K.jsx)("span", {
				className: u("text-muted-foreground text-xs italic"),
				"data-streamdown": "image-fallback",
				children: g.imageNotAvailable
			}),
			o && (0, K.jsx)("div", {
				className: u("pointer-events-none absolute inset-0 hidden rounded-lg bg-black/10 group-hover:block"),
				"data-streamdown": "image-overlay"
			}),
			v && (0, K.jsx)("button", {
				className: u("absolute right-2 bottom-2 flex h-8 w-8 cursor-pointer items-center justify-center rounded-md border border-border bg-background/90 shadow-sm backdrop-blur-sm transition-all duration-200 hover:bg-background", "opacity-0 group-hover:opacity-100"),
				onClick: ee,
				title: g.downloadImage,
				type: "button",
				children: (0, K.jsx)(l, { size: 14 })
			})
		]
	}) : null;
}, xg = (e) => (typeof e == "function" ? e() : e) ?? document.body, Sg = 0, Cg = () => {
	Sg += 1, Sg === 1 && (document.body.style.overflow = "hidden");
}, wg = () => {
	Sg = Math.max(0, Sg - 1), Sg === 0 && (document.body.style.overflow = "");
}, Tg = ({ url: e, isOpen: t, onClose: n, onConfirm: r }) => {
	let { CheckIcon: i, CopyIcon: a, ExternalLinkIcon: o, XIcon: s } = lg(), c = Z(), [l, u] = (0, k.useState)(!1), d = fg(), { portal: f } = (0, k.useContext)($), p = (0, k.useCallback)(async () => {
		try {
			await navigator.clipboard.writeText(e), u(!0), setTimeout(() => u(!1), 2e3);
		} catch {}
	}, [e]), m = (0, k.useCallback)(() => {
		r(), n();
	}, [r, n]);
	if ((0, k.useEffect)(() => {
		if (t) {
			Cg();
			let e = (e) => {
				e.key === "Escape" && n();
			};
			return document.addEventListener("keydown", e), () => {
				document.removeEventListener("keydown", e), wg();
			};
		}
	}, [t, n]), !t || typeof document > "u") return null;
	let h = (0, K.jsx)("div", {
		className: c("fixed inset-0 z-50 flex items-center justify-center bg-background/50 backdrop-blur-sm"),
		"data-streamdown": "link-safety-modal",
		onClick: n,
		onKeyDown: (e) => {
			e.key === "Escape" && n();
		},
		role: "button",
		tabIndex: 0,
		children: (0, K.jsxs)("div", {
			className: c("relative mx-4 flex w-full max-w-md flex-col gap-4 rounded-xl border bg-background p-6 shadow-lg"),
			onClick: (e) => e.stopPropagation(),
			onKeyDown: (e) => e.stopPropagation(),
			role: "presentation",
			children: [
				(0, K.jsx)("button", {
					className: c("absolute top-4 right-4 rounded-md p-1 text-muted-foreground transition-all hover:bg-muted hover:text-foreground"),
					onClick: n,
					title: d.close,
					type: "button",
					children: (0, K.jsx)(s, { size: 16 })
				}),
				(0, K.jsxs)("div", {
					className: c("flex flex-col gap-2"),
					children: [(0, K.jsxs)("div", {
						className: c("flex items-center gap-2 font-semibold text-lg"),
						children: [(0, K.jsx)(o, { size: 20 }), (0, K.jsx)("span", { children: d.openExternalLink })]
					}), (0, K.jsx)("p", {
						className: c("text-muted-foreground text-sm"),
						children: d.externalLinkWarning
					})]
				}),
				(0, K.jsx)("div", {
					className: c("break-all rounded-md bg-muted p-3 font-mono text-sm", e.length > 100 && "max-h-32 overflow-y-auto"),
					children: e
				}),
				(0, K.jsxs)("div", {
					className: c("flex gap-2"),
					children: [(0, K.jsx)("button", {
						className: c("flex flex-1 items-center justify-center gap-2 rounded-md border bg-background px-4 py-2 font-medium text-sm transition-all hover:bg-muted"),
						onClick: p,
						type: "button",
						children: l ? (0, K.jsxs)(K.Fragment, { children: [(0, K.jsx)(i, { size: 14 }), (0, K.jsx)("span", { children: d.copied })] }) : (0, K.jsxs)(K.Fragment, { children: [(0, K.jsx)(a, { size: 14 }), (0, K.jsx)("span", { children: d.copyLink })] })
					}), (0, K.jsxs)("button", {
						className: c("flex flex-1 items-center justify-center gap-2 rounded-md bg-primary px-4 py-2 font-medium text-primary-foreground text-sm transition-all hover:bg-primary/90"),
						onClick: m,
						type: "button",
						children: [(0, K.jsx)(o, { size: 14 }), (0, K.jsx)("span", { children: d.openLink })]
					})]
				})
			]
		})
	});
	return (0, op.createPortal)(h, xg(f));
}, Eg = (0, k.createContext)(null), Dg = () => (0, k.useContext)(Eg), Og = () => Dg()?.code ?? null, kg = () => Dg()?.mermaid ?? null, Ag = (e) => {
	var t;
	let n = Dg();
	return n != null && n.renderers && e && (t = n.renderers.find((t) => Array.isArray(t.language) ? t.language.includes(e) : t.language === e)) != null ? t : null;
}, jg = /<svg\b[^>]*>/i, Mg = /\bviewBox=(['"])(.*?)\1/i, Ng = /[\s,]+/, Pg = /\swidth=(['"]).*?\1/gi, Fg = /\sheight=(['"]).*?\1/gi, Ig = /\sstyle=(['"])(.*?)\1/i, Lg = /^width\s*:/i, Rg = /^height\s*:/i, zg = /^max-width\s*:/i, Bg = /^<svg/i, Vg = (e) => {
	let t = e.match(jg);
	if (!t) return null;
	let n = t[0].match(Mg)?.[2];
	if (!n) return null;
	let r = n.trim().split(Ng).map((e) => Number.parseFloat(e));
	if (r.length < 4 || r.slice(0, 4).some(Number.isNaN)) return null;
	let i = r[2], a = r[3];
	return i > 0 && a > 0 ? {
		height: a,
		width: i
	} : null;
}, Hg = (e) => {
	let t = e.match(jg);
	if (!t) return e;
	try {
		let n = t[0], r = Vg(e);
		if (!r) return e;
		let { width: i, height: a } = r, o = n.replace(Pg, "").replace(Fg, ""), s = o.match(Ig), c = `width:${i}px;height:${a}px;max-width:none;`;
		if (s) {
			let e = s[1], t = s[2].split(";").map((e) => e.trim()).filter(Boolean).filter((e) => !(Lg.test(e) || Rg.test(e) || zg.test(e))).join(";"), n = `${c}${t ? `${t};` : ""}`;
			o = o.replace(Ig, ` style=${e}${n}${e}`);
		} else o = o.replace(Bg, `<svg style="${c}"`);
		return o = o.replace(Bg, `<svg width="${i}" height="${a}"`), e.replace(n, o);
	} catch {
		return e;
	}
}, Ug = (e) => {
	if (typeof DOMParser > "u" || typeof XMLSerializer > "u") return e;
	let t = new DOMParser().parseFromString(e, "text/html").querySelector("svg");
	return t ? new XMLSerializer().serializeToString(t) : e;
}, Wg = (e, t) => new Promise((t, n) => {
	let r = "data:image/svg+xml;base64," + btoa(unescape(encodeURIComponent(e))), i = new Image();
	i.crossOrigin = "anonymous", i.onload = () => {
		let e = document.createElement("canvas"), r = i.width * 5, a = i.height * 5;
		e.width = r, e.height = a;
		let o = e.getContext("2d");
		o ? (o.drawImage(i, 0, 0, r, a), e.toBlob((e) => {
			e ? t(e) : n(/* @__PURE__ */ Error("Failed to create PNG blob"));
		}, "image/png")) : n(/* @__PURE__ */ Error("Failed to create 2D canvas context for PNG export"));
	}, i.onerror = () => n(/* @__PURE__ */ Error("Failed to load SVG image")), i.src = r;
}), Gg = ({ chart: e, children: t, className: n, onDownload: r, config: i, onError: a }) => {
	let o = Z(), [s, c] = (0, k.useState)(!1), l = (0, k.useRef)(null), { isAnimating: u, controls: d } = (0, k.useContext)($), f = lg(), p = kg(), m = fg(), h = mg(d, "mermaid", "diagram"), g = async (t) => {
		try {
			if (t === "mmd") {
				Uh(`${h}.mmd`, e, "text/plain"), c(!1), r?.(t);
				return;
			}
			if (!p) {
				a?.(/* @__PURE__ */ Error("Mermaid plugin not available"));
				return;
			}
			let n = p.getMermaid(i), o = e.split("").reduce((e, t) => (e << 5) - e + t.charCodeAt(0) | 0, 0), s = `mermaid-${Math.abs(o)}-${Date.now()}-${Math.random().toString(36).substring(2, 9)}`, { svg: l } = await n.render(s, e);
			if (!l) {
				a?.(/* @__PURE__ */ Error("SVG not found. Please wait for the diagram to render."));
				return;
			}
			let u = Ug(l);
			if (t === "svg") {
				Uh(`${h}.svg`, u, "image/svg+xml"), c(!1), r?.(t);
				return;
			}
			if (t === "png") {
				let e = await Wg(u);
				Uh(`${h}.png`, e, "image/png"), r?.(t), c(!1);
				return;
			}
		} catch (e) {
			a?.(e);
		}
	};
	return (0, k.useEffect)(() => {
		let e = (e) => {
			let t = e.composedPath();
			l.current && !t.includes(l.current) && c(!1);
		};
		return document.addEventListener("mousedown", e), () => {
			document.removeEventListener("mousedown", e);
		};
	}, []), (0, K.jsxs)("div", {
		className: o("relative"),
		ref: l,
		children: [(0, K.jsx)("button", {
			"aria-label": m.downloadDiagram,
			className: o("cursor-pointer p-1 text-muted-foreground transition-all hover:text-foreground disabled:cursor-not-allowed disabled:opacity-50", n),
			disabled: u,
			onClick: () => c(!s),
			title: m.downloadDiagram,
			type: "button",
			children: t ?? (0, K.jsx)(f.DownloadIcon, {
				"aria-hidden": "true",
				size: 14
			})
		}), s ? (0, K.jsxs)("div", {
			className: o("absolute top-full right-0 z-10 mt-1 min-w-[120px] overflow-hidden rounded-md border border-border bg-background shadow-lg"),
			children: [
				(0, K.jsx)("button", {
					"aria-label": m.downloadDiagramAsSvg,
					className: o("w-full px-3 py-2 text-left text-sm transition-colors hover:bg-muted/40"),
					onClick: () => g("svg"),
					title: m.downloadDiagramAsSvg,
					type: "button",
					children: m.mermaidFormatSvg
				}),
				(0, K.jsx)("button", {
					"aria-label": m.downloadDiagramAsPng,
					className: o("w-full px-3 py-2 text-left text-sm transition-colors hover:bg-muted/40"),
					onClick: () => g("png"),
					title: m.downloadDiagramAsPng,
					type: "button",
					children: m.mermaidFormatPng
				}),
				(0, K.jsx)("button", {
					"aria-label": m.downloadDiagramAsMmd,
					className: o("w-full px-3 py-2 text-left text-sm transition-colors hover:bg-muted/40"),
					onClick: () => g("mmd"),
					title: m.downloadDiagramAsMmd,
					type: "button",
					children: m.mermaidFormatMmd
				})
			]
		}) : null]
	});
}, Kg = ({ chart: e, config: t, onFullscreen: n, onExit: r, className: i, ...a }) => {
	let { Maximize2Icon: o, XIcon: s } = lg(), c = Z(), [l, u] = (0, k.useState)(!1), { isAnimating: d, controls: f, portal: p } = (0, k.useContext)($), m = fg(), h = (() => {
		if (typeof f == "boolean") return f;
		let e = f.mermaid;
		return e === !1 ? !1 : e === !0 || e === void 0 || e.panZoom !== !1;
	})(), g = (() => {
		if (typeof f == "boolean") return f;
		let e = f.mermaid;
		return e === !1 ? !1 : e === !0 || e === void 0 || e.download !== !1;
	})(), _ = (() => {
		if (typeof f == "boolean") return f;
		let e = f.mermaid;
		return e === !1 ? !1 : e === !0 || e === void 0 || e.copy !== !1;
	})(), v = () => {
		u(!l);
	};
	return (0, k.useEffect)(() => {
		if (l) {
			Cg();
			let e = (e) => {
				e.key === "Escape" && u(!1);
			};
			return document.addEventListener("keydown", e), () => {
				document.removeEventListener("keydown", e), wg();
			};
		}
	}, [l]), (0, k.useEffect)(() => {
		l ? n?.() : r && r();
	}, [
		l,
		n,
		r
	]), (0, K.jsxs)(K.Fragment, { children: [(0, K.jsx)("button", {
		className: c("cursor-pointer p-1 text-muted-foreground transition-all hover:text-foreground disabled:cursor-not-allowed disabled:opacity-50", i),
		disabled: d,
		onClick: v,
		title: m.viewFullscreen,
		type: "button",
		...a,
		"aria-label": m.viewFullscreen,
		children: (0, K.jsx)(o, {
			"aria-hidden": "true",
			size: 14
		})
	}), l ? (0, op.createPortal)((0, K.jsxs)("div", {
		"aria-label": m.viewFullscreen,
		"aria-modal": "true",
		className: c("fixed inset-0 z-50 flex items-center justify-center bg-background/95 backdrop-blur-sm"),
		"data-streamdown": "mermaid-fullscreen",
		onClick: v,
		onKeyDown: (e) => {
			e.key === "Escape" && v();
		},
		role: "dialog",
		children: [(0, K.jsxs)("div", {
			className: c("absolute top-4 right-4 z-10 flex items-center gap-1"),
			onClick: (e) => e.stopPropagation(),
			onKeyDown: (e) => e.stopPropagation(),
			role: "presentation",
			children: [
				g ? (0, K.jsx)(Gg, {
					chart: e,
					config: t
				}) : null,
				_ ? (0, K.jsx)(pg, { code: e }) : null,
				(0, K.jsx)("button", {
					"aria-label": m.exitFullscreen,
					className: c("rounded-md p-2 text-muted-foreground transition-all hover:bg-muted hover:text-foreground"),
					onClick: v,
					title: m.exitFullscreen,
					type: "button",
					children: (0, K.jsx)(s, {
						"aria-hidden": "true",
						size: 20
					})
				})
			]
		}), (0, K.jsx)("div", {
			className: c("flex size-full items-center justify-center p-4"),
			onClick: (e) => e.stopPropagation(),
			onKeyDown: (e) => e.stopPropagation(),
			role: "presentation",
			children: (0, K.jsx)(Ty, {
				chart: e,
				className: c("size-full [&_svg]:h-auto [&_svg]:w-auto"),
				config: t,
				fullscreen: !0,
				showControls: h
			})
		})]
	}), xg(p)) : null] });
};
function qg(e) {
	if (e.nodeType === Node.TEXT_NODE) return e.textContent ?? "";
	if (e.nodeType !== Node.ELEMENT_NODE) return "";
	let t = e;
	return t.tagName === "BR" ? "\n" : Array.from(t.childNodes).map(qg).join("");
}
var Jg = (e) => {
	let t = [], n = [], r = e.querySelectorAll("thead th");
	for (let e of r) t.push(qg(e).trim());
	let i = e.querySelectorAll("tbody tr");
	for (let e of i) {
		let t = [], r = e.querySelectorAll("td");
		for (let e of r) t.push(qg(e).trim());
		n.push(t);
	}
	return {
		headers: t,
		rows: n
	};
}, Yg = (e) => {
	if (typeof e != "object") return ",";
	let t = e.table;
	return typeof t == "object" ? t.csvSeparator ?? "," : ",";
}, Xg = (e, t = ",") => {
	let n;
	n = t === "auto" ? Intl.NumberFormat().format(1.1).includes(",") ? ";" : "," : t;
	let { headers: r, rows: i } = e, a = (e) => {
		let t = !1;
		for (let r of e) if (r === n || r === "\"" || r === "\n" || r === "\r") {
			t = !0;
			break;
		}
		return t ? `"${e.replace(/"/g, "\"\"")}"` : e;
	}, o = r.length > 0 ? i.length + 1 : i.length, s = Array(o), c = 0;
	r.length > 0 && (s[c] = r.map(a).join(n), c += 1);
	for (let e of i) s[c] = e.map(a).join(n), c += 1;
	return s.join("\n");
}, Zg = (e) => {
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
}, Qg = (e) => {
	let t = !1;
	for (let n of e) if (n === "\\" || n === "|" || n === "\n" || n === "&" || n === "<" || n === ">") {
		t = !0;
		break;
	}
	if (!t) return e;
	let n = [];
	for (let t of e) t === "\\" ? n.push("\\\\") : t === "|" ? n.push("\\|") : t === "\n" ? n.push("<br>") : t === "&" ? n.push("&amp;") : t === "<" ? n.push("&lt;") : t === ">" ? n.push("&gt;") : n.push(t);
	return n.join("");
}, $g = (e) => {
	let { headers: t, rows: n } = e;
	if (t.length === 0) return "";
	let r = Array(n.length + 2), i = 0, a = t.map((e) => Qg(e));
	r[i] = `| ${a.join(" | ")} |`, i += 1;
	let o = Array(t.length);
	for (let e = 0; e < t.length; e += 1) o[e] = "---";
	r[i] = `| ${o.join(" | ")} |`, i += 1;
	for (let e of n) if (e.length < t.length) {
		let n = Array(t.length);
		for (let r = 0; r < t.length; r += 1) n[r] = r < e.length ? Qg(e[r]) : "";
		r[i] = `| ${n.join(" | ")} |`, i += 1;
	} else {
		let t = e.map((e) => Qg(e));
		r[i] = `| ${t.join(" | ")} |`, i += 1;
	}
	return r.join("\n");
}, e_ = ({ children: e, className: t, onCopy: n, onError: r, timeout: i = 2e3 }) => {
	let a = Z(), [o, s] = (0, k.useState)(!1), [c, l] = (0, k.useState)(!1), u = (0, k.useRef)(null), d = (0, k.useRef)(0), { isAnimating: f, controls: p } = (0, k.useContext)($), m = fg(), h = Yg(p), g = async (e) => {
		var t;
		if (typeof window > "u" || !((t = navigator == null ? void 0 : navigator.clipboard) != null && t.write)) r?.(/* @__PURE__ */ Error("Clipboard API not available"));
		else try {
			let t = (u.current?.closest("[data-streamdown=\"table-wrapper\"]"))?.querySelector("table");
			if (!t) {
				r?.(/* @__PURE__ */ Error("Table not found"));
				return;
			}
			let a = Jg(t), o = "";
			o = e === "csv" ? Xg(a, h) : e === "tsv" ? Zg(a) : $g(a);
			let c = new ClipboardItem({
				"text/plain": new Blob([o], { type: "text/plain" }),
				"text/html": new Blob([t.outerHTML], { type: "text/html" })
			});
			await navigator.clipboard.write([c]), l(!0), s(!1), n?.(e), d.current = window.setTimeout(() => l(!1), i);
		} catch (e) {
			r?.(e);
		}
	};
	(0, k.useEffect)(() => {
		let e = (e) => {
			let t = e.composedPath();
			u.current && !t.includes(u.current) && s(!1);
		};
		return document.addEventListener("mousedown", e), () => {
			document.removeEventListener("mousedown", e), window.clearTimeout(d.current);
		};
	}, []);
	let _ = lg(), v = c ? _.CheckIcon : _.CopyIcon;
	return (0, K.jsxs)("div", {
		className: a("relative"),
		ref: u,
		children: [(0, K.jsx)("button", {
			className: a("cursor-pointer p-1 text-muted-foreground transition-all hover:text-foreground disabled:cursor-not-allowed disabled:opacity-50", t),
			disabled: f,
			onClick: () => s(!o),
			title: m.copyTable,
			type: "button",
			children: e ?? (0, K.jsx)(v, {
				height: 14,
				width: 14
			})
		}), o ? (0, K.jsxs)("div", {
			className: a("absolute top-full right-0 z-20 mt-1 min-w-[120px] overflow-hidden rounded-md border border-border bg-background shadow-lg"),
			children: [
				(0, K.jsx)("button", {
					className: a("w-full px-3 py-2 text-left text-sm transition-colors hover:bg-muted/40"),
					onClick: () => g("md"),
					title: m.copyTableAsMarkdown,
					type: "button",
					children: m.tableFormatMarkdown
				}),
				(0, K.jsx)("button", {
					className: a("w-full px-3 py-2 text-left text-sm transition-colors hover:bg-muted/40"),
					onClick: () => g("csv"),
					title: m.copyTableAsCsv,
					type: "button",
					children: m.tableFormatCsv
				}),
				(0, K.jsx)("button", {
					className: a("w-full px-3 py-2 text-left text-sm transition-colors hover:bg-muted/40"),
					onClick: () => g("tsv"),
					title: m.copyTableAsTsv,
					type: "button",
					children: m.tableFormatTsv
				})
			]
		}) : null]
	});
}, t_ = ({ children: e, className: t, onDownload: n, onError: r }) => {
	let i = Z(), [a, o] = (0, k.useState)(!1), s = (0, k.useRef)(null), { isAnimating: c, controls: l } = (0, k.useContext)($), u = fg(), d = lg(), f = Yg(l), p = (e) => {
		try {
			let t = (s.current?.closest("[data-streamdown=\"table-wrapper\"]"))?.querySelector("table");
			if (!t) {
				r?.(/* @__PURE__ */ Error("Table not found"));
				return;
			}
			let i = Jg(t), a = e === "csv" ? Xg(i, f) : $g(i), c = e === "csv" ? "csv" : "md";
			Uh(`${mg(l, "table", "table")}.${c}`, a, e === "csv" ? "text/csv" : "text/markdown"), o(!1), n?.(e);
		} catch (e) {
			r?.(e);
		}
	};
	return (0, k.useEffect)(() => {
		let e = (e) => {
			let t = e.composedPath();
			s.current && !t.includes(s.current) && o(!1);
		};
		return document.addEventListener("mousedown", e), () => {
			document.removeEventListener("mousedown", e);
		};
	}, []), (0, K.jsxs)("div", {
		className: i("relative"),
		ref: s,
		children: [(0, K.jsx)("button", {
			className: i("cursor-pointer p-1 text-muted-foreground transition-all hover:text-foreground disabled:cursor-not-allowed disabled:opacity-50", t),
			disabled: c,
			onClick: () => o(!a),
			title: u.downloadTable,
			type: "button",
			children: e ?? (0, K.jsx)(d.DownloadIcon, { size: 14 })
		}), a ? (0, K.jsxs)("div", {
			className: i("absolute top-full right-0 z-20 mt-1 min-w-[120px] overflow-hidden rounded-md border border-border bg-background shadow-lg"),
			children: [(0, K.jsx)("button", {
				className: i("w-full px-3 py-2 text-left text-sm transition-colors hover:bg-muted/40"),
				onClick: () => p("csv"),
				title: u.downloadTableAsCsv,
				type: "button",
				children: u.tableFormatCsv
			}), (0, K.jsx)("button", {
				className: i("w-full px-3 py-2 text-left text-sm transition-colors hover:bg-muted/40"),
				onClick: () => p("markdown"),
				title: u.downloadTableAsMarkdown,
				type: "button",
				children: u.tableFormatMarkdown
			})]
		}) : null]
	});
}, n_ = ({ children: e, className: t, showCopy: n = !0, showDownload: r = !0 }) => {
	let { Maximize2Icon: i, XIcon: a } = lg(), o = Z(), [s, c] = (0, k.useState)(!1), { isAnimating: l, portal: u } = (0, k.useContext)($), d = fg(), f = () => {
		c(!0);
	}, p = () => {
		c(!1);
	};
	return (0, k.useEffect)(() => {
		if (s) {
			Cg();
			let e = (e) => {
				e.key === "Escape" && c(!1);
			};
			return document.addEventListener("keydown", e), () => {
				document.removeEventListener("keydown", e), wg();
			};
		}
	}, [s]), (0, K.jsxs)(K.Fragment, { children: [(0, K.jsx)("button", {
		className: o("cursor-pointer p-1 text-muted-foreground transition-all hover:text-foreground disabled:cursor-not-allowed disabled:opacity-50", t),
		disabled: l,
		onClick: f,
		title: d.viewFullscreen,
		type: "button",
		children: (0, K.jsx)(i, { size: 14 })
	}), s ? (0, op.createPortal)((0, K.jsx)("div", {
		"aria-label": d.viewFullscreen,
		"aria-modal": "true",
		className: o("fixed inset-0 z-50 flex flex-col bg-background"),
		"data-streamdown": "table-fullscreen",
		onClick: p,
		onKeyDown: (e) => {
			e.key === "Escape" && p();
		},
		role: "dialog",
		children: (0, K.jsxs)("div", {
			className: o("flex h-full flex-col"),
			"data-streamdown": "table-wrapper",
			onClick: (e) => e.stopPropagation(),
			onKeyDown: (e) => e.stopPropagation(),
			role: "presentation",
			children: [(0, K.jsxs)("div", {
				className: o("flex items-center justify-end gap-1 p-4"),
				children: [
					n ? (0, K.jsx)(e_, {}) : null,
					r ? (0, K.jsx)(t_, {}) : null,
					(0, K.jsx)("button", {
						className: o("rounded-md p-1 text-muted-foreground transition-all hover:bg-muted hover:text-foreground"),
						onClick: p,
						title: d.exitFullscreen,
						type: "button",
						children: (0, K.jsx)(a, { size: 20 })
					})
				]
			}), (0, K.jsx)("div", {
				className: o("flex-1 overflow-auto p-4 pt-0 [&_thead]:sticky [&_thead]:top-0 [&_thead]:z-10"),
				children: (0, K.jsx)("table", {
					className: o("w-full border-collapse border border-border"),
					"data-streamdown": "table",
					children: e
				})
			})]
		})
	}), xg(u)) : null] });
}, r_ = ({ children: e, className: t, maxHeight: n, showControls: r, showCopy: i = !0, showDownload: a = !0, showFullscreen: o = !0, ...s }) => {
	let c = Z(), { isAnimating: l } = (0, k.useContext)($), u = Kh(n), d = qh(l, !!u, e), f = r && i, p = r && a, m = r && o, h = f || p || m;
	return (0, K.jsxs)("div", {
		className: c("my-4 flex flex-col gap-2 rounded-lg border border-border bg-sidebar p-2"),
		"data-streamdown": "table-wrapper",
		children: [h ? (0, K.jsxs)("div", {
			className: c("flex items-center justify-end gap-1"),
			children: [
				f ? (0, K.jsx)(e_, {}) : null,
				p ? (0, K.jsx)(t_, {}) : null,
				m ? (0, K.jsx)(n_, {
					showCopy: f,
					showDownload: p,
					children: e
				}) : null
			]
		}) : null, (0, K.jsx)("div", {
			className: c("border-collapse overflow-x-auto overflow-y-auto rounded-md border border-border bg-background"),
			ref: d,
			style: u ? { maxHeight: u } : void 0,
			children: (0, K.jsx)("table", {
				className: c("w-full divide-y divide-border", t),
				"data-streamdown": "table",
				...s,
				children: e
			})
		})]
	});
}, i_ = /startLine=(\d+)/, a_ = /\bnoLineNumbers\b/, o_ = (0, k.lazy)(() => import("./mermaid-WFAXDIZX-CzOZGoUk.js").then((e) => ({ default: e.Mermaid }))), s_ = /language-([^\s]+)/, c_ = "node";
function Q(e, t) {
	let n = Object.keys(e);
	if (n.length !== Object.keys(t).length) return !1;
	let r = e, i = t;
	for (let e of n) if (e !== c_ && !Object.is(r[e], i[e])) return !1;
	return !0;
}
function l_(e, t) {
	return e?.properties?.metastring === t?.properties?.metastring;
}
var u_ = (e, t) => typeof e == "boolean" ? e : e[t] !== !1, d_ = (e, t) => {
	if (typeof e == "boolean") return e;
	let n = e.table;
	return n === !1 ? !1 : n === !0 || n === void 0 || n[t] !== !1;
}, f_ = (e, t) => {
	if (typeof e == "boolean") return e;
	let n = e.code;
	return n === !1 ? !1 : n === !0 || n === void 0 || n[t] !== !1;
}, p_ = (e, t) => {
	if (typeof e == "boolean") return e;
	let n = e.mermaid;
	return n === !1 ? !1 : n === !0 || n === void 0 || n[t] !== !1;
}, m_ = (e, t) => {
	if (typeof e == "boolean") return e;
	let n = e.image;
	return n === !1 ? !1 : n === !0 || n === void 0 || n[t] !== !1;
}, h_ = (0, k.memo)(({ children: e, className: t, node: n, ...r }) => {
	let i = Z();
	return (0, K.jsx)("ol", {
		className: i("list-inside list-decimal whitespace-normal [li_&]:pl-6", t),
		"data-streamdown": "ordered-list",
		...r,
		children: e
	});
}, (e, t) => Q(e, t));
h_.displayName = "MarkdownOl";
var g_ = (0, k.memo)(({ children: e, className: t, node: n, ...r }) => {
	let i = Z(), a = Array.isArray(e) ? e.filter((e) => e !== "\n" && e !== "") : [e], o = a.length === 1 && (0, k.isValidElement)(a[0]) && (a[0].type === z_ || a[0].props.node?.tagname === "p") ? a[0].props.children : e;
	return (0, K.jsx)("li", {
		className: i("py-1 [&>p]:inline", t),
		"data-streamdown": "list-item",
		...r,
		children: o
	});
}, (e, t) => Q(e, t));
g_.displayName = "MarkdownLi";
var __ = (0, k.memo)(({ children: e, className: t, node: n, ...r }) => {
	let i = Z();
	return (0, K.jsx)("ul", {
		className: i("list-inside list-disc whitespace-normal [li_&]:pl-6", t),
		"data-streamdown": "unordered-list",
		...r,
		children: e
	});
}, (e, t) => Q(e, t));
__.displayName = "MarkdownUl";
var v_ = (0, k.memo)(({ className: e, node: t, ...n }) => {
	let r = Z();
	return (0, K.jsx)("hr", {
		className: r("my-6 border-border", e),
		"data-streamdown": "horizontal-rule",
		...n
	});
}, (e, t) => Q(e, t));
v_.displayName = "MarkdownHr";
var y_ = (0, k.memo)(({ children: e, className: t, node: n, ...r }) => {
	let i = Z();
	return (0, K.jsx)("span", {
		className: i("font-semibold", t),
		"data-streamdown": "strong",
		...r,
		children: e
	});
}, (e, t) => Q(e, t));
y_.displayName = "MarkdownStrong";
var b_ = (0, k.memo)(({ children: e, className: t, href: n, node: r, ...i }) => {
	let a = Z(), { linkSafety: o } = (0, k.useContext)($), [s, c] = (0, k.useState)(!1), l = n === "streamdown:incomplete-link", u = (0, k.useCallback)(async (e) => {
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
	]), d = (0, k.useCallback)(() => {
		n && window.open(n, "_blank", "noreferrer");
	}, [n]), f = (0, k.useCallback)(() => {
		c(!1);
	}, []), p = {
		url: n ?? "",
		isOpen: s,
		onClose: f,
		onConfirm: d
	};
	return o != null && o.enabled && n ? (0, K.jsxs)(K.Fragment, { children: [(0, K.jsx)("button", {
		className: a("wrap-anywhere appearance-none text-left font-medium text-primary underline", t),
		"data-incomplete": l,
		"data-streamdown": "link",
		onClick: u,
		type: "button",
		children: e
	}), o.renderModal ? o.renderModal(p) : (0, K.jsx)(Tg, { ...p })] }) : (0, K.jsx)("a", {
		className: a("wrap-anywhere font-medium text-primary underline", t),
		"data-incomplete": l,
		"data-streamdown": "link",
		href: n,
		rel: "noreferrer",
		target: "_blank",
		...i,
		children: e
	});
}, (e, t) => Q(e, t));
b_.displayName = "MarkdownA";
var x_ = (0, k.memo)(({ children: e, className: t, node: n, ...r }) => {
	let i = Z();
	return (0, K.jsx)("h1", {
		className: i("mt-6 mb-2 font-semibold text-3xl", t),
		"data-streamdown": "heading-1",
		...r,
		children: e
	});
}, (e, t) => Q(e, t));
x_.displayName = "MarkdownH1";
var S_ = (0, k.memo)(({ children: e, className: t, node: n, ...r }) => {
	let i = Z();
	return (0, K.jsx)("h2", {
		className: i("mt-6 mb-2 font-semibold text-2xl", t),
		"data-streamdown": "heading-2",
		...r,
		children: e
	});
}, (e, t) => Q(e, t));
S_.displayName = "MarkdownH2";
var C_ = (0, k.memo)(({ children: e, className: t, node: n, ...r }) => {
	let i = Z();
	return (0, K.jsx)("h3", {
		className: i("mt-6 mb-2 font-semibold text-xl", t),
		"data-streamdown": "heading-3",
		...r,
		children: e
	});
}, (e, t) => Q(e, t));
C_.displayName = "MarkdownH3";
var w_ = (0, k.memo)(({ children: e, className: t, node: n, ...r }) => {
	let i = Z();
	return (0, K.jsx)("h4", {
		className: i("mt-6 mb-2 font-semibold text-lg", t),
		"data-streamdown": "heading-4",
		...r,
		children: e
	});
}, (e, t) => Q(e, t));
w_.displayName = "MarkdownH4";
var T_ = (0, k.memo)(({ children: e, className: t, node: n, ...r }) => {
	let i = Z();
	return (0, K.jsx)("h5", {
		className: i("mt-6 mb-2 font-semibold text-base", t),
		"data-streamdown": "heading-5",
		...r,
		children: e
	});
}, (e, t) => Q(e, t));
T_.displayName = "MarkdownH5";
var E_ = (0, k.memo)(({ children: e, className: t, node: n, ...r }) => {
	let i = Z();
	return (0, K.jsx)("h6", {
		className: i("mt-6 mb-2 font-semibold text-sm", t),
		"data-streamdown": "heading-6",
		...r,
		children: e
	});
}, (e, t) => Q(e, t));
E_.displayName = "MarkdownH6";
var D_ = (0, k.memo)(({ children: e, className: t, node: n, ...r }) => {
	let { controls: i, tableMaxHeight: a } = (0, k.useContext)($), o = u_(i, "table"), s = d_(i, "copy"), c = d_(i, "download"), l = d_(i, "fullscreen");
	return (0, K.jsx)(r_, {
		className: t,
		maxHeight: a,
		showControls: o,
		showCopy: s,
		showDownload: c,
		showFullscreen: l,
		...r,
		children: e
	});
}, (e, t) => Q(e, t));
D_.displayName = "MarkdownTable";
var O_ = (0, k.memo)(({ children: e, className: t, node: n, ...r }) => {
	let i = Z();
	return (0, K.jsx)("thead", {
		className: i("bg-muted/80", t),
		"data-streamdown": "table-header",
		...r,
		children: e
	});
}, (e, t) => Q(e, t));
O_.displayName = "MarkdownThead";
var k_ = (0, k.memo)(({ children: e, className: t, node: n, ...r }) => {
	let i = Z();
	return (0, K.jsx)("tbody", {
		className: i("divide-y divide-border", t),
		"data-streamdown": "table-body",
		...r,
		children: e
	});
}, (e, t) => Q(e, t));
k_.displayName = "MarkdownTbody";
var A_ = (0, k.memo)(({ children: e, className: t, node: n, ...r }) => {
	let i = Z();
	return (0, K.jsx)("tr", {
		className: i("border-border", t),
		"data-streamdown": "table-row",
		...r,
		children: e
	});
}, (e, t) => Q(e, t));
A_.displayName = "MarkdownTr";
var j_ = (0, k.memo)(({ children: e, className: t, node: n, ...r }) => {
	let i = Z();
	return (0, K.jsx)("th", {
		className: i("whitespace-nowrap px-4 py-2 text-left font-semibold text-sm", t),
		"data-streamdown": "table-header-cell",
		...r,
		children: e
	});
}, (e, t) => Q(e, t));
j_.displayName = "MarkdownTh";
var M_ = (0, k.memo)(({ children: e, className: t, node: n, ...r }) => {
	let i = Z();
	return (0, K.jsx)("td", {
		className: i("px-4 py-2 text-sm", t),
		"data-streamdown": "table-cell",
		...r,
		children: e
	});
}, (e, t) => Q(e, t));
M_.displayName = "MarkdownTd";
var N_ = (0, k.memo)(({ children: e, className: t, node: n, ...r }) => {
	let i = Z();
	return (0, K.jsx)("blockquote", {
		className: i("my-4 border-muted-foreground/30 border-l-4 pl-4 text-muted-foreground italic", t),
		"data-streamdown": "blockquote",
		...r,
		children: e
	});
}, (e, t) => Q(e, t));
N_.displayName = "MarkdownBlockquote";
var P_ = (0, k.memo)(({ children: e, className: t, node: n, ...r }) => {
	let i = Z();
	return (0, K.jsx)("sup", {
		className: i("text-sm", t),
		"data-streamdown": "superscript",
		...r,
		children: e
	});
}, (e, t) => Q(e, t));
P_.displayName = "MarkdownSup";
var F_ = (0, k.memo)(({ children: e, className: t, node: n, ...r }) => {
	let i = Z();
	return (0, K.jsx)("sub", {
		className: i("text-sm", t),
		"data-streamdown": "subscript",
		...r,
		children: e
	});
}, (e, t) => Q(e, t));
F_.displayName = "MarkdownSub";
var I_ = (0, k.memo)(({ children: e, className: t, node: n, ...r }) => {
	if ("data-footnotes" in r) {
		let n = (e) => {
			if (!(0, k.isValidElement)(e)) return !1;
			let t = Array.isArray(e.props.children) ? e.props.children : [e.props.children], n = !1, r = !1;
			for (let e of t) if (e) {
				if (typeof e == "string") e.trim() !== "" && (n = !0);
				else if ((0, k.isValidElement)(e)) {
					if (e.props?.["data-footnote-backref"] !== void 0) r = !0;
					else {
						let t = Array.isArray(e.props.children) ? e.props.children : [e.props.children];
						for (let e of t) if (typeof e == "string" && e.trim() !== "" || (0, k.isValidElement)(e) && e.props?.["data-footnote-backref"] === void 0) {
							n = !0;
							break;
						}
					}
				}
			}
			return r && !n;
		}, i = Array.isArray(e) ? e.map((e) => {
			if (!(0, k.isValidElement)(e)) return e;
			if (e.type === h_) {
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
		return (Array.isArray(i) ? i.some((e) => e !== null) : i !== null) ? (0, K.jsx)("section", {
			className: t,
			...r,
			children: i
		}) : null;
	}
	return (0, K.jsx)("section", {
		className: t,
		...r,
		children: e
	});
}, (e, t) => Q(e, t));
I_.displayName = "MarkdownSection";
var L_ = (0, k.memo)(({ node: e, className: t, children: n, ...r }) => {
	let i = Z(), a = !("data-block" in r), { mermaid: o, controls: s, lineNumbers: c } = (0, k.useContext)($), l = kg(), u = zh(), d = (t?.match(s_))?.at(1) ?? "", f = Ag(d);
	if (a) return (0, K.jsx)("code", {
		className: i("rounded bg-muted px-1.5 py-0.5 font-mono text-sm", t),
		"data-streamdown": "inline-code",
		...r,
		children: n
	});
	let p = e?.properties?.metastring, m = p?.match(i_), h = m ? Number.parseInt(m[1], 10) : void 0, g = h !== void 0 && h >= 1 ? h : void 0, _ = !(p && a_.test(p)) && c !== !1, v = "";
	if ((0, k.isValidElement)(n) && n.props && typeof n.props == "object" && "children" in n.props && typeof n.props.children == "string" ? v = n.props.children : typeof n == "string" && (v = n), f) {
		let e = f.component;
		return (0, K.jsx)(k.Suspense, {
			fallback: (0, K.jsx)(vg, {}),
			children: (0, K.jsx)(e, {
				code: v,
				isIncomplete: u,
				language: d,
				meta: p
			})
		});
	}
	if (d === "mermaid" && l) {
		let e = u_(s, "mermaid"), n = p_(s, "download"), r = p_(s, "copy"), a = p_(s, "fullscreen"), c = p_(s, "panZoom"), l = e && (n || r || a);
		return (0, K.jsx)(k.Suspense, {
			fallback: (0, K.jsx)(vg, {}),
			children: (0, K.jsxs)("div", {
				className: i("group relative my-4 flex w-full flex-col gap-2 rounded-xl border border-border bg-sidebar p-2", t),
				"data-streamdown": "mermaid-block",
				children: [
					(0, K.jsx)("div", {
						className: i("flex h-8 items-center text-muted-foreground text-xs"),
						children: (0, K.jsx)("span", {
							className: i("ml-1 font-mono lowercase"),
							children: "mermaid"
						})
					}),
					l ? (0, K.jsx)("div", {
						className: i("pointer-events-none absolute top-2 right-2 z-10 flex items-center"),
						children: (0, K.jsxs)("div", {
							className: i("pointer-events-auto flex shrink-0 items-center gap-2 rounded-md border border-sidebar bg-sidebar/80 px-1.5 py-1 supports-[backdrop-filter]:bg-sidebar/70 supports-[backdrop-filter]:backdrop-blur"),
							"data-streamdown": "mermaid-block-actions",
							children: [
								n ? (0, K.jsx)(Gg, {
									chart: v,
									config: o?.config
								}) : null,
								r ? (0, K.jsx)(pg, {
									code: v,
									...hg(s, "mermaid")
								}) : null,
								a ? (0, K.jsx)(Kg, {
									chart: v,
									config: o?.config
								}) : null
							]
						})
					}) : null,
					(0, K.jsx)("div", {
						className: i("overflow-hidden rounded-md border border-border bg-background"),
						children: (0, K.jsx)(o_, {
							chart: v,
							config: o?.config,
							showControls: c
						})
					})
				]
			})
		});
	}
	let y = u_(s, "code"), b = f_(s, "download"), x = f_(s, "copy"), { "data-block": ee, ...S } = r;
	return (0, K.jsx)(ig, {
		className: t,
		code: v,
		isIncomplete: u,
		language: d,
		lineNumbers: _,
		startLine: g,
		...S,
		children: y ? (0, K.jsxs)(K.Fragment, { children: [b ? (0, K.jsx)(_g, {
			code: v,
			language: d
		}) : null, x ? (0, K.jsx)(pg, { ...hg(s, "code") }) : null] }) : null
	});
}, (e, t) => Q(e, t) && l_(e.node, t.node));
L_.displayName = "MarkdownCode";
var R_ = (0, k.memo)(({ node: e, className: t, ...n }) => {
	let { controls: r } = (0, k.useContext)($), i = u_(r, "image"), a = i && m_(r, "download");
	return (0, K.jsx)(bg, {
		className: t,
		node: e,
		showControls: i,
		showDownloadControl: a,
		...n
	});
}, (e, t) => Q(e, t));
R_.displayName = "MarkdownImg";
var z_ = (0, k.memo)(({ children: e, node: t, ...n }) => {
	let r = (Array.isArray(e) ? e : [e]).filter((e) => e != null && e !== "");
	if (r.length === 1 && (0, k.isValidElement)(r[0])) {
		let t = r[0].props.node?.tagName;
		if (t === "img" || t === "code" && "data-block" in r[0].props) return (0, K.jsx)(K.Fragment, { children: e });
	}
	return (0, K.jsx)("p", {
		...n,
		children: e
	});
}, (e, t) => Q(e, t));
z_.displayName = "MarkdownParagraph";
var B_ = {
	ol: h_,
	li: g_,
	ul: __,
	hr: v_,
	strong: y_,
	a: b_,
	h1: x_,
	h2: S_,
	h3: C_,
	h4: w_,
	h5: T_,
	h6: E_,
	table: D_,
	thead: O_,
	tbody: k_,
	tr: A_,
	th: j_,
	td: M_,
	blockquote: N_,
	code: L_,
	img: R_,
	pre: ({ children: e }) => (0, k.isValidElement)(e) ? (0, k.cloneElement)(e, { "data-block": "true" }) : e,
	sup: P_,
	sub: F_,
	p: z_,
	section: I_
}, V_ = /[\u0590-\u08FF\uFB1D-\uFDFF\uFE70-\uFEFF]/, H_ = /\p{L}/u;
function U_(e) {
	let t = e.replace(/(```|~~~)[\s\S]*?\1/g, "").replace(/^#{1,6}\s+/gm, "").replace(/(\*{1,3}|_{1,3})/g, "").replace(/`[^`]*`/g, "").replace(/\[([^\]]*)\]\([^)]*\)/g, "$1").replace(/^[\s>*\-+\d.]+/gm, ""), n, r = 0, i = 0;
	for (let e of t) V_.test(e) ? (n ??= "rtl", i += 1) : H_.test(e) && (n ??= "ltr", r += 1);
	return i > r ? "rtl" : r > i ? "ltr" : n ?? "ltr";
}
var W_ = /^[ \t]{0,3}(`{3,}|~{3,})/, G_ = /^\|?[ \t]*:?-{1,}:?[ \t]*(\|[ \t]*:?-{1,}:?[ \t]*)*\|?$/, K_ = (e) => {
	let t = e.split("\n"), n = null, r = 0;
	for (let e of t) {
		let t = W_.exec(e);
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
}, q_ = (e) => {
	let t = e.split("\n");
	for (let e of t) {
		let t = e.trim();
		if (t.length > 0 && t.includes("|") && G_.test(t)) return !0;
	}
	return !1;
}, J_ = () => (e) => {
	x(e, "html", (e, t, n) => {
		!n || typeof t != "number" || (n.children[t] = {
			type: "text",
			value: e.value
		});
	});
}, Y_ = [], X_ = { allowDangerousHtml: !0 }, Z_ = /* @__PURE__ */ new WeakMap(), Q_ = new class {
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
						let n = Z_.get(e);
						n || (n = e.name, Z_.set(e, n)), t += n;
					} else t += String(e);
					t += ":", t += JSON.stringify(n);
				} else if (typeof r == "function") {
					let e = Z_.get(r);
					e || (e = r.name, Z_.set(r, e)), t += e;
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
}(), $_ = (e) => {
	let t = ev(e), n = e.children || "";
	return sv(t.runSync(t.parse(n), n), e);
}, ev = (e) => {
	let t = Q_.get(e);
	if (t) return t;
	let n = nv(e);
	return Q_.set(e, n), n;
}, tv = (e) => e.some((e) => Array.isArray(e) ? e[0] === Fo : e === Fo), nv = (e) => {
	let t = e.rehypePlugins || Y_, n = e.remarkPlugins || Y_, r = tv(t) ? n : [...n, J_], i = e.remarkRehypeOptions ? {
		...X_,
		...e.remarkRehypeOptions
	} : X_;
	return Xf().use(Ad).use(r).use(Sf, i).use(t);
}, rv = (e) => e, iv = (e, t, n, r) => {
	n ? e.children.splice(t, 1) : e.children[t] = {
		type: "text",
		value: r
	};
}, av = (e, t) => {
	for (let n in kd) if (Object.hasOwn(kd, n) && Object.hasOwn(e.properties, n)) {
		let r = e.properties[n], i = kd[n];
		(i === null || i.includes(e.tagName)) && (e.properties[n] = t(String(r || ""), n, e) ?? void 0);
	}
}, ov = (e, t, n, r, i, a) => {
	let o = !1;
	return r ? o = !r.includes(e.tagName) : i && (o = i.includes(e.tagName)), !o && a && typeof t == "number" && (o = !a(e, t, n)), o;
}, sv = (e, t) => {
	let { allowElement: n, allowedElements: r, disallowedElements: i, skipHtml: a, unwrapDisallowed: o, urlTransform: s } = t;
	if (n || r || i || a || s) {
		let t = s || rv;
		x(e, (e, s, c) => {
			if (e.type === "raw" && c && typeof s == "number") return iv(c, s, a, e.value), s;
			if (e.type === "element" && (av(e, t), ov(e, s, c, r, i, n) && c && typeof s == "number")) return o && e.children ? c.children.splice(s, 1, ...e.children) : c.children.splice(s, 1), s;
		});
	}
	return od(e, {
		Fragment: K.Fragment,
		components: t.components,
		ignoreInvalidStyle: !0,
		jsx: K.jsx,
		jsxs: K.jsxs,
		passKeys: !0,
		passNode: !0
	});
}, cv = /\[\^[\w-]{1,200}\](?!:)/, lv = /\[\^[\w-]{1,200}\]:/, uv = /<([A-Za-z][\w:-]*)[\s>/]/, dv = /^ {0,3}\[[^\]]+\]:/m, fv = /* @__PURE__ */ new Set([
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
]), pv = /* @__PURE__ */ new Map(), mv = /* @__PURE__ */ new Map(), hv = (e) => {
	let t = e.toLowerCase(), n = pv.get(t);
	if (n) return n;
	let r = RegExp(`<${t}(?=[\\s>/])[^>]*>`, "gi");
	return pv.set(t, r), r;
}, gv = (e) => {
	let t = e.toLowerCase(), n = mv.get(t);
	if (n) return n;
	let r = RegExp(`</${t}(?=[\\s>])[^>]*>`, "gi");
	return mv.set(t, r), r;
}, _v = (e, t) => {
	if (fv.has(t.toLowerCase())) return 0;
	let n = e.match(hv(t));
	if (!n) return 0;
	let r = 0;
	for (let e of n) e.trimEnd().endsWith("/>") || (r += 1);
	return r;
}, vv = (e, t) => {
	let n = e.match(gv(t));
	return n ? n.length : 0;
}, yv = (e) => {
	let t = 0;
	for (let n = 0; n < e.length - 1; n += 1) e[n] === "$" && e[n + 1] === "$" && (t += 1, n += 1);
	return t;
}, bv = /\r\n|\r/g, xv = (e) => {
	let t = new Ym({ gfm: !0 });
	return {
		tokens: t.blockTokens(e),
		hasLinkDefinitions: Object.keys(t.tokens.links).length > 0
	};
}, Sv = null, Cv = "\n\n", wv = (e) => {
	for (let t = e.length - 3; t >= 0; --t) if (e[t].endsWith(Cv)) return t + 1;
	return 0;
}, Tv = (e) => ` ${e}`.slice(1), Ev = (e) => e.map(Tv), Dv = (e) => {
	let t = [], n = [], r = !1;
	for (let i of e) {
		let e = i.raw, a = t.length;
		if (n.length > 0) {
			t[a - 1] += e;
			let r = n.at(-1), i = _v(e, r), o = vv(e, r);
			for (let e = 0; e < i; e += 1) n.push(r);
			for (let e = 0; e < o; e += 1) n.length > 0 && n.at(-1) === r && n.pop();
		} else {
			if (i.type === "html" && i.block) {
				let t = e.match(uv);
				if (t) {
					let r = t[1];
					_v(e, r) > vv(e, r) && n.push(r);
				}
			}
			if (i.type === "space" && a > 0) t[a - 1] += e;
			else {
				if (a > 0 && !r) {
					let n = t[a - 1];
					if (yv(n) % 2 == 1) {
						t[a - 1] = n + e;
						continue;
					}
				}
				t.push(e), i.type !== "space" && (r = i.type === "code");
			}
		}
	}
	return t;
}, Ov = (e, t) => {
	if (t.length <= e.input.length || t.slice(0, e.input.length) !== e.input || dv.test(t)) return null;
	let n = wv(e.blocks);
	if (n === 0) return null;
	let r = e.verifiedCount, i = e.verifiedLength;
	if (r > n) {
		r = n, i = 0;
		for (let t = 0; t < r; t += 1) i += e.blocks[t].length;
	}
	for (; r < n && t.startsWith(e.blocks[r], i);) i += e.blocks[r].length, r += 1;
	if (r !== n) return null;
	let a = xv(t.slice(i));
	if (a.hasLinkDefinitions) return null;
	let o = Ev(Dv(a.tokens));
	return {
		input: t,
		blocks: e.blocks.slice(0, n).concat(o),
		verifiedCount: r,
		verifiedLength: i
	};
}, kv = (e) => {
	let t = cv.test(e), n = lv.test(e);
	if (t || n) return [e];
	let r = e.includes("\r") ? e.replace(bv, "\n") : e;
	if (Sv?.input === r) return Sv.blocks;
	let i = Sv ? Ov(Sv, r) : null;
	if (!i) {
		let { tokens: e, hasLinkDefinitions: t } = xv(r);
		i = {
			input: r,
			blocks: t ? [Tv(r)] : Ev(Dv(e)),
			verifiedCount: 0,
			verifiedLength: 0
		};
	}
	return Sv = i, i.blocks;
}, Av = /^\n*/, jv = /\n*$/, Mv = /\n\n/g, Nv = (e, t, n) => t.includes("\n") ? `${e}${t.replace(Mv, "\n<!---->\n").replace(Av, "\n\n").replace(jv, "\n\n")}${n}

` : e + t + n, Pv = (e, t) => {
	let n = RegExp(`<(${t})(?=[\\s>/])([^>]*)>`, "gi"), r = "", i = 0, a = n.exec(e);
	for (; a;) {
		let t = a[0], o = a[1], s = a.index + t.length, c = e.slice(s);
		if (RegExp(`</${o}\\s*>`, "i").test(c)) {
			a = n.exec(e);
			continue;
		}
		if (r += e.slice(i, a.index), c.length === 0) {
			r += t, i = s, a = n.exec(e);
			continue;
		}
		if (c.startsWith("\n\n")) {
			let n = c.slice(2).replace(Mv, "\n<!---->\n");
			r += `${t}

${n}`, i = e.length;
			break;
		}
		if (!c.startsWith("\n")) {
			r += t, i = s, a = n.exec(e);
			continue;
		}
		let l = c.slice(1);
		if (l.trim().length === 0) {
			r += t, i = s, a = n.exec(e);
			continue;
		}
		let u = l.replace(Mv, "\n<!---->\n");
		r += `${t}

${u}`, i = e.length;
		break;
	}
	return i === 0 ? e : r + e.slice(i);
}, Fv = (e, t) => {
	if (!t.length) return e;
	let n = e;
	for (let e of t) {
		let t = RegExp(`(<${e}(?=[\\s>/])[^>]*>)([\\s\\S]*?)(</${e}\\s*>)`, "gi");
		n = n.replace(t, (e, t, n, r) => Nv(t, n, r)), n = Pv(n, e);
	}
	return n;
}, Iv = /([\\`*_~[\]|])/g, Lv = (e) => e.replace(Iv, "\\$1"), Rv = (e, t) => {
	if (!t.length) return e;
	let n = e;
	for (let e of t) {
		let t = RegExp(`(<${e}(?=[\\s>/])[^>]*>)([\\s\\S]*?)(</${e}\\s*>)`, "gi");
		n = n.replace(t, (e, t, n, r) => t + Lv(n).replace(/\n\n/g, "&#10;&#10;") + r);
	}
	return n;
}, zv = /* @__PURE__ */ new Set([
	"blockquote",
	"dd",
	"dt",
	"figcaption",
	"h1",
	"h2",
	"h3",
	"h4",
	"h5",
	"h6",
	"li",
	"p",
	"td",
	"th"
]), Bv = /* @__PURE__ */ new Set([
	"code",
	"kbd",
	"pre",
	"samp",
	"var"
]);
function Vv(e) {
	return e.children.map((e) => e.type === "text" ? e.value : e.type === "element" && !Bv.has(e.tagName) ? Vv(e) : "").join("");
}
function Hv() {
	return (e) => {
		x(e, "element", (e) => {
			Bv.has(e.tagName) ? (e.properties ??= {}, e.properties.dir = "ltr") : zv.has(e.tagName) && (e.properties ??= {}, typeof e.properties.dir != "string" && (e.properties.dir = U_(Vv(e))));
		});
	};
}
var Uv = (e) => e.type === "text" ? e.value : "children" in e && Array.isArray(e.children) ? e.children.map(Uv).join("") : "", Wv = (e) => (t) => {
	if (!e || e.length === 0) return;
	let n = new Set(e.map((e) => e.toLowerCase()));
	x(t, "element", (e) => {
		if (n.has(e.tagName.toLowerCase())) {
			let t = Uv(e);
			e.children = t ? [{
				type: "text",
				value: t
			}] : [];
		}
	});
}, Gv = () => (e) => {
	x(e, "code", (e) => {
		e.meta && (e.data = e.data ?? {}, e.data.hProperties = {
			...e.data.hProperties ?? {},
			metastring: e.meta
		});
	});
}, Kv = /^([a-zA-Z][a-zA-Z\d+\-.]*:)/, qv = (e) => new Set(e.map((e) => e.trim().toLowerCase()).filter((e) => e.length > 0).map((e) => e.endsWith(":") ? e : `${e}:`));
function Jv(e, t, n) {
	let r = e.position?.start?.offset;
	if (typeof r == "number" && n.charCodeAt(r) === 91 || e.children.length !== 1) return !1;
	let [i] = e.children;
	if (i.type !== "text") return !1;
	let a = Kv.exec(e.url);
	if (!a) return !1;
	let o = a[1].toLowerCase();
	return t.has(o) ? o === "mailto:" ? e.url === `mailto:${i.value}` : e.url === i.value || e.url === `${o}//${i.value}` : !1;
}
var Yv = (e = []) => {
	let t = qv(e);
	return (e, n) => {
		if (t.size === 0) return;
		let r = String(n.value ?? "");
		x(e, "link", (e, n, i) => {
			if (i && n !== void 0 && Jv(e, t, r)) return i.children.splice(n, 1, ...e.children), n;
		});
	};
}, Xv = "data-sd-caret-hidden", Zv = () => {}, Qv = () => Zv, $v = (e) => {
	let t = (0, k.useRef)(null), n = (0, k.useRef)(null), r = (0, k.useSyncExternalStore)(Qv, () => !0, () => !1);
	return (0, k.useLayoutEffect)(() => {
		var r;
		let i = e && (r = t.current?.lastElementChild) != null ? r : null, a = n.current;
		a !== i && (a?.removeAttribute(Xv), i?.setAttribute(Xv, ""), n.current = i);
	}), {
		containerRef: t,
		showCaret: r || !e
	};
}, ey = /^[a-z]/, ty = /^[ \t]*<[\w!/?-]/, ny = /(^|\n)[ \t]{4,}(?=<[\w!/?-])/g, ry = (e) => typeof e != "string" || e.length === 0 || !ty.test(e) ? e : e.replace(ny, "$1"), iy = {
	...Lo,
	clobberPrefix: "",
	protocols: {
		...Lo.protocols,
		href: [
			...Lo.protocols?.href ?? [],
			"tel",
			"streamdown"
		],
		src: [...Lo.protocols?.src ?? [], "streamdown"]
	},
	attributes: {
		...Lo.attributes,
		code: [...Lo.attributes?.code ?? [], "metastring"]
	}
}, ay = {
	raw: Fo,
	sanitize: [es, iy],
	harden: [ye, {
		allowedImagePrefixes: ["*"],
		allowedLinkPrefixes: ["*"],
		allowedProtocols: ["*"],
		defaultOrigin: void 0,
		allowDataImages: !0
	}]
}, oy = {
	gfm: [xc, {}],
	codeMeta: Gv
}, sy = Object.values(ay), cy = Object.values(oy), ly = {
	block: " ▋",
	circle: " ●"
}, uy = ["github-light", "github-dark"], dy = { enabled: !0 }, $ = (0, k.createContext)({
	codeBlockMaxHeight: 400,
	shikiTheme: uy,
	controls: !0,
	isAnimating: !1,
	lineNumbers: !0,
	mode: "streaming",
	mermaid: void 0,
	linkSafety: dy,
	portal: void 0,
	tableMaxHeight: 300
}), fy = (0, k.memo)(({ content: e, shouldParseIncompleteMarkdown: t, shouldNormalizeHtmlIndentation: n, index: r, isIncomplete: i, dir: a, animatePlugin: o, ...s }) => {
	(0, k.useLayoutEffect)(() => {
		o?.commit();
	});
	let c = typeof e == "string" && n ? ry(e) : e, l = (0, K.jsx)($_, {
		...s,
		children: c
	});
	return (0, K.jsx)(Rh.Provider, {
		value: i,
		children: a ? (0, K.jsx)("div", {
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
	return e.rehypePlugins === t.rehypePlugins && e.remarkPlugins === t.remarkPlugins && !!e.animatePlugin == !!t.animatePlugin;
});
fy.displayName = "Block";
var py = (e) => {
	let t = {};
	for (let [n, r] of Object.entries(e)) (n.startsWith("data-") || n.startsWith("aria-")) && (t[n] = r);
	return t;
}, my = (e, t) => {
	let n = py(e), r = py(t), i = Object.keys(n);
	return i.length === Object.keys(r).length && i.every((e) => Object.hasOwn(r, e) && n[e] === r[e]);
}, hy = (0, k.memo)(({ children: e, mode: t = "streaming", dir: n, parseIncompleteMarkdown: r = !0, normalizeHtmlIndentation: i = !1, components: a, rehypePlugins: o = sy, remarkPlugins: s = cy, className: c, shikiTheme: l, mermaid: u, codeBlockMaxHeight: d = 400, controls: f = !0, isAnimating: p = !1, tableMaxHeight: m = 300, animated: h, BlockComponent: g = fy, parseMarkdownIntoBlocksFn: _ = kv, caret: v, plugins: y, remend: b, linkSafety: x = dy, portal: ee, lineNumbers: S = !0, allowedTags: C, fallbackComponent: w, literalTagContent: T, disableAutolinkProtocols: E, translations: te, icons: D, prefix: ne, onAnimationStart: re, onAnimationEnd: ie, ...ae }) => {
	let oe = (0, k.useId)(), se = py(ae), ce = (0, k.useMemo)(() => Hh(ne), [ne]), le = (0, k.useRef)(null), ue = (0, k.useRef)(re), de = (0, k.useRef)(ie);
	ue.current = re, de.current = ie, (0, k.useEffect)(() => {
		var e, n, r;
		if (t === "static") return;
		let i = le.current;
		le.current = p, i === null ? p && ((e = ue.current) == null || e.call(ue)) : p && !i ? (n = ue.current) == null || n.call(ue) : !p && i && ((r = de.current) == null || r.call(de));
	}, [p, t]);
	let fe = (0, k.useMemo)(() => C ? Object.keys(C) : [], [C]), pe = (0, k.useMemo)(() => {
		if (typeof e != "string") return "";
		let n = t === "streaming" && r ? Lu(e, b) : e;
		return T && T.length > 0 && (n = Rv(n, T)), fe.length > 0 && (n = Fv(n, fe)), n;
	}, [
		e,
		t,
		r,
		b,
		fe,
		T
	]), O = (0, k.useMemo)(() => _(pe), [pe, _]), me = (0, k.useMemo)(() => n === "auto" ? O.map(U_) : void 0, [O, n]), he = (0, k.useMemo)(() => O.map((e, t) => `${oe}-${t}`), [O.length, oe]), ge = (0, k.useMemo)(() => h === !0 ? "true" : h ? JSON.stringify(h) : "", [h]), _e = (0, k.useRef)(null), ve = (0, k.useRef)([]), ye = (0, k.useRef)([]), be = (0, k.useRef)(null), xe = (0, k.useRef)("");
	ge ? (xe.current === ge ? _e.current ||= uh() : (xe.current = ge, _e.current = uh({ maxBacklogMs: ge === "true" ? void 0 : h.maxBacklogMs }), ve.current = [], ye.current = []), p && _e.current && _e.current.beginPass(_e.current.now())) : (_e.current = null, ve.current = [], ye.current = [], xe.current = ""), (0, k.useLayoutEffect)(() => {
		var e;
		p && ((e = _e.current) == null || e.commitPass());
	});
	let Se = (0, k.useMemo)(() => ({
		codeBlockMaxHeight: d,
		shikiTheme: l ?? (y?.code)?.getThemes() ?? uy,
		controls: f,
		isAnimating: p,
		lineNumbers: S,
		mode: t,
		mermaid: u,
		linkSafety: x,
		portal: ee,
		tableMaxHeight: m
	}), [
		d,
		l,
		f,
		p,
		S,
		t,
		u,
		x,
		ee,
		y?.code,
		m
	]), Ce = (0, k.useMemo)(() => te ? JSON.stringify(te) : "", [te]), we = (0, k.useMemo)(() => ({
		...ug,
		...te
	}), [Ce]), Te = (0, k.useMemo)(() => {
		let { inlineCode: e, ...t } = a ?? {}, n = {
			...B_,
			...t
		};
		if (e) {
			let t = n.code;
			n.code = (n) => "data-block" in n ? t ? (0, k.createElement)(t, n) : null : (0, k.createElement)(e, n);
		}
		if (w) {
			if (C) for (let e of Object.keys(C)) Object.hasOwn(n, e) || (n[e] = w);
			let e = {
				configurable: !0,
				enumerable: !1,
				value: w,
				writable: !1
			};
			return new Proxy(n, {
				getOwnPropertyDescriptor(t, n) {
					let r = Object.getOwnPropertyDescriptor(t, n);
					if (r) return r;
					if (typeof n == "string" && ey.test(n)) return e;
				},
				get(e, t, n) {
					return typeof t == "string" && ey.test(t) && !Object.hasOwn(e, t) ? w : Reflect.get(e, t, n);
				}
			});
		}
		return n;
	}, [
		a,
		w,
		C
	]), Ee = (0, k.useMemo)(() => {
		let e = [];
		return y != null && y.cjk && (e = [...e, ...y.cjk.remarkPluginsBefore]), e = [...e, ...s], E && E.length > 0 && (e = [...e, [Yv, E]]), y != null && y.cjk && (e = [...e, ...y.cjk.remarkPluginsAfter]), y != null && y.math && (e = [...e, y.math.remarkPlugin]), e;
	}, [
		s,
		y?.math,
		y?.cjk,
		E
	]), De = (0, k.useMemo)(() => {
		let e = o;
		if (C && Object.keys(C).length > 0 && o === sy) {
			let t = {
				...iy,
				tagNames: [...iy.tagNames ?? [], ...Object.keys(C)],
				attributes: {
					...iy.attributes,
					...C
				}
			};
			e = [
				ay.raw,
				[es, t],
				ay.harden
			];
		}
		return T && T.length > 0 && (e = [...e, [Wv, T]]), y != null && y.math && (e = [...e, y.math.rehypePlugin]), n === "auto" && t === "static" && (e = [...e, Hv]), e;
	}, [
		o,
		y?.math,
		C,
		T,
		n,
		t
	]), Oe = (0, k.useMemo)(() => v && p ? { "--streamdown-caret": `"${ly[v]}"` } : void 0, [v, p]), { containerRef: ke, showCaret: Ae } = $v((0, k.useMemo)(() => {
		if (!p || O.length === 0) return !1;
		let e = O.at(-1);
		return K_(e) || q_(e);
	}, [p, O])), je = (e) => {
		let t = null;
		if (_e.current && p) {
			if (!ve.current[e]) {
				let { maxBacklogMs: t, ...n } = ge && ge !== "true" ? h : {};
				ve.current[e] = Ih({
					...n,
					timeline: _e.current
				});
			}
			t = ve.current[e];
		}
		be.current !== De && (ye.current = [], be.current = De), t && !ye.current[e] && (ye.current[e] = [...De, t.rehypePlugin]);
		let n = t && ye.current[e] ? ye.current[e] : De;
		return {
			blockAnimatePlugin: t,
			blockRehypePlugins: n
		};
	};
	return t === "static" ? (0, K.jsx)(dg.Provider, {
		value: we,
		children: (0, K.jsx)(Eg.Provider, {
			value: y ?? null,
			children: (0, K.jsx)($.Provider, {
				value: Se,
				children: (0, K.jsx)(cg, {
					icons: D,
					children: (0, K.jsx)(Wh.Provider, {
						value: ce,
						children: (0, K.jsx)("div", {
							...se,
							className: ce("space-y-4 whitespace-normal [&>*:first-child]:mt-0 [&>*:last-child]:mb-0", c),
							dir: n === "auto" ? void 0 : n,
							children: (0, K.jsx)($_, {
								components: Te,
								rehypePlugins: De,
								remarkPlugins: Ee,
								...ae,
								children: pe
							})
						})
					})
				})
			})
		})
	}) : (0, K.jsx)(dg.Provider, {
		value: we,
		children: (0, K.jsx)(Eg.Provider, {
			value: y ?? null,
			children: (0, K.jsx)($.Provider, {
				value: Se,
				children: (0, K.jsx)(cg, {
					icons: D,
					children: (0, K.jsx)(Wh.Provider, {
						value: ce,
						children: (0, K.jsxs)("div", {
							...se,
							className: ce("space-y-4 whitespace-normal [&>*:first-child]:mt-0 [&>*:last-child]:mb-0", v && Ae ? "[&>*:last-child:not([data-sd-caret-hidden])]:after:inline [&>*:last-child:not([data-sd-caret-hidden])]:after:align-baseline [&>*:last-child:not([data-sd-caret-hidden])]:after:content-[var(--streamdown-caret)]" : null, c),
							ref: ke,
							style: Oe,
							children: [O.length === 0 && v && p && (0, K.jsx)("span", {}), O.map((e, t) => {
								let a = t === O.length - 1, o = p && a && K_(e), { blockAnimatePlugin: s, blockRehypePlugins: c } = je(t);
								return (0, K.jsx)(g, {
									animatePlugin: s,
									components: Te,
									content: e,
									dir: me?.[t] ?? (n === "auto" ? void 0 : n),
									index: t,
									isIncomplete: o,
									rehypePlugins: c,
									remarkPlugins: Ee,
									shouldNormalizeHtmlIndentation: i,
									shouldParseIncompleteMarkdown: r,
									...ae
								}, he[t]);
							})]
						})
					})
				})
			})
		})
	});
}, (e, t) => e.children === t.children && e.shikiTheme === t.shikiTheme && e.isAnimating === t.isAnimating && e.animated === t.animated && e.mode === t.mode && e.plugins === t.plugins && e.className === t.className && e.linkSafety === t.linkSafety && e.lineNumbers === t.lineNumbers && e.codeBlockMaxHeight === t.codeBlockMaxHeight && e.tableMaxHeight === t.tableMaxHeight && e.normalizeHtmlIndentation === t.normalizeHtmlIndentation && e.literalTagContent === t.literalTagContent && e.disableAutolinkProtocols === t.disableAutolinkProtocols && JSON.stringify(e.translations) === JSON.stringify(t.translations) && e.prefix === t.prefix && e.dir === t.dir && e.fallbackComponent === t.fallbackComponent && my(e, t));
hy.displayName = "Streamdown";
var gy = (e) => e > 0 ? e : null, _y = (e, t) => {
	if (e.endsWith("px")) return gy(Number.parseFloat(e));
	if (e.endsWith("rem")) {
		let t = Number.parseFloat(getComputedStyle(document.documentElement).fontSize || "16");
		return gy(Number.parseFloat(e) * t);
	}
	if (e.endsWith("em")) {
		let n = Number.parseFloat(getComputedStyle(t).fontSize || "16");
		return gy(Number.parseFloat(e) * n);
	}
	if (e.endsWith("vh")) return gy(Number.parseFloat(e) / 100 * window.innerHeight);
	if (e.endsWith("%")) {
		let n = t.parentElement;
		return n && n.clientHeight > 0 ? gy(Number.parseFloat(e) / 100 * n.clientHeight) : null;
	}
	return null;
}, vy = (e, t) => {
	let n = t.trim().toLowerCase();
	return !n || n === "none" || n === "auto" ? null : _y(n, e);
}, yy = /^min\(\s*([^,]+)\s*,\s*([^)]+)\s*\)$/i, by = (e, t) => {
	let n = vy(e, t);
	if (n != null) return n;
	let r = t.trim().match(yy);
	if (!r) return null;
	let i = vy(e, r[1].trim()), a = vy(e, r[2].trim());
	return i == null || a == null ? null : Math.min(i, a);
}, xy = (e, t, n) => {
	let r = t.clientWidth;
	if (!(r > 0)) return null;
	if (n) {
		let n = t.clientHeight;
		if (!(n > 0)) return null;
		let i = Math.min(r / e.width, n / e.height, 1);
		return !(i > 0) || Number.isNaN(i) ? null : {
			fitZoom: i,
			viewportHeight: null
		};
	}
	let i = Math.min(r / e.width, 1);
	if (!(i > 0) || Number.isNaN(i)) return null;
	let a = e.height * i, o = null, s = t;
	for (; s && o == null;) {
		let e = getComputedStyle(s);
		o = by(s, e.maxHeight), s = s.parentElement;
	}
	let c = o == null ? a : Math.min(a, o), l = Math.min(i, c / e.height, 1);
	return !(l > 0) || Number.isNaN(l) ? null : {
		fitZoom: l,
		viewportHeight: c
	};
}, Sy = ({ children: e, className: t, contentSize: n, fitKey: r, minZoom: i = .5, maxZoom: a = 3, zoomStep: o = .1, showControls: s = !0, initialZoom: c = 1, isAutoFit: l = !1, fullscreen: u = !1 }) => {
	let { RotateCcwIcon: d, ZoomInIcon: f, ZoomOutIcon: p } = lg(), m = Z(), h = fg(), g = (0, k.useRef)(null), _ = (0, k.useRef)(null), v = (0, k.useRef)(!1), [y, b] = (0, k.useState)(c), [x, ee] = (0, k.useState)(i), [S, C] = (0, k.useState)(c), [w, T] = (0, k.useState)({
		x: 0,
		y: 0
	}), [E, te] = (0, k.useState)(!1), [D, ne] = (0, k.useState)({
		x: 0,
		y: 0
	}), [re, ie] = (0, k.useState)({
		x: 0,
		y: 0
	}), [ae, oe] = (0, k.useState)(null), se = (0, k.useCallback)((e) => {
		let t = g.current;
		if (!t) return;
		let n = xy(e, t, u);
		if (!n) return;
		let { fitZoom: r, viewportHeight: a } = n;
		oe(a), b(r), ee(Math.min(i, r)), v.current || (C(r), T({
			x: 0,
			y: 0
		}));
	}, [u, i]), ce = (0, k.useCallback)((e) => {
		C((t) => Math.max(x, Math.min(a, t + e))), v.current = !0;
	}, [x, a]), le = (0, k.useCallback)(() => {
		ce(o);
	}, [ce, o]), ue = (0, k.useCallback)(() => {
		ce(-o);
	}, [ce, o]), de = (0, k.useCallback)(() => {
		C(y), T({
			x: 0,
			y: 0
		}), v.current = !1;
	}, [y]), fe = (0, k.useCallback)((e) => {
		e.preventDefault();
		let t = e.deltaY > 0 ? -o : o;
		ce(t);
	}, [ce, o]), pe = (0, k.useCallback)((e) => {
		if (e.button !== 0 || e.isPrimary === !1) return;
		te(!0), v.current = !0, ne({
			x: e.clientX,
			y: e.clientY
		}), ie(w);
		let t = e.currentTarget;
		t instanceof HTMLElement && t.setPointerCapture(e.pointerId);
	}, [w]), O = (0, k.useCallback)((e) => {
		if (!E) return;
		e.preventDefault();
		let t = e.clientX - D.x, n = e.clientY - D.y;
		T({
			x: re.x + t,
			y: re.y + n
		});
	}, [
		E,
		D,
		re
	]), me = (0, k.useCallback)((e) => {
		te(!1);
		let t = e.currentTarget;
		t instanceof HTMLElement && t.releasePointerCapture(e.pointerId);
	}, []);
	(0, k.useEffect)(() => {
		ee(i), l || (b(c), C(c), oe(null));
	}, [
		c,
		l,
		i
	]), (0, k.useLayoutEffect)(() => {
		if (!(l && n)) return;
		se(n);
		let e = g.current;
		if (!e || typeof ResizeObserver > "u") return;
		let t = new ResizeObserver(() => {
			se(n);
		});
		return t.observe(e), () => {
			t.disconnect();
		};
	}, [
		se,
		n,
		l
	]), (0, k.useEffect)(() => {
		l && (v.current = !1);
	}, [r, l]), (0, k.useEffect)(() => {
		let e = g.current;
		if (e) return e.addEventListener("wheel", fe, { passive: !1 }), () => {
			e.removeEventListener("wheel", fe);
		};
	}, [fe]), (0, k.useEffect)(() => {
		let e = _.current;
		if (e && E) return document.body.style.userSelect = "none", e.addEventListener("pointermove", O, { passive: !1 }), e.addEventListener("pointerup", me), e.addEventListener("pointercancel", me), () => {
			document.body.style.userSelect = "", e.removeEventListener("pointermove", O), e.removeEventListener("pointerup", me), e.removeEventListener("pointercancel", me);
		};
	}, [
		E,
		O,
		me
	]);
	let he = !u && l && ae != null ? { height: ae } : void 0;
	return (0, K.jsxs)("div", {
		className: m("relative flex flex-col", u ? "h-full w-full" : "min-h-28 w-full", t),
		ref: g,
		style: {
			cursor: E ? "grabbing" : "grab",
			...he
		},
		children: [s ? (0, K.jsxs)("div", {
			className: m("absolute z-10 flex flex-col gap-1 rounded-md border border-border bg-background/80 p-1 supports-[backdrop-filter]:bg-background/70 supports-[backdrop-filter]:backdrop-blur-sm", u ? "bottom-4 left-4" : "bottom-2 left-2"),
			children: [
				(0, K.jsx)("button", {
					"aria-label": h.zoomIn,
					className: m("flex items-center justify-center rounded p-1.5 text-muted-foreground transition-colors hover:bg-muted hover:text-foreground disabled:cursor-not-allowed disabled:opacity-50"),
					disabled: S >= a,
					onClick: le,
					title: h.zoomIn,
					type: "button",
					children: (0, K.jsx)(f, {
						"aria-hidden": "true",
						size: 16
					})
				}),
				(0, K.jsx)("button", {
					"aria-label": h.zoomOut,
					className: m("flex items-center justify-center rounded p-1.5 text-muted-foreground transition-colors hover:bg-muted hover:text-foreground disabled:cursor-not-allowed disabled:opacity-50"),
					disabled: S <= x,
					onClick: ue,
					title: h.zoomOut,
					type: "button",
					children: (0, K.jsx)(p, {
						"aria-hidden": "true",
						size: 16
					})
				}),
				(0, K.jsx)("button", {
					"aria-label": h.resetView,
					className: m("flex items-center justify-center rounded p-1.5 text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"),
					onClick: de,
					title: h.resetView,
					type: "button",
					children: (0, K.jsx)(d, {
						"aria-hidden": "true",
						size: 16
					})
				})
			]
		}) : null, (0, K.jsx)("div", {
			className: m("flex h-full w-full flex-1 origin-center items-center justify-center transition-transform duration-150 ease-out"),
			onPointerDown: pe,
			ref: _,
			role: "application",
			style: {
				transform: `translate(${w.x}px, ${w.y}px) scale(${S})`,
				transformOrigin: "center center",
				touchAction: "none",
				willChange: "transform"
			},
			children: e
		})]
	});
}, Cy = (e, t) => new Promise((n) => {
	let r = setTimeout(n, e);
	t(() => {
		clearTimeout(r), n();
	});
}), wy = async (e, { chart: t, config: n }, r, i) => {
	try {
		let i = e.getMermaid(n), a = t.split("").reduce((e, t) => (e << 5) - e + t.charCodeAt(0) | 0, 0), o = `mermaid-${Math.abs(a)}-${Date.now()}-${Math.random().toString(36).substring(2, 9)}`, { svg: s } = await i.render(o, t), c = Vg(s);
		return {
			ok: !0,
			svg: r ? s : Hg(s),
			size: c
		};
	} catch (e) {
		return {
			ok: !1,
			error: e instanceof Error ? e.message : i
		};
	}
}, Ty = ({ chart: e, className: t, config: n, fullscreen: r = !1, showControls: i = !0 }) => {
	let a = Z(), [o, s] = (0, k.useState)(null), [c, l] = (0, k.useState)(!1), [u, d] = (0, k.useState)(""), [f, p] = (0, k.useState)(null), m = fg(), [h, g] = (0, k.useState)(""), [_, v] = (0, k.useState)(0), { mermaid: y } = (0, k.useContext)($), b = kg(), x = y?.errorComponent, { shouldRender: ee, containerRef: S } = oh({ immediate: r }), C = (0, k.useRef)(null), w = (0, k.useRef)(!1), T = (0, k.useRef)(!1), E = zh(), te = (0, k.useRef)(E), D = (0, k.useRef)(null);
	(0, k.useEffect)(() => (T.current = !0, () => {
		var e;
		T.current = !1, (e = D.current) == null || e.call(D);
	}), []), (0, k.useEffect)(() => {
		var e;
		te.current = E, E || (e = D.current) == null || e.call(D);
	}, [E]);
	let ne = async (e) => {
		te.current && (await Cy(e, (e) => {
			D.current = e;
		}), D.current = null);
	};
	if ((0, k.useEffect)(() => {
		if (!ee) return;
		if (!b) {
			s("Mermaid plugin not available. Please add the mermaid plugin to enable diagram rendering.");
			return;
		}
		if (C.current = {
			chart: e,
			config: n
		}, w.current) return;
		let t = (e, t) => {
			t.ok ? (d(t.svg), p(t.size), g(t.svg)) : e === C.current && s(t.error);
		};
		(async () => {
			w.current = !0, s(null), l(!0);
			try {
				let e = C.current;
				for (; e;) {
					let n = performance.now(), i = await wy(b, e, r, m.errorMermaid);
					if (!T.current || (t(e, i), i.ok && await ne(performance.now() - n), !T.current || e === C.current)) return;
					e = C.current;
				}
			} finally {
				w.current = !1, T.current && l(!1);
			}
		})();
	}, [
		e,
		n,
		_,
		ee,
		b
	]), !(ee || u || h)) return (0, K.jsx)("div", {
		className: a("my-4 min-h-[200px]", t),
		ref: S
	});
	if (c && !u && !h) return (0, K.jsx)("div", {
		className: a("my-4 flex justify-center p-4", t),
		ref: S,
		children: (0, K.jsxs)("div", {
			className: a("flex items-center space-x-2 text-muted-foreground"),
			children: [(0, K.jsx)("div", { className: a("h-4 w-4 animate-spin rounded-full border-current border-b-2") }), (0, K.jsx)("span", {
				className: a("text-sm"),
				children: "Loading diagram..."
			})]
		})
	});
	if (o && !u && !h) return x ? (0, K.jsx)("div", {
		ref: S,
		children: (0, K.jsx)(x, {
			chart: e,
			error: o,
			retry: () => v((e) => e + 1)
		})
	}) : (0, K.jsxs)("div", {
		className: a("rounded-md bg-red-50 p-4", t),
		ref: S,
		children: [(0, K.jsxs)("p", {
			className: a("font-mono text-red-700 text-sm"),
			children: ["Mermaid Error: ", o]
		}), (0, K.jsxs)("details", {
			className: a("mt-2"),
			children: [(0, K.jsx)("summary", {
				className: a("cursor-pointer text-red-600 text-xs"),
				children: "Show Code"
			}), (0, K.jsx)("pre", {
				className: a("mt-2 overflow-x-auto rounded bg-red-100 p-2 text-red-800 text-xs"),
				children: e
			})]
		})]
	});
	let re = u || h;
	return (0, K.jsx)("div", {
		className: a(r ? "size-full" : "max-h-[min(70vh,40rem)] w-full", t),
		"data-streamdown": "mermaid",
		ref: S,
		children: (0, K.jsx)(Sy, {
			className: a(r ? "size-full overflow-hidden" : "max-h-[min(70vh,40rem)] overflow-hidden", t),
			contentSize: f,
			fitKey: e,
			fullscreen: r,
			isAutoFit: !0,
			maxZoom: 3,
			minZoom: .1,
			showControls: i,
			zoomStep: .1,
			children: (0, K.jsx)("div", {
				"aria-label": m.mermaidAriaLabel,
				className: a("flex justify-center", r ? "size-full items-center" : null),
				dangerouslySetInnerHTML: { __html: re },
				role: "img"
			})
		})
	});
};
export { hy as Streamdown, ay as defaultRehypePlugins, Ty as i, $ as n, Og as r, Zh as t };
