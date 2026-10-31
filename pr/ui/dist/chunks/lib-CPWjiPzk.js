import { n as e } from "./rolldown-runtime-QGRDbDUS.js";
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/unist-util-is/6.0.1/d52968e5f99f6b91f4bf587c0182bad2af768f306043a546097f2dce25379f7b/node_modules/unist-util-is/lib/index.js
var t = (function(e) {
	if (e == null) return o;
	if (typeof e == "function") return a(e);
	if (typeof e == "object") return Array.isArray(e) ? n(e) : r(e);
	if (typeof e == "string") return i(e);
	throw Error("Expected function, string, or object as test");
});
function n(e) {
	let n = [], r = -1;
	for (; ++r < e.length;) n[r] = t(e[r]);
	return a(i);
	function i(...e) {
		let t = -1;
		for (; ++t < n.length;) if (n[t].apply(this, e)) return !0;
		return !1;
	}
}
function r(e) {
	let t = e;
	return a(n);
	function n(n) {
		let r = n, i;
		for (i in e) if (r[i] !== t[i]) return !1;
		return !0;
	}
}
function i(e) {
	return a(t);
	function t(t) {
		return t && t.type === e;
	}
}
function a(e) {
	return t;
	function t(t, n, r) {
		return !!(s(t) && e.call(this, t, typeof n == "number" ? n : void 0, r || void 0));
	}
}
function o() {
	return !0;
}
function s(e) {
	return typeof e == "object" && !!e && "type" in e;
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/unist-util-visit-parents/6.0.2/a55922569a25dabfa9558ac0d2c6c019e17bc7fcf5313b2e0d42dd4a96be77c0/node_modules/unist-util-visit-parents/lib/color.js
function c(e) {
	return e;
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/unist-util-visit-parents/6.0.2/a55922569a25dabfa9558ac0d2c6c019e17bc7fcf5313b2e0d42dd4a96be77c0/node_modules/unist-util-visit-parents/lib/index.js
var l = [], u = "skip";
function d(e, n, r, i) {
	let a;
	typeof n == "function" && typeof r != "function" ? (i = r, r = n) : a = n;
	let o = t(a), s = i ? -1 : 1;
	u(e, void 0, [])();
	function u(e, t, a) {
		let d = e && typeof e == "object" ? e : {};
		if (typeof d.type == "string") {
			let t = typeof d.tagName == "string" ? d.tagName : typeof d.name == "string" ? d.name : void 0;
			Object.defineProperty(p, "name", { value: "node (" + c(e.type + (t ? "<" + t + ">" : "")) + ")" });
		}
		return p;
		function p() {
			let c = l, d, p, m;
			if ((!n || o(e, t, a[a.length - 1] || void 0)) && (c = f(r(e, a)), c[0] === !1)) return c;
			if ("children" in e && e.children) {
				let t = e;
				if (t.children && c[0] !== "skip") for (p = (i ? t.children.length : -1) + s, m = a.concat(t); p > -1 && p < t.children.length;) {
					let e = t.children[p];
					if (d = u(e, p, m)(), d[0] === !1) return d;
					p = typeof d[1] == "number" ? d[1] : p + s;
				}
			}
			return c;
		}
	}
}
function f(e) {
	return Array.isArray(e) ? e : typeof e == "number" ? [!0, e] : e == null ? l : [e];
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/unist-util-visit/5.1.0/33b95b306dd1d9e0108177f06c061fed46d3a6f21dd24b5a0171f12190d3b07a/node_modules/unist-util-visit/lib/index.js
function p(e, t, n, r) {
	let i, a, o;
	typeof t == "function" && typeof n != "function" ? (a = void 0, o = t, i = n) : (a = t, o = n, i = r), d(e, a, s, i);
	function s(e, t) {
		let n = t[t.length - 1], r = n ? n.children.indexOf(e) : void 0;
		return o(e, r, n);
	}
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/devlop/1.1.0/10663eb7e9f3c716a550e236adb7386132eab4b1ea4d001e6ae3b0eae82b8f01/node_modules/devlop/lib/default.js
function m() {}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/micromark-util-character/2.1.1/1fbc36800dd09487322bddfb36d588a1aaa2b8f466cf8a4dd6d460933154d03e/node_modules/micromark-util-character/index.js
var h = D(/[A-Za-z]/), g = D(/[\dA-Za-z]/), _ = D(/[#-'*+\--9=?A-Z^-~]/);
function v(e) {
	return e !== null && (e < 32 || e === 127);
}
var y = D(/\d/), b = D(/[\dA-Fa-f]/), x = D(/[!-/:-@[-`{-~]/);
function S(e) {
	return e !== null && e < -2;
}
function C(e) {
	return e !== null && (e < 0 || e === 32);
}
function w(e) {
	return e === -2 || e === -1 || e === 32;
}
var T = D(/\p{P}|\p{S}/u), E = D(/\s/);
function D(e) {
	return t;
	function t(t) {
		return t !== null && t > -1 && e.test(String.fromCharCode(t));
	}
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/escape-string-regexp/5.0.0/975308e2f12ad393ccebb3882840c3c1e975368995bfe20e83579ff1c006e7af/node_modules/escape-string-regexp/index.js
function O(e) {
	if (typeof e != "string") throw TypeError("Expected a string");
	return e.replace(/[|\\{}()[\]^$+*?.]/g, "\\$&").replace(/-/g, "\\x2d");
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/micromark-util-normalize-identifier/2.0.1/f4d483adf339598516c01a12f23dec0189b6c043ba07fe1e7270107c06818288/node_modules/micromark-util-normalize-identifier/index.js
function k(e) {
	return e.replace(/[\t\n\r ]+/g, " ").replace(/^ | $/g, "").toLowerCase().toUpperCase();
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/mdast-util-gfm-strikethrough/2.0.0/7db69bb882519e50dd56becff8ccad60fdedffd5a86cadc5bd2fe4dc3f635445/node_modules/mdast-util-gfm-strikethrough/lib/index.js
var A = [
	"autolink",
	"destinationLiteral",
	"destinationRaw",
	"reference",
	"titleQuote",
	"titleApostrophe"
];
F.peek = ee;
function j() {
	return {
		canContainEols: ["delete"],
		enter: { strikethrough: N },
		exit: { strikethrough: P }
	};
}
function M() {
	return {
		unsafe: [{
			character: "~",
			inConstruct: "phrasing",
			notInConstruct: A
		}],
		handlers: { delete: F }
	};
}
function N(e) {
	this.enter({
		type: "delete",
		children: []
	}, e);
}
function P(e) {
	this.exit(e);
}
function F(e, t, n, r) {
	let i = n.createTracker(r), a = n.enter("strikethrough"), o = i.move("~~");
	return o += n.containerPhrasing(e, {
		...i.current(),
		before: o,
		after: "~"
	}), o += i.move("~~"), a(), o;
}
function ee() {
	return "~";
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/markdown-table/3.0.4/9f40d56831a72ec47bc8a8523f02dc9508fdeab2ad0a22ac84ebc060924982b0/node_modules/markdown-table/index.js
function I(e) {
	return e.length;
}
function L(e, t) {
	let n = t || {}, r = (n.align || []).concat(), i = n.stringLength || I, a = [], o = [], s = [], c = [], l = 0, u = -1;
	for (; ++u < e.length;) {
		let t = [], r = [], a = -1;
		for (e[u].length > l && (l = e[u].length); ++a < e[u].length;) {
			let o = te(e[u][a]);
			if (n.alignDelimiters !== !1) {
				let e = i(o);
				r[a] = e, (c[a] === void 0 || e > c[a]) && (c[a] = e);
			}
			t.push(o);
		}
		o[u] = t, s[u] = r;
	}
	let d = -1;
	if (typeof r == "object" && "length" in r) for (; ++d < l;) a[d] = R(r[d]);
	else {
		let e = R(r);
		for (; ++d < l;) a[d] = e;
	}
	d = -1;
	let f = [], p = [];
	for (; ++d < l;) {
		let e = a[d], t = "", r = "";
		e === 99 ? (t = ":", r = ":") : e === 108 ? t = ":" : e === 114 && (r = ":");
		let i = n.alignDelimiters === !1 ? 1 : Math.max(1, c[d] - t.length - r.length), o = t + "-".repeat(i) + r;
		n.alignDelimiters !== !1 && (i = t.length + i + r.length, i > c[d] && (c[d] = i), p[d] = i), f[d] = o;
	}
	o.splice(1, 0, f), s.splice(1, 0, p), u = -1;
	let m = [];
	for (; ++u < o.length;) {
		let e = o[u], t = s[u];
		d = -1;
		let r = [];
		for (; ++d < l;) {
			let i = e[d] || "", o = "", s = "";
			if (n.alignDelimiters !== !1) {
				let e = c[d] - (t[d] || 0), n = a[d];
				n === 114 ? o = " ".repeat(e) : n === 99 ? e % 2 ? (o = " ".repeat(e / 2 + .5), s = " ".repeat(e / 2 - .5)) : (o = " ".repeat(e / 2), s = o) : s = " ".repeat(e);
			}
			n.delimiterStart !== !1 && !d && r.push("|"), n.padding === !1 || n.alignDelimiters === !1 && i === "" || n.delimiterStart === !1 && !d || r.push(" "), n.alignDelimiters !== !1 && r.push(o), r.push(i), n.alignDelimiters !== !1 && r.push(s), n.padding !== !1 && r.push(" "), (n.delimiterEnd !== !1 || d !== l - 1) && r.push("|");
		}
		m.push(n.delimiterEnd === !1 ? r.join("").replace(/ +$/, "") : r.join(""));
	}
	return m.join("\n");
}
function te(e) {
	return e == null ? "" : String(e);
}
function R(e) {
	let t = typeof e == "string" ? e.codePointAt(0) : 0;
	return t === 67 || t === 99 ? 99 : t === 76 || t === 108 ? 108 : t === 82 || t === 114 ? 114 : 0;
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/mdast-util-to-markdown/2.1.2/bf6627bebb2b304f751c9c5e7d9b2000845c3a65bcdc58c5070d516b6bbcc50c/node_modules/mdast-util-to-markdown/lib/handle/blockquote.js
function z(e, t, n, r) {
	let i = n.enter("blockquote"), a = n.createTracker(r);
	a.move("> "), a.shift(2);
	let o = n.indentLines(n.containerFlow(e, a.current()), B);
	return i(), o;
}
function B(e, t, n) {
	return ">" + (n ? "" : " ") + e;
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/mdast-util-to-markdown/2.1.2/bf6627bebb2b304f751c9c5e7d9b2000845c3a65bcdc58c5070d516b6bbcc50c/node_modules/mdast-util-to-markdown/lib/util/pattern-in-scope.js
function V(e, t) {
	return ne(e, t.inConstruct, !0) && !ne(e, t.notInConstruct, !1);
}
function ne(e, t, n) {
	if (typeof t == "string" && (t = [t]), !t || t.length === 0) return n;
	let r = -1;
	for (; ++r < t.length;) if (e.includes(t[r])) return !0;
	return !1;
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/mdast-util-to-markdown/2.1.2/bf6627bebb2b304f751c9c5e7d9b2000845c3a65bcdc58c5070d516b6bbcc50c/node_modules/mdast-util-to-markdown/lib/handle/break.js
function re(e, t, n, r) {
	let i = -1;
	for (; ++i < n.unsafe.length;) if (n.unsafe[i].character === "\n" && V(n.stack, n.unsafe[i])) return /[ \t]/.test(r.before) ? "" : " ";
	return "\\\n";
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/longest-streak/3.1.0/88910b4232726d9d30993f803635471bbdd113f509560f4153c3fd3833766652/node_modules/longest-streak/index.js
function ie(e, t) {
	let n = String(e), r = n.indexOf(t), i = r, a = 0, o = 0;
	if (typeof t != "string") throw TypeError("Expected substring");
	for (; r !== -1;) r === i ? ++a > o && (o = a) : a = 1, i = r + t.length, r = n.indexOf(t, i);
	return o;
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/mdast-util-to-markdown/2.1.2/bf6627bebb2b304f751c9c5e7d9b2000845c3a65bcdc58c5070d516b6bbcc50c/node_modules/mdast-util-to-markdown/lib/util/format-code-as-indented.js
function ae(e, t) {
	return !(t.options.fences !== !1 || !e.value || e.lang || !/[^ \r\n]/.test(e.value) || /^[\t ]*(?:[\r\n]|$)|(?:^|[\r\n])[\t ]*$/.test(e.value));
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/mdast-util-to-markdown/2.1.2/bf6627bebb2b304f751c9c5e7d9b2000845c3a65bcdc58c5070d516b6bbcc50c/node_modules/mdast-util-to-markdown/lib/util/check-fence.js
function oe(e) {
	let t = e.options.fence || "`";
	if (t !== "`" && t !== "~") throw Error("Cannot serialize code with `" + t + "` for `options.fence`, expected `` ` `` or `~`");
	return t;
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/mdast-util-to-markdown/2.1.2/bf6627bebb2b304f751c9c5e7d9b2000845c3a65bcdc58c5070d516b6bbcc50c/node_modules/mdast-util-to-markdown/lib/handle/code.js
function se(e, t, n, r) {
	let i = oe(n), a = e.value || "", o = i === "`" ? "GraveAccent" : "Tilde";
	if (ae(e, n)) {
		let e = n.enter("codeIndented"), t = n.indentLines(a, ce);
		return e(), t;
	}
	let s = n.createTracker(r), c = i.repeat(Math.max(ie(a, i) + 1, 3)), l = n.enter("codeFenced"), u = s.move(c);
	if (e.lang) {
		let t = n.enter(`codeFencedLang${o}`);
		u += s.move(n.safe(e.lang, {
			before: u,
			after: " ",
			encode: ["`"],
			...s.current()
		})), t();
	}
	if (e.lang && e.meta) {
		let t = n.enter(`codeFencedMeta${o}`);
		u += s.move(" "), u += s.move(n.safe(e.meta, {
			before: u,
			after: "\n",
			encode: ["`"],
			...s.current()
		})), t();
	}
	return u += s.move("\n"), a && (u += s.move(a + "\n")), u += s.move(c), l(), u;
}
function ce(e, t, n) {
	return (n ? "" : "    ") + e;
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/mdast-util-to-markdown/2.1.2/bf6627bebb2b304f751c9c5e7d9b2000845c3a65bcdc58c5070d516b6bbcc50c/node_modules/mdast-util-to-markdown/lib/util/check-quote.js
function le(e) {
	let t = e.options.quote || "\"";
	if (t !== "\"" && t !== "'") throw Error("Cannot serialize title with `" + t + "` for `options.quote`, expected `\"`, or `'`");
	return t;
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/mdast-util-to-markdown/2.1.2/bf6627bebb2b304f751c9c5e7d9b2000845c3a65bcdc58c5070d516b6bbcc50c/node_modules/mdast-util-to-markdown/lib/handle/definition.js
function ue(e, t, n, r) {
	let i = le(n), a = i === "\"" ? "Quote" : "Apostrophe", o = n.enter("definition"), s = n.enter("label"), c = n.createTracker(r), l = c.move("[");
	return l += c.move(n.safe(n.associationId(e), {
		before: l,
		after: "]",
		...c.current()
	})), l += c.move("]: "), s(), !e.url || /[\0- \u007F]/.test(e.url) ? (s = n.enter("destinationLiteral"), l += c.move("<"), l += c.move(n.safe(e.url, {
		before: l,
		after: ">",
		...c.current()
	})), l += c.move(">")) : (s = n.enter("destinationRaw"), l += c.move(n.safe(e.url, {
		before: l,
		after: e.title ? " " : "\n",
		...c.current()
	}))), s(), e.title && (s = n.enter(`title${a}`), l += c.move(" " + i), l += c.move(n.safe(e.title, {
		before: l,
		after: i,
		...c.current()
	})), l += c.move(i), s()), o(), l;
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/mdast-util-to-markdown/2.1.2/bf6627bebb2b304f751c9c5e7d9b2000845c3a65bcdc58c5070d516b6bbcc50c/node_modules/mdast-util-to-markdown/lib/util/check-emphasis.js
function de(e) {
	let t = e.options.emphasis || "*";
	if (t !== "*" && t !== "_") throw Error("Cannot serialize emphasis with `" + t + "` for `options.emphasis`, expected `*`, or `_`");
	return t;
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/mdast-util-to-markdown/2.1.2/bf6627bebb2b304f751c9c5e7d9b2000845c3a65bcdc58c5070d516b6bbcc50c/node_modules/mdast-util-to-markdown/lib/util/encode-character-reference.js
function H(e) {
	return "&#x" + e.toString(16).toUpperCase() + ";";
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/micromark-util-classify-character/2.0.1/381abde4ec23ae89adecf366e37f7e24f39dddd7afcd4d7ae1dcb0b1d68395f7/node_modules/micromark-util-classify-character/index.js
function U(e) {
	if (e === null || C(e) || E(e)) return 1;
	if (T(e)) return 2;
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/mdast-util-to-markdown/2.1.2/bf6627bebb2b304f751c9c5e7d9b2000845c3a65bcdc58c5070d516b6bbcc50c/node_modules/mdast-util-to-markdown/lib/util/encode-info.js
function W(e, t, n) {
	let r = U(e), i = U(t);
	return r === void 0 ? i === void 0 ? n === "_" ? {
		inside: !0,
		outside: !0
	} : {
		inside: !1,
		outside: !1
	} : i === 1 ? {
		inside: !0,
		outside: !0
	} : {
		inside: !1,
		outside: !0
	} : r === 1 ? i === void 0 ? {
		inside: !1,
		outside: !1
	} : i === 1 ? {
		inside: !0,
		outside: !0
	} : {
		inside: !1,
		outside: !1
	} : i === void 0 ? {
		inside: !1,
		outside: !1
	} : i === 1 ? {
		inside: !0,
		outside: !1
	} : {
		inside: !1,
		outside: !1
	};
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/mdast-util-to-markdown/2.1.2/bf6627bebb2b304f751c9c5e7d9b2000845c3a65bcdc58c5070d516b6bbcc50c/node_modules/mdast-util-to-markdown/lib/handle/emphasis.js
fe.peek = pe;
function fe(e, t, n, r) {
	let i = de(n), a = n.enter("emphasis"), o = n.createTracker(r), s = o.move(i), c = o.move(n.containerPhrasing(e, {
		after: i,
		before: s,
		...o.current()
	})), l = c.charCodeAt(0), u = W(r.before.charCodeAt(r.before.length - 1), l, i);
	u.inside && (c = H(l) + c.slice(1));
	let d = c.charCodeAt(c.length - 1), f = W(r.after.charCodeAt(0), d, i);
	f.inside && (c = c.slice(0, -1) + H(d));
	let p = o.move(i);
	return a(), n.attentionEncodeSurroundingInfo = {
		after: f.outside,
		before: u.outside
	}, s + c + p;
}
function pe(e, t, n) {
	return n.options.emphasis || "*";
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/mdast-util-to-string/4.0.0/3001db021a790b37432b62272503499c0eaa2fb2c605e9dc3d1a5e4620151f66/node_modules/mdast-util-to-string/lib/index.js
var me = {};
function he(e, t) {
	let n = t || me;
	return ge(e, typeof n.includeImageAlt != "boolean" || n.includeImageAlt, typeof n.includeHtml != "boolean" || n.includeHtml);
}
function ge(e, t, n) {
	if (ve(e)) {
		if ("value" in e) return e.type === "html" && !n ? "" : e.value;
		if (t && "alt" in e && e.alt) return e.alt;
		if ("children" in e) return _e(e.children, t, n);
	}
	return Array.isArray(e) ? _e(e, t, n) : "";
}
function _e(e, t, n) {
	let r = [], i = -1;
	for (; ++i < e.length;) r[i] = ge(e[i], t, n);
	return r.join("");
}
function ve(e) {
	return !!(e && typeof e == "object");
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/mdast-util-to-markdown/2.1.2/bf6627bebb2b304f751c9c5e7d9b2000845c3a65bcdc58c5070d516b6bbcc50c/node_modules/mdast-util-to-markdown/lib/util/format-heading-as-setext.js
function ye(e, t) {
	let n = !1;
	return p(e, function(e) {
		if ("value" in e && /\r?\n|\r/.test(e.value) || e.type === "break") return n = !0, !1;
	}), !!((!e.depth || e.depth < 3) && he(e) && (t.options.setext || n));
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/mdast-util-to-markdown/2.1.2/bf6627bebb2b304f751c9c5e7d9b2000845c3a65bcdc58c5070d516b6bbcc50c/node_modules/mdast-util-to-markdown/lib/handle/heading.js
function be(e, t, n, r) {
	let i = Math.max(Math.min(6, e.depth || 1), 1), a = n.createTracker(r);
	if (ye(e, n)) {
		let t = n.enter("headingSetext"), r = n.enter("phrasing"), o = n.containerPhrasing(e, {
			...a.current(),
			before: "\n",
			after: "\n"
		});
		return r(), t(), o + "\n" + (i === 1 ? "=" : "-").repeat(o.length - (Math.max(o.lastIndexOf("\r"), o.lastIndexOf("\n")) + 1));
	}
	let o = "#".repeat(i), s = n.enter("headingAtx"), c = n.enter("phrasing");
	a.move(o + " ");
	let l = n.containerPhrasing(e, {
		before: "# ",
		after: "\n",
		...a.current()
	});
	return /^[\t ]/.test(l) && (l = H(l.charCodeAt(0)) + l.slice(1)), l = l ? o + " " + l : o, n.options.closeAtx && (l += " " + o), c(), s(), l;
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/mdast-util-to-markdown/2.1.2/bf6627bebb2b304f751c9c5e7d9b2000845c3a65bcdc58c5070d516b6bbcc50c/node_modules/mdast-util-to-markdown/lib/handle/html.js
xe.peek = Se;
function xe(e) {
	return e.value || "";
}
function Se() {
	return "<";
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/mdast-util-to-markdown/2.1.2/bf6627bebb2b304f751c9c5e7d9b2000845c3a65bcdc58c5070d516b6bbcc50c/node_modules/mdast-util-to-markdown/lib/handle/image.js
Ce.peek = we;
function Ce(e, t, n, r) {
	let i = le(n), a = i === "\"" ? "Quote" : "Apostrophe", o = n.enter("image"), s = n.enter("label"), c = n.createTracker(r), l = c.move("![");
	return l += c.move(n.safe(e.alt, {
		before: l,
		after: "]",
		...c.current()
	})), l += c.move("]("), s(), !e.url && e.title || /[\0- \u007F]/.test(e.url) ? (s = n.enter("destinationLiteral"), l += c.move("<"), l += c.move(n.safe(e.url, {
		before: l,
		after: ">",
		...c.current()
	})), l += c.move(">")) : (s = n.enter("destinationRaw"), l += c.move(n.safe(e.url, {
		before: l,
		after: e.title ? " " : ")",
		...c.current()
	}))), s(), e.title && (s = n.enter(`title${a}`), l += c.move(" " + i), l += c.move(n.safe(e.title, {
		before: l,
		after: i,
		...c.current()
	})), l += c.move(i), s()), l += c.move(")"), o(), l;
}
function we() {
	return "!";
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/mdast-util-to-markdown/2.1.2/bf6627bebb2b304f751c9c5e7d9b2000845c3a65bcdc58c5070d516b6bbcc50c/node_modules/mdast-util-to-markdown/lib/handle/image-reference.js
Te.peek = Ee;
function Te(e, t, n, r) {
	let i = e.referenceType, a = n.enter("imageReference"), o = n.enter("label"), s = n.createTracker(r), c = s.move("!["), l = n.safe(e.alt, {
		before: c,
		after: "]",
		...s.current()
	});
	c += s.move(l + "]["), o();
	let u = n.stack;
	n.stack = [], o = n.enter("reference");
	let d = n.safe(n.associationId(e), {
		before: c,
		after: "]",
		...s.current()
	});
	return o(), n.stack = u, a(), i === "full" || !l || l !== d ? c += s.move(d + "]") : i === "shortcut" ? c = c.slice(0, -1) : c += s.move("]"), c;
}
function Ee() {
	return "!";
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/mdast-util-to-markdown/2.1.2/bf6627bebb2b304f751c9c5e7d9b2000845c3a65bcdc58c5070d516b6bbcc50c/node_modules/mdast-util-to-markdown/lib/handle/inline-code.js
De.peek = Oe;
function De(e, t, n) {
	let r = e.value || "", i = "`", a = -1;
	for (; RegExp("(^|[^`])" + i + "([^`]|$)").test(r);) i += "`";
	for (/[^ \r\n]/.test(r) && (/^[ \r\n]/.test(r) && /[ \r\n]$/.test(r) || /^`|`$/.test(r)) && (r = " " + r + " "); ++a < n.unsafe.length;) {
		let e = n.unsafe[a], t = n.compilePattern(e), i;
		if (e.atBreak) for (; i = t.exec(r);) {
			let e = i.index;
			r.charCodeAt(e) === 10 && r.charCodeAt(e - 1) === 13 && e--, r = r.slice(0, e) + " " + r.slice(i.index + 1);
		}
	}
	return i + r + i;
}
function Oe() {
	return "`";
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/mdast-util-to-markdown/2.1.2/bf6627bebb2b304f751c9c5e7d9b2000845c3a65bcdc58c5070d516b6bbcc50c/node_modules/mdast-util-to-markdown/lib/util/format-link-as-autolink.js
function ke(e, t) {
	let n = he(e);
	return !(t.options.resourceLink || !e.url || e.title || !e.children || e.children.length !== 1 || e.children[0].type !== "text" || n !== e.url && "mailto:" + n !== e.url || !/^[a-z][a-z+.-]+:/i.test(e.url) || /[\0- <>\u007F]/.test(e.url));
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/mdast-util-to-markdown/2.1.2/bf6627bebb2b304f751c9c5e7d9b2000845c3a65bcdc58c5070d516b6bbcc50c/node_modules/mdast-util-to-markdown/lib/handle/link.js
Ae.peek = je;
function Ae(e, t, n, r) {
	let i = le(n), a = i === "\"" ? "Quote" : "Apostrophe", o = n.createTracker(r), s, c;
	if (ke(e, n)) {
		let t = n.stack;
		n.stack = [], s = n.enter("autolink");
		let r = o.move("<");
		return r += o.move(n.containerPhrasing(e, {
			before: r,
			after: ">",
			...o.current()
		})), r += o.move(">"), s(), n.stack = t, r;
	}
	s = n.enter("link"), c = n.enter("label");
	let l = o.move("[");
	return l += o.move(n.containerPhrasing(e, {
		before: l,
		after: "](",
		...o.current()
	})), l += o.move("]("), c(), !e.url && e.title || /[\0- \u007F]/.test(e.url) ? (c = n.enter("destinationLiteral"), l += o.move("<"), l += o.move(n.safe(e.url, {
		before: l,
		after: ">",
		...o.current()
	})), l += o.move(">")) : (c = n.enter("destinationRaw"), l += o.move(n.safe(e.url, {
		before: l,
		after: e.title ? " " : ")",
		...o.current()
	}))), c(), e.title && (c = n.enter(`title${a}`), l += o.move(" " + i), l += o.move(n.safe(e.title, {
		before: l,
		after: i,
		...o.current()
	})), l += o.move(i), c()), l += o.move(")"), s(), l;
}
function je(e, t, n) {
	return ke(e, n) ? "<" : "[";
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/mdast-util-to-markdown/2.1.2/bf6627bebb2b304f751c9c5e7d9b2000845c3a65bcdc58c5070d516b6bbcc50c/node_modules/mdast-util-to-markdown/lib/handle/link-reference.js
Me.peek = Ne;
function Me(e, t, n, r) {
	let i = e.referenceType, a = n.enter("linkReference"), o = n.enter("label"), s = n.createTracker(r), c = s.move("["), l = n.containerPhrasing(e, {
		before: c,
		after: "]",
		...s.current()
	});
	c += s.move(l + "]["), o();
	let u = n.stack;
	n.stack = [], o = n.enter("reference");
	let d = n.safe(n.associationId(e), {
		before: c,
		after: "]",
		...s.current()
	});
	return o(), n.stack = u, a(), i === "full" || !l || l !== d ? c += s.move(d + "]") : i === "shortcut" ? c = c.slice(0, -1) : c += s.move("]"), c;
}
function Ne() {
	return "[";
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/mdast-util-to-markdown/2.1.2/bf6627bebb2b304f751c9c5e7d9b2000845c3a65bcdc58c5070d516b6bbcc50c/node_modules/mdast-util-to-markdown/lib/util/check-bullet.js
function Pe(e) {
	let t = e.options.bullet || "*";
	if (t !== "*" && t !== "+" && t !== "-") throw Error("Cannot serialize items with `" + t + "` for `options.bullet`, expected `*`, `+`, or `-`");
	return t;
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/mdast-util-to-markdown/2.1.2/bf6627bebb2b304f751c9c5e7d9b2000845c3a65bcdc58c5070d516b6bbcc50c/node_modules/mdast-util-to-markdown/lib/util/check-bullet-other.js
function Fe(e) {
	let t = Pe(e), n = e.options.bulletOther;
	if (!n) return t === "*" ? "-" : "*";
	if (n !== "*" && n !== "+" && n !== "-") throw Error("Cannot serialize items with `" + n + "` for `options.bulletOther`, expected `*`, `+`, or `-`");
	if (n === t) throw Error("Expected `bullet` (`" + t + "`) and `bulletOther` (`" + n + "`) to be different");
	return n;
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/mdast-util-to-markdown/2.1.2/bf6627bebb2b304f751c9c5e7d9b2000845c3a65bcdc58c5070d516b6bbcc50c/node_modules/mdast-util-to-markdown/lib/util/check-bullet-ordered.js
function Ie(e) {
	let t = e.options.bulletOrdered || ".";
	if (t !== "." && t !== ")") throw Error("Cannot serialize items with `" + t + "` for `options.bulletOrdered`, expected `.` or `)`");
	return t;
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/mdast-util-to-markdown/2.1.2/bf6627bebb2b304f751c9c5e7d9b2000845c3a65bcdc58c5070d516b6bbcc50c/node_modules/mdast-util-to-markdown/lib/util/check-rule.js
function Le(e) {
	let t = e.options.rule || "*";
	if (t !== "*" && t !== "-" && t !== "_") throw Error("Cannot serialize rules with `" + t + "` for `options.rule`, expected `*`, `-`, or `_`");
	return t;
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/mdast-util-to-markdown/2.1.2/bf6627bebb2b304f751c9c5e7d9b2000845c3a65bcdc58c5070d516b6bbcc50c/node_modules/mdast-util-to-markdown/lib/handle/list.js
function Re(e, t, n, r) {
	let i = n.enter("list"), a = n.bulletCurrent, o = e.ordered ? Ie(n) : Pe(n), s = e.ordered ? o === "." ? ")" : "." : Fe(n), c = t && n.bulletLastUsed ? o === n.bulletLastUsed : !1;
	if (!e.ordered) {
		let t = e.children ? e.children[0] : void 0;
		if (o !== "*" && o !== "-" || !t || t.children && t.children[0] || n.stack[n.stack.length - 1] !== "list" || n.stack[n.stack.length - 2] !== "listItem" || n.stack[n.stack.length - 3] !== "list" || n.stack[n.stack.length - 4] !== "listItem" || n.indexStack[n.indexStack.length - 1] !== 0 || n.indexStack[n.indexStack.length - 2] !== 0 || n.indexStack[n.indexStack.length - 3] !== 0 || (c = !0), Le(n) === o && t) {
			let t = -1;
			for (; ++t < e.children.length;) {
				let n = e.children[t];
				if (n && n.type === "listItem" && n.children && n.children[0] && n.children[0].type === "thematicBreak") {
					c = !0;
					break;
				}
			}
		}
	}
	c && (o = s), n.bulletCurrent = o;
	let l = n.containerFlow(e, r);
	return n.bulletLastUsed = o, n.bulletCurrent = a, i(), l;
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/mdast-util-to-markdown/2.1.2/bf6627bebb2b304f751c9c5e7d9b2000845c3a65bcdc58c5070d516b6bbcc50c/node_modules/mdast-util-to-markdown/lib/util/check-list-item-indent.js
function ze(e) {
	let t = e.options.listItemIndent || "one";
	if (t !== "tab" && t !== "one" && t !== "mixed") throw Error("Cannot serialize items with `" + t + "` for `options.listItemIndent`, expected `tab`, `one`, or `mixed`");
	return t;
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/mdast-util-to-markdown/2.1.2/bf6627bebb2b304f751c9c5e7d9b2000845c3a65bcdc58c5070d516b6bbcc50c/node_modules/mdast-util-to-markdown/lib/handle/list-item.js
function Be(e, t, n, r) {
	let i = ze(n), a = n.bulletCurrent || Pe(n);
	t && t.type === "list" && t.ordered && (a = (typeof t.start == "number" && t.start > -1 ? t.start : 1) + (n.options.incrementListMarker === !1 ? 0 : t.children.indexOf(e)) + a);
	let o = a.length + 1;
	(i === "tab" || i === "mixed" && (t && t.type === "list" && t.spread || e.spread)) && (o = Math.ceil(o / 4) * 4);
	let s = n.createTracker(r);
	s.move(a + " ".repeat(o - a.length)), s.shift(o);
	let c = n.enter("listItem"), l = n.indentLines(n.containerFlow(e, s.current()), u);
	return c(), l;
	function u(e, t, n) {
		return t ? (n ? "" : " ".repeat(o)) + e : (n ? a : a + " ".repeat(o - a.length)) + e;
	}
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/mdast-util-to-markdown/2.1.2/bf6627bebb2b304f751c9c5e7d9b2000845c3a65bcdc58c5070d516b6bbcc50c/node_modules/mdast-util-to-markdown/lib/handle/paragraph.js
function Ve(e, t, n, r) {
	let i = n.enter("paragraph"), a = n.enter("phrasing"), o = n.containerPhrasing(e, r);
	return a(), i(), o;
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/mdast-util-phrasing/4.1.0/7838c5490456302bb7bb8fbe67836e69f2a3dd7b2d881ebdda14a037651398ff/node_modules/mdast-util-phrasing/lib/index.js
var He = t([
	"break",
	"delete",
	"emphasis",
	"footnote",
	"footnoteReference",
	"image",
	"imageReference",
	"inlineCode",
	"inlineMath",
	"link",
	"linkReference",
	"mdxJsxTextElement",
	"mdxTextExpression",
	"strong",
	"text",
	"textDirective"
]);
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/mdast-util-to-markdown/2.1.2/bf6627bebb2b304f751c9c5e7d9b2000845c3a65bcdc58c5070d516b6bbcc50c/node_modules/mdast-util-to-markdown/lib/handle/root.js
function Ue(e, t, n, r) {
	return (e.children.some(function(e) {
		return He(e);
	}) ? n.containerPhrasing : n.containerFlow).call(n, e, r);
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/mdast-util-to-markdown/2.1.2/bf6627bebb2b304f751c9c5e7d9b2000845c3a65bcdc58c5070d516b6bbcc50c/node_modules/mdast-util-to-markdown/lib/util/check-strong.js
function We(e) {
	let t = e.options.strong || "*";
	if (t !== "*" && t !== "_") throw Error("Cannot serialize strong with `" + t + "` for `options.strong`, expected `*`, or `_`");
	return t;
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/mdast-util-to-markdown/2.1.2/bf6627bebb2b304f751c9c5e7d9b2000845c3a65bcdc58c5070d516b6bbcc50c/node_modules/mdast-util-to-markdown/lib/handle/strong.js
Ge.peek = Ke;
function Ge(e, t, n, r) {
	let i = We(n), a = n.enter("strong"), o = n.createTracker(r), s = o.move(i + i), c = o.move(n.containerPhrasing(e, {
		after: i,
		before: s,
		...o.current()
	})), l = c.charCodeAt(0), u = W(r.before.charCodeAt(r.before.length - 1), l, i);
	u.inside && (c = H(l) + c.slice(1));
	let d = c.charCodeAt(c.length - 1), f = W(r.after.charCodeAt(0), d, i);
	f.inside && (c = c.slice(0, -1) + H(d));
	let p = o.move(i + i);
	return a(), n.attentionEncodeSurroundingInfo = {
		after: f.outside,
		before: u.outside
	}, s + c + p;
}
function Ke(e, t, n) {
	return n.options.strong || "*";
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/mdast-util-to-markdown/2.1.2/bf6627bebb2b304f751c9c5e7d9b2000845c3a65bcdc58c5070d516b6bbcc50c/node_modules/mdast-util-to-markdown/lib/handle/text.js
function qe(e, t, n, r) {
	return n.safe(e.value, r);
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/mdast-util-to-markdown/2.1.2/bf6627bebb2b304f751c9c5e7d9b2000845c3a65bcdc58c5070d516b6bbcc50c/node_modules/mdast-util-to-markdown/lib/util/check-rule-repetition.js
function Je(e) {
	let t = e.options.ruleRepetition || 3;
	if (t < 3) throw Error("Cannot serialize rules with repetition `" + t + "` for `options.ruleRepetition`, expected `3` or more");
	return t;
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/mdast-util-to-markdown/2.1.2/bf6627bebb2b304f751c9c5e7d9b2000845c3a65bcdc58c5070d516b6bbcc50c/node_modules/mdast-util-to-markdown/lib/handle/thematic-break.js
function Ye(e, t, n) {
	let r = (Le(n) + (n.options.ruleSpaces ? " " : "")).repeat(Je(n));
	return n.options.ruleSpaces ? r.slice(0, -1) : r;
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/mdast-util-to-markdown/2.1.2/bf6627bebb2b304f751c9c5e7d9b2000845c3a65bcdc58c5070d516b6bbcc50c/node_modules/mdast-util-to-markdown/lib/handle/index.js
var Xe = {
	blockquote: z,
	break: re,
	code: se,
	definition: ue,
	emphasis: fe,
	hardBreak: re,
	heading: be,
	html: xe,
	image: Ce,
	imageReference: Te,
	inlineCode: De,
	link: Ae,
	linkReference: Me,
	list: Re,
	listItem: Be,
	paragraph: Ve,
	root: Ue,
	strong: Ge,
	text: qe,
	thematicBreak: Ye
}, Ze = document.createElement("i");
function Qe(e) {
	let t = "&" + e + ";";
	Ze.innerHTML = t;
	let n = Ze.textContent;
	return n.charCodeAt(n.length - 1) === 59 && e !== "semi" ? !1 : n !== t && n;
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/micromark-util-decode-numeric-character-reference/2.0.2/2f367c0e1b15bf77ce33033a7bf92715227cb1dd04f3a9a248f48df6e3b18ef5/node_modules/micromark-util-decode-numeric-character-reference/index.js
function $e(e, t) {
	let n = Number.parseInt(e, t);
	return n < 9 || n === 11 || n > 13 && n < 32 || n > 126 && n < 160 || n > 55295 && n < 57344 || n > 64975 && n < 65008 || (n & 65535) == 65535 || (n & 65535) == 65534 || n > 1114111 ? "�" : String.fromCodePoint(n);
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/micromark-util-decode-string/2.0.1/41338bd8900bae35a1562df2eead2f8e2bae1fe288b45aeb7b1bda00ba32c898/node_modules/micromark-util-decode-string/index.js
var et = /\\([!-/:-@[-`{-~])|&(#(?:\d{1,7}|x[\da-f]{1,6})|[\da-z]{1,31});/gi;
function tt(e) {
	return e.replace(et, nt);
}
function nt(e, t, n) {
	if (t) return t;
	if (n.charCodeAt(0) === 35) {
		let e = n.charCodeAt(1), t = e === 120 || e === 88;
		return $e(n.slice(t ? 2 : 1), t ? 16 : 10);
	}
	return Qe(n) || e;
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/mdast-util-gfm-table/2.0.0/cbcd0cf53f80de3bc7a3a47c38b1ab37f76ea11564a29da9ad50cc523fcc7887/node_modules/mdast-util-gfm-table/lib/index.js
function rt() {
	return {
		enter: {
			table: it,
			tableData: ct,
			tableHeader: ct,
			tableRow: ot
		},
		exit: {
			codeText: lt,
			table: at,
			tableData: st,
			tableHeader: st,
			tableRow: st
		}
	};
}
function it(e) {
	let t = e._align;
	this.enter({
		type: "table",
		align: t.map(function(e) {
			return e === "none" ? null : e;
		}),
		children: []
	}, e), this.data.inTable = !0;
}
function at(e) {
	this.exit(e), this.data.inTable = void 0;
}
function ot(e) {
	this.enter({
		type: "tableRow",
		children: []
	}, e);
}
function st(e) {
	this.exit(e);
}
function ct(e) {
	this.enter({
		type: "tableCell",
		children: []
	}, e);
}
function lt(e) {
	let t = this.resume();
	this.data.inTable && (t = t.replace(/\\([\\|])/g, ut));
	let n = this.stack[this.stack.length - 1];
	n.type, n.value = t, this.exit(e);
}
function ut(e, t) {
	return t === "|" ? t : e;
}
function dt(e) {
	let t = e || {}, n = t.tableCellPadding, r = t.tablePipeAlign, i = t.stringLength, a = n ? " " : "|";
	return {
		unsafe: [
			{
				character: "\r",
				inConstruct: "tableCell"
			},
			{
				character: "\n",
				inConstruct: "tableCell"
			},
			{
				atBreak: !0,
				character: "|",
				after: "[	 :-]"
			},
			{
				character: "|",
				inConstruct: "tableCell"
			},
			{
				atBreak: !0,
				character: ":",
				after: "-"
			},
			{
				atBreak: !0,
				character: "-",
				after: "[:|-]"
			}
		],
		handlers: {
			inlineCode: f,
			table: o,
			tableCell: c,
			tableRow: s
		}
	};
	function o(e, t, n, r) {
		return l(u(e, n, r), e.align);
	}
	function s(e, t, n, r) {
		let i = l([d(e, n, r)]);
		return i.slice(0, i.indexOf("\n"));
	}
	function c(e, t, n, r) {
		let i = n.enter("tableCell"), o = n.enter("phrasing"), s = n.containerPhrasing(e, {
			...r,
			before: a,
			after: a
		});
		return o(), i(), s;
	}
	function l(e, t) {
		return L(e, {
			align: t,
			alignDelimiters: r,
			padding: n,
			stringLength: i
		});
	}
	function u(e, t, n) {
		let r = e.children, i = -1, a = [], o = t.enter("table");
		for (; ++i < r.length;) a[i] = d(r[i], t, n);
		return o(), a;
	}
	function d(e, t, n) {
		let r = e.children, i = -1, a = [], o = t.enter("tableRow");
		for (; ++i < r.length;) a[i] = c(r[i], e, t, n);
		return o(), a;
	}
	function f(e, t, n) {
		let r = Xe.inlineCode(e, t, n);
		return n.stack.includes("tableCell") && (r = r.replace(/\|/g, "\\$&")), r;
	}
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/mdast-util-gfm-task-list-item/2.0.0/0f3250a2c33fca61c2896d2b8111fb1503979f5b5726875cd6838609be5ce427/node_modules/mdast-util-gfm-task-list-item/lib/index.js
function ft() {
	return { exit: {
		taskListCheckValueChecked: mt,
		taskListCheckValueUnchecked: mt,
		paragraph: ht
	} };
}
function pt() {
	return {
		unsafe: [{
			atBreak: !0,
			character: "-",
			after: "[:|-]"
		}],
		handlers: { listItem: gt }
	};
}
function mt(e) {
	let t = this.stack[this.stack.length - 2];
	t.type, t.checked = e.type === "taskListCheckValueChecked";
}
function ht(e) {
	let t = this.stack[this.stack.length - 2];
	if (t && t.type === "listItem" && typeof t.checked == "boolean") {
		let e = this.stack[this.stack.length - 1];
		e.type;
		let n = e.children[0];
		if (n && n.type === "text") {
			let r = t.children, i = -1, a;
			for (; ++i < r.length;) {
				let e = r[i];
				if (e.type === "paragraph") {
					a = e;
					break;
				}
			}
			a === e && (n.value = n.value.slice(1), n.value.length === 0 ? e.children.shift() : e.position && n.position && typeof n.position.start.offset == "number" && (n.position.start.column++, n.position.start.offset++, e.position.start = Object.assign({}, n.position.start)));
		}
	}
	this.exit(e);
}
function gt(e, t, n, r) {
	let i = e.children[0], a = typeof e.checked == "boolean" && i && i.type === "paragraph", o = "[" + (e.checked ? "x" : " ") + "] ", s = n.createTracker(r);
	a && s.move(o);
	let c = Xe.listItem(e, t, n, {
		...r,
		...s.current()
	});
	return a && (c = c.replace(/^(?:[*+-]|\d+\.)([\r\n]| {1,3})/, l)), c;
	function l(e) {
		return e + o;
	}
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/micromark-util-chunked/2.0.1/5556f402c8dd7f1c5dc24994f18cc3360671b7719ff8fbff86f737513a2650e0/node_modules/micromark-util-chunked/index.js
function G(e, t, n, r) {
	let i = e.length, a = 0, o;
	if (t = t < 0 ? -t > i ? 0 : i + t : t > i ? i : t, n = n > 0 ? n : 0, r.length < 1e4) o = Array.from(r), o.unshift(t, n), e.splice(...o);
	else for (n && e.splice(t, n); a < r.length;) o = r.slice(a, a + 1e4), o.unshift(t, 0), e.splice(...o), a += 1e4, t += 1e4;
}
function K(e, t) {
	return e.length > 0 ? (G(e, e.length, 0, t), e) : t;
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/micromark-util-combine-extensions/2.0.1/68c93e8838c173b07bc90f6082a7d0f2ff14f95894305fef8387956cd4cb4023/node_modules/micromark-util-combine-extensions/index.js
var _t = {}.hasOwnProperty;
function vt(e) {
	let t = {}, n = -1;
	for (; ++n < e.length;) yt(t, e[n]);
	return t;
}
function yt(e, t) {
	let n;
	for (n in t) {
		let r = (_t.call(e, n) ? e[n] : void 0) || (e[n] = {}), i = t[n], a;
		if (i) for (a in i) {
			_t.call(r, a) || (r[a] = []);
			let e = i[a];
			bt(r[a], Array.isArray(e) ? e : e ? [e] : []);
		}
	}
}
function bt(e, t) {
	let n = -1, r = [];
	for (; ++n < t.length;) (t[n].add === "after" ? e : r).push(t[n]);
	G(e, 0, 0, r);
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/micromark-util-resolve-all/2.0.1/2bce0c71e2614073cad22a013e9d283ed9515802e6a71cf27f865503cd1287f6/node_modules/micromark-util-resolve-all/index.js
function xt(e, t, n) {
	let r = [], i = -1;
	for (; ++i < e.length;) {
		let a = e[i].resolveAll;
		a && !r.includes(a) && (t = a(t, n), r.push(a));
	}
	return t;
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/micromark-core-commonmark/2.0.3/a1b4dd720a345ea7669bd6eac1b98d97da85d9a8fd797edbe457df3db914e918/node_modules/micromark-core-commonmark/lib/attention.js
var St = {
	name: "attention",
	resolveAll: Ct,
	tokenize: wt
};
function Ct(e, t) {
	let n = -1, r, i, a, o, s, c, l, u;
	for (; ++n < e.length;) if (e[n][0] === "enter" && e[n][1].type === "attentionSequence" && e[n][1]._close) {
		for (r = n; r--;) if (e[r][0] === "exit" && e[r][1].type === "attentionSequence" && e[r][1]._open && t.sliceSerialize(e[r][1]).charCodeAt(0) === t.sliceSerialize(e[n][1]).charCodeAt(0)) {
			if ((e[r][1]._close || e[n][1]._open) && (e[n][1].end.offset - e[n][1].start.offset) % 3 && !((e[r][1].end.offset - e[r][1].start.offset + e[n][1].end.offset - e[n][1].start.offset) % 3)) continue;
			c = e[r][1].end.offset - e[r][1].start.offset > 1 && e[n][1].end.offset - e[n][1].start.offset > 1 ? 2 : 1;
			let d = { ...e[r][1].end }, f = { ...e[n][1].start };
			Tt(d, -c), Tt(f, c), o = {
				type: c > 1 ? "strongSequence" : "emphasisSequence",
				start: d,
				end: { ...e[r][1].end }
			}, s = {
				type: c > 1 ? "strongSequence" : "emphasisSequence",
				start: { ...e[n][1].start },
				end: f
			}, a = {
				type: c > 1 ? "strongText" : "emphasisText",
				start: { ...e[r][1].end },
				end: { ...e[n][1].start }
			}, i = {
				type: c > 1 ? "strong" : "emphasis",
				start: { ...o.start },
				end: { ...s.end }
			}, e[r][1].end = { ...o.start }, e[n][1].start = { ...s.end }, l = [], e[r][1].end.offset - e[r][1].start.offset && (l = K(l, [[
				"enter",
				e[r][1],
				t
			], [
				"exit",
				e[r][1],
				t
			]])), l = K(l, [
				[
					"enter",
					i,
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
					"enter",
					a,
					t
				]
			]), l = K(l, xt(t.parser.constructs.insideSpan.null, e.slice(r + 1, n), t)), l = K(l, [
				[
					"exit",
					a,
					t
				],
				[
					"enter",
					s,
					t
				],
				[
					"exit",
					s,
					t
				],
				[
					"exit",
					i,
					t
				]
			]), e[n][1].end.offset - e[n][1].start.offset ? (u = 2, l = K(l, [[
				"enter",
				e[n][1],
				t
			], [
				"exit",
				e[n][1],
				t
			]])) : u = 0, G(e, r - 1, n - r + 3, l), n = r + l.length - u - 2;
			break;
		}
	}
	for (n = -1; ++n < e.length;) e[n][1].type === "attentionSequence" && (e[n][1].type = "data");
	return e;
}
function wt(e, t) {
	let n = this.parser.constructs.attentionMarkers.null, r = this.previous, i = U(r), a;
	return o;
	function o(t) {
		return a = t, e.enter("attentionSequence"), s(t);
	}
	function s(o) {
		if (o === a) return e.consume(o), s;
		let c = e.exit("attentionSequence"), l = U(o), u = !l || l === 2 && i || n.includes(o), d = !i || i === 2 && l || n.includes(r);
		return c._open = !!(a === 42 ? u : u && (i || !d)), c._close = !!(a === 42 ? d : d && (l || !u)), t(o);
	}
}
function Tt(e, t) {
	e.column += t, e.offset += t, e._bufferIndex += t;
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/micromark-core-commonmark/2.0.3/a1b4dd720a345ea7669bd6eac1b98d97da85d9a8fd797edbe457df3db914e918/node_modules/micromark-core-commonmark/lib/autolink.js
var Et = {
	name: "autolink",
	tokenize: Dt
};
function Dt(e, t, n) {
	let r = 0;
	return i;
	function i(t) {
		return e.enter("autolink"), e.enter("autolinkMarker"), e.consume(t), e.exit("autolinkMarker"), e.enter("autolinkProtocol"), a;
	}
	function a(t) {
		return h(t) ? (e.consume(t), o) : t === 64 ? n(t) : l(t);
	}
	function o(e) {
		return e === 43 || e === 45 || e === 46 || g(e) ? (r = 1, s(e)) : l(e);
	}
	function s(t) {
		return t === 58 ? (e.consume(t), r = 0, c) : (t === 43 || t === 45 || t === 46 || g(t)) && r++ < 32 ? (e.consume(t), s) : (r = 0, l(t));
	}
	function c(r) {
		return r === 62 ? (e.exit("autolinkProtocol"), e.enter("autolinkMarker"), e.consume(r), e.exit("autolinkMarker"), e.exit("autolink"), t) : r === null || r === 32 || r === 60 || v(r) ? n(r) : (e.consume(r), c);
	}
	function l(t) {
		return t === 64 ? (e.consume(t), u) : _(t) ? (e.consume(t), l) : n(t);
	}
	function u(e) {
		return g(e) ? d(e) : n(e);
	}
	function d(n) {
		return n === 46 ? (e.consume(n), r = 0, u) : n === 62 ? (e.exit("autolinkProtocol").type = "autolinkEmail", e.enter("autolinkMarker"), e.consume(n), e.exit("autolinkMarker"), e.exit("autolink"), t) : f(n);
	}
	function f(t) {
		if ((t === 45 || g(t)) && r++ < 63) {
			let n = t === 45 ? f : d;
			return e.consume(t), n;
		}
		return n(t);
	}
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/micromark-factory-space/2.0.1/400e2bc1ce17e213f989a3e39c3727555d92389777dc9db84afa396691e0775a/node_modules/micromark-factory-space/index.js
function q(e, t, n, r) {
	let i = r ? r - 1 : Infinity, a = 0;
	return o;
	function o(r) {
		return w(r) ? (e.enter(n), s(r)) : t(r);
	}
	function s(r) {
		return w(r) && a++ < i ? (e.consume(r), s) : (e.exit(n), t(r));
	}
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/micromark-core-commonmark/2.0.3/a1b4dd720a345ea7669bd6eac1b98d97da85d9a8fd797edbe457df3db914e918/node_modules/micromark-core-commonmark/lib/blank-line.js
var Ot = {
	partial: !0,
	tokenize: kt
};
function kt(e, t, n) {
	return r;
	function r(t) {
		return w(t) ? q(e, i, "linePrefix")(t) : i(t);
	}
	function i(e) {
		return e === null || S(e) ? t(e) : n(e);
	}
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/micromark-core-commonmark/2.0.3/a1b4dd720a345ea7669bd6eac1b98d97da85d9a8fd797edbe457df3db914e918/node_modules/micromark-core-commonmark/lib/block-quote.js
var At = {
	continuation: { tokenize: Mt },
	exit: Nt,
	name: "blockQuote",
	tokenize: jt
};
function jt(e, t, n) {
	let r = this;
	return i;
	function i(t) {
		if (t === 62) {
			let n = r.containerState;
			return n.open ||= (e.enter("blockQuote", { _container: !0 }), !0), e.enter("blockQuotePrefix"), e.enter("blockQuoteMarker"), e.consume(t), e.exit("blockQuoteMarker"), a;
		}
		return n(t);
	}
	function a(n) {
		return w(n) ? (e.enter("blockQuotePrefixWhitespace"), e.consume(n), e.exit("blockQuotePrefixWhitespace"), e.exit("blockQuotePrefix"), t) : (e.exit("blockQuotePrefix"), t(n));
	}
}
function Mt(e, t, n) {
	let r = this;
	return i;
	function i(t) {
		return w(t) ? q(e, a, "linePrefix", r.parser.constructs.disable.null.includes("codeIndented") ? void 0 : 4)(t) : a(t);
	}
	function a(r) {
		return e.attempt(At, t, n)(r);
	}
}
function Nt(e) {
	e.exit("blockQuote");
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/micromark-core-commonmark/2.0.3/a1b4dd720a345ea7669bd6eac1b98d97da85d9a8fd797edbe457df3db914e918/node_modules/micromark-core-commonmark/lib/character-escape.js
var Pt = {
	name: "characterEscape",
	tokenize: Ft
};
function Ft(e, t, n) {
	return r;
	function r(t) {
		return e.enter("characterEscape"), e.enter("escapeMarker"), e.consume(t), e.exit("escapeMarker"), i;
	}
	function i(r) {
		return x(r) ? (e.enter("characterEscapeValue"), e.consume(r), e.exit("characterEscapeValue"), e.exit("characterEscape"), t) : n(r);
	}
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/micromark-core-commonmark/2.0.3/a1b4dd720a345ea7669bd6eac1b98d97da85d9a8fd797edbe457df3db914e918/node_modules/micromark-core-commonmark/lib/character-reference.js
var It = {
	name: "characterReference",
	tokenize: Lt
};
function Lt(e, t, n) {
	let r = this, i = 0, a, o;
	return s;
	function s(t) {
		return e.enter("characterReference"), e.enter("characterReferenceMarker"), e.consume(t), e.exit("characterReferenceMarker"), c;
	}
	function c(t) {
		return t === 35 ? (e.enter("characterReferenceMarkerNumeric"), e.consume(t), e.exit("characterReferenceMarkerNumeric"), l) : (e.enter("characterReferenceValue"), a = 31, o = g, u(t));
	}
	function l(t) {
		return t === 88 || t === 120 ? (e.enter("characterReferenceMarkerHexadecimal"), e.consume(t), e.exit("characterReferenceMarkerHexadecimal"), e.enter("characterReferenceValue"), a = 6, o = b, u) : (e.enter("characterReferenceValue"), a = 7, o = y, u(t));
	}
	function u(s) {
		if (s === 59 && i) {
			let i = e.exit("characterReferenceValue");
			return o === g && !Qe(r.sliceSerialize(i)) ? n(s) : (e.enter("characterReferenceMarker"), e.consume(s), e.exit("characterReferenceMarker"), e.exit("characterReference"), t);
		}
		return o(s) && i++ < a ? (e.consume(s), u) : n(s);
	}
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/micromark-core-commonmark/2.0.3/a1b4dd720a345ea7669bd6eac1b98d97da85d9a8fd797edbe457df3db914e918/node_modules/micromark-core-commonmark/lib/code-fenced.js
var Rt = {
	partial: !0,
	tokenize: Vt
}, zt = {
	concrete: !0,
	name: "codeFenced",
	tokenize: Bt
};
function Bt(e, t, n) {
	let r = this, i = {
		partial: !0,
		tokenize: x
	}, a = 0, o = 0, s;
	return c;
	function c(e) {
		return l(e);
	}
	function l(t) {
		let n = r.events[r.events.length - 1];
		return a = n && n[1].type === "linePrefix" ? n[2].sliceSerialize(n[1], !0).length : 0, s = t, e.enter("codeFenced"), e.enter("codeFencedFence"), e.enter("codeFencedFenceSequence"), u(t);
	}
	function u(t) {
		return t === s ? (o++, e.consume(t), u) : o < 3 ? n(t) : (e.exit("codeFencedFenceSequence"), w(t) ? q(e, d, "whitespace")(t) : d(t));
	}
	function d(n) {
		return n === null || S(n) ? (e.exit("codeFencedFence"), r.interrupt ? t(n) : e.check(Rt, h, b)(n)) : (e.enter("codeFencedFenceInfo"), e.enter("chunkString", { contentType: "string" }), f(n));
	}
	function f(t) {
		return t === null || S(t) ? (e.exit("chunkString"), e.exit("codeFencedFenceInfo"), d(t)) : w(t) ? (e.exit("chunkString"), e.exit("codeFencedFenceInfo"), q(e, p, "whitespace")(t)) : t === 96 && t === s ? n(t) : (e.consume(t), f);
	}
	function p(t) {
		return t === null || S(t) ? d(t) : (e.enter("codeFencedFenceMeta"), e.enter("chunkString", { contentType: "string" }), m(t));
	}
	function m(t) {
		return t === null || S(t) ? (e.exit("chunkString"), e.exit("codeFencedFenceMeta"), d(t)) : t === 96 && t === s ? n(t) : (e.consume(t), m);
	}
	function h(t) {
		return e.attempt(i, b, g)(t);
	}
	function g(t) {
		return e.enter("lineEnding"), e.consume(t), e.exit("lineEnding"), _;
	}
	function _(t) {
		return a > 0 && w(t) ? q(e, v, "linePrefix", a + 1)(t) : v(t);
	}
	function v(t) {
		return t === null || S(t) ? e.check(Rt, h, b)(t) : (e.enter("codeFlowValue"), y(t));
	}
	function y(t) {
		return t === null || S(t) ? (e.exit("codeFlowValue"), v(t)) : (e.consume(t), y);
	}
	function b(n) {
		return e.exit("codeFenced"), t(n);
	}
	function x(e, t, n) {
		let i = 0;
		return a;
		function a(t) {
			return e.enter("lineEnding"), e.consume(t), e.exit("lineEnding"), c;
		}
		function c(t) {
			return e.enter("codeFencedFence"), w(t) ? q(e, l, "linePrefix", r.parser.constructs.disable.null.includes("codeIndented") ? void 0 : 4)(t) : l(t);
		}
		function l(t) {
			return t === s ? (e.enter("codeFencedFenceSequence"), u(t)) : n(t);
		}
		function u(t) {
			return t === s ? (i++, e.consume(t), u) : i >= o ? (e.exit("codeFencedFenceSequence"), w(t) ? q(e, d, "whitespace")(t) : d(t)) : n(t);
		}
		function d(r) {
			return r === null || S(r) ? (e.exit("codeFencedFence"), t(r)) : n(r);
		}
	}
}
function Vt(e, t, n) {
	let r = this;
	return i;
	function i(t) {
		return t === null ? n(t) : (e.enter("lineEnding"), e.consume(t), e.exit("lineEnding"), a);
	}
	function a(e) {
		return r.parser.lazy[r.now().line] ? n(e) : t(e);
	}
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/micromark-core-commonmark/2.0.3/a1b4dd720a345ea7669bd6eac1b98d97da85d9a8fd797edbe457df3db914e918/node_modules/micromark-core-commonmark/lib/code-indented.js
var Ht = {
	name: "codeIndented",
	tokenize: Wt
}, Ut = {
	partial: !0,
	tokenize: Gt
};
function Wt(e, t, n) {
	let r = this;
	return i;
	function i(t) {
		return e.enter("codeIndented"), q(e, a, "linePrefix", 5)(t);
	}
	function a(e) {
		let t = r.events[r.events.length - 1];
		return t && t[1].type === "linePrefix" && t[2].sliceSerialize(t[1], !0).length >= 4 ? o(e) : n(e);
	}
	function o(t) {
		return t === null ? c(t) : S(t) ? e.attempt(Ut, o, c)(t) : (e.enter("codeFlowValue"), s(t));
	}
	function s(t) {
		return t === null || S(t) ? (e.exit("codeFlowValue"), o(t)) : (e.consume(t), s);
	}
	function c(n) {
		return e.exit("codeIndented"), t(n);
	}
}
function Gt(e, t, n) {
	let r = this;
	return i;
	function i(t) {
		return r.parser.lazy[r.now().line] ? n(t) : S(t) ? (e.enter("lineEnding"), e.consume(t), e.exit("lineEnding"), i) : q(e, a, "linePrefix", 5)(t);
	}
	function a(e) {
		let a = r.events[r.events.length - 1];
		return a && a[1].type === "linePrefix" && a[2].sliceSerialize(a[1], !0).length >= 4 ? t(e) : S(e) ? i(e) : n(e);
	}
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/micromark-core-commonmark/2.0.3/a1b4dd720a345ea7669bd6eac1b98d97da85d9a8fd797edbe457df3db914e918/node_modules/micromark-core-commonmark/lib/code-text.js
var Kt = {
	name: "codeText",
	previous: Jt,
	resolve: qt,
	tokenize: Yt
};
function qt(e) {
	let t = e.length - 4, n = 3, r, i;
	if ((e[n][1].type === "lineEnding" || e[n][1].type === "space") && (e[t][1].type === "lineEnding" || e[t][1].type === "space")) {
		for (r = n; ++r < t;) if (e[r][1].type === "codeTextData") {
			e[n][1].type = "codeTextPadding", e[t][1].type = "codeTextPadding", n += 2, t -= 2;
			break;
		}
	}
	for (r = n - 1, t++; ++r <= t;) i === void 0 ? r !== t && e[r][1].type !== "lineEnding" && (i = r) : (r === t || e[r][1].type === "lineEnding") && (e[i][1].type = "codeTextData", r !== i + 2 && (e[i][1].end = e[r - 1][1].end, e.splice(i + 2, r - i - 2), t -= r - i - 2, r = i + 2), i = void 0);
	return e;
}
function Jt(e) {
	return e !== 96 || this.events[this.events.length - 1][1].type === "characterEscape";
}
function Yt(e, t, n) {
	let r = 0, i, a;
	return o;
	function o(t) {
		return e.enter("codeText"), e.enter("codeTextSequence"), s(t);
	}
	function s(t) {
		return t === 96 ? (e.consume(t), r++, s) : (e.exit("codeTextSequence"), c(t));
	}
	function c(t) {
		return t === null ? n(t) : t === 32 ? (e.enter("space"), e.consume(t), e.exit("space"), c) : t === 96 ? (a = e.enter("codeTextSequence"), i = 0, u(t)) : S(t) ? (e.enter("lineEnding"), e.consume(t), e.exit("lineEnding"), c) : (e.enter("codeTextData"), l(t));
	}
	function l(t) {
		return t === null || t === 32 || t === 96 || S(t) ? (e.exit("codeTextData"), c(t)) : (e.consume(t), l);
	}
	function u(n) {
		return n === 96 ? (e.consume(n), i++, u) : i === r ? (e.exit("codeTextSequence"), e.exit("codeText"), t(n)) : (a.type = "codeTextData", l(n));
	}
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/micromark-util-subtokenize/2.1.0/696a1fbc26eea3e93d7e962ddd6e844e2e5d13fa526ec8c9d5c529752e1db321/node_modules/micromark-util-subtokenize/lib/splice-buffer.js
var Xt = class {
	constructor(e) {
		this.left = e ? [...e] : [], this.right = [];
	}
	get(e) {
		if (e < 0 || e >= this.left.length + this.right.length) throw RangeError("Cannot access index `" + e + "` in a splice buffer of size `" + (this.left.length + this.right.length) + "`");
		return e < this.left.length ? this.left[e] : this.right[this.right.length - e + this.left.length - 1];
	}
	get length() {
		return this.left.length + this.right.length;
	}
	shift() {
		return this.setCursor(0), this.right.pop();
	}
	slice(e, t) {
		let n = t ?? Infinity;
		return n < this.left.length ? this.left.slice(e, n) : e > this.left.length ? this.right.slice(this.right.length - n + this.left.length, this.right.length - e + this.left.length).reverse() : this.left.slice(e).concat(this.right.slice(this.right.length - n + this.left.length).reverse());
	}
	splice(e, t, n) {
		let r = t || 0;
		this.setCursor(Math.trunc(e));
		let i = this.right.splice(this.right.length - r, Infinity);
		return n && Zt(this.left, n), i.reverse();
	}
	pop() {
		return this.setCursor(Infinity), this.left.pop();
	}
	push(e) {
		this.setCursor(Infinity), this.left.push(e);
	}
	pushMany(e) {
		this.setCursor(Infinity), Zt(this.left, e);
	}
	unshift(e) {
		this.setCursor(0), this.right.push(e);
	}
	unshiftMany(e) {
		this.setCursor(0), Zt(this.right, e.reverse());
	}
	setCursor(e) {
		if (!(e === this.left.length || e > this.left.length && this.right.length === 0 || e < 0 && this.left.length === 0)) {
			if (e < this.left.length) {
				let t = this.left.splice(e, Infinity);
				Zt(this.right, t.reverse());
			} else {
				let t = this.right.splice(this.left.length + this.right.length - e, Infinity);
				Zt(this.left, t.reverse());
			}
		}
	}
};
function Zt(e, t) {
	let n = 0;
	if (t.length < 1e4) e.push(...t);
	else for (; n < t.length;) e.push(...t.slice(n, n + 1e4)), n += 1e4;
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/micromark-util-subtokenize/2.1.0/696a1fbc26eea3e93d7e962ddd6e844e2e5d13fa526ec8c9d5c529752e1db321/node_modules/micromark-util-subtokenize/index.js
function Qt(e) {
	let t = {}, n = -1, r, i, a, o, s, c, l, u = new Xt(e);
	for (; ++n < u.length;) {
		for (; n in t;) n = t[n];
		if (r = u.get(n), n && r[1].type === "chunkFlow" && u.get(n - 1)[1].type === "listItemPrefix" && (c = r[1]._tokenizer.events, a = 0, a < c.length && c[a][1].type === "lineEndingBlank" && (a += 2), a < c.length && c[a][1].type === "content")) for (; ++a < c.length && c[a][1].type !== "content";) c[a][1].type === "chunkText" && (c[a][1]._isInFirstContentOfListItem = !0, a++);
		if (r[0] === "enter") r[1].contentType && (Object.assign(t, $t(u, n)), n = t[n], l = !0);
		else if (r[1]._container) {
			for (a = n, i = void 0; a--;) if (o = u.get(a), o[1].type === "lineEnding" || o[1].type === "lineEndingBlank") o[0] === "enter" && (i && (u.get(i)[1].type = "lineEndingBlank"), o[1].type = "lineEnding", i = a);
			else if (o[1].type !== "linePrefix" && o[1].type !== "listItemIndent") break;
			i && (r[1].end = { ...u.get(i)[1].start }, s = u.slice(i, n), s.unshift(r), u.splice(i, n - i + 1, s));
		}
	}
	return G(e, 0, Infinity, u.slice(0)), !l;
}
function $t(e, t) {
	let n = e.get(t)[1], r = e.get(t)[2], i = t - 1, a = [], o = n._tokenizer;
	o || (o = r.parser[n.contentType](n.start), n._contentTypeTextTrailing && (o._contentTypeTextTrailing = !0));
	let s = o.events, c = [], l = {}, u, d, f = -1, p = n, m = 0, h = 0, g = [h];
	for (; p;) {
		for (; e.get(++i)[1] !== p;);
		a.push(i), p._tokenizer || (u = r.sliceStream(p), p.next || u.push(null), d && o.defineSkip(p.start), p._isInFirstContentOfListItem && (o._gfmTasklistFirstContentOfListItem = !0), o.write(u), p._isInFirstContentOfListItem && (o._gfmTasklistFirstContentOfListItem = void 0)), d = p, p = p.next;
	}
	for (p = n; ++f < s.length;) s[f][0] === "exit" && s[f - 1][0] === "enter" && s[f][1].type === s[f - 1][1].type && s[f][1].start.line !== s[f][1].end.line && (h = f + 1, g.push(h), p._tokenizer = void 0, p.previous = void 0, p = p.next);
	for (o.events = [], p ? (p._tokenizer = void 0, p.previous = void 0) : g.pop(), f = g.length; f--;) {
		let t = s.slice(g[f], g[f + 1]), n = a.pop();
		c.push([n, n + t.length - 1]), e.splice(n, 2, t);
	}
	for (c.reverse(), f = -1; ++f < c.length;) l[m + c[f][0]] = m + c[f][1], m += c[f][1] - c[f][0] - 1;
	return l;
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/micromark-core-commonmark/2.0.3/a1b4dd720a345ea7669bd6eac1b98d97da85d9a8fd797edbe457df3db914e918/node_modules/micromark-core-commonmark/lib/content.js
var en = {
	resolve: nn,
	tokenize: rn
}, tn = {
	partial: !0,
	tokenize: an
};
function nn(e) {
	return Qt(e), e;
}
function rn(e, t) {
	let n;
	return r;
	function r(t) {
		return e.enter("content"), n = e.enter("chunkContent", { contentType: "content" }), i(t);
	}
	function i(t) {
		return t === null ? a(t) : S(t) ? e.check(tn, o, a)(t) : (e.consume(t), i);
	}
	function a(n) {
		return e.exit("chunkContent"), e.exit("content"), t(n);
	}
	function o(t) {
		return e.consume(t), e.exit("chunkContent"), n.next = e.enter("chunkContent", {
			contentType: "content",
			previous: n
		}), n = n.next, i;
	}
}
function an(e, t, n) {
	let r = this;
	return i;
	function i(t) {
		return e.exit("chunkContent"), e.enter("lineEnding"), e.consume(t), e.exit("lineEnding"), q(e, a, "linePrefix");
	}
	function a(i) {
		if (i === null || S(i)) return n(i);
		let a = r.events[r.events.length - 1];
		return !r.parser.constructs.disable.null.includes("codeIndented") && a && a[1].type === "linePrefix" && a[2].sliceSerialize(a[1], !0).length >= 4 ? t(i) : e.interrupt(r.parser.constructs.flow, n, t)(i);
	}
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/micromark-factory-destination/2.0.1/26df63127760915841dc48387910b7046389fd4b7c1856c2fd7f975f9ce7eb01/node_modules/micromark-factory-destination/index.js
function on(e, t, n, r, i, a, o, s, c) {
	let l = c || Infinity, u = 0;
	return d;
	function d(t) {
		return t === 60 ? (e.enter(r), e.enter(i), e.enter(a), e.consume(t), e.exit(a), f) : t === null || t === 32 || t === 41 || v(t) ? n(t) : (e.enter(r), e.enter(o), e.enter(s), e.enter("chunkString", { contentType: "string" }), h(t));
	}
	function f(n) {
		return n === 62 ? (e.enter(a), e.consume(n), e.exit(a), e.exit(i), e.exit(r), t) : (e.enter(s), e.enter("chunkString", { contentType: "string" }), p(n));
	}
	function p(t) {
		return t === 62 ? (e.exit("chunkString"), e.exit(s), f(t)) : t === null || t === 60 || S(t) ? n(t) : (e.consume(t), t === 92 ? m : p);
	}
	function m(t) {
		return t === 60 || t === 62 || t === 92 ? (e.consume(t), p) : p(t);
	}
	function h(i) {
		return !u && (i === null || i === 41 || C(i)) ? (e.exit("chunkString"), e.exit(s), e.exit(o), e.exit(r), t(i)) : u < l && i === 40 ? (e.consume(i), u++, h) : i === 41 ? (e.consume(i), u--, h) : i === null || i === 32 || i === 40 || v(i) ? n(i) : (e.consume(i), i === 92 ? g : h);
	}
	function g(t) {
		return t === 40 || t === 41 || t === 92 ? (e.consume(t), h) : h(t);
	}
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/micromark-factory-label/2.0.1/432033e48060bda7b13079caf4c3445e0dbd05c86e60e81931fe7b945b7a3598/node_modules/micromark-factory-label/index.js
function sn(e, t, n, r, i, a) {
	let o = this, s = 0, c;
	return l;
	function l(t) {
		return e.enter(r), e.enter(i), e.consume(t), e.exit(i), e.enter(a), u;
	}
	function u(l) {
		return s > 999 || l === null || l === 91 || l === 93 && !c || 
		/* c8 ignore next 3 */
		l === 94 && !s && "_hiddenFootnoteSupport" in o.parser.constructs ? n(l) : l === 93 ? (e.exit(a), e.enter(i), e.consume(l), e.exit(i), e.exit(r), t) : S(l) ? (e.enter("lineEnding"), e.consume(l), e.exit("lineEnding"), u) : (e.enter("chunkString", { contentType: "string" }), d(l));
	}
	function d(t) {
		return t === null || t === 91 || t === 93 || S(t) || s++ > 999 ? (e.exit("chunkString"), u(t)) : (e.consume(t), c ||= !w(t), t === 92 ? f : d);
	}
	function f(t) {
		return t === 91 || t === 92 || t === 93 ? (e.consume(t), s++, d) : d(t);
	}
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/micromark-factory-title/2.0.1/ecf7a38be71cf17c2e5e05efe1ed990796ffbdac8c63b191725890526b30a3f4/node_modules/micromark-factory-title/index.js
function cn(e, t, n, r, i, a) {
	let o;
	return s;
	function s(t) {
		return t === 34 || t === 39 || t === 40 ? (e.enter(r), e.enter(i), e.consume(t), e.exit(i), o = t === 40 ? 41 : t, c) : n(t);
	}
	function c(n) {
		return n === o ? (e.enter(i), e.consume(n), e.exit(i), e.exit(r), t) : (e.enter(a), l(n));
	}
	function l(t) {
		return t === o ? (e.exit(a), c(o)) : t === null ? n(t) : S(t) ? (e.enter("lineEnding"), e.consume(t), e.exit("lineEnding"), q(e, l, "linePrefix")) : (e.enter("chunkString", { contentType: "string" }), u(t));
	}
	function u(t) {
		return t === o || t === null || S(t) ? (e.exit("chunkString"), l(t)) : (e.consume(t), t === 92 ? d : u);
	}
	function d(t) {
		return t === o || t === 92 ? (e.consume(t), u) : u(t);
	}
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/micromark-factory-whitespace/2.0.1/69595bdfaf7baa9cd7d25461c9981fe1c1329373ae351088e5efc1e19ed3ef43/node_modules/micromark-factory-whitespace/index.js
function J(e, t) {
	let n;
	return r;
	function r(i) {
		return S(i) ? (e.enter("lineEnding"), e.consume(i), e.exit("lineEnding"), n = !0, r) : w(i) ? q(e, r, n ? "linePrefix" : "lineSuffix")(i) : t(i);
	}
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/micromark-core-commonmark/2.0.3/a1b4dd720a345ea7669bd6eac1b98d97da85d9a8fd797edbe457df3db914e918/node_modules/micromark-core-commonmark/lib/definition.js
var ln = {
	name: "definition",
	tokenize: dn
}, un = {
	partial: !0,
	tokenize: fn
};
function dn(e, t, n) {
	let r = this, i;
	return a;
	function a(t) {
		return e.enter("definition"), o(t);
	}
	function o(t) {
		return sn.call(r, e, s, n, "definitionLabel", "definitionLabelMarker", "definitionLabelString")(t);
	}
	function s(t) {
		return i = k(r.sliceSerialize(r.events[r.events.length - 1][1]).slice(1, -1)), t === 58 ? (e.enter("definitionMarker"), e.consume(t), e.exit("definitionMarker"), c) : n(t);
	}
	function c(t) {
		return C(t) ? J(e, l)(t) : l(t);
	}
	function l(t) {
		return on(e, u, n, "definitionDestination", "definitionDestinationLiteral", "definitionDestinationLiteralMarker", "definitionDestinationRaw", "definitionDestinationString")(t);
	}
	function u(t) {
		return e.attempt(un, d, d)(t);
	}
	function d(t) {
		return w(t) ? q(e, f, "whitespace")(t) : f(t);
	}
	function f(a) {
		return a === null || S(a) ? (e.exit("definition"), r.parser.defined.push(i), t(a)) : n(a);
	}
}
function fn(e, t, n) {
	return r;
	function r(t) {
		return C(t) ? J(e, i)(t) : n(t);
	}
	function i(t) {
		return cn(e, a, n, "definitionTitle", "definitionTitleMarker", "definitionTitleString")(t);
	}
	function a(t) {
		return w(t) ? q(e, o, "whitespace")(t) : o(t);
	}
	function o(e) {
		return e === null || S(e) ? t(e) : n(e);
	}
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/micromark-core-commonmark/2.0.3/a1b4dd720a345ea7669bd6eac1b98d97da85d9a8fd797edbe457df3db914e918/node_modules/micromark-core-commonmark/lib/hard-break-escape.js
var pn = {
	name: "hardBreakEscape",
	tokenize: mn
};
function mn(e, t, n) {
	return r;
	function r(t) {
		return e.enter("hardBreakEscape"), e.consume(t), i;
	}
	function i(r) {
		return S(r) ? (e.exit("hardBreakEscape"), t(r)) : n(r);
	}
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/micromark-core-commonmark/2.0.3/a1b4dd720a345ea7669bd6eac1b98d97da85d9a8fd797edbe457df3db914e918/node_modules/micromark-core-commonmark/lib/heading-atx.js
var hn = {
	name: "headingAtx",
	resolve: gn,
	tokenize: _n
};
function gn(e, t) {
	let n = e.length - 2, r = 3, i, a;
	return e[r][1].type === "whitespace" && (r += 2), n - 2 > r && e[n][1].type === "whitespace" && (n -= 2), e[n][1].type === "atxHeadingSequence" && (r === n - 1 || n - 4 > r && e[n - 2][1].type === "whitespace") && (n -= r + 1 === n ? 2 : 4), n > r && (i = {
		type: "atxHeadingText",
		start: e[r][1].start,
		end: e[n][1].end
	}, a = {
		type: "chunkText",
		start: e[r][1].start,
		end: e[n][1].end,
		contentType: "text"
	}, G(e, r, n - r + 1, [
		[
			"enter",
			i,
			t
		],
		[
			"enter",
			a,
			t
		],
		[
			"exit",
			a,
			t
		],
		[
			"exit",
			i,
			t
		]
	])), e;
}
function _n(e, t, n) {
	let r = 0;
	return i;
	function i(t) {
		return e.enter("atxHeading"), a(t);
	}
	function a(t) {
		return e.enter("atxHeadingSequence"), o(t);
	}
	function o(t) {
		return t === 35 && r++ < 6 ? (e.consume(t), o) : t === null || C(t) ? (e.exit("atxHeadingSequence"), s(t)) : n(t);
	}
	function s(n) {
		return n === 35 ? (e.enter("atxHeadingSequence"), c(n)) : n === null || S(n) ? (e.exit("atxHeading"), t(n)) : w(n) ? q(e, s, "whitespace")(n) : (e.enter("atxHeadingText"), l(n));
	}
	function c(t) {
		return t === 35 ? (e.consume(t), c) : (e.exit("atxHeadingSequence"), s(t));
	}
	function l(t) {
		return t === null || t === 35 || C(t) ? (e.exit("atxHeadingText"), s(t)) : (e.consume(t), l);
	}
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/micromark-util-html-tag-name/2.0.1/c0e9cae0a674b32b4cce94c294171bb954c1d2633ecb0af45deeca4fdcd935ce/node_modules/micromark-util-html-tag-name/index.js
var vn = /* @__PURE__ */ "address.article.aside.base.basefont.blockquote.body.caption.center.col.colgroup.dd.details.dialog.dir.div.dl.dt.fieldset.figcaption.figure.footer.form.frame.frameset.h1.h2.h3.h4.h5.h6.head.header.hr.html.iframe.legend.li.link.main.menu.menuitem.nav.noframes.ol.optgroup.option.p.param.search.section.summary.table.tbody.td.tfoot.th.thead.title.tr.track.ul".split("."), yn = [
	"pre",
	"script",
	"style",
	"textarea"
], bn = {
	concrete: !0,
	name: "htmlFlow",
	resolveTo: Cn,
	tokenize: wn
}, xn = {
	partial: !0,
	tokenize: En
}, Sn = {
	partial: !0,
	tokenize: Tn
};
function Cn(e) {
	let t = e.length;
	for (; t-- && (e[t][0] !== "enter" || e[t][1].type !== "htmlFlow"););
	return t > 1 && e[t - 2][1].type === "linePrefix" && (e[t][1].start = e[t - 2][1].start, e[t + 1][1].start = e[t - 2][1].start, e.splice(t - 2, 2)), e;
}
function wn(e, t, n) {
	let r = this, i, a, o, s, c;
	return l;
	function l(e) {
		return u(e);
	}
	function u(t) {
		return e.enter("htmlFlow"), e.enter("htmlFlowData"), e.consume(t), d;
	}
	function d(s) {
		return s === 33 ? (e.consume(s), f) : s === 47 ? (e.consume(s), a = !0, _) : s === 63 ? (e.consume(s), i = 3, r.interrupt ? t : z) : h(s) ? (e.consume(s), o = String.fromCharCode(s), v) : n(s);
	}
	function f(a) {
		return a === 45 ? (e.consume(a), i = 2, p) : a === 91 ? (e.consume(a), i = 5, s = 0, m) : h(a) ? (e.consume(a), i = 4, r.interrupt ? t : z) : n(a);
	}
	function p(i) {
		return i === 45 ? (e.consume(i), r.interrupt ? t : z) : n(i);
	}
	function m(i) {
		return i === "CDATA[".charCodeAt(s++) ? (e.consume(i), s === 6 ? r.interrupt ? t : N : m) : n(i);
	}
	function _(t) {
		return h(t) ? (e.consume(t), o = String.fromCharCode(t), v) : n(t);
	}
	function v(s) {
		if (s === null || s === 47 || s === 62 || C(s)) {
			let c = s === 47, l = o.toLowerCase();
			return !c && !a && yn.includes(l) ? (i = 1, r.interrupt ? t(s) : N(s)) : vn.includes(o.toLowerCase()) ? (i = 6, c ? (e.consume(s), y) : r.interrupt ? t(s) : N(s)) : (i = 7, r.interrupt && !r.parser.lazy[r.now().line] ? n(s) : a ? b(s) : x(s));
		}
		return s === 45 || g(s) ? (e.consume(s), o += String.fromCharCode(s), v) : n(s);
	}
	function y(i) {
		return i === 62 ? (e.consume(i), r.interrupt ? t : N) : n(i);
	}
	function b(t) {
		return w(t) ? (e.consume(t), b) : j(t);
	}
	function x(t) {
		return t === 47 ? (e.consume(t), j) : t === 58 || t === 95 || h(t) ? (e.consume(t), T) : w(t) ? (e.consume(t), x) : j(t);
	}
	function T(t) {
		return t === 45 || t === 46 || t === 58 || t === 95 || g(t) ? (e.consume(t), T) : E(t);
	}
	function E(t) {
		return t === 61 ? (e.consume(t), D) : w(t) ? (e.consume(t), E) : x(t);
	}
	function D(t) {
		return t === null || t === 60 || t === 61 || t === 62 || t === 96 ? n(t) : t === 34 || t === 39 ? (e.consume(t), c = t, O) : w(t) ? (e.consume(t), D) : k(t);
	}
	function O(t) {
		return t === c ? (e.consume(t), c = null, A) : t === null || S(t) ? n(t) : (e.consume(t), O);
	}
	function k(t) {
		return t === null || t === 34 || t === 39 || t === 47 || t === 60 || t === 61 || t === 62 || t === 96 || C(t) ? E(t) : (e.consume(t), k);
	}
	function A(e) {
		return e === 47 || e === 62 || w(e) ? x(e) : n(e);
	}
	function j(t) {
		return t === 62 ? (e.consume(t), M) : n(t);
	}
	function M(t) {
		return t === null || S(t) ? N(t) : w(t) ? (e.consume(t), M) : n(t);
	}
	function N(t) {
		return t === 45 && i === 2 ? (e.consume(t), I) : t === 60 && i === 1 ? (e.consume(t), L) : t === 62 && i === 4 ? (e.consume(t), B) : t === 63 && i === 3 ? (e.consume(t), z) : t === 93 && i === 5 ? (e.consume(t), R) : S(t) && (i === 6 || i === 7) ? (e.exit("htmlFlowData"), e.check(xn, V, P)(t)) : t === null || S(t) ? (e.exit("htmlFlowData"), P(t)) : (e.consume(t), N);
	}
	function P(t) {
		return e.check(Sn, F, V)(t);
	}
	function F(t) {
		return e.enter("lineEnding"), e.consume(t), e.exit("lineEnding"), ee;
	}
	function ee(t) {
		return t === null || S(t) ? P(t) : (e.enter("htmlFlowData"), N(t));
	}
	function I(t) {
		return t === 45 ? (e.consume(t), z) : N(t);
	}
	function L(t) {
		return t === 47 ? (e.consume(t), o = "", te) : N(t);
	}
	function te(t) {
		if (t === 62) {
			let n = o.toLowerCase();
			return yn.includes(n) ? (e.consume(t), B) : N(t);
		}
		return h(t) && o.length < 8 ? (e.consume(t), o += String.fromCharCode(t), te) : N(t);
	}
	function R(t) {
		return t === 93 ? (e.consume(t), z) : N(t);
	}
	function z(t) {
		return t === 62 ? (e.consume(t), B) : t === 45 && i === 2 ? (e.consume(t), z) : N(t);
	}
	function B(t) {
		return t === null || S(t) ? (e.exit("htmlFlowData"), V(t)) : (e.consume(t), B);
	}
	function V(n) {
		return e.exit("htmlFlow"), t(n);
	}
}
function Tn(e, t, n) {
	let r = this;
	return i;
	function i(t) {
		return S(t) ? (e.enter("lineEnding"), e.consume(t), e.exit("lineEnding"), a) : n(t);
	}
	function a(e) {
		return r.parser.lazy[r.now().line] ? n(e) : t(e);
	}
}
function En(e, t, n) {
	return r;
	function r(r) {
		return e.enter("lineEnding"), e.consume(r), e.exit("lineEnding"), e.attempt(Ot, t, n);
	}
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/micromark-core-commonmark/2.0.3/a1b4dd720a345ea7669bd6eac1b98d97da85d9a8fd797edbe457df3db914e918/node_modules/micromark-core-commonmark/lib/html-text.js
var Dn = {
	name: "htmlText",
	tokenize: On
};
function On(e, t, n) {
	let r = this, i, a, o;
	return s;
	function s(t) {
		return e.enter("htmlText"), e.enter("htmlTextData"), e.consume(t), c;
	}
	function c(t) {
		return t === 33 ? (e.consume(t), l) : t === 47 ? (e.consume(t), E) : t === 63 ? (e.consume(t), x) : h(t) ? (e.consume(t), k) : n(t);
	}
	function l(t) {
		return t === 45 ? (e.consume(t), u) : t === 91 ? (e.consume(t), a = 0, m) : h(t) ? (e.consume(t), b) : n(t);
	}
	function u(t) {
		return t === 45 ? (e.consume(t), p) : n(t);
	}
	function d(t) {
		return t === null ? n(t) : t === 45 ? (e.consume(t), f) : S(t) ? (o = d, L(t)) : (e.consume(t), d);
	}
	function f(t) {
		return t === 45 ? (e.consume(t), p) : d(t);
	}
	function p(e) {
		return e === 62 ? I(e) : e === 45 ? f(e) : d(e);
	}
	function m(t) {
		return t === "CDATA[".charCodeAt(a++) ? (e.consume(t), a === 6 ? _ : m) : n(t);
	}
	function _(t) {
		return t === null ? n(t) : t === 93 ? (e.consume(t), v) : S(t) ? (o = _, L(t)) : (e.consume(t), _);
	}
	function v(t) {
		return t === 93 ? (e.consume(t), y) : _(t);
	}
	function y(t) {
		return t === 62 ? I(t) : t === 93 ? (e.consume(t), y) : _(t);
	}
	function b(t) {
		return t === null || t === 62 ? I(t) : S(t) ? (o = b, L(t)) : (e.consume(t), b);
	}
	function x(t) {
		return t === null ? n(t) : t === 63 ? (e.consume(t), T) : S(t) ? (o = x, L(t)) : (e.consume(t), x);
	}
	function T(e) {
		return e === 62 ? I(e) : x(e);
	}
	function E(t) {
		return h(t) ? (e.consume(t), D) : n(t);
	}
	function D(t) {
		return t === 45 || g(t) ? (e.consume(t), D) : O(t);
	}
	function O(t) {
		return S(t) ? (o = O, L(t)) : w(t) ? (e.consume(t), O) : I(t);
	}
	function k(t) {
		return t === 45 || g(t) ? (e.consume(t), k) : t === 47 || t === 62 || C(t) ? A(t) : n(t);
	}
	function A(t) {
		return t === 47 ? (e.consume(t), I) : t === 58 || t === 95 || h(t) ? (e.consume(t), j) : S(t) ? (o = A, L(t)) : w(t) ? (e.consume(t), A) : I(t);
	}
	function j(t) {
		return t === 45 || t === 46 || t === 58 || t === 95 || g(t) ? (e.consume(t), j) : M(t);
	}
	function M(t) {
		return t === 61 ? (e.consume(t), N) : S(t) ? (o = M, L(t)) : w(t) ? (e.consume(t), M) : A(t);
	}
	function N(t) {
		return t === null || t === 60 || t === 61 || t === 62 || t === 96 ? n(t) : t === 34 || t === 39 ? (e.consume(t), i = t, P) : S(t) ? (o = N, L(t)) : w(t) ? (e.consume(t), N) : (e.consume(t), F);
	}
	function P(t) {
		return t === i ? (e.consume(t), i = void 0, ee) : t === null ? n(t) : S(t) ? (o = P, L(t)) : (e.consume(t), P);
	}
	function F(t) {
		return t === null || t === 34 || t === 39 || t === 60 || t === 61 || t === 96 ? n(t) : t === 47 || t === 62 || C(t) ? A(t) : (e.consume(t), F);
	}
	function ee(e) {
		return e === 47 || e === 62 || C(e) ? A(e) : n(e);
	}
	function I(r) {
		return r === 62 ? (e.consume(r), e.exit("htmlTextData"), e.exit("htmlText"), t) : n(r);
	}
	function L(t) {
		return e.exit("htmlTextData"), e.enter("lineEnding"), e.consume(t), e.exit("lineEnding"), te;
	}
	function te(t) {
		return w(t) ? q(e, R, "linePrefix", r.parser.constructs.disable.null.includes("codeIndented") ? void 0 : 4)(t) : R(t);
	}
	function R(t) {
		return e.enter("htmlTextData"), o(t);
	}
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/micromark-core-commonmark/2.0.3/a1b4dd720a345ea7669bd6eac1b98d97da85d9a8fd797edbe457df3db914e918/node_modules/micromark-core-commonmark/lib/label-end.js
var kn = {
	name: "labelEnd",
	resolveAll: Nn,
	resolveTo: Pn,
	tokenize: Fn
}, An = { tokenize: In }, jn = { tokenize: Ln }, Mn = { tokenize: Rn };
function Nn(e) {
	let t = -1, n = [];
	for (; ++t < e.length;) {
		let r = e[t][1];
		if (n.push(e[t]), r.type === "labelImage" || r.type === "labelLink" || r.type === "labelEnd") {
			let e = r.type === "labelImage" ? 4 : 2;
			r.type = "data", t += e;
		}
	}
	return e.length !== n.length && G(e, 0, e.length, n), e;
}
function Pn(e, t) {
	let n = e.length, r = 0, i, a, o, s;
	for (; n--;) if (i = e[n][1], a) {
		if (i.type === "link" || i.type === "labelLink" && i._inactive) break;
		e[n][0] === "enter" && i.type === "labelLink" && (i._inactive = !0);
	} else if (o) {
		if (e[n][0] === "enter" && (i.type === "labelImage" || i.type === "labelLink") && !i._balanced && (a = n, i.type !== "labelLink")) {
			r = 2;
			break;
		}
	} else i.type === "labelEnd" && (o = n);
	let c = {
		type: e[a][1].type === "labelLink" ? "link" : "image",
		start: { ...e[a][1].start },
		end: { ...e[e.length - 1][1].end }
	}, l = {
		type: "label",
		start: { ...e[a][1].start },
		end: { ...e[o][1].end }
	}, u = {
		type: "labelText",
		start: { ...e[a + r + 2][1].end },
		end: { ...e[o - 2][1].start }
	};
	return s = [[
		"enter",
		c,
		t
	], [
		"enter",
		l,
		t
	]], s = K(s, e.slice(a + 1, a + r + 3)), s = K(s, [[
		"enter",
		u,
		t
	]]), s = K(s, xt(t.parser.constructs.insideSpan.null, e.slice(a + r + 4, o - 3), t)), s = K(s, [
		[
			"exit",
			u,
			t
		],
		e[o - 2],
		e[o - 1],
		[
			"exit",
			l,
			t
		]
	]), s = K(s, e.slice(o + 1)), s = K(s, [[
		"exit",
		c,
		t
	]]), G(e, a, e.length, s), e;
}
function Fn(e, t, n) {
	let r = this, i = r.events.length, a, o;
	for (; i--;) if ((r.events[i][1].type === "labelImage" || r.events[i][1].type === "labelLink") && !r.events[i][1]._balanced) {
		a = r.events[i][1];
		break;
	}
	return s;
	function s(t) {
		return a ? a._inactive ? d(t) : (o = r.parser.defined.includes(k(r.sliceSerialize({
			start: a.end,
			end: r.now()
		}))), e.enter("labelEnd"), e.enter("labelMarker"), e.consume(t), e.exit("labelMarker"), e.exit("labelEnd"), c) : n(t);
	}
	function c(t) {
		return t === 40 ? e.attempt(An, u, o ? u : d)(t) : t === 91 ? e.attempt(jn, u, o ? l : d)(t) : o ? u(t) : d(t);
	}
	function l(t) {
		return e.attempt(Mn, u, d)(t);
	}
	function u(e) {
		return t(e);
	}
	function d(e) {
		return a._balanced = !0, n(e);
	}
}
function In(e, t, n) {
	return r;
	function r(t) {
		return e.enter("resource"), e.enter("resourceMarker"), e.consume(t), e.exit("resourceMarker"), i;
	}
	function i(t) {
		return C(t) ? J(e, a)(t) : a(t);
	}
	function a(t) {
		return t === 41 ? u(t) : on(e, o, s, "resourceDestination", "resourceDestinationLiteral", "resourceDestinationLiteralMarker", "resourceDestinationRaw", "resourceDestinationString", 32)(t);
	}
	function o(t) {
		return C(t) ? J(e, c)(t) : u(t);
	}
	function s(e) {
		return n(e);
	}
	function c(t) {
		return t === 34 || t === 39 || t === 40 ? cn(e, l, n, "resourceTitle", "resourceTitleMarker", "resourceTitleString")(t) : u(t);
	}
	function l(t) {
		return C(t) ? J(e, u)(t) : u(t);
	}
	function u(r) {
		return r === 41 ? (e.enter("resourceMarker"), e.consume(r), e.exit("resourceMarker"), e.exit("resource"), t) : n(r);
	}
}
function Ln(e, t, n) {
	let r = this;
	return i;
	function i(t) {
		return sn.call(r, e, a, o, "reference", "referenceMarker", "referenceString")(t);
	}
	function a(e) {
		return r.parser.defined.includes(k(r.sliceSerialize(r.events[r.events.length - 1][1]).slice(1, -1))) ? t(e) : n(e);
	}
	function o(e) {
		return n(e);
	}
}
function Rn(e, t, n) {
	return r;
	function r(t) {
		return e.enter("reference"), e.enter("referenceMarker"), e.consume(t), e.exit("referenceMarker"), i;
	}
	function i(r) {
		return r === 93 ? (e.enter("referenceMarker"), e.consume(r), e.exit("referenceMarker"), e.exit("reference"), t) : n(r);
	}
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/micromark-core-commonmark/2.0.3/a1b4dd720a345ea7669bd6eac1b98d97da85d9a8fd797edbe457df3db914e918/node_modules/micromark-core-commonmark/lib/label-start-image.js
var zn = {
	name: "labelStartImage",
	resolveAll: kn.resolveAll,
	tokenize: Bn
};
function Bn(e, t, n) {
	let r = this;
	return i;
	function i(t) {
		return e.enter("labelImage"), e.enter("labelImageMarker"), e.consume(t), e.exit("labelImageMarker"), a;
	}
	function a(t) {
		return t === 91 ? (e.enter("labelMarker"), e.consume(t), e.exit("labelMarker"), e.exit("labelImage"), o) : n(t);
	}
	function o(e) {
		/* c8 ignore next 3 */
		return e === 94 && "_hiddenFootnoteSupport" in r.parser.constructs ? n(e) : t(e);
	}
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/micromark-core-commonmark/2.0.3/a1b4dd720a345ea7669bd6eac1b98d97da85d9a8fd797edbe457df3db914e918/node_modules/micromark-core-commonmark/lib/label-start-link.js
var Vn = {
	name: "labelStartLink",
	resolveAll: kn.resolveAll,
	tokenize: Hn
};
function Hn(e, t, n) {
	let r = this;
	return i;
	function i(t) {
		return e.enter("labelLink"), e.enter("labelMarker"), e.consume(t), e.exit("labelMarker"), e.exit("labelLink"), a;
	}
	function a(e) {
		/* c8 ignore next 3 */
		return e === 94 && "_hiddenFootnoteSupport" in r.parser.constructs ? n(e) : t(e);
	}
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/micromark-core-commonmark/2.0.3/a1b4dd720a345ea7669bd6eac1b98d97da85d9a8fd797edbe457df3db914e918/node_modules/micromark-core-commonmark/lib/line-ending.js
var Un = {
	name: "lineEnding",
	tokenize: Wn
};
function Wn(e, t) {
	return n;
	function n(n) {
		return e.enter("lineEnding"), e.consume(n), e.exit("lineEnding"), q(e, t, "linePrefix");
	}
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/micromark-core-commonmark/2.0.3/a1b4dd720a345ea7669bd6eac1b98d97da85d9a8fd797edbe457df3db914e918/node_modules/micromark-core-commonmark/lib/thematic-break.js
var Gn = {
	name: "thematicBreak",
	tokenize: Kn
};
function Kn(e, t, n) {
	let r = 0, i;
	return a;
	function a(t) {
		return e.enter("thematicBreak"), o(t);
	}
	function o(e) {
		return i = e, s(e);
	}
	function s(a) {
		return a === i ? (e.enter("thematicBreakSequence"), c(a)) : r >= 3 && (a === null || S(a)) ? (e.exit("thematicBreak"), t(a)) : n(a);
	}
	function c(t) {
		return t === i ? (e.consume(t), r++, c) : (e.exit("thematicBreakSequence"), w(t) ? q(e, s, "whitespace")(t) : s(t));
	}
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/micromark-core-commonmark/2.0.3/a1b4dd720a345ea7669bd6eac1b98d97da85d9a8fd797edbe457df3db914e918/node_modules/micromark-core-commonmark/lib/list.js
var Y = {
	continuation: { tokenize: Xn },
	exit: Qn,
	name: "list",
	tokenize: Yn
}, qn = {
	partial: !0,
	tokenize: $n
}, Jn = {
	partial: !0,
	tokenize: Zn
};
function Yn(e, t, n) {
	let r = this, i = r.events[r.events.length - 1], a = i && i[1].type === "linePrefix" ? i[2].sliceSerialize(i[1], !0).length : 0, o = 0;
	return s;
	function s(t) {
		let i = r.containerState.type || (t === 42 || t === 43 || t === 45 ? "listUnordered" : "listOrdered");
		if (i === "listUnordered" ? !r.containerState.marker || t === r.containerState.marker : y(t)) {
			if (r.containerState.type || (r.containerState.type = i, e.enter(i, { _container: !0 })), i === "listUnordered") return e.enter("listItemPrefix"), t === 42 || t === 45 ? e.check(Gn, n, l)(t) : l(t);
			if (!r.interrupt || t === 49) return e.enter("listItemPrefix"), e.enter("listItemValue"), c(t);
		}
		return n(t);
	}
	function c(t) {
		return y(t) && ++o < 10 ? (e.consume(t), c) : (!r.interrupt || o < 2) && (r.containerState.marker ? t === r.containerState.marker : t === 41 || t === 46) ? (e.exit("listItemValue"), l(t)) : n(t);
	}
	function l(t) {
		return e.enter("listItemMarker"), e.consume(t), e.exit("listItemMarker"), r.containerState.marker = r.containerState.marker || t, e.check(Ot, r.interrupt ? n : u, e.attempt(qn, f, d));
	}
	function u(e) {
		return r.containerState.initialBlankLine = !0, a++, f(e);
	}
	function d(t) {
		return w(t) ? (e.enter("listItemPrefixWhitespace"), e.consume(t), e.exit("listItemPrefixWhitespace"), f) : n(t);
	}
	function f(n) {
		return r.containerState.size = a + r.sliceSerialize(e.exit("listItemPrefix"), !0).length, t(n);
	}
}
function Xn(e, t, n) {
	let r = this;
	return r.containerState._closeFlow = void 0, e.check(Ot, i, a);
	function i(n) {
		return r.containerState.furtherBlankLines = r.containerState.furtherBlankLines || r.containerState.initialBlankLine, q(e, t, "listItemIndent", r.containerState.size + 1)(n);
	}
	function a(n) {
		return r.containerState.furtherBlankLines || !w(n) ? (r.containerState.furtherBlankLines = void 0, r.containerState.initialBlankLine = void 0, o(n)) : (r.containerState.furtherBlankLines = void 0, r.containerState.initialBlankLine = void 0, e.attempt(Jn, t, o)(n));
	}
	function o(i) {
		return r.containerState._closeFlow = !0, r.interrupt = void 0, q(e, e.attempt(Y, t, n), "linePrefix", r.parser.constructs.disable.null.includes("codeIndented") ? void 0 : 4)(i);
	}
}
function Zn(e, t, n) {
	let r = this;
	return q(e, i, "listItemIndent", r.containerState.size + 1);
	function i(e) {
		let i = r.events[r.events.length - 1];
		return i && i[1].type === "listItemIndent" && i[2].sliceSerialize(i[1], !0).length === r.containerState.size ? t(e) : n(e);
	}
}
function Qn(e) {
	e.exit(this.containerState.type);
}
function $n(e, t, n) {
	let r = this;
	return q(e, i, "listItemPrefixWhitespace", r.parser.constructs.disable.null.includes("codeIndented") ? void 0 : 5);
	function i(e) {
		let i = r.events[r.events.length - 1];
		return !w(e) && i && i[1].type === "listItemPrefixWhitespace" ? t(e) : n(e);
	}
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/micromark-core-commonmark/2.0.3/a1b4dd720a345ea7669bd6eac1b98d97da85d9a8fd797edbe457df3db914e918/node_modules/micromark-core-commonmark/lib/setext-underline.js
var er = {
	name: "setextUnderline",
	resolveTo: tr,
	tokenize: nr
};
function tr(e, t) {
	let n = e.length, r, i, a;
	for (; n--;) if (e[n][0] === "enter") {
		if (e[n][1].type === "content") {
			r = n;
			break;
		}
		e[n][1].type === "paragraph" && (i = n);
	} else e[n][1].type === "content" && e.splice(n, 1), !a && e[n][1].type === "definition" && (a = n);
	let o = {
		type: "setextHeading",
		start: { ...e[r][1].start },
		end: { ...e[e.length - 1][1].end }
	};
	return e[i][1].type = "setextHeadingText", a ? (e.splice(i, 0, [
		"enter",
		o,
		t
	]), e.splice(a + 1, 0, [
		"exit",
		e[r][1],
		t
	]), e[r][1].end = { ...e[a][1].end }) : e[r][1] = o, e.push([
		"exit",
		o,
		t
	]), e;
}
function nr(e, t, n) {
	let r = this, i;
	return a;
	function a(t) {
		let a = r.events.length, s;
		for (; a--;) if (r.events[a][1].type !== "lineEnding" && r.events[a][1].type !== "linePrefix" && r.events[a][1].type !== "content") {
			s = r.events[a][1].type === "paragraph";
			break;
		}
		return !r.parser.lazy[r.now().line] && (r.interrupt || s) ? (e.enter("setextHeadingLine"), i = t, o(t)) : n(t);
	}
	function o(t) {
		return e.enter("setextHeadingLineSequence"), s(t);
	}
	function s(t) {
		return t === i ? (e.consume(t), s) : (e.exit("setextHeadingLineSequence"), w(t) ? q(e, c, "lineSuffix")(t) : c(t));
	}
	function c(r) {
		return r === null || S(r) ? (e.exit("setextHeadingLine"), t(r)) : n(r);
	}
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/micromark-extension-gfm-strikethrough/2.1.0/58caef4d27cbd6958d93419dbb686f215399e2f36c59a802c5af5264698caa17/node_modules/micromark-extension-gfm-strikethrough/lib/syntax.js
function rr(e) {
	let t = (e || {}).singleTilde, n = {
		name: "strikethrough",
		tokenize: i,
		resolveAll: r
	};
	return t ??= !0, {
		text: { 126: n },
		insideSpan: { null: [n] },
		attentionMarkers: { null: [126] }
	};
	function r(e, t) {
		let n = -1;
		for (; ++n < e.length;) if (e[n][0] === "enter" && e[n][1].type === "strikethroughSequenceTemporary" && e[n][1]._close) {
			let r = n;
			for (; r--;) if (e[r][0] === "exit" && e[r][1].type === "strikethroughSequenceTemporary" && e[r][1]._open && e[n][1].end.offset - e[n][1].start.offset === e[r][1].end.offset - e[r][1].start.offset) {
				e[n][1].type = "strikethroughSequence", e[r][1].type = "strikethroughSequence";
				let i = {
					type: "strikethrough",
					start: Object.assign({}, e[r][1].start),
					end: Object.assign({}, e[n][1].end)
				}, a = {
					type: "strikethroughText",
					start: Object.assign({}, e[r][1].end),
					end: Object.assign({}, e[n][1].start)
				}, o = [
					[
						"enter",
						i,
						t
					],
					[
						"enter",
						e[r][1],
						t
					],
					[
						"exit",
						e[r][1],
						t
					],
					[
						"enter",
						a,
						t
					]
				], s = t.parser.constructs.insideSpan.null;
				s && G(o, o.length, 0, xt(s, e.slice(r + 1, n), t)), G(o, o.length, 0, [
					[
						"exit",
						a,
						t
					],
					[
						"enter",
						e[n][1],
						t
					],
					[
						"exit",
						e[n][1],
						t
					],
					[
						"exit",
						i,
						t
					]
				]), G(e, r - 1, n - r + 3, o), n = r + o.length - 2;
				break;
			}
		}
		for (n = -1; ++n < e.length;) e[n][1].type === "strikethroughSequenceTemporary" && (e[n][1].type = "data");
		return e;
	}
	function i(e, n, r) {
		let i = this.previous, a = this.events, o = 0;
		return s;
		function s(t) {
			return i === 126 && a[a.length - 1][1].type !== "characterEscape" ? r(t) : (e.enter("strikethroughSequenceTemporary"), c(t));
		}
		function c(a) {
			let s = U(i);
			if (a === 126) return o > 1 ? r(a) : (e.consume(a), o++, c);
			if (o < 2 && !t) return r(a);
			let l = e.exit("strikethroughSequenceTemporary"), u = U(a);
			return l._open = !u || u === 2 && !!s, l._close = !s || s === 2 && !!u, n(a);
		}
	}
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/micromark-extension-gfm-table/2.1.1/878c34944f5a913c0570ff1b00a0685e46c9408f5a55d50e5bb512bbbc8966d6/node_modules/micromark-extension-gfm-table/lib/edit-map.js
var ir = class {
	constructor() {
		this.map = [];
	}
	add(e, t, n) {
		ar(this, e, t, n);
	}
	consume(e) {
		/* c8 ignore next 3 -- `resolve` is never called without tables, so without edits. */
		if (this.map.sort(function(e, t) {
			return e[0] - t[0];
		}), this.map.length === 0) return;
		let t = this.map.length, n = [];
		for (; t > 0;) --t, n.push(e.slice(this.map[t][0] + this.map[t][1]), this.map[t][2]), e.length = this.map[t][0];
		n.push(e.slice()), e.length = 0;
		let r = n.pop();
		for (; r;) {
			for (let t of r) e.push(t);
			r = n.pop();
		}
		this.map.length = 0;
	}
};
function ar(e, t, n, r) {
	let i = 0;
	/* c8 ignore next 3 -- `resolve` is never called without tables, so without edits. */
	if (n !== 0 || r.length !== 0) {
		for (; i < e.map.length;) {
			if (e.map[i][0] === t) {
				e.map[i][1] += n, e.map[i][2].push(...r);
				return;
			}
			i += 1;
		}
		e.map.push([
			t,
			n,
			r
		]);
	}
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/micromark-extension-gfm-table/2.1.1/878c34944f5a913c0570ff1b00a0685e46c9408f5a55d50e5bb512bbbc8966d6/node_modules/micromark-extension-gfm-table/lib/infer.js
function or(e, t) {
	let n = !1, r = [];
	for (; t < e.length;) {
		let i = e[t];
		if (n) {
			if (i[0] === "enter") i[1].type === "tableContent" && r.push(e[t + 1][1].type === "tableDelimiterMarker" ? "left" : "none");
			else if (i[1].type === "tableContent") {
				if (e[t - 1][1].type === "tableDelimiterMarker") {
					let e = r.length - 1;
					r[e] = r[e] === "left" ? "center" : "right";
				}
			} else if (i[1].type === "tableDelimiterRow") break;
		} else i[0] === "enter" && i[1].type === "tableDelimiterRow" && (n = !0);
		t += 1;
	}
	return r;
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/micromark-extension-gfm-table/2.1.1/878c34944f5a913c0570ff1b00a0685e46c9408f5a55d50e5bb512bbbc8966d6/node_modules/micromark-extension-gfm-table/lib/syntax.js
function sr() {
	return { flow: { null: {
		name: "table",
		tokenize: cr,
		resolveAll: lr
	} } };
}
function cr(e, t, n) {
	let r = this, i = 0, a = 0, o;
	return s;
	function s(e) {
		let t = r.events.length - 1;
		for (; t > -1;) {
			let e = r.events[t][1].type;
			if (e === "lineEnding" || e === "linePrefix") t--;
			else break;
		}
		let i = t > -1 ? r.events[t][1].type : null, a = i === "tableHead" || i === "tableRow" ? T : c;
		return a === T && r.parser.lazy[r.now().line] ? n(e) : a(e);
	}
	function c(t) {
		return e.enter("tableHead"), e.enter("tableRow"), l(t);
	}
	function l(e) {
		return e === 124 || (o = !0, a += 1), u(e);
	}
	function u(t) {
		return t === null ? n(t) : S(t) ? a > 1 ? (a = 0, r.interrupt = !0, e.exit("tableRow"), e.enter("lineEnding"), e.consume(t), e.exit("lineEnding"), p) : n(t) : w(t) ? q(e, u, "whitespace")(t) : (a += 1, o && (o = !1, i += 1), t === 124 ? (e.enter("tableCellDivider"), e.consume(t), e.exit("tableCellDivider"), o = !0, u) : (e.enter("data"), d(t)));
	}
	function d(t) {
		return t === null || t === 124 || C(t) ? (e.exit("data"), u(t)) : (e.consume(t), t === 92 ? f : d);
	}
	function f(t) {
		return t === 92 || t === 124 ? (e.consume(t), d) : d(t);
	}
	function p(t) {
		return r.interrupt = !1, r.parser.lazy[r.now().line] ? n(t) : (e.enter("tableDelimiterRow"), o = !1, w(t) ? q(e, m, "linePrefix", r.parser.constructs.disable.null.includes("codeIndented") ? void 0 : 4)(t) : m(t));
	}
	function m(t) {
		return t === 45 || t === 58 ? g(t) : t === 124 ? (o = !0, e.enter("tableCellDivider"), e.consume(t), e.exit("tableCellDivider"), h) : x(t);
	}
	function h(t) {
		return w(t) ? q(e, g, "whitespace")(t) : g(t);
	}
	function g(t) {
		return t === 58 ? (a += 1, o = !0, e.enter("tableDelimiterMarker"), e.consume(t), e.exit("tableDelimiterMarker"), _) : t === 45 ? (a += 1, _(t)) : t === null || S(t) ? b(t) : x(t);
	}
	function _(t) {
		return t === 45 ? (e.enter("tableDelimiterFiller"), v(t)) : x(t);
	}
	function v(t) {
		return t === 45 ? (e.consume(t), v) : t === 58 ? (o = !0, e.exit("tableDelimiterFiller"), e.enter("tableDelimiterMarker"), e.consume(t), e.exit("tableDelimiterMarker"), y) : (e.exit("tableDelimiterFiller"), y(t));
	}
	function y(t) {
		return w(t) ? q(e, b, "whitespace")(t) : b(t);
	}
	function b(n) {
		return n === 124 ? m(n) : n === null || S(n) ? !o || i !== a ? x(n) : (e.exit("tableDelimiterRow"), e.exit("tableHead"), t(n)) : x(n);
	}
	function x(e) {
		return n(e);
	}
	function T(t) {
		return e.enter("tableRow"), E(t);
	}
	function E(n) {
		return n === 124 ? (e.enter("tableCellDivider"), e.consume(n), e.exit("tableCellDivider"), E) : n === null || S(n) ? (e.exit("tableRow"), t(n)) : w(n) ? q(e, E, "whitespace")(n) : (e.enter("data"), D(n));
	}
	function D(t) {
		return t === null || t === 124 || C(t) ? (e.exit("data"), E(t)) : (e.consume(t), t === 92 ? O : D);
	}
	function O(t) {
		return t === 92 || t === 124 ? (e.consume(t), D) : D(t);
	}
}
function lr(e, t) {
	let n = -1, r = !0, i = 0, a = [
		0,
		0,
		0,
		0
	], o = [
		0,
		0,
		0,
		0
	], s = !1, c = 0, l, u, d, f = new ir();
	for (; ++n < e.length;) {
		let p = e[n], m = p[1];
		p[0] === "enter" ? m.type === "tableHead" ? (s = !1, c !== 0 && (dr(f, t, c, l, u), u = void 0, c = 0), l = {
			type: "table",
			start: Object.assign({}, m.start),
			end: Object.assign({}, m.end)
		}, f.add(n, 0, [[
			"enter",
			l,
			t
		]])) : m.type === "tableRow" || m.type === "tableDelimiterRow" ? (r = !0, d = void 0, a = [
			0,
			0,
			0,
			0
		], o = [
			0,
			n + 1,
			0,
			0
		], s && (s = !1, u = {
			type: "tableBody",
			start: Object.assign({}, m.start),
			end: Object.assign({}, m.end)
		}, f.add(n, 0, [[
			"enter",
			u,
			t
		]])), i = m.type === "tableDelimiterRow" ? 2 : u ? 3 : 1) : i && (m.type === "data" || m.type === "tableDelimiterMarker" || m.type === "tableDelimiterFiller") ? (r = !1, o[2] === 0 && (a[1] !== 0 && (o[0] = o[1], d = ur(f, t, a, i, void 0, d), a = [
			0,
			0,
			0,
			0
		]), o[2] = n)) : m.type === "tableCellDivider" && (r ? r = !1 : (a[1] !== 0 && (o[0] = o[1], d = ur(f, t, a, i, void 0, d)), a = o, o = [
			a[1],
			n,
			0,
			0
		])) : m.type === "tableHead" ? (s = !0, c = n) : m.type === "tableRow" || m.type === "tableDelimiterRow" ? (c = n, a[1] === 0 ? o[1] !== 0 && (d = ur(f, t, o, i, n, d)) : (o[0] = o[1], d = ur(f, t, a, i, n, d)), i = 0) : i && (m.type === "data" || m.type === "tableDelimiterMarker" || m.type === "tableDelimiterFiller") && (o[3] = n);
	}
	for (c !== 0 && dr(f, t, c, l, u), f.consume(t.events), n = -1; ++n < t.events.length;) {
		let e = t.events[n];
		e[0] === "enter" && e[1].type === "table" && (e[1]._align = or(t.events, n));
	}
	return e;
}
function ur(e, t, n, r, i, a) {
	let o = r === 1 ? "tableHeader" : r === 2 ? "tableDelimiter" : "tableData";
	n[0] !== 0 && (a.end = Object.assign({}, X(t.events, n[0])), e.add(n[0], 0, [[
		"exit",
		a,
		t
	]]));
	let s = X(t.events, n[1]);
	if (a = {
		type: o,
		start: Object.assign({}, s),
		end: Object.assign({}, s)
	}, e.add(n[1], 0, [[
		"enter",
		a,
		t
	]]), n[2] !== 0) {
		let i = X(t.events, n[2]), a = X(t.events, n[3]), o = {
			type: "tableContent",
			start: Object.assign({}, i),
			end: Object.assign({}, a)
		};
		if (e.add(n[2], 0, [[
			"enter",
			o,
			t
		]]), r !== 2) {
			let r = t.events[n[2]], i = t.events[n[3]];
			if (r[1].end = Object.assign({}, i[1].end), r[1].type = "chunkText", r[1].contentType = "text", n[3] > n[2] + 1) {
				let t = n[2] + 1, r = n[3] - n[2] - 1;
				e.add(t, r, []);
			}
		}
		e.add(n[3] + 1, 0, [[
			"exit",
			o,
			t
		]]);
	}
	return i !== void 0 && (a.end = Object.assign({}, X(t.events, i)), e.add(i, 0, [[
		"exit",
		a,
		t
	]]), a = void 0), a;
}
function dr(e, t, n, r, i) {
	let a = [], o = X(t.events, n);
	i && (i.end = Object.assign({}, o), a.push([
		"exit",
		i,
		t
	])), r.end = Object.assign({}, o), a.push([
		"exit",
		r,
		t
	]), e.add(n + 1, 0, a);
}
function X(e, t) {
	let n = e[t], r = n[0] === "enter" ? "start" : "end";
	return n[1][r];
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/micromark-extension-gfm-task-list-item/2.1.0/4dd3a60d240c34d70a4bf59dbce767bd31373a117900c3d5b732b02138780950/node_modules/micromark-extension-gfm-task-list-item/lib/syntax.js
var fr = {
	name: "tasklistCheck",
	tokenize: mr
};
function pr() {
	return { text: { 91: fr } };
}
function mr(e, t, n) {
	let r = this;
	return i;
	function i(t) {
		return r.previous !== null || !r._gfmTasklistFirstContentOfListItem ? n(t) : (e.enter("taskListCheck"), e.enter("taskListCheckMarker"), e.consume(t), e.exit("taskListCheckMarker"), a);
	}
	function a(t) {
		return C(t) ? (e.enter("taskListCheckValueUnchecked"), e.consume(t), e.exit("taskListCheckValueUnchecked"), o) : t === 88 || t === 120 ? (e.enter("taskListCheckValueChecked"), e.consume(t), e.exit("taskListCheckValueChecked"), o) : n(t);
	}
	function o(t) {
		return t === 93 ? (e.enter("taskListCheckMarker"), e.consume(t), e.exit("taskListCheckMarker"), e.exit("taskListCheck"), s) : n(t);
	}
	function s(r) {
		return S(r) ? t(r) : w(r) ? e.check({ tokenize: hr }, t, n)(r) : n(r);
	}
}
function hr(e, t, n) {
	return q(e, r, "whitespace");
	function r(e) {
		return e === null ? n(e) : t(e);
	}
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/estree-util-is-identifier-name/3.0.0/2f63975b60cb39f7f0b629ab8cb41eac581112b2a65bb77adc24beee5b4ee23a/node_modules/estree-util-is-identifier-name/lib/index.js
var gr = /[$_\p{ID_Start}]/u, _r = /[$_\u{200C}\u{200D}\p{ID_Continue}]/u, vr = /[-$_\u{200C}\u{200D}\p{ID_Continue}]/u, yr = /^[$_\p{ID_Start}][$_\u{200C}\u{200D}\p{ID_Continue}]*$/u, br = /^[$_\p{ID_Start}][-$_\u{200C}\u{200D}\p{ID_Continue}]*$/u, xr = {};
function Sr(e) {
	return e ? gr.test(String.fromCodePoint(e)) : !1;
}
function Cr(e, t) {
	let n = (t || xr).jsx ? vr : _r;
	return e ? n.test(String.fromCodePoint(e)) : !1;
}
function wr(e, t) {
	return ((t || xr).jsx ? br : yr).test(e);
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/unist-util-stringify-position/4.0.0/f40519889373bb6fe3243d645f519d458de197c110c7dcf685d0ba277eec75c5/node_modules/unist-util-stringify-position/lib/index.js
function Z(e) {
	return !e || typeof e != "object" ? "" : "position" in e || "type" in e ? Er(e.position) : "start" in e || "end" in e ? Er(e) : "line" in e || "column" in e ? Tr(e) : "";
}
function Tr(e) {
	return Dr(e && e.line) + ":" + Dr(e && e.column);
}
function Er(e) {
	return Tr(e && e.start) + "-" + Tr(e && e.end);
}
function Dr(e) {
	return e && typeof e == "number" ? e : 1;
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/vfile-message/4.0.3/f806703989e3b8d2b44edd6d79953d7492cd7bed236aca86ac3785f2cd530ad9/node_modules/vfile-message/lib/index.js
var Q = class extends Error {
	constructor(e, t, n) {
		super(), typeof t == "string" && (n = t, t = void 0);
		let r = "", i = {}, a = !1;
		if (t && (i = "line" in t && "column" in t || "start" in t && "end" in t ? { place: t } : "type" in t ? {
			ancestors: [t],
			place: t.position
		} : { ...t }), typeof e == "string" ? r = e : !i.cause && e && (a = !0, r = e.message, i.cause = e), !i.ruleId && !i.source && typeof n == "string") {
			let e = n.indexOf(":");
			e === -1 ? i.ruleId = n : (i.source = n.slice(0, e), i.ruleId = n.slice(e + 1));
		}
		if (!i.place && i.ancestors && i.ancestors) {
			let e = i.ancestors[i.ancestors.length - 1];
			e && (i.place = e.position);
		}
		let o = i.place && "start" in i.place ? i.place.start : i.place;
		this.ancestors = i.ancestors || void 0, this.cause = i.cause || void 0, this.column = o ? o.column : void 0, this.fatal = void 0, this.file = "", this.message = r, this.line = o ? o.line : void 0, this.name = Z(i.place) || "1:1", this.place = i.place || void 0, this.reason = this.message, this.ruleId = i.ruleId || void 0, this.source = i.source || void 0, this.stack = a && i.cause && typeof i.cause.stack == "string" ? i.cause.stack : "", this.actual = void 0, this.expected = void 0, this.note = void 0, this.url = void 0;
	}
};
Q.prototype.file = "", Q.prototype.name = "", Q.prototype.reason = "", Q.prototype.message = "", Q.prototype.stack = "", Q.prototype.column = void 0, Q.prototype.line = void 0, Q.prototype.ancestors = void 0, Q.prototype.cause = void 0, Q.prototype.fatal = void 0, Q.prototype.place = void 0, Q.prototype.ruleId = void 0, Q.prototype.source = void 0;
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/micromark/4.0.2/33f446d065a1a40872b9e101893829fcd9a593224746530399c4e443f7fb881f/node_modules/micromark/lib/initialize/content.js
var Or = { tokenize: kr };
function kr(e) {
	let t = e.attempt(this.parser.constructs.contentInitial, r, i), n;
	return t;
	function r(n) {
		if (n === null) e.consume(n);
		else return e.enter("lineEnding"), e.consume(n), e.exit("lineEnding"), q(e, t, "linePrefix");
	}
	function i(t) {
		return e.enter("paragraph"), a(t);
	}
	function a(t) {
		let r = e.enter("chunkText", {
			contentType: "text",
			previous: n
		});
		return n && (n.next = r), n = r, o(t);
	}
	function o(t) {
		if (t === null) e.exit("chunkText"), e.exit("paragraph"), e.consume(t);
		else return S(t) ? (e.consume(t), e.exit("chunkText"), a) : (e.consume(t), o);
	}
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/micromark/4.0.2/33f446d065a1a40872b9e101893829fcd9a593224746530399c4e443f7fb881f/node_modules/micromark/lib/initialize/document.js
var Ar = { tokenize: Mr }, jr = { tokenize: Nr };
function Mr(e) {
	let t = this, n = [], r = 0, i, a, o;
	return s;
	function s(i) {
		if (r < n.length) {
			let a = n[r];
			return t.containerState = a[1], e.attempt(a[0].continuation, c, l)(i);
		}
		return l(i);
	}
	function c(e) {
		if (r++, t.containerState._closeFlow) {
			t.containerState._closeFlow = void 0, i && v();
			let n = t.events.length, a = n, o;
			for (; a--;) if (t.events[a][0] === "exit" && t.events[a][1].type === "chunkFlow") {
				o = t.events[a][1].end;
				break;
			}
			_(r);
			let s = n;
			for (; s < t.events.length;) t.events[s][1].end = { ...o }, s++;
			return G(t.events, a + 1, 0, t.events.slice(n)), t.events.length = s, l(e);
		}
		return s(e);
	}
	function l(a) {
		if (r === n.length) {
			if (!i) return f(a);
			if (i.currentConstruct && i.currentConstruct.concrete) return m(a);
			t.interrupt = !(!i.currentConstruct || i._gfmTableDynamicInterruptHack);
		}
		return t.containerState = {}, e.check(jr, u, d)(a);
	}
	function u(e) {
		return i && v(), _(r), f(e);
	}
	function d(e) {
		return t.parser.lazy[t.now().line] = r !== n.length, o = t.now().offset, m(e);
	}
	function f(n) {
		return t.containerState = {}, e.attempt(jr, p, m)(n);
	}
	function p(e) {
		return r++, n.push([t.currentConstruct, t.containerState]), f(e);
	}
	function m(n) {
		if (n === null) i && v(), _(0), e.consume(n);
		else return i ||= t.parser.flow(t.now()), e.enter("chunkFlow", {
			_tokenizer: i,
			contentType: "flow",
			previous: a
		}), h(n);
	}
	function h(n) {
		if (n === null) g(e.exit("chunkFlow"), !0), _(0), e.consume(n);
		else return S(n) ? (e.consume(n), g(e.exit("chunkFlow")), r = 0, t.interrupt = void 0, s) : (e.consume(n), h);
	}
	function g(e, n) {
		let s = t.sliceStream(e);
		if (n && s.push(null), e.previous = a, a && (a.next = e), a = e, i.defineSkip(e.start), i.write(s), t.parser.lazy[e.start.line]) {
			let e = i.events.length;
			for (; e--;) if (i.events[e][1].start.offset < o && (!i.events[e][1].end || i.events[e][1].end.offset > o)) return;
			let n = t.events.length, a = n, s, c;
			for (; a--;) if (t.events[a][0] === "exit" && t.events[a][1].type === "chunkFlow") {
				if (s) {
					c = t.events[a][1].end;
					break;
				}
				s = !0;
			}
			for (_(r), e = n; e < t.events.length;) t.events[e][1].end = { ...c }, e++;
			G(t.events, a + 1, 0, t.events.slice(n)), t.events.length = e;
		}
	}
	function _(r) {
		let i = n.length;
		for (; i-- > r;) {
			let r = n[i];
			t.containerState = r[1], r[0].exit.call(t, e);
		}
		n.length = r;
	}
	function v() {
		i.write([null]), a = void 0, i = void 0, t.containerState._closeFlow = void 0;
	}
}
function Nr(e, t, n) {
	return q(e, e.attempt(this.parser.constructs.document, t, n), "linePrefix", this.parser.constructs.disable.null.includes("codeIndented") ? void 0 : 4);
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/micromark/4.0.2/33f446d065a1a40872b9e101893829fcd9a593224746530399c4e443f7fb881f/node_modules/micromark/lib/initialize/flow.js
var Pr = { tokenize: Fr };
function Fr(e) {
	let t = this, n = e.attempt(Ot, r, e.attempt(this.parser.constructs.flowInitial, i, q(e, e.attempt(this.parser.constructs.flow, i, e.attempt(en, i)), "linePrefix")));
	return n;
	function r(r) {
		if (r === null) e.consume(r);
		else return e.enter("lineEndingBlank"), e.consume(r), e.exit("lineEndingBlank"), t.currentConstruct = void 0, n;
	}
	function i(r) {
		if (r === null) e.consume(r);
		else return e.enter("lineEnding"), e.consume(r), e.exit("lineEnding"), t.currentConstruct = void 0, n;
	}
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/micromark/4.0.2/33f446d065a1a40872b9e101893829fcd9a593224746530399c4e443f7fb881f/node_modules/micromark/lib/initialize/text.js
var Ir = { resolveAll: Br() }, Lr = zr("string"), Rr = zr("text");
function zr(e) {
	return {
		resolveAll: Br(e === "text" ? Vr : void 0),
		tokenize: t
	};
	function t(t) {
		let n = this, r = this.parser.constructs[e], i = t.attempt(r, a, o);
		return a;
		function a(e) {
			return c(e) ? i(e) : o(e);
		}
		function o(e) {
			if (e === null) t.consume(e);
			else return t.enter("data"), t.consume(e), s;
		}
		function s(e) {
			return c(e) ? (t.exit("data"), i(e)) : (t.consume(e), s);
		}
		function c(e) {
			if (e === null) return !0;
			let t = r[e], i = -1;
			if (t) for (; ++i < t.length;) {
				let e = t[i];
				if (!e.previous || e.previous.call(n, n.previous)) return !0;
			}
			return !1;
		}
	}
}
function Br(e) {
	return t;
	function t(t, n) {
		let r = -1, i;
		for (; ++r <= t.length;) i === void 0 ? t[r] && t[r][1].type === "data" && (i = r, r++) : t[r] && t[r][1].type === "data" || (r !== i + 2 && (t[i][1].end = t[r - 1][1].end, t.splice(i + 2, r - i - 2), r = i + 2), i = void 0);
		return e ? e(t, n) : t;
	}
}
function Vr(e, t) {
	let n = 0;
	for (; ++n <= e.length;) if ((n === e.length || e[n][1].type === "lineEnding") && e[n - 1][1].type === "data") {
		let r = e[n - 1][1], i = t.sliceStream(r), a = i.length, o = -1, s = 0, c;
		for (; a--;) {
			let e = i[a];
			if (typeof e == "string") {
				for (o = e.length; e.charCodeAt(o - 1) === 32;) s++, o--;
				if (o) break;
				o = -1;
			} else if (e === -2) c = !0, s++;
			else if (e !== -1) {
				a++;
				break;
			}
		}
		if (t._contentTypeTextTrailing && n === e.length && (s = 0), s) {
			let i = {
				type: n === e.length || c || s < 2 ? "lineSuffix" : "hardBreakTrailing",
				start: {
					_bufferIndex: a ? o : r.start._bufferIndex + o,
					_index: r.start._index + a,
					line: r.end.line,
					column: r.end.column - s,
					offset: r.end.offset - s
				},
				end: { ...r.end }
			};
			r.end = { ...i.start }, r.start.offset === r.end.offset ? Object.assign(r, i) : (e.splice(n, 0, [
				"enter",
				i,
				t
			], [
				"exit",
				i,
				t
			]), n += 2);
		}
		n++;
	}
	return e;
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/micromark/4.0.2/33f446d065a1a40872b9e101893829fcd9a593224746530399c4e443f7fb881f/node_modules/micromark/lib/constructs.js
var Hr = /* @__PURE__ */ e({
	attentionMarkers: () => Xr,
	contentInitial: () => Wr,
	disable: () => Zr,
	document: () => Ur,
	flow: () => Kr,
	flowInitial: () => Gr,
	insideSpan: () => Yr,
	string: () => qr,
	text: () => Jr
}), Ur = {
	42: Y,
	43: Y,
	45: Y,
	48: Y,
	49: Y,
	50: Y,
	51: Y,
	52: Y,
	53: Y,
	54: Y,
	55: Y,
	56: Y,
	57: Y,
	62: At
}, Wr = { 91: ln }, Gr = {
	[-2]: Ht,
	[-1]: Ht,
	32: Ht
}, Kr = {
	35: hn,
	42: Gn,
	45: [er, Gn],
	60: bn,
	61: er,
	95: Gn,
	96: zt,
	126: zt
}, qr = {
	38: It,
	92: Pt
}, Jr = {
	[-5]: Un,
	[-4]: Un,
	[-3]: Un,
	33: zn,
	38: It,
	42: St,
	60: [Et, Dn],
	91: Vn,
	92: [pn, Pt],
	93: kn,
	95: St,
	96: Kt
}, Yr = { null: [St, Ir] }, Xr = { null: [42, 95] }, Zr = { null: [] };
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/micromark/4.0.2/33f446d065a1a40872b9e101893829fcd9a593224746530399c4e443f7fb881f/node_modules/micromark/lib/create-tokenizer.js
function Qr(e, t, n) {
	let r = {
		_bufferIndex: -1,
		_index: 0,
		line: n && n.line || 1,
		column: n && n.column || 1,
		offset: n && n.offset || 0
	}, i = {}, a = [], o = [], s = [], c = {
		attempt: w(x),
		check: w(C),
		consume: v,
		enter: y,
		exit: b,
		interrupt: w(C, { interrupt: !0 })
	}, l = {
		code: null,
		containerState: {},
		defineSkip: h,
		events: [],
		now: m,
		parser: e,
		previous: null,
		sliceSerialize: f,
		sliceStream: p,
		write: d
	}, u = t.tokenize.call(l, c);
	return t.resolveAll && a.push(t), l;
	function d(e) {
		return o = K(o, e), g(), o[o.length - 1] === null ? (T(t, 0), l.events = xt(a, l.events, l), l.events) : [];
	}
	function f(e, t) {
		return ei(p(e), t);
	}
	function p(e) {
		return $r(o, e);
	}
	function m() {
		let { _bufferIndex: e, _index: t, line: n, column: i, offset: a } = r;
		return {
			_bufferIndex: e,
			_index: t,
			line: n,
			column: i,
			offset: a
		};
	}
	function h(e) {
		i[e.line] = e.column, D();
	}
	function g() {
		let e;
		for (; r._index < o.length;) {
			let t = o[r._index];
			if (typeof t == "string") for (e = r._index, r._bufferIndex < 0 && (r._bufferIndex = 0); r._index === e && r._bufferIndex < t.length;) _(t.charCodeAt(r._bufferIndex));
			else _(t);
		}
	}
	function _(e) {
		u = u(e);
	}
	function v(e) {
		S(e) ? (r.line++, r.column = 1, r.offset += e === -3 ? 2 : 1, D()) : e !== -1 && (r.column++, r.offset++), r._bufferIndex < 0 ? r._index++ : (r._bufferIndex++, r._bufferIndex === o[r._index].length && (r._bufferIndex = -1, r._index++)), l.previous = e;
	}
	function y(e, t) {
		let n = t || {};
		return n.type = e, n.start = m(), l.events.push([
			"enter",
			n,
			l
		]), s.push(n), n;
	}
	function b(e) {
		let t = s.pop();
		return t.end = m(), l.events.push([
			"exit",
			t,
			l
		]), t;
	}
	function x(e, t) {
		T(e, t.from);
	}
	function C(e, t) {
		t.restore();
	}
	function w(e, t) {
		return n;
		function n(n, r, i) {
			let a, o, s, u;
			return Array.isArray(n) ? f(n) : "tokenize" in n ? f([n]) : d(n);
			function d(e) {
				return t;
				function t(t) {
					let n = t !== null && e[t], r = t !== null && e.null;
					return f([...Array.isArray(n) ? n : n ? [n] : [], ...Array.isArray(r) ? r : r ? [r] : []])(t);
				}
			}
			function f(e) {
				return a = e, o = 0, e.length === 0 ? i : p(e[o]);
			}
			function p(e) {
				return n;
				function n(n) {
					return u = E(), s = e, e.partial || (l.currentConstruct = e), e.name && l.parser.constructs.disable.null.includes(e.name) ? h(n) : e.tokenize.call(t ? Object.assign(Object.create(l), t) : l, c, m, h)(n);
				}
			}
			function m(t) {
				return e(s, u), r;
			}
			function h(e) {
				return u.restore(), ++o < a.length ? p(a[o]) : i;
			}
		}
	}
	function T(e, t) {
		e.resolveAll && !a.includes(e) && a.push(e), e.resolve && G(l.events, t, l.events.length - t, e.resolve(l.events.slice(t), l)), e.resolveTo && (l.events = e.resolveTo(l.events, l));
	}
	function E() {
		let e = m(), t = l.previous, n = l.currentConstruct, i = l.events.length, a = Array.from(s);
		return {
			from: i,
			restore: o
		};
		function o() {
			r = e, l.previous = t, l.currentConstruct = n, l.events.length = i, s = a, D();
		}
	}
	function D() {
		r.line in i && r.column < 2 && (r.column = i[r.line], r.offset += i[r.line] - 1);
	}
}
function $r(e, t) {
	let n = t.start._index, r = t.start._bufferIndex, i = t.end._index, a = t.end._bufferIndex, o;
	if (n === i) o = [e[n].slice(r, a)];
	else {
		if (o = e.slice(n, i), r > -1) {
			let e = o[0];
			typeof e == "string" ? o[0] = e.slice(r) : o.shift();
		}
		a > 0 && o.push(e[i].slice(0, a));
	}
	return o;
}
function ei(e, t) {
	let n = -1, r = [], i;
	for (; ++n < e.length;) {
		let a = e[n], o;
		if (typeof a == "string") o = a;
		else switch (a) {
			case -5:
				o = "\r";
				break;
			case -4:
				o = "\n";
				break;
			case -3:
				o = "\r\n";
				break;
			case -2:
				o = t ? " " : "	";
				break;
			case -1:
				if (!t && i) continue;
				o = " ";
				break;
			default: o = String.fromCharCode(a);
		}
		i = a === -2, r.push(o);
	}
	return r.join("");
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/micromark/4.0.2/33f446d065a1a40872b9e101893829fcd9a593224746530399c4e443f7fb881f/node_modules/micromark/lib/parse.js
function ti(e) {
	let t = {
		constructs: vt([Hr, ...(e || {}).extensions || []]),
		content: n(Or),
		defined: [],
		document: n(Ar),
		flow: n(Pr),
		lazy: {},
		string: n(Lr),
		text: n(Rr)
	};
	return t;
	function n(e) {
		return n;
		function n(n) {
			return Qr(t, e, n);
		}
	}
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/micromark/4.0.2/33f446d065a1a40872b9e101893829fcd9a593224746530399c4e443f7fb881f/node_modules/micromark/lib/postprocess.js
function ni(e) {
	for (; !Qt(e););
	return e;
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/micromark/4.0.2/33f446d065a1a40872b9e101893829fcd9a593224746530399c4e443f7fb881f/node_modules/micromark/lib/preprocess.js
var ri = /[\0\t\n\r]/g;
function ii() {
	let e = 1, t = "", n = !0, r;
	return i;
	function i(i, a, o) {
		let s = [], c, l, u, d, f;
		for (i = t + (typeof i == "string" ? i.toString() : new TextDecoder(a || void 0).decode(i)), u = 0, t = "", n &&= (i.charCodeAt(0) === 65279 && u++, void 0); u < i.length;) {
			if (ri.lastIndex = u, c = ri.exec(i), d = c && c.index !== void 0 ? c.index : i.length, f = i.charCodeAt(d), !c) {
				t = i.slice(u);
				break;
			}
			if (f === 10 && u === d && r) s.push(-3), r = void 0;
			else switch (r &&= (s.push(-5), void 0), u < d && (s.push(i.slice(u, d)), e += d - u), f) {
				case 0:
					s.push(65533), e++;
					break;
				case 9:
					for (l = Math.ceil(e / 4) * 4, s.push(-2); e++ < l;) s.push(-1);
					break;
				case 10:
					s.push(-4), e = 1;
					break;
				default: r = !0, e = 1;
			}
			u = d + 1;
		}
		return o && (r && s.push(-5), t && s.push(t), s.push(null)), s;
	}
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/mdast-util-from-markdown/2.0.3/7ddac1f2652fe32048a52b2ba515080918dff1daf152dfbe84babc8cd329295c/node_modules/mdast-util-from-markdown/lib/index.js
var ai = {}.hasOwnProperty;
function oi(e, t, n) {
	return t && typeof t == "object" && (n = t, t = void 0), si(n)(ni(ti(n).document().write(ii()(e, t, !0))));
}
function si(e) {
	let t = {
		transforms: [],
		canContainEols: [
			"emphasis",
			"fragment",
			"heading",
			"paragraph",
			"strong"
		],
		enter: {
			autolink: a(W),
			autolinkProtocol: T,
			autolinkEmail: T,
			atxHeading: a(ue),
			blockQuote: a(ae),
			characterEscape: T,
			characterReference: T,
			codeFenced: a(oe),
			codeFencedFenceInfo: o,
			codeFencedFenceMeta: o,
			codeIndented: a(oe, o),
			codeText: a(se, o),
			codeTextData: T,
			data: T,
			codeFlowValue: T,
			definition: a(ce),
			definitionDestinationString: o,
			definitionLabelString: o,
			definitionTitleString: o,
			emphasis: a(le),
			hardBreakEscape: a(de),
			hardBreakTrailing: a(de),
			htmlFlow: a(H, o),
			htmlFlowData: T,
			htmlText: a(H, o),
			htmlTextData: T,
			image: a(U),
			label: o,
			link: a(W),
			listItem: a(pe),
			listItemValue: f,
			listOrdered: a(fe, d),
			listUnordered: a(fe),
			paragraph: a(me),
			reference: R,
			referenceString: o,
			resourceDestinationString: o,
			resourceTitleString: o,
			setextHeading: a(ue),
			strong: a(ge),
			thematicBreak: a(ve)
		},
		exit: {
			atxHeading: c(),
			atxHeadingSequence: x,
			autolink: c(),
			autolinkEmail: ie,
			autolinkProtocol: re,
			blockQuote: c(),
			characterEscapeValue: E,
			characterReferenceMarkerHexadecimal: B,
			characterReferenceMarkerNumeric: B,
			characterReferenceValue: V,
			characterReference: ne,
			codeFenced: c(g),
			codeFencedFence: h,
			codeFencedFenceInfo: p,
			codeFencedFenceMeta: m,
			codeFlowValue: E,
			codeIndented: c(_),
			codeText: c(M),
			codeTextData: E,
			data: E,
			definition: c(),
			definitionDestinationString: b,
			definitionLabelString: v,
			definitionTitleString: y,
			emphasis: c(),
			hardBreakEscape: c(O),
			hardBreakTrailing: c(O),
			htmlFlow: c(A),
			htmlFlowData: E,
			htmlText: c(j),
			htmlTextData: E,
			image: c(P),
			label: ee,
			labelText: F,
			lineEnding: D,
			link: c(N),
			listItem: c(),
			listOrdered: c(),
			listUnordered: c(),
			paragraph: c(),
			referenceString: z,
			resourceDestinationString: I,
			resourceTitleString: L,
			resource: te,
			setextHeading: c(w),
			setextHeadingLineSequence: C,
			setextHeadingText: S,
			strong: c(),
			thematicBreak: c()
		}
	};
	ci(t, (e || {}).mdastExtensions || []);
	let n = {};
	return r;
	function r(e) {
		let r = {
			type: "root",
			children: []
		}, a = {
			stack: [r],
			tokenStack: [],
			config: t,
			enter: s,
			exit: l,
			buffer: o,
			resume: u,
			data: n
		}, c = [], d = -1;
		for (; ++d < e.length;) (e[d][1].type === "listOrdered" || e[d][1].type === "listUnordered") && (e[d][0] === "enter" ? c.push(d) : d = i(e, c.pop(), d));
		for (d = -1; ++d < e.length;) {
			let n = t[e[d][0]];
			ai.call(n, e[d][1].type) && n[e[d][1].type].call(Object.assign({ sliceSerialize: e[d][2].sliceSerialize }, a), e[d][1]);
		}
		if (a.tokenStack.length > 0) {
			let e = a.tokenStack[a.tokenStack.length - 1];
			(e[1] || ui).call(a, void 0, e[0]);
		}
		for (r.position = {
			start: $(e.length > 0 ? e[0][1].start : {
				line: 1,
				column: 1,
				offset: 0
			}),
			end: $(e.length > 0 ? e[e.length - 2][1].end : {
				line: 1,
				column: 1,
				offset: 0
			})
		}, d = -1; ++d < t.transforms.length;) r = t.transforms[d](r) || r;
		return r;
	}
	function i(e, t, n) {
		let r = t - 1, i = -1, a = !1, o, s, c, l;
		for (; ++r <= n;) {
			let t = e[r];
			switch (t[1].type) {
				case "listUnordered":
				case "listOrdered":
				case "blockQuote":
					t[0] === "enter" ? i++ : i--, l = void 0;
					break;
				case "lineEndingBlank":
					t[0] === "enter" && (o && !l && !i && !c && (c = r), l = void 0);
					break;
				case "linePrefix":
				case "listItemValue":
				case "listItemMarker":
				case "listItemPrefix":
				case "listItemPrefixWhitespace": break;
				default: l = void 0;
			}
			if (!i && t[0] === "enter" && t[1].type === "listItemPrefix" || i === -1 && t[0] === "exit" && (t[1].type === "listUnordered" || t[1].type === "listOrdered")) {
				if (o) {
					let i = r;
					for (s = void 0; i--;) {
						let t = e[i];
						if (t[1].type === "lineEnding" || t[1].type === "lineEndingBlank") {
							if (t[0] === "exit") continue;
							s && (e[s][1].type = "lineEndingBlank", a = !0), t[1].type = "lineEnding", s = i;
						} else if (t[1].type !== "linePrefix" && t[1].type !== "blockQuotePrefix" && t[1].type !== "blockQuotePrefixWhitespace" && t[1].type !== "blockQuoteMarker" && t[1].type !== "listItemIndent") break;
					}
					c && (!s || c < s) && (o._spread = !0), o.end = Object.assign({}, s ? e[s][1].start : t[1].end), e.splice(s || r, 0, [
						"exit",
						o,
						t[2]
					]), r++, n++;
				}
				if (t[1].type === "listItemPrefix") {
					let i = {
						type: "listItem",
						_spread: !1,
						start: Object.assign({}, t[1].start),
						end: void 0
					};
					o = i, e.splice(r, 0, [
						"enter",
						i,
						t[2]
					]), r++, n++, c = void 0, l = !0;
				}
			}
		}
		return e[t][1]._spread = a, n;
	}
	function a(e, t) {
		return n;
		function n(n) {
			s.call(this, e(n), n), t && t.call(this, n);
		}
	}
	function o() {
		this.stack.push({
			type: "fragment",
			children: []
		});
	}
	function s(e, t, n) {
		this.stack[this.stack.length - 1].children.push(e), this.stack.push(e), this.tokenStack.push([t, n || void 0]), e.position = {
			start: $(t.start),
			end: void 0
		};
	}
	function c(e) {
		return t;
		function t(t) {
			e && e.call(this, t), l.call(this, t);
		}
	}
	function l(e, t) {
		let n = this.stack.pop(), r = this.tokenStack.pop();
		if (r) r[0].type !== e.type && (t ? t.call(this, e, r[0]) : (r[1] || ui).call(this, e, r[0]));
		else throw Error("Cannot close `" + e.type + "` (" + Z({
			start: e.start,
			end: e.end
		}) + "): it’s not open");
		n.position.end = $(e.end);
	}
	function u() {
		return he(this.stack.pop());
	}
	function d() {
		this.data.expectingFirstListItemValue = !0;
	}
	function f(e) {
		if (this.data.expectingFirstListItemValue) {
			let t = this.stack[this.stack.length - 2];
			t.start = Number.parseInt(this.sliceSerialize(e), 10), this.data.expectingFirstListItemValue = void 0;
		}
	}
	function p() {
		let e = this.resume(), t = this.stack[this.stack.length - 1];
		t.lang = e;
	}
	function m() {
		let e = this.resume(), t = this.stack[this.stack.length - 1];
		t.meta = e;
	}
	function h() {
		this.data.flowCodeInside || (this.buffer(), this.data.flowCodeInside = !0);
	}
	function g() {
		let e = this.resume(), t = this.stack[this.stack.length - 1];
		t.value = e.replace(/^(\r?\n|\r)|(\r?\n|\r)$/g, ""), this.data.flowCodeInside = void 0;
	}
	function _() {
		let e = this.resume(), t = this.stack[this.stack.length - 1];
		t.value = e.replace(/(\r?\n|\r)$/g, "");
	}
	function v(e) {
		let t = this.resume(), n = this.stack[this.stack.length - 1];
		n.label = t, n.identifier = k(this.sliceSerialize(e)).toLowerCase();
	}
	function y() {
		let e = this.resume(), t = this.stack[this.stack.length - 1];
		t.title = e;
	}
	function b() {
		let e = this.resume(), t = this.stack[this.stack.length - 1];
		t.url = e;
	}
	function x(e) {
		let t = this.stack[this.stack.length - 1];
		t.depth ||= this.sliceSerialize(e).length;
	}
	function S() {
		this.data.setextHeadingSlurpLineEnding = !0;
	}
	function C(e) {
		let t = this.stack[this.stack.length - 1];
		t.depth = this.sliceSerialize(e).codePointAt(0) === 61 ? 1 : 2;
	}
	function w() {
		this.data.setextHeadingSlurpLineEnding = void 0;
	}
	function T(e) {
		let t = this.stack[this.stack.length - 1].children, n = t[t.length - 1];
		n && n.type === "text" || (n = _e(), n.position = {
			start: $(e.start),
			end: void 0
		}, t.push(n)), this.stack.push(n);
	}
	function E(e) {
		let t = this.stack.pop();
		t.value += this.sliceSerialize(e), t.position.end = $(e.end);
	}
	function D(e) {
		let n = this.stack[this.stack.length - 1];
		if (this.data.atHardBreak) {
			let t = n.children[n.children.length - 1];
			t.position.end = $(e.end), this.data.atHardBreak = void 0;
		} else !this.data.setextHeadingSlurpLineEnding && t.canContainEols.includes(n.type) && (T.call(this, e), E.call(this, e));
	}
	function O() {
		this.data.atHardBreak = !0;
	}
	function A() {
		let e = this.resume(), t = this.stack[this.stack.length - 1];
		t.value = e;
	}
	function j() {
		let e = this.resume(), t = this.stack[this.stack.length - 1];
		t.value = e;
	}
	function M() {
		let e = this.resume(), t = this.stack[this.stack.length - 1];
		t.value = e;
	}
	function N() {
		let e = this.stack[this.stack.length - 1];
		if (this.data.inReference) {
			let t = this.data.referenceType || "shortcut";
			e.type += "Reference", e.referenceType = t, delete e.url, delete e.title;
		} else delete e.identifier, delete e.label;
		this.data.referenceType = void 0;
	}
	function P() {
		let e = this.stack[this.stack.length - 1];
		if (this.data.inReference) {
			let t = this.data.referenceType || "shortcut";
			e.type += "Reference", e.referenceType = t, delete e.url, delete e.title;
		} else delete e.identifier, delete e.label;
		this.data.referenceType = void 0;
	}
	function F(e) {
		let t = this.sliceSerialize(e), n = this.stack[this.stack.length - 2];
		n.label = tt(t), n.identifier = k(t).toLowerCase();
	}
	function ee() {
		let e = this.stack[this.stack.length - 1], t = this.resume(), n = this.stack[this.stack.length - 1];
		this.data.inReference = !0, n.type === "link" ? n.children = e.children : n.alt = t;
	}
	function I() {
		let e = this.resume(), t = this.stack[this.stack.length - 1];
		t.url = e;
	}
	function L() {
		let e = this.resume(), t = this.stack[this.stack.length - 1];
		t.title = e;
	}
	function te() {
		this.data.inReference = void 0;
	}
	function R() {
		this.data.referenceType = "collapsed";
	}
	function z(e) {
		let t = this.resume(), n = this.stack[this.stack.length - 1];
		n.label = t, n.identifier = k(this.sliceSerialize(e)).toLowerCase(), this.data.referenceType = "full";
	}
	function B(e) {
		this.data.characterReferenceType = e.type;
	}
	function V(e) {
		let t = this.sliceSerialize(e), n = this.data.characterReferenceType, r;
		n ? (r = $e(t, n === "characterReferenceMarkerNumeric" ? 10 : 16), this.data.characterReferenceType = void 0) : r = Qe(t);
		let i = this.stack[this.stack.length - 1];
		i.value += r;
	}
	function ne(e) {
		let t = this.stack.pop();
		t.position.end = $(e.end);
	}
	function re(e) {
		E.call(this, e);
		let t = this.stack[this.stack.length - 1];
		t.url = this.sliceSerialize(e);
	}
	function ie(e) {
		E.call(this, e);
		let t = this.stack[this.stack.length - 1];
		t.url = "mailto:" + this.sliceSerialize(e);
	}
	function ae() {
		return {
			type: "blockquote",
			children: []
		};
	}
	function oe() {
		return {
			type: "code",
			lang: null,
			meta: null,
			value: ""
		};
	}
	function se() {
		return {
			type: "inlineCode",
			value: ""
		};
	}
	function ce() {
		return {
			type: "definition",
			identifier: "",
			label: null,
			title: null,
			url: ""
		};
	}
	function le() {
		return {
			type: "emphasis",
			children: []
		};
	}
	function ue() {
		return {
			type: "heading",
			depth: 0,
			children: []
		};
	}
	function de() {
		return { type: "break" };
	}
	function H() {
		return {
			type: "html",
			value: ""
		};
	}
	function U() {
		return {
			type: "image",
			title: null,
			url: "",
			alt: null
		};
	}
	function W() {
		return {
			type: "link",
			title: null,
			url: "",
			children: []
		};
	}
	function fe(e) {
		return {
			type: "list",
			ordered: e.type === "listOrdered",
			start: null,
			spread: e._spread,
			children: []
		};
	}
	function pe(e) {
		return {
			type: "listItem",
			spread: e._spread,
			checked: null,
			children: []
		};
	}
	function me() {
		return {
			type: "paragraph",
			children: []
		};
	}
	function ge() {
		return {
			type: "strong",
			children: []
		};
	}
	function _e() {
		return {
			type: "text",
			value: ""
		};
	}
	function ve() {
		return { type: "thematicBreak" };
	}
}
function $(e) {
	return {
		line: e.line,
		column: e.column,
		offset: e.offset
	};
}
function ci(e, t) {
	let n = -1;
	for (; ++n < t.length;) {
		let r = t[n];
		Array.isArray(r) ? ci(e, r) : li(e, r);
	}
}
function li(e, t) {
	let n;
	for (n in t) if (ai.call(t, n)) switch (n) {
		case "canContainEols": {
			let r = t[n];
			r && e[n].push(...r);
			break;
		}
		case "transforms": {
			let r = t[n];
			r && e[n].push(...r);
			break;
		}
		case "enter":
		case "exit": {
			let r = t[n];
			r && Object.assign(e[n], r);
			break;
		}
	}
}
function ui(e, t) {
	throw Error(e ? "Cannot close `" + e.type + "` (" + Z({
		start: e.start,
		end: e.end
	}) + "): a different token (`" + t.type + "`, " + Z({
		start: t.start,
		end: t.end
	}) + ") is open" : "Cannot close document, a token (`" + t.type + "`, " + Z({
		start: t.start,
		end: t.end
	}) + ") is still open");
}
export { k as A, m as B, ye as C, V as D, ae as E, S as F, u as H, C as I, w as L, h as M, g as N, j as O, v as P, T as R, Xe as S, H as T, d as U, p as V, t as W, pt as _, wr as a, tt as b, sr as c, Ot as d, q as f, ft as g, G as h, Cr as i, O as j, M as k, rr as l, vt as m, Q as n, Sr as o, xt as p, Z as r, pr as s, oi as t, J as u, rt as v, U as w, Qe as x, dt as y, E as z };
