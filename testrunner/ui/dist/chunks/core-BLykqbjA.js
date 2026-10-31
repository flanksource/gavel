import { i as e, r as t, t as n } from "./dist-CMfY_7V8.js";
import { a as r, c as i, o as a, r as o, s, t as c, u as l } from "./lib-BeutSJOr.js";
import { n as u, t as d } from "./ccount-ClnF6NWF.js";
import { t as f } from "./lib-2bY02OrG.js";
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@shikijs/types/1.29.2/fa976e740007c0f2ff2f2d7b178001eeed9062b87a5a109e24f7e05ccdbf6731/node_modules/@shikijs/types/dist/index.mjs
var p = class extends Error {
	constructor(e) {
		super(e), this.name = "ShikiError";
	}
}, m = !1, h = !1;
function g(e = !0, t = !1) {
	m = e, h = t;
}
function _(e, t = 3) {
	if (m && !(typeof m == "number" && t > m)) {
		if (h) throw Error(`[SHIKI DEPRECATE]: ${e}`);
		console.trace(`[SHIKI DEPRECATE]: ${e}`);
	}
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@shikijs/vscode-textmate/10.0.2/f17adac2cc9ab9608c41c21074366fbd90a2bd3c3e8f0484114d3779a9657073/node_modules/@shikijs/vscode-textmate/dist/index.js
function ee(e) {
	return te(e);
}
function te(e) {
	return Array.isArray(e) ? ne(e) : e instanceof RegExp ? e : typeof e == "object" ? re(e) : e;
}
function ne(e) {
	let t = [];
	for (let n = 0, r = e.length; n < r; n++) t[n] = te(e[n]);
	return t;
}
function re(e) {
	let t = {};
	for (let n in e) t[n] = te(e[n]);
	return t;
}
function ie(e, ...t) {
	return t.forEach((t) => {
		for (let n in t) e[n] = t[n];
	}), e;
}
function ae(e) {
	let t = ~e.lastIndexOf("/") || ~e.lastIndexOf("\\");
	return t === 0 ? e : ~t === e.length - 1 ? ae(e.substring(0, e.length - 1)) : e.substr(~t + 1);
}
var oe = /\$(\d+)|\${(\d+):\/(downcase|upcase)}/g, se = class {
	static hasCaptures(e) {
		return e !== null && (oe.lastIndex = 0, oe.test(e));
	}
	static replaceCaptures(e, t, n) {
		return e.replace(oe, (e, r, i, a) => {
			let o = n[parseInt(r || i, 10)];
			if (o) {
				let e = t.substring(o.start, o.end);
				for (; e[0] === ".";) e = e.substring(1);
				switch (a) {
					case "downcase": return e.toLowerCase();
					case "upcase": return e.toUpperCase();
					default: return e;
				}
			} else return e;
		});
	}
};
function ce(e, t) {
	return e < t ? -1 : +(e > t);
}
function le(e, t) {
	if (e === null && t === null) return 0;
	if (!e) return -1;
	if (!t) return 1;
	let n = e.length, r = t.length;
	if (n === r) {
		for (let r = 0; r < n; r++) {
			let n = ce(e[r], t[r]);
			if (n !== 0) return n;
		}
		return 0;
	}
	return n - r;
}
function ue(e) {
	return !!(/^#[0-9a-f]{6}$/i.test(e) || /^#[0-9a-f]{8}$/i.test(e) || /^#[0-9a-f]{3}$/i.test(e) || /^#[0-9a-f]{4}$/i.test(e));
}
function de(e) {
	return e.replace(/[\-\\\{\}\*\+\?\|\^\$\.\,\[\]\(\)\#\s]/g, "\\$&");
}
var fe = class {
	constructor(e) {
		this.fn = e;
	}
	cache = /* @__PURE__ */ new Map();
	get(e) {
		if (this.cache.has(e)) return this.cache.get(e);
		let t = this.fn(e);
		return this.cache.set(e, t), t;
	}
}, pe = class {
	constructor(e, t, n) {
		this._colorMap = e, this._defaults = t, this._root = n;
	}
	static createFromRawTheme(e, t) {
		return this.createFromParsedTheme(ve(e), t);
	}
	static createFromParsedTheme(e, t) {
		return be(e, t);
	}
	_cachedMatchRoot = new fe((e) => this._root.match(e));
	getColorMap() {
		return this._colorMap.getColorMap();
	}
	getDefaults() {
		return this._defaults;
	}
	match(e) {
		if (e === null) return this._defaults;
		let t = e.scopeName, n = this._cachedMatchRoot.get(t).find((t) => he(e.parent, t.parentScopes));
		return n ? new _e(n.fontStyle, n.foreground, n.background) : null;
	}
}, me = class e {
	constructor(e, t) {
		this.parent = e, this.scopeName = t;
	}
	static push(t, n) {
		for (let r of n) t = new e(t, r);
		return t;
	}
	static from(...t) {
		let n = null;
		for (let r = 0; r < t.length; r++) n = new e(n, t[r]);
		return n;
	}
	push(t) {
		return new e(this, t);
	}
	getSegments() {
		let e = this, t = [];
		for (; e;) t.push(e.scopeName), e = e.parent;
		return t.reverse(), t;
	}
	toString() {
		return this.getSegments().join(" ");
	}
	extends(e) {
		return this === e || this.parent !== null && this.parent.extends(e);
	}
	getExtensionIfDefined(e) {
		let t = [], n = this;
		for (; n && n !== e;) t.push(n.scopeName), n = n.parent;
		return n === e ? t.reverse() : void 0;
	}
};
function he(e, t) {
	if (t.length === 0) return !0;
	for (let n = 0; n < t.length; n++) {
		let r = t[n], i = !1;
		if (r === ">") {
			if (n === t.length - 1) return !1;
			r = t[++n], i = !0;
		}
		for (; e && !ge(e.scopeName, r);) {
			if (i) return !1;
			e = e.parent;
		}
		if (!e) return !1;
		e = e.parent;
	}
	return !0;
}
function ge(e, t) {
	return t === e || e.startsWith(t) && e[t.length] === ".";
}
var _e = class {
	constructor(e, t, n) {
		this.fontStyle = e, this.foregroundId = t, this.backgroundId = n;
	}
};
function ve(e) {
	if (!e || !e.settings || !Array.isArray(e.settings)) return [];
	let t = e.settings, n = [], r = 0;
	for (let e = 0, i = t.length; e < i; e++) {
		let i = t[e];
		if (!i.settings) continue;
		let a;
		if (typeof i.scope == "string") {
			let e = i.scope;
			e = e.replace(/^[,]+/, ""), e = e.replace(/[,]+$/, ""), a = e.split(",");
		} else a = Array.isArray(i.scope) ? i.scope : [""];
		let o = -1;
		if (typeof i.settings.fontStyle == "string") {
			o = 0;
			let e = i.settings.fontStyle.split(" ");
			for (let t = 0, n = e.length; t < n; t++) switch (e[t]) {
				case "italic":
					o |= 1;
					break;
				case "bold":
					o |= 2;
					break;
				case "underline":
					o |= 4;
					break;
				case "strikethrough": o |= 8;
			}
		}
		let s = null;
		typeof i.settings.foreground == "string" && ue(i.settings.foreground) && (s = i.settings.foreground);
		let c = null;
		typeof i.settings.background == "string" && ue(i.settings.background) && (c = i.settings.background);
		for (let t = 0, i = a.length; t < i; t++) {
			let i = a[t].trim().split(" "), l = i[i.length - 1], u = null;
			i.length > 1 && (u = i.slice(0, i.length - 1), u.reverse()), n[r++] = new ye(l, u, e, o, s, c);
		}
	}
	return n;
}
var ye = class {
	constructor(e, t, n, r, i, a) {
		this.scope = e, this.parentScopes = t, this.index = n, this.fontStyle = r, this.foreground = i, this.background = a;
	}
}, v = /* @__PURE__ */ ((e) => (e[e.NotSet = -1] = "NotSet", e[e.None = 0] = "None", e[e.Italic = 1] = "Italic", e[e.Bold = 2] = "Bold", e[e.Underline = 4] = "Underline", e[e.Strikethrough = 8] = "Strikethrough", e))(v || {});
function be(e, t) {
	e.sort((e, t) => {
		let n = ce(e.scope, t.scope);
		return n !== 0 || (n = le(e.parentScopes, t.parentScopes), n !== 0) ? n : e.index - t.index;
	});
	let n = 0, r = "#000000", i = "#ffffff";
	for (; e.length >= 1 && e[0].scope === "";) {
		let t = e.shift();
		t.fontStyle !== -1 && (n = t.fontStyle), t.foreground !== null && (r = t.foreground), t.background !== null && (i = t.background);
	}
	let a = new xe(t), o = new _e(n, a.getId(r), a.getId(i)), s = new we(new Ce(0, null, -1, 0, 0), []);
	for (let t = 0, n = e.length; t < n; t++) {
		let n = e[t];
		s.insert(0, n.scope, n.parentScopes, n.fontStyle, a.getId(n.foreground), a.getId(n.background));
	}
	return new pe(a, o, s);
}
var xe = class {
	_isFrozen;
	_lastColorId;
	_id2color;
	_color2id;
	constructor(e) {
		if (this._lastColorId = 0, this._id2color = [], this._color2id = /* @__PURE__ */ Object.create(null), Array.isArray(e)) {
			this._isFrozen = !0;
			for (let t = 0, n = e.length; t < n; t++) this._color2id[e[t]] = t, this._id2color[t] = e[t];
		} else this._isFrozen = !1;
	}
	getId(e) {
		if (e === null) return 0;
		e = e.toUpperCase();
		let t = this._color2id[e];
		if (t) return t;
		if (this._isFrozen) throw Error(`Missing color in color map - ${e}`);
		return t = ++this._lastColorId, this._color2id[e] = t, this._id2color[t] = e, t;
	}
	getColorMap() {
		return this._id2color.slice(0);
	}
}, Se = Object.freeze([]), Ce = class e {
	scopeDepth;
	parentScopes;
	fontStyle;
	foreground;
	background;
	constructor(e, t, n, r, i) {
		this.scopeDepth = e, this.parentScopes = t || Se, this.fontStyle = n, this.foreground = r, this.background = i;
	}
	clone() {
		return new e(this.scopeDepth, this.parentScopes, this.fontStyle, this.foreground, this.background);
	}
	static cloneArr(e) {
		let t = [];
		for (let n = 0, r = e.length; n < r; n++) t[n] = e[n].clone();
		return t;
	}
	acceptOverwrite(e, t, n, r) {
		this.scopeDepth > e ? console.log("how did this happen?") : this.scopeDepth = e, t !== -1 && (this.fontStyle = t), n !== 0 && (this.foreground = n), r !== 0 && (this.background = r);
	}
}, we = class e {
	constructor(e, t = [], n = {}) {
		this._mainRule = e, this._children = n, this._rulesWithParentScopes = t;
	}
	_rulesWithParentScopes;
	static _cmpBySpecificity(e, t) {
		if (e.scopeDepth !== t.scopeDepth) return t.scopeDepth - e.scopeDepth;
		let n = 0, r = 0;
		for (; e.parentScopes[n] === ">" && n++, t.parentScopes[r] === ">" && r++, !(n >= e.parentScopes.length || r >= t.parentScopes.length);) {
			let i = t.parentScopes[r].length - e.parentScopes[n].length;
			if (i !== 0) return i;
			n++, r++;
		}
		return t.parentScopes.length - e.parentScopes.length;
	}
	match(t) {
		if (t !== "") {
			let e = t.indexOf("."), n, r;
			if (e === -1 ? (n = t, r = "") : (n = t.substring(0, e), r = t.substring(e + 1)), this._children.hasOwnProperty(n)) return this._children[n].match(r);
		}
		let n = this._rulesWithParentScopes.concat(this._mainRule);
		return n.sort(e._cmpBySpecificity), n;
	}
	insert(t, n, r, i, a, o) {
		if (n === "") {
			this._doInsertHere(t, r, i, a, o);
			return;
		}
		let s = n.indexOf("."), c, l;
		s === -1 ? (c = n, l = "") : (c = n.substring(0, s), l = n.substring(s + 1));
		let u;
		this._children.hasOwnProperty(c) ? u = this._children[c] : (u = new e(this._mainRule.clone(), Ce.cloneArr(this._rulesWithParentScopes)), this._children[c] = u), u.insert(t + 1, l, r, i, a, o);
	}
	_doInsertHere(e, t, n, r, i) {
		if (t === null) this._mainRule.acceptOverwrite(e, n, r, i);
		else {
			for (let a = 0, o = this._rulesWithParentScopes.length; a < o; a++) {
				let o = this._rulesWithParentScopes[a];
				if (le(o.parentScopes, t) === 0) {
					o.acceptOverwrite(e, n, r, i);
					return;
				}
			}
			n === -1 && (n = this._mainRule.fontStyle), r === 0 && (r = this._mainRule.foreground), i === 0 && (i = this._mainRule.background), this._rulesWithParentScopes.push(new Ce(e, t, n, r, i));
		}
	}
}, y = class e {
	static toBinaryStr(e) {
		return e.toString(2).padStart(32, "0");
	}
	static print(t) {
		let n = e.getLanguageId(t), r = e.getTokenType(t), i = e.getFontStyle(t), a = e.getForeground(t), o = e.getBackground(t);
		console.log({
			languageId: n,
			tokenType: r,
			fontStyle: i,
			foreground: a,
			background: o
		});
	}
	static getLanguageId(e) {
		return (e & 255) >>> 0;
	}
	static getTokenType(e) {
		return (e & 768) >>> 8;
	}
	static containsBalancedBrackets(e) {
		return !!(e & 1024);
	}
	static getFontStyle(e) {
		return (e & 30720) >>> 11;
	}
	static getForeground(e) {
		return (e & 16744448) >>> 15;
	}
	static getBackground(e) {
		return (e & 4278190080) >>> 24;
	}
	static set(t, n, r, i, a, o, s) {
		let c = e.getLanguageId(t), l = e.getTokenType(t), u = +!!e.containsBalancedBrackets(t), d = e.getFontStyle(t), f = e.getForeground(t), p = e.getBackground(t);
		return n !== 0 && (c = n), r !== 8 && (l = Ee(r)), i !== null && (u = +!!i), a !== -1 && (d = a), o !== 0 && (f = o), s !== 0 && (p = s), (c | 0 | l << 8 | u << 10 | d << 11 | f << 15 | p << 24) >>> 0;
	}
};
function Te(e) {
	return e;
}
function Ee(e) {
	return e;
}
function De(e, t) {
	let n = [], r = ke(e), i = r.next();
	for (; i !== null;) {
		let e = 0;
		if (i.length === 2 && i.charAt(1) === ":") {
			switch (i.charAt(0)) {
				case "R":
					e = 1;
					break;
				case "L":
					e = -1;
					break;
				default: console.log(`Unknown priority ${i} in scope selector`);
			}
			i = r.next();
		}
		let t = o();
		if (n.push({
			matcher: t,
			priority: e
		}), i !== ",") break;
		i = r.next();
	}
	return n;
	function a() {
		if (i === "-") {
			i = r.next();
			let e = a();
			return (t) => !!e && !e(t);
		}
		if (i === "(") {
			i = r.next();
			let e = s();
			return i === ")" && (i = r.next()), e;
		}
		if (Oe(i)) {
			let e = [];
			do
				e.push(i), i = r.next();
			while (Oe(i));
			return (n) => t(e, n);
		}
		return null;
	}
	function o() {
		let e = [], t = a();
		for (; t;) e.push(t), t = a();
		return (t) => e.every((e) => e(t));
	}
	function s() {
		let e = [], t = o();
		for (; t && (e.push(t), i === "|" || i === ",");) {
			do
				i = r.next();
			while (i === "|" || i === ",");
			t = o();
		}
		return (t) => e.some((e) => e(t));
	}
}
function Oe(e) {
	return !!e && !!e.match(/[\w\.:]+/);
}
function ke(e) {
	let t = /([LR]:|[\w\.:][\w\.:\-]*|[\,\|\-\(\)])/g, n = t.exec(e);
	return { next: () => {
		if (!n) return null;
		let r = n[0];
		return n = t.exec(e), r;
	} };
}
function Ae(e) {
	typeof e.dispose == "function" && e.dispose();
}
var b = class {
	constructor(e) {
		this.scopeName = e;
	}
	toKey() {
		return this.scopeName;
	}
}, je = class {
	constructor(e, t) {
		this.scopeName = e, this.ruleName = t;
	}
	toKey() {
		return `${this.scopeName}#${this.ruleName}`;
	}
}, Me = class {
	_references = [];
	_seenReferenceKeys = /* @__PURE__ */ new Set();
	get references() {
		return this._references;
	}
	visitedRule = /* @__PURE__ */ new Set();
	add(e) {
		let t = e.toKey();
		this._seenReferenceKeys.has(t) || (this._seenReferenceKeys.add(t), this._references.push(e));
	}
}, Ne = class {
	constructor(e, t) {
		this.repo = e, this.initialScopeName = t, this.seenFullScopeRequests.add(this.initialScopeName), this.Q = [new b(this.initialScopeName)];
	}
	seenFullScopeRequests = /* @__PURE__ */ new Set();
	seenPartialScopeRequests = /* @__PURE__ */ new Set();
	Q;
	processQueue() {
		let e = this.Q;
		this.Q = [];
		let t = new Me();
		for (let n of e) Pe(n, this.initialScopeName, this.repo, t);
		for (let e of t.references) if (e instanceof b) {
			if (this.seenFullScopeRequests.has(e.scopeName)) continue;
			this.seenFullScopeRequests.add(e.scopeName), this.Q.push(e);
		} else {
			if (this.seenFullScopeRequests.has(e.scopeName) || this.seenPartialScopeRequests.has(e.toKey())) continue;
			this.seenPartialScopeRequests.add(e.toKey()), this.Q.push(e);
		}
	}
};
function Pe(e, t, n, r) {
	let i = n.lookup(e.scopeName);
	if (!i) {
		if (e.scopeName === t) throw Error(`No grammar provided for <${t}>`);
		return;
	}
	let a = n.lookup(t);
	e instanceof b ? Ie({
		baseGrammar: a,
		selfGrammar: i
	}, r) : Fe(e.ruleName, {
		baseGrammar: a,
		selfGrammar: i,
		repository: i.repository
	}, r);
	let o = n.injections(e.scopeName);
	if (o) for (let e of o) r.add(new b(e));
}
function Fe(e, t, n) {
	if (t.repository && t.repository[e]) {
		let r = t.repository[e];
		Le([r], t, n);
	}
}
function Ie(e, t) {
	e.selfGrammar.patterns && Array.isArray(e.selfGrammar.patterns) && Le(e.selfGrammar.patterns, {
		...e,
		repository: e.selfGrammar.repository
	}, t), e.selfGrammar.injections && Le(Object.values(e.selfGrammar.injections), {
		...e,
		repository: e.selfGrammar.repository
	}, t);
}
function Le(e, t, n) {
	for (let r of e) {
		if (n.visitedRule.has(r)) continue;
		n.visitedRule.add(r);
		let e = r.repository ? ie({}, t.repository, r.repository) : t.repository;
		Array.isArray(r.patterns) && Le(r.patterns, {
			...t,
			repository: e
		}, n);
		let i = r.include;
		if (!i) continue;
		let a = Ue(i);
		switch (a.kind) {
			case 0:
				Ie({
					...t,
					selfGrammar: t.baseGrammar
				}, n);
				break;
			case 1:
				Ie(t, n);
				break;
			case 2:
				Fe(a.ruleName, {
					...t,
					repository: e
				}, n);
				break;
			case 3:
			case 4:
				let r = a.scopeName === t.selfGrammar.scopeName ? t.selfGrammar : a.scopeName === t.baseGrammar.scopeName ? t.baseGrammar : void 0;
				if (r) {
					let i = {
						baseGrammar: t.baseGrammar,
						selfGrammar: r,
						repository: e
					};
					a.kind === 4 ? Fe(a.ruleName, i, n) : Ie(i, n);
				} else a.kind === 4 ? n.add(new je(a.scopeName, a.ruleName)) : n.add(new b(a.scopeName));
		}
	}
}
var Re = class {
	kind = 0;
}, ze = class {
	kind = 1;
}, Be = class {
	constructor(e) {
		this.ruleName = e;
	}
	kind = 2;
}, Ve = class {
	constructor(e) {
		this.scopeName = e;
	}
	kind = 3;
}, He = class {
	constructor(e, t) {
		this.scopeName = e, this.ruleName = t;
	}
	kind = 4;
};
function Ue(e) {
	if (e === "$base") return new Re();
	if (e === "$self") return new ze();
	let t = e.indexOf("#");
	return t === -1 ? new Ve(e) : t === 0 ? new Be(e.substring(1)) : new He(e.substring(0, t), e.substring(t + 1));
}
var We = /\\(\d+)/, Ge = /\\(\d+)/g, Ke = -1, qe = -2;
function Je(e) {
	return e;
}
function Ye(e) {
	return e;
}
var x = class {
	$location;
	id;
	_nameIsCapturing;
	_name;
	_contentNameIsCapturing;
	_contentName;
	constructor(e, t, n, r) {
		this.$location = e, this.id = t, this._name = n || null, this._nameIsCapturing = se.hasCaptures(this._name), this._contentName = r || null, this._contentNameIsCapturing = se.hasCaptures(this._contentName);
	}
	get debugName() {
		let e = this.$location ? `${ae(this.$location.filename)}:${this.$location.line}` : "unknown";
		return `${this.constructor.name}#${this.id} @ ${e}`;
	}
	getName(e, t) {
		return !this._nameIsCapturing || this._name === null || e === null || t === null ? this._name : se.replaceCaptures(this._name, e, t);
	}
	getContentName(e, t) {
		return !this._contentNameIsCapturing || this._contentName === null ? this._contentName : se.replaceCaptures(this._contentName, e, t);
	}
}, Xe = class extends x {
	retokenizeCapturedWithRuleId;
	constructor(e, t, n, r, i) {
		super(e, t, n, r), this.retokenizeCapturedWithRuleId = i;
	}
	dispose() {}
	collectPatterns(e, t) {
		throw Error("Not supported!");
	}
	compile(e, t) {
		throw Error("Not supported!");
	}
	compileAG(e, t, n, r) {
		throw Error("Not supported!");
	}
}, Ze = class extends x {
	_match;
	captures;
	_cachedCompiledPatterns;
	constructor(e, t, n, r, i) {
		super(e, t, n, null), this._match = new S(r, this.id), this.captures = i, this._cachedCompiledPatterns = null;
	}
	dispose() {
		this._cachedCompiledPatterns &&= (this._cachedCompiledPatterns.dispose(), null);
	}
	get debugMatchRegExp() {
		return `${this._match.source}`;
	}
	collectPatterns(e, t) {
		t.push(this._match);
	}
	compile(e, t) {
		return this._getCachedCompiledPatterns(e).compile(e);
	}
	compileAG(e, t, n, r) {
		return this._getCachedCompiledPatterns(e).compileAG(e, n, r);
	}
	_getCachedCompiledPatterns(e) {
		return this._cachedCompiledPatterns || (this._cachedCompiledPatterns = new nt(), this.collectPatterns(e, this._cachedCompiledPatterns)), this._cachedCompiledPatterns;
	}
}, Qe = class extends x {
	hasMissingPatterns;
	patterns;
	_cachedCompiledPatterns;
	constructor(e, t, n, r, i) {
		super(e, t, n, r), this.patterns = i.patterns, this.hasMissingPatterns = i.hasMissingPatterns, this._cachedCompiledPatterns = null;
	}
	dispose() {
		this._cachedCompiledPatterns &&= (this._cachedCompiledPatterns.dispose(), null);
	}
	collectPatterns(e, t) {
		for (let n of this.patterns) e.getRule(n).collectPatterns(e, t);
	}
	compile(e, t) {
		return this._getCachedCompiledPatterns(e).compile(e);
	}
	compileAG(e, t, n, r) {
		return this._getCachedCompiledPatterns(e).compileAG(e, n, r);
	}
	_getCachedCompiledPatterns(e) {
		return this._cachedCompiledPatterns || (this._cachedCompiledPatterns = new nt(), this.collectPatterns(e, this._cachedCompiledPatterns)), this._cachedCompiledPatterns;
	}
}, $e = class extends x {
	_begin;
	beginCaptures;
	_end;
	endHasBackReferences;
	endCaptures;
	applyEndPatternLast;
	hasMissingPatterns;
	patterns;
	_cachedCompiledPatterns;
	constructor(e, t, n, r, i, a, o, s, c, l) {
		super(e, t, n, r), this._begin = new S(i, this.id), this.beginCaptures = a, this._end = new S(o || "￿", -1), this.endHasBackReferences = this._end.hasBackReferences, this.endCaptures = s, this.applyEndPatternLast = c || !1, this.patterns = l.patterns, this.hasMissingPatterns = l.hasMissingPatterns, this._cachedCompiledPatterns = null;
	}
	dispose() {
		this._cachedCompiledPatterns &&= (this._cachedCompiledPatterns.dispose(), null);
	}
	get debugBeginRegExp() {
		return `${this._begin.source}`;
	}
	get debugEndRegExp() {
		return `${this._end.source}`;
	}
	getEndWithResolvedBackReferences(e, t) {
		return this._end.resolveBackReferences(e, t);
	}
	collectPatterns(e, t) {
		t.push(this._begin);
	}
	compile(e, t) {
		return this._getCachedCompiledPatterns(e, t).compile(e);
	}
	compileAG(e, t, n, r) {
		return this._getCachedCompiledPatterns(e, t).compileAG(e, n, r);
	}
	_getCachedCompiledPatterns(e, t) {
		if (!this._cachedCompiledPatterns) {
			this._cachedCompiledPatterns = new nt();
			for (let t of this.patterns) e.getRule(t).collectPatterns(e, this._cachedCompiledPatterns);
			this.applyEndPatternLast ? this._cachedCompiledPatterns.push(this._end.hasBackReferences ? this._end.clone() : this._end) : this._cachedCompiledPatterns.unshift(this._end.hasBackReferences ? this._end.clone() : this._end);
		}
		return this._end.hasBackReferences && (this.applyEndPatternLast ? this._cachedCompiledPatterns.setSource(this._cachedCompiledPatterns.length() - 1, t) : this._cachedCompiledPatterns.setSource(0, t)), this._cachedCompiledPatterns;
	}
}, et = class extends x {
	_begin;
	beginCaptures;
	whileCaptures;
	_while;
	whileHasBackReferences;
	hasMissingPatterns;
	patterns;
	_cachedCompiledPatterns;
	_cachedCompiledWhilePatterns;
	constructor(e, t, n, r, i, a, o, s, c) {
		super(e, t, n, r), this._begin = new S(i, this.id), this.beginCaptures = a, this.whileCaptures = s, this._while = new S(o, qe), this.whileHasBackReferences = this._while.hasBackReferences, this.patterns = c.patterns, this.hasMissingPatterns = c.hasMissingPatterns, this._cachedCompiledPatterns = null, this._cachedCompiledWhilePatterns = null;
	}
	dispose() {
		this._cachedCompiledPatterns &&= (this._cachedCompiledPatterns.dispose(), null), this._cachedCompiledWhilePatterns &&= (this._cachedCompiledWhilePatterns.dispose(), null);
	}
	get debugBeginRegExp() {
		return `${this._begin.source}`;
	}
	get debugWhileRegExp() {
		return `${this._while.source}`;
	}
	getWhileWithResolvedBackReferences(e, t) {
		return this._while.resolveBackReferences(e, t);
	}
	collectPatterns(e, t) {
		t.push(this._begin);
	}
	compile(e, t) {
		return this._getCachedCompiledPatterns(e).compile(e);
	}
	compileAG(e, t, n, r) {
		return this._getCachedCompiledPatterns(e).compileAG(e, n, r);
	}
	_getCachedCompiledPatterns(e) {
		if (!this._cachedCompiledPatterns) {
			this._cachedCompiledPatterns = new nt();
			for (let t of this.patterns) e.getRule(t).collectPatterns(e, this._cachedCompiledPatterns);
		}
		return this._cachedCompiledPatterns;
	}
	compileWhile(e, t) {
		return this._getCachedCompiledWhilePatterns(e, t).compile(e);
	}
	compileWhileAG(e, t, n, r) {
		return this._getCachedCompiledWhilePatterns(e, t).compileAG(e, n, r);
	}
	_getCachedCompiledWhilePatterns(e, t) {
		return this._cachedCompiledWhilePatterns || (this._cachedCompiledWhilePatterns = new nt(), this._cachedCompiledWhilePatterns.push(this._while.hasBackReferences ? this._while.clone() : this._while)), this._while.hasBackReferences && this._cachedCompiledWhilePatterns.setSource(0, t || "￿"), this._cachedCompiledWhilePatterns;
	}
}, tt = class e {
	static createCaptureRule(e, t, n, r, i) {
		return e.registerRule((e) => new Xe(t, e, n, r, i));
	}
	static getCompiledRuleId(t, n, r) {
		return t.id || n.registerRule((i) => {
			if (t.id = i, t.match) return new Ze(t.$vscodeTextmateLocation, t.id, t.name, t.match, e._compileCaptures(t.captures, n, r));
			if (t.begin === void 0) {
				t.repository && (r = ie({}, r, t.repository));
				let i = t.patterns;
				return i === void 0 && t.include && (i = [{ include: t.include }]), new Qe(t.$vscodeTextmateLocation, t.id, t.name, t.contentName, e._compilePatterns(i, n, r));
			}
			return t.while ? new et(t.$vscodeTextmateLocation, t.id, t.name, t.contentName, t.begin, e._compileCaptures(t.beginCaptures || t.captures, n, r), t.while, e._compileCaptures(t.whileCaptures || t.captures, n, r), e._compilePatterns(t.patterns, n, r)) : new $e(t.$vscodeTextmateLocation, t.id, t.name, t.contentName, t.begin, e._compileCaptures(t.beginCaptures || t.captures, n, r), t.end, e._compileCaptures(t.endCaptures || t.captures, n, r), t.applyEndPatternLast, e._compilePatterns(t.patterns, n, r));
		}), t.id;
	}
	static _compileCaptures(t, n, r) {
		let i = [];
		if (t) {
			let a = 0;
			for (let e in t) {
				if (e === "$vscodeTextmateLocation") continue;
				let t = parseInt(e, 10);
				t > a && (a = t);
			}
			for (let e = 0; e <= a; e++) i[e] = null;
			for (let a in t) {
				if (a === "$vscodeTextmateLocation") continue;
				let o = parseInt(a, 10), s = 0;
				t[a].patterns && (s = e.getCompiledRuleId(t[a], n, r)), i[o] = e.createCaptureRule(n, t[a].$vscodeTextmateLocation, t[a].name, t[a].contentName, s);
			}
		}
		return i;
	}
	static _compilePatterns(t, n, r) {
		let i = [];
		if (t) for (let a = 0, o = t.length; a < o; a++) {
			let o = t[a], s = -1;
			if (o.include) {
				let t = Ue(o.include);
				switch (t.kind) {
					case 0:
					case 1:
						s = e.getCompiledRuleId(r[o.include], n, r);
						break;
					case 2:
						let i = r[t.ruleName];
						i && (s = e.getCompiledRuleId(i, n, r));
						break;
					case 3:
					case 4:
						let a = t.scopeName, c = t.kind === 4 ? t.ruleName : null, l = n.getExternalGrammar(a, r);
						if (l) {
							if (c) {
								let t = l.repository[c];
								t && (s = e.getCompiledRuleId(t, n, l.repository));
							} else s = e.getCompiledRuleId(l.repository.$self, n, l.repository);
						}
				}
			} else s = e.getCompiledRuleId(o, n, r);
			if (s !== -1) {
				let e = n.getRule(s), t = !1;
				if ((e instanceof Qe || e instanceof $e || e instanceof et) && e.hasMissingPatterns && e.patterns.length === 0 && (t = !0), t) continue;
				i.push(s);
			}
		}
		return {
			patterns: i,
			hasMissingPatterns: (t ? t.length : 0) !== i.length
		};
	}
}, S = class e {
	source;
	ruleId;
	hasAnchor;
	hasBackReferences;
	_anchorCache;
	constructor(e, t) {
		if (e && typeof e == "string") {
			let t = e.length, n = 0, r = [], i = !1;
			for (let a = 0; a < t; a++) if (e.charAt(a) === "\\" && a + 1 < t) {
				let t = e.charAt(a + 1);
				t === "z" ? (r.push(e.substring(n, a)), r.push("$(?!\\n)(?<!\\n)"), n = a + 2) : (t === "A" || t === "G") && (i = !0), a++;
			}
			this.hasAnchor = i, n === 0 ? this.source = e : (r.push(e.substring(n, t)), this.source = r.join(""));
		} else this.hasAnchor = !1, this.source = e;
		this._anchorCache = this.hasAnchor ? this._buildAnchorCache() : null, this.ruleId = t, this.hasBackReferences = typeof this.source == "string" && We.test(this.source);
	}
	clone() {
		return new e(this.source, this.ruleId);
	}
	setSource(e) {
		this.source !== e && (this.source = e, this.hasAnchor && (this._anchorCache = this._buildAnchorCache()));
	}
	resolveBackReferences(e, t) {
		if (typeof this.source != "string") throw Error("This method should only be called if the source is a string");
		let n = t.map((t) => e.substring(t.start, t.end));
		return Ge.lastIndex = 0, this.source.replace(Ge, (e, t) => de(n[parseInt(t, 10)] || ""));
	}
	_buildAnchorCache() {
		if (typeof this.source != "string") throw Error("This method should only be called if the source is a string");
		let e = [], t = [], n = [], r = [], i, a, o, s;
		for (i = 0, a = this.source.length; i < a; i++) o = this.source.charAt(i), e[i] = o, t[i] = o, n[i] = o, r[i] = o, o === "\\" && i + 1 < a && (s = this.source.charAt(i + 1), s === "A" ? (e[i + 1] = "￿", t[i + 1] = "￿", n[i + 1] = "A", r[i + 1] = "A") : s === "G" ? (e[i + 1] = "￿", t[i + 1] = "G", n[i + 1] = "￿", r[i + 1] = "G") : (e[i + 1] = s, t[i + 1] = s, n[i + 1] = s, r[i + 1] = s), i++);
		return {
			A0_G0: e.join(""),
			A0_G1: t.join(""),
			A1_G0: n.join(""),
			A1_G1: r.join("")
		};
	}
	resolveAnchors(e, t) {
		return !this.hasAnchor || !this._anchorCache || typeof this.source != "string" ? this.source : e ? t ? this._anchorCache.A1_G1 : this._anchorCache.A1_G0 : t ? this._anchorCache.A0_G1 : this._anchorCache.A0_G0;
	}
}, nt = class {
	_items;
	_hasAnchors;
	_cached;
	_anchorCache;
	constructor() {
		this._items = [], this._hasAnchors = !1, this._cached = null, this._anchorCache = {
			A0_G0: null,
			A0_G1: null,
			A1_G0: null,
			A1_G1: null
		};
	}
	dispose() {
		this._disposeCaches();
	}
	_disposeCaches() {
		this._cached &&= (this._cached.dispose(), null), this._anchorCache.A0_G0 && (this._anchorCache.A0_G0.dispose(), this._anchorCache.A0_G0 = null), this._anchorCache.A0_G1 && (this._anchorCache.A0_G1.dispose(), this._anchorCache.A0_G1 = null), this._anchorCache.A1_G0 && (this._anchorCache.A1_G0.dispose(), this._anchorCache.A1_G0 = null), this._anchorCache.A1_G1 && (this._anchorCache.A1_G1.dispose(), this._anchorCache.A1_G1 = null);
	}
	push(e) {
		this._items.push(e), this._hasAnchors = this._hasAnchors || e.hasAnchor;
	}
	unshift(e) {
		this._items.unshift(e), this._hasAnchors = this._hasAnchors || e.hasAnchor;
	}
	length() {
		return this._items.length;
	}
	setSource(e, t) {
		this._items[e].source !== t && (this._disposeCaches(), this._items[e].setSource(t));
	}
	compile(e) {
		if (!this._cached) {
			let t = this._items.map((e) => e.source);
			this._cached = new rt(e, t, this._items.map((e) => e.ruleId));
		}
		return this._cached;
	}
	compileAG(e, t, n) {
		return this._hasAnchors ? t ? n ? (this._anchorCache.A1_G1 || (this._anchorCache.A1_G1 = this._resolveAnchors(e, t, n)), this._anchorCache.A1_G1) : (this._anchorCache.A1_G0 || (this._anchorCache.A1_G0 = this._resolveAnchors(e, t, n)), this._anchorCache.A1_G0) : n ? (this._anchorCache.A0_G1 || (this._anchorCache.A0_G1 = this._resolveAnchors(e, t, n)), this._anchorCache.A0_G1) : (this._anchorCache.A0_G0 || (this._anchorCache.A0_G0 = this._resolveAnchors(e, t, n)), this._anchorCache.A0_G0) : this.compile(e);
	}
	_resolveAnchors(e, t, n) {
		return new rt(e, this._items.map((e) => e.resolveAnchors(t, n)), this._items.map((e) => e.ruleId));
	}
}, rt = class {
	constructor(e, t, n) {
		this.regExps = t, this.rules = n, this.scanner = e.createOnigScanner(t);
	}
	scanner;
	dispose() {
		typeof this.scanner.dispose == "function" && this.scanner.dispose();
	}
	toString() {
		let e = [];
		for (let t = 0, n = this.rules.length; t < n; t++) e.push("   - " + this.rules[t] + ": " + this.regExps[t]);
		return e.join("\n");
	}
	findNextMatchSync(e, t, n) {
		let r = this.scanner.findNextMatchSync(e, t, n);
		return r ? {
			ruleId: this.rules[r.index],
			captureIndices: r.captureIndices
		} : null;
	}
}, it = class {
	constructor(e, t) {
		this.languageId = e, this.tokenType = t;
	}
}, at = class e {
	_defaultAttributes;
	_embeddedLanguagesMatcher;
	constructor(e, t) {
		this._defaultAttributes = new it(e, 8), this._embeddedLanguagesMatcher = new ot(Object.entries(t || {}));
	}
	getDefaultAttributes() {
		return this._defaultAttributes;
	}
	getBasicScopeAttributes(t) {
		return t === null ? e._NULL_SCOPE_METADATA : this._getBasicScopeAttributes.get(t);
	}
	static _NULL_SCOPE_METADATA = new it(0, 0);
	_getBasicScopeAttributes = new fe((e) => new it(this._scopeToLanguage(e), this._toStandardTokenType(e)));
	_scopeToLanguage(e) {
		return this._embeddedLanguagesMatcher.match(e) || 0;
	}
	_toStandardTokenType(t) {
		let n = t.match(e.STANDARD_TOKEN_TYPE_REGEXP);
		if (!n) return 8;
		switch (n[1]) {
			case "comment": return 1;
			case "string": return 2;
			case "regex": return 3;
			case "meta.embedded": return 0;
		}
		throw Error("Unexpected match for standard token type!");
	}
	static STANDARD_TOKEN_TYPE_REGEXP = /\b(comment|string|regex|meta\.embedded)\b/;
}, ot = class {
	values;
	scopesRegExp;
	constructor(e) {
		if (e.length === 0) this.values = null, this.scopesRegExp = null;
		else {
			this.values = new Map(e);
			let t = e.map(([e, t]) => de(e));
			t.sort(), t.reverse(), this.scopesRegExp = RegExp(`^((${t.join(")|(")}))($|\\.)`, "");
		}
	}
	match(e) {
		if (!this.scopesRegExp) return;
		let t = e.match(this.scopesRegExp);
		if (t) return this.values.get(t[1]);
	}
};
typeof process < "u" && process.env.VSCODE_TEXTMATE_DEBUG;
var st = class {
	constructor(e, t) {
		this.stack = e, this.stoppedEarly = t;
	}
};
function ct(e, t, n, r, i, a, o, s) {
	let c = t.content.length, l = !1, u = -1;
	if (o) {
		let o = lt(e, t, n, r, i, a);
		i = o.stack, r = o.linePos, n = o.isFirstLine, u = o.anchorPosition;
	}
	let d = Date.now();
	for (; !l;) {
		if (s !== 0 && Date.now() - d > s) return new st(i, !0);
		f();
	}
	return new st(i, !1);
	function f() {
		let o = ut(e, t, n, r, i, u);
		if (!o) {
			a.produce(i, c), l = !0;
			return;
		}
		let s = o.captureIndices, d = o.matchedRuleId, f = s && s.length > 0 ? s[0].end > r : !1;
		if (d === Ke) {
			let o = i.getRule(e);
			a.produce(i, s[0].start), i = i.withContentNameScopesList(i.nameScopesList), ht(e, t, n, i, a, o.endCaptures, s), a.produce(i, s[0].end);
			let d = i;
			if (i = i.parent, u = d.getAnchorPos(), !f && d.getEnterPos() === r) {
				i = d, a.produce(i, c), l = !0;
				return;
			}
		} else {
			let o = e.getRule(d);
			a.produce(i, s[0].start);
			let p = i, m = o.getName(t.content, s), h = i.contentNameScopesList.pushAttributed(m, e);
			if (i = i.push(d, r, u, s[0].end === c, null, h, h), o instanceof $e) {
				let r = o;
				ht(e, t, n, i, a, r.beginCaptures, s), a.produce(i, s[0].end), u = s[0].end;
				let d = r.getContentName(t.content, s), m = h.pushAttributed(d, e);
				if (i = i.withContentNameScopesList(m), r.endHasBackReferences && (i = i.withEndRule(r.getEndWithResolvedBackReferences(t.content, s))), !f && p.hasSameRuleAs(i)) {
					i = i.pop(), a.produce(i, c), l = !0;
					return;
				}
			} else if (o instanceof et) {
				let r = o;
				ht(e, t, n, i, a, r.beginCaptures, s), a.produce(i, s[0].end), u = s[0].end;
				let d = r.getContentName(t.content, s), m = h.pushAttributed(d, e);
				if (i = i.withContentNameScopesList(m), r.whileHasBackReferences && (i = i.withEndRule(r.getWhileWithResolvedBackReferences(t.content, s))), !f && p.hasSameRuleAs(i)) {
					i = i.pop(), a.produce(i, c), l = !0;
					return;
				}
			} else if (ht(e, t, n, i, a, o.captures, s), a.produce(i, s[0].end), i = i.pop(), !f) {
				i = i.safePop(), a.produce(i, c), l = !0;
				return;
			}
		}
		s[0].end > r && (r = s[0].end, n = !1);
	}
}
function lt(e, t, n, r, i, a) {
	let o = i.beginRuleCapturedEOL ? 0 : -1, s = [];
	for (let t = i; t; t = t.pop()) {
		let n = t.getRule(e);
		n instanceof et && s.push({
			rule: n,
			stack: t
		});
	}
	for (let c = s.pop(); c; c = s.pop()) {
		let { ruleScanner: s, findOptions: l } = mt(c.rule, e, c.stack.endRule, n, r === o), u = s.findNextMatchSync(t, r, l);
		if (u) {
			if (u.ruleId !== qe) {
				i = c.stack.pop();
				break;
			}
			u.captureIndices && u.captureIndices.length && (a.produce(c.stack, u.captureIndices[0].start), ht(e, t, n, c.stack, a, c.rule.whileCaptures, u.captureIndices), a.produce(c.stack, u.captureIndices[0].end), o = u.captureIndices[0].end, u.captureIndices[0].end > r && (r = u.captureIndices[0].end, n = !1));
		} else {
			i = c.stack.pop();
			break;
		}
	}
	return {
		stack: i,
		linePos: r,
		anchorPosition: o,
		isFirstLine: n
	};
}
function ut(e, t, n, r, i, a) {
	let o = dt(e, t, n, r, i, a), s = e.getInjections();
	if (s.length === 0) return o;
	let c = ft(s, e, t, n, r, i, a);
	if (!c) return o;
	if (!o) return c;
	let l = o.captureIndices[0].start, u = c.captureIndices[0].start;
	return u < l || c.priorityMatch && u === l ? c : o;
}
function dt(e, t, n, r, i, a) {
	let { ruleScanner: o, findOptions: s } = pt(i.getRule(e), e, i.endRule, n, r === a), c = o.findNextMatchSync(t, r, s);
	return c ? {
		captureIndices: c.captureIndices,
		matchedRuleId: c.ruleId
	} : null;
}
function ft(e, t, n, r, i, a, o) {
	let s = Number.MAX_VALUE, c = null, l, u = 0, d = a.contentNameScopesList.getScopeNames();
	for (let a = 0, f = e.length; a < f; a++) {
		let f = e[a];
		if (!f.matcher(d)) continue;
		let { ruleScanner: p, findOptions: m } = pt(t.getRule(f.ruleId), t, null, r, i === o), h = p.findNextMatchSync(n, i, m);
		if (!h) continue;
		let g = h.captureIndices[0].start;
		if (!(g >= s) && (s = g, c = h.captureIndices, l = h.ruleId, u = f.priority, s === i)) break;
	}
	return c ? {
		priorityMatch: u === -1,
		captureIndices: c,
		matchedRuleId: l
	} : null;
}
function pt(e, t, n, r, i) {
	return {
		ruleScanner: e.compileAG(t, n, r, i),
		findOptions: 0
	};
}
function mt(e, t, n, r, i) {
	return {
		ruleScanner: e.compileWhileAG(t, n, r, i),
		findOptions: 0
	};
}
function ht(e, t, n, r, i, a, o) {
	if (a.length === 0) return;
	let s = t.content, c = Math.min(a.length, o.length), l = [], u = o[0].end;
	for (let t = 0; t < c; t++) {
		let c = a[t];
		if (c === null) continue;
		let d = o[t];
		if (d.length === 0) continue;
		if (d.start > u) break;
		for (; l.length > 0 && l[l.length - 1].endPos <= d.start;) i.produceFromScopes(l[l.length - 1].scopes, l[l.length - 1].endPos), l.pop();
		if (l.length > 0 ? i.produceFromScopes(l[l.length - 1].scopes, d.start) : i.produce(r, d.start), c.retokenizeCapturedWithRuleId) {
			let t = c.getName(s, o), a = r.contentNameScopesList.pushAttributed(t, e), l = c.getContentName(s, o), u = a.pushAttributed(l, e), f = r.push(c.retokenizeCapturedWithRuleId, d.start, -1, !1, null, a, u), p = e.createOnigString(s.substring(0, d.end));
			ct(e, p, n && d.start === 0, d.start, f, i, !1, 0), Ae(p);
			continue;
		}
		let f = c.getName(s, o);
		if (f !== null) {
			let t = (l.length > 0 ? l[l.length - 1].scopes : r.contentNameScopesList).pushAttributed(f, e);
			l.push(new gt(t, d.end));
		}
	}
	for (; l.length > 0;) i.produceFromScopes(l[l.length - 1].scopes, l[l.length - 1].endPos), l.pop();
}
var gt = class {
	scopes;
	endPos;
	constructor(e, t) {
		this.scopes = e, this.endPos = t;
	}
};
function _t(e, t, n, r, i, a, o, s) {
	return new xt(e, t, n, r, i, a, o, s);
}
function vt(e, t, n, r, i) {
	let a = De(t, yt), o = tt.getCompiledRuleId(n, r, i.repository);
	for (let n of a) e.push({
		debugSelector: t,
		matcher: n.matcher,
		ruleId: o,
		grammar: i,
		priority: n.priority
	});
}
function yt(e, t) {
	if (t.length < e.length) return !1;
	let n = 0;
	return e.every((e) => {
		for (let r = n; r < t.length; r++) if (bt(t[r], e)) return n = r + 1, !0;
		return !1;
	});
}
function bt(e, t) {
	if (!e) return !1;
	if (e === t) return !0;
	let n = t.length;
	return e.length > n && e.substr(0, n) === t && e[n] === ".";
}
var xt = class {
	constructor(e, t, n, r, i, a, o, s) {
		if (this._rootScopeName = e, this.balancedBracketSelectors = a, this._onigLib = s, this._basicScopeAttributesProvider = new at(n, r), this._rootId = -1, this._lastRuleId = 0, this._ruleId2desc = [null], this._includedGrammars = {}, this._grammarRepository = o, this._grammar = St(t, null), this._injections = null, this._tokenTypeMatchers = [], i) for (let e of Object.keys(i)) {
			let t = De(e, yt);
			for (let n of t) this._tokenTypeMatchers.push({
				matcher: n.matcher,
				type: i[e]
			});
		}
	}
	_rootId;
	_lastRuleId;
	_ruleId2desc;
	_includedGrammars;
	_grammarRepository;
	_grammar;
	_injections;
	_basicScopeAttributesProvider;
	_tokenTypeMatchers;
	get themeProvider() {
		return this._grammarRepository;
	}
	dispose() {
		for (let e of this._ruleId2desc) e && e.dispose();
	}
	createOnigScanner(e) {
		return this._onigLib.createOnigScanner(e);
	}
	createOnigString(e) {
		return this._onigLib.createOnigString(e);
	}
	getMetadataForScope(e) {
		return this._basicScopeAttributesProvider.getBasicScopeAttributes(e);
	}
	_collectInjections() {
		let e = {
			lookup: (e) => e === this._rootScopeName ? this._grammar : this.getExternalGrammar(e),
			injections: (e) => this._grammarRepository.injections(e)
		}, t = [], n = this._rootScopeName, r = e.lookup(n);
		if (r) {
			let e = r.injections;
			if (e) for (let n in e) vt(t, n, e[n], this, r);
			let i = this._grammarRepository.injections(n);
			i && i.forEach((e) => {
				let n = this.getExternalGrammar(e);
				if (n) {
					let e = n.injectionSelector;
					e && vt(t, e, n, this, n);
				}
			});
		}
		return t.sort((e, t) => e.priority - t.priority), t;
	}
	getInjections() {
		return this._injections === null && (this._injections = this._collectInjections()), this._injections;
	}
	registerRule(e) {
		let t = ++this._lastRuleId, n = e(Je(t));
		return this._ruleId2desc[t] = n, n;
	}
	getRule(e) {
		return this._ruleId2desc[Ye(e)];
	}
	getExternalGrammar(e, t) {
		if (this._includedGrammars[e]) return this._includedGrammars[e];
		if (this._grammarRepository) {
			let n = this._grammarRepository.lookup(e);
			if (n) return this._includedGrammars[e] = St(n, t && t.$base), this._includedGrammars[e];
		}
	}
	tokenizeLine(e, t, n = 0) {
		let r = this._tokenize(e, t, !1, n);
		return {
			tokens: r.lineTokens.getResult(r.ruleStack, r.lineLength),
			ruleStack: r.ruleStack,
			stoppedEarly: r.stoppedEarly
		};
	}
	tokenizeLine2(e, t, n = 0) {
		let r = this._tokenize(e, t, !0, n);
		return {
			tokens: r.lineTokens.getBinaryResult(r.ruleStack, r.lineLength),
			ruleStack: r.ruleStack,
			stoppedEarly: r.stoppedEarly
		};
	}
	_tokenize(e, t, n, r) {
		this._rootId === -1 && (this._rootId = tt.getCompiledRuleId(this._grammar.repository.$self, this, this._grammar.repository), this.getInjections());
		let i;
		if (!t || t === Ct.NULL) {
			i = !0;
			let e = this._basicScopeAttributesProvider.getDefaultAttributes(), n = this.themeProvider.getDefaults(), r = y.set(0, e.languageId, e.tokenType, null, n.fontStyle, n.foregroundId, n.backgroundId), a = this.getRule(this._rootId).getName(null, null), o;
			o = a ? C.createRootAndLookUpScopeName(a, r, this) : C.createRoot("unknown", r), t = new Ct(null, this._rootId, -1, -1, !1, null, o, o);
		} else i = !1, t.reset();
		e += "\n";
		let a = this.createOnigString(e), o = a.content.length, s = new Tt(n, e, this._tokenTypeMatchers, this.balancedBracketSelectors), c = ct(this, a, i, 0, t, s, !0, r);
		return Ae(a), {
			lineLength: o,
			lineTokens: s,
			ruleStack: c.stack,
			stoppedEarly: c.stoppedEarly
		};
	}
};
function St(e, t) {
	return e = ee(e), e.repository = e.repository || {}, e.repository.$self = {
		$vscodeTextmateLocation: e.$vscodeTextmateLocation,
		patterns: e.patterns,
		name: e.scopeName
	}, e.repository.$base = t || e.repository.$self, e;
}
var C = class e {
	constructor(e, t, n) {
		this.parent = e, this.scopePath = t, this.tokenAttributes = n;
	}
	static fromExtension(t, n) {
		let r = t, i = t?.scopePath ?? null;
		for (let t of n) i = me.push(i, t.scopeNames), r = new e(r, i, t.encodedTokenAttributes);
		return r;
	}
	static createRoot(t, n) {
		return new e(null, new me(null, t), n);
	}
	static createRootAndLookUpScopeName(t, n, r) {
		let i = r.getMetadataForScope(t), a = new me(null, t), o = r.themeProvider.themeMatch(a), s = e.mergeAttributes(n, i, o);
		return new e(null, a, s);
	}
	get scopeName() {
		return this.scopePath.scopeName;
	}
	toString() {
		return this.getScopeNames().join(" ");
	}
	equals(t) {
		return e.equals(this, t);
	}
	static equals(e, t) {
		do {
			if (e === t || !e && !t) return !0;
			if (!e || !t || e.scopeName !== t.scopeName || e.tokenAttributes !== t.tokenAttributes) return !1;
			e = e.parent, t = t.parent;
		} while (1);
	}
	static mergeAttributes(e, t, n) {
		let r = -1, i = 0, a = 0;
		return n !== null && (r = n.fontStyle, i = n.foregroundId, a = n.backgroundId), y.set(e, t.languageId, t.tokenType, null, r, i, a);
	}
	pushAttributed(t, n) {
		if (t === null) return this;
		if (t.indexOf(" ") === -1) return e._pushAttributed(this, t, n);
		let r = t.split(/ /g), i = this;
		for (let t of r) i = e._pushAttributed(i, t, n);
		return i;
	}
	static _pushAttributed(t, n, r) {
		let i = r.getMetadataForScope(n), a = t.scopePath.push(n), o = r.themeProvider.themeMatch(a), s = e.mergeAttributes(t.tokenAttributes, i, o);
		return new e(t, a, s);
	}
	getScopeNames() {
		return this.scopePath.getSegments();
	}
	getExtensionIfDefined(e) {
		let t = [], n = this;
		for (; n && n !== e;) t.push({
			encodedTokenAttributes: n.tokenAttributes,
			scopeNames: n.scopePath.getExtensionIfDefined(n.parent?.scopePath ?? null)
		}), n = n.parent;
		return n === e ? t.reverse() : void 0;
	}
}, Ct = class e {
	constructor(e, t, n, r, i, a, o, s) {
		this.parent = e, this.ruleId = t, this.beginRuleCapturedEOL = i, this.endRule = a, this.nameScopesList = o, this.contentNameScopesList = s, this.depth = this.parent ? this.parent.depth + 1 : 1, this._enterPos = n, this._anchorPos = r;
	}
	_stackElementBrand = void 0;
	static NULL = new e(null, 0, 0, 0, !1, null, null, null);
	_enterPos;
	_anchorPos;
	depth;
	equals(t) {
		return t !== null && e._equals(this, t);
	}
	static _equals(e, t) {
		return e === t ? !0 : this._structuralEquals(e, t) ? C.equals(e.contentNameScopesList, t.contentNameScopesList) : !1;
	}
	static _structuralEquals(e, t) {
		do {
			if (e === t || !e && !t) return !0;
			if (!e || !t || e.depth !== t.depth || e.ruleId !== t.ruleId || e.endRule !== t.endRule) return !1;
			e = e.parent, t = t.parent;
		} while (1);
	}
	clone() {
		return this;
	}
	static _reset(e) {
		for (; e;) e._enterPos = -1, e._anchorPos = -1, e = e.parent;
	}
	reset() {
		e._reset(this);
	}
	pop() {
		return this.parent;
	}
	safePop() {
		return this.parent ? this.parent : this;
	}
	push(t, n, r, i, a, o, s) {
		return new e(this, t, n, r, i, a, o, s);
	}
	getEnterPos() {
		return this._enterPos;
	}
	getAnchorPos() {
		return this._anchorPos;
	}
	getRule(e) {
		return e.getRule(this.ruleId);
	}
	toString() {
		let e = [];
		return this._writeString(e, 0), "[" + e.join(",") + "]";
	}
	_writeString(e, t) {
		return this.parent && (t = this.parent._writeString(e, t)), e[t++] = `(${this.ruleId}, ${this.nameScopesList?.toString()}, ${this.contentNameScopesList?.toString()})`, t;
	}
	withContentNameScopesList(e) {
		return this.contentNameScopesList === e ? this : this.parent.push(this.ruleId, this._enterPos, this._anchorPos, this.beginRuleCapturedEOL, this.endRule, this.nameScopesList, e);
	}
	withEndRule(t) {
		return this.endRule === t ? this : new e(this.parent, this.ruleId, this._enterPos, this._anchorPos, this.beginRuleCapturedEOL, t, this.nameScopesList, this.contentNameScopesList);
	}
	hasSameRuleAs(e) {
		let t = this;
		for (; t && t._enterPos === e._enterPos;) {
			if (t.ruleId === e.ruleId) return !0;
			t = t.parent;
		}
		return !1;
	}
	toStateStackFrame() {
		return {
			ruleId: Ye(this.ruleId),
			beginRuleCapturedEOL: this.beginRuleCapturedEOL,
			endRule: this.endRule,
			nameScopesList: this.nameScopesList?.getExtensionIfDefined(this.parent?.nameScopesList ?? null) ?? [],
			contentNameScopesList: this.contentNameScopesList?.getExtensionIfDefined(this.nameScopesList) ?? []
		};
	}
	static pushFrame(t, n) {
		let r = C.fromExtension(t?.nameScopesList ?? null, n.nameScopesList);
		return new e(t, Je(n.ruleId), n.enterPos ?? -1, n.anchorPos ?? -1, n.beginRuleCapturedEOL, n.endRule, r, C.fromExtension(r, n.contentNameScopesList));
	}
}, wt = class {
	balancedBracketScopes;
	unbalancedBracketScopes;
	allowAny = !1;
	constructor(e, t) {
		this.balancedBracketScopes = e.flatMap((e) => e === "*" ? (this.allowAny = !0, []) : De(e, yt).map((e) => e.matcher)), this.unbalancedBracketScopes = t.flatMap((e) => De(e, yt).map((e) => e.matcher));
	}
	get matchesAlways() {
		return this.allowAny && this.unbalancedBracketScopes.length === 0;
	}
	get matchesNever() {
		return this.balancedBracketScopes.length === 0 && !this.allowAny;
	}
	match(e) {
		for (let t of this.unbalancedBracketScopes) if (t(e)) return !1;
		for (let t of this.balancedBracketScopes) if (t(e)) return !0;
		return this.allowAny;
	}
}, Tt = class {
	constructor(e, t, n, r) {
		this.balancedBracketSelectors = r, this._emitBinaryTokens = e, this._tokenTypeOverrides = n, this._lineText = null, this._tokens = [], this._binaryTokens = [], this._lastTokenEndIndex = 0;
	}
	_emitBinaryTokens;
	_lineText;
	_tokens;
	_binaryTokens;
	_lastTokenEndIndex;
	_tokenTypeOverrides;
	produce(e, t) {
		this.produceFromScopes(e.contentNameScopesList, t);
	}
	produceFromScopes(e, t) {
		if (this._lastTokenEndIndex >= t) return;
		if (this._emitBinaryTokens) {
			let n = e?.tokenAttributes ?? 0, r = !1;
			if (this.balancedBracketSelectors?.matchesAlways && (r = !0), this._tokenTypeOverrides.length > 0 || this.balancedBracketSelectors && !this.balancedBracketSelectors.matchesAlways && !this.balancedBracketSelectors.matchesNever) {
				let t = e?.getScopeNames() ?? [];
				for (let e of this._tokenTypeOverrides) e.matcher(t) && (n = y.set(n, 0, Te(e.type), null, -1, 0, 0));
				this.balancedBracketSelectors && (r = this.balancedBracketSelectors.match(t));
			}
			if (r && (n = y.set(n, 0, 8, r, -1, 0, 0)), this._binaryTokens.length > 0 && this._binaryTokens[this._binaryTokens.length - 1] === n) {
				this._lastTokenEndIndex = t;
				return;
			}
			this._binaryTokens.push(this._lastTokenEndIndex), this._binaryTokens.push(n), this._lastTokenEndIndex = t;
			return;
		}
		let n = e?.getScopeNames() ?? [];
		this._tokens.push({
			startIndex: this._lastTokenEndIndex,
			endIndex: t,
			scopes: n
		}), this._lastTokenEndIndex = t;
	}
	getResult(e, t) {
		return this._tokens.length > 0 && this._tokens[this._tokens.length - 1].startIndex === t - 1 && this._tokens.pop(), this._tokens.length === 0 && (this._lastTokenEndIndex = -1, this.produce(e, t), this._tokens[this._tokens.length - 1].startIndex = 0), this._tokens;
	}
	getBinaryResult(e, t) {
		this._binaryTokens.length > 0 && this._binaryTokens[this._binaryTokens.length - 2] === t - 1 && (this._binaryTokens.pop(), this._binaryTokens.pop()), this._binaryTokens.length === 0 && (this._lastTokenEndIndex = -1, this.produce(e, t), this._binaryTokens[this._binaryTokens.length - 2] = 0);
		let n = new Uint32Array(this._binaryTokens.length);
		for (let e = 0, t = this._binaryTokens.length; e < t; e++) n[e] = this._binaryTokens[e];
		return n;
	}
}, Et = class {
	constructor(e, t) {
		this._onigLib = t, this._theme = e;
	}
	_grammars = /* @__PURE__ */ new Map();
	_rawGrammars = /* @__PURE__ */ new Map();
	_injectionGrammars = /* @__PURE__ */ new Map();
	_theme;
	dispose() {
		for (let e of this._grammars.values()) e.dispose();
	}
	setTheme(e) {
		this._theme = e;
	}
	getColorMap() {
		return this._theme.getColorMap();
	}
	addGrammar(e, t) {
		this._rawGrammars.set(e.scopeName, e), t && this._injectionGrammars.set(e.scopeName, t);
	}
	lookup(e) {
		return this._rawGrammars.get(e);
	}
	injections(e) {
		return this._injectionGrammars.get(e);
	}
	getDefaults() {
		return this._theme.getDefaults();
	}
	themeMatch(e) {
		return this._theme.match(e);
	}
	grammarForScopeName(e, t, n, r, i) {
		if (!this._grammars.has(e)) {
			let a = this._rawGrammars.get(e);
			if (!a) return null;
			this._grammars.set(e, _t(e, a, t, n, r, i, this, this._onigLib));
		}
		return this._grammars.get(e);
	}
}, Dt = class {
	_options;
	_syncRegistry;
	_ensureGrammarCache;
	constructor(e) {
		this._options = e, this._syncRegistry = new Et(pe.createFromRawTheme(e.theme, e.colorMap), e.onigLib), this._ensureGrammarCache = /* @__PURE__ */ new Map();
	}
	dispose() {
		this._syncRegistry.dispose();
	}
	setTheme(e, t) {
		this._syncRegistry.setTheme(pe.createFromRawTheme(e, t));
	}
	getColorMap() {
		return this._syncRegistry.getColorMap();
	}
	loadGrammarWithEmbeddedLanguages(e, t, n) {
		return this.loadGrammarWithConfiguration(e, t, { embeddedLanguages: n });
	}
	loadGrammarWithConfiguration(e, t, n) {
		return this._loadGrammar(e, t, n.embeddedLanguages, n.tokenTypes, new wt(n.balancedBracketSelectors || [], n.unbalancedBracketSelectors || []));
	}
	loadGrammar(e) {
		return this._loadGrammar(e, 0, null, null, null);
	}
	_loadGrammar(e, t, n, r, i) {
		let a = new Ne(this._syncRegistry, e);
		for (; a.Q.length > 0;) a.Q.map((e) => this._loadSingleGrammar(e.scopeName)), a.processQueue();
		return this._grammarForScopeName(e, t, n, r, i);
	}
	_loadSingleGrammar(e) {
		this._ensureGrammarCache.has(e) || (this._doLoadSingleGrammar(e), this._ensureGrammarCache.set(e, !0));
	}
	_doLoadSingleGrammar(e) {
		let t = this._options.loadGrammar(e);
		if (t) {
			let n = typeof this._options.getInjections == "function" ? this._options.getInjections(e) : void 0;
			this._syncRegistry.addGrammar(t, n);
		}
	}
	addGrammar(e, t = [], n = 0, r = null) {
		return this._syncRegistry.addGrammar(e, t), this._grammarForScopeName(e.scopeName, n, r);
	}
	_grammarForScopeName(e, t = 0, n = null, r = null, i = null) {
		return this._syncRegistry.grammarForScopeName(e, t, n, r, i);
	}
}, Ot = Ct.NULL, kt = /^>|^->|<!--|-->|--!>|<!-$/g, At = [">"], jt = ["<", ">"];
function Mt(e, t, n, r) {
	return r.settings.bogusComments ? "<?" + f(e.value, Object.assign({}, r.settings.characterReferences, { subset: At })) + ">" : "<!--" + e.value.replace(kt, i) + "-->";
	function i(e) {
		return f(e, Object.assign({}, r.settings.characterReferences, { subset: jt }));
	}
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/hast-util-to-html/9.0.5/d974b148a338f62647e86468060113373485d7b344318d975a7e678af9482f05/node_modules/hast-util-to-html/lib/handle/doctype.js
function Nt(e, t, n, r) {
	return "<!" + (r.settings.upperDoctype ? "DOCTYPE" : "doctype") + (r.settings.tightDoctype ? "" : " ") + "html>";
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/hast-util-to-html/9.0.5/d974b148a338f62647e86468060113373485d7b344318d975a7e678af9482f05/node_modules/hast-util-to-html/lib/omission/util/siblings.js
var w = It(1), Pt = It(-1), Ft = [];
function It(e) {
	return t;
	function t(t, n, r) {
		let i = t ? t.children : Ft, a = (n || 0) + e, o = i[a];
		if (!r) for (; o && c(o);) a += e, o = i[a];
		return o;
	}
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/hast-util-to-html/9.0.5/d974b148a338f62647e86468060113373485d7b344318d975a7e678af9482f05/node_modules/hast-util-to-html/lib/omission/omission.js
var Lt = {}.hasOwnProperty;
function Rt(e) {
	return t;
	function t(t, n, r) {
		return Lt.call(e, t.tagName) && e[t.tagName](t, n, r);
	}
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/hast-util-to-html/9.0.5/d974b148a338f62647e86468060113373485d7b344318d975a7e678af9482f05/node_modules/hast-util-to-html/lib/omission/closing.js
var zt = Rt({
	body: Ht,
	caption: Bt,
	colgroup: Bt,
	dd: Kt,
	dt: Gt,
	head: Bt,
	html: Vt,
	li: Wt,
	optgroup: Jt,
	option: Yt,
	p: Ut,
	rp: qt,
	rt: qt,
	tbody: Zt,
	td: en,
	tfoot: Qt,
	th: en,
	thead: Xt,
	tr: $t
});
function Bt(e, t, n) {
	let r = w(n, t, !0);
	return !r || r.type !== "comment" && !(r.type === "text" && c(r.value.charAt(0)));
}
function Vt(e, t, n) {
	let r = w(n, t);
	return !r || r.type !== "comment";
}
function Ht(e, t, n) {
	let r = w(n, t);
	return !r || r.type !== "comment";
}
function Ut(e, t, n) {
	let r = w(n, t);
	return r ? r.type === "element" && (r.tagName === "address" || r.tagName === "article" || r.tagName === "aside" || r.tagName === "blockquote" || r.tagName === "details" || r.tagName === "div" || r.tagName === "dl" || r.tagName === "fieldset" || r.tagName === "figcaption" || r.tagName === "figure" || r.tagName === "footer" || r.tagName === "form" || r.tagName === "h1" || r.tagName === "h2" || r.tagName === "h3" || r.tagName === "h4" || r.tagName === "h5" || r.tagName === "h6" || r.tagName === "header" || r.tagName === "hgroup" || r.tagName === "hr" || r.tagName === "main" || r.tagName === "menu" || r.tagName === "nav" || r.tagName === "ol" || r.tagName === "p" || r.tagName === "pre" || r.tagName === "section" || r.tagName === "table" || r.tagName === "ul") : !n || n.type !== "element" || n.tagName !== "a" && n.tagName !== "audio" && n.tagName !== "del" && n.tagName !== "ins" && n.tagName !== "map" && n.tagName !== "noscript" && n.tagName !== "video";
}
function Wt(e, t, n) {
	let r = w(n, t);
	return !r || r.type === "element" && r.tagName === "li";
}
function Gt(e, t, n) {
	let r = w(n, t);
	return !(!r || r.type !== "element" || r.tagName !== "dt" && r.tagName !== "dd");
}
function Kt(e, t, n) {
	let r = w(n, t);
	return !r || r.type === "element" && (r.tagName === "dt" || r.tagName === "dd");
}
function qt(e, t, n) {
	let r = w(n, t);
	return !r || r.type === "element" && (r.tagName === "rp" || r.tagName === "rt");
}
function Jt(e, t, n) {
	let r = w(n, t);
	return !r || r.type === "element" && r.tagName === "optgroup";
}
function Yt(e, t, n) {
	let r = w(n, t);
	return !r || r.type === "element" && (r.tagName === "option" || r.tagName === "optgroup");
}
function Xt(e, t, n) {
	let r = w(n, t);
	return !(!r || r.type !== "element" || r.tagName !== "tbody" && r.tagName !== "tfoot");
}
function Zt(e, t, n) {
	let r = w(n, t);
	return !r || r.type === "element" && (r.tagName === "tbody" || r.tagName === "tfoot");
}
function Qt(e, t, n) {
	return !w(n, t);
}
function $t(e, t, n) {
	let r = w(n, t);
	return !r || r.type === "element" && r.tagName === "tr";
}
function en(e, t, n) {
	let r = w(n, t);
	return !r || r.type === "element" && (r.tagName === "td" || r.tagName === "th");
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/hast-util-to-html/9.0.5/d974b148a338f62647e86468060113373485d7b344318d975a7e678af9482f05/node_modules/hast-util-to-html/lib/omission/opening.js
var tn = Rt({
	body: an,
	colgroup: on,
	head: rn,
	html: nn,
	tbody: sn
});
function nn(e) {
	let t = w(e, -1);
	return !t || t.type !== "comment";
}
function rn(e) {
	let t = /* @__PURE__ */ new Set();
	for (let n of e.children) if (n.type === "element" && (n.tagName === "base" || n.tagName === "title")) {
		if (t.has(n.tagName)) return !1;
		t.add(n.tagName);
	}
	let n = e.children[0];
	return !n || n.type === "element";
}
function an(e) {
	let t = w(e, -1, !0);
	return !t || t.type !== "comment" && !(t.type === "text" && c(t.value.charAt(0))) && (t.type !== "element" || t.tagName !== "meta" && t.tagName !== "link" && t.tagName !== "script" && t.tagName !== "style" && t.tagName !== "template");
}
function on(e, t, n) {
	let r = Pt(n, t), i = w(e, -1, !0);
	return n && r && r.type === "element" && r.tagName === "colgroup" && zt(r, n.children.indexOf(r), n) ? !1 : !!(i && i.type === "element" && i.tagName === "col");
}
function sn(e, t, n) {
	let r = Pt(n, t), i = w(e, -1);
	return n && r && r.type === "element" && (r.tagName === "thead" || r.tagName === "tbody") && zt(r, n.children.indexOf(r), n) ? !1 : !!(i && i.type === "element" && i.tagName === "tr");
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/hast-util-to-html/9.0.5/d974b148a338f62647e86468060113373485d7b344318d975a7e678af9482f05/node_modules/hast-util-to-html/lib/handle/element.js
var cn = {
	name: [["	\n\f\r &/=>".split(""), "	\n\f\r \"&'/=>`".split("")], ["\0	\n\f\r \"&'/<=>".split(""), "\0	\n\f\r \"&'/<=>`".split("")]],
	unquoted: [["	\n\f\r &>".split(""), "\0	\n\f\r \"&'<=>`".split("")], ["\0	\n\f\r \"&'<=>`".split(""), "\0	\n\f\r \"&'<=>`".split("")]],
	single: [["&'".split(""), "\"&'`".split("")], ["\0&'".split(""), "\0\"&'`".split("")]],
	double: [["\"&".split(""), "\"&'`".split("")], ["\0\"&".split(""), "\0\"&'`".split("")]]
};
function ln(e, t, n, r) {
	let i = r.schema, a = i.space !== "svg" && r.settings.omitOptionalTags, o = i.space === "svg" ? r.settings.closeEmptyElements : r.settings.voids.includes(e.tagName.toLowerCase()), c = [], l;
	i.space === "html" && e.tagName === "svg" && (r.schema = s);
	let u = un(r, e.properties), d = r.all(i.space === "html" && e.tagName === "template" ? e.content : e);
	return r.schema = i, d && (o = !1), !u && a && tn(e, t, n) || (c.push("<", e.tagName, u ? " " + u : ""), o && (i.space === "svg" || r.settings.closeSelfClosing) && (l = u.charAt(u.length - 1), (!r.settings.tightSelfClosing || l === "/" || l && l !== "\"" && l !== "'") && c.push(" "), c.push("/")), c.push(">")), c.push(d), o || a && zt(e, t, n) || c.push("</" + e.tagName + ">"), c.join("");
}
function un(e, t) {
	let n = [], r = -1, i;
	if (t) {
		for (i in t) if (t[i] !== null && t[i] !== void 0) {
			let r = dn(e, i, t[i]);
			r && n.push(r);
		}
	}
	for (; ++r < n.length;) {
		let t = e.settings.tightAttributes ? n[r].charAt(n[r].length - 1) : void 0;
		r !== n.length - 1 && t !== "\"" && t !== "'" && (n[r] += " ");
	}
	return n.join("");
}
function dn(e, t, n) {
	let a = i(e.schema, t), s = e.settings.allowParseErrors && e.schema.space === "html" ? 0 : 1, c = +!e.settings.allowDangerousCharacters, l = e.quote, u;
	if (a.overloadedBoolean && (n === a.attribute || n === "") ? n = !0 : (a.boolean || a.overloadedBoolean) && (typeof n != "string" || n === a.attribute || n === "") && (n = !!n), n == null || n === !1 || typeof n == "number" && Number.isNaN(n)) return "";
	let p = f(a.attribute, Object.assign({}, e.settings.characterReferences, { subset: cn.name[s][c] }));
	return n === !0 || (n = Array.isArray(n) ? (a.commaSeparated ? r : o)(n, { padLeft: !e.settings.tightCommaSeparatedLists }) : String(n), e.settings.collapseEmptyAttributes && !n) ? p : (e.settings.preferUnquoted && (u = f(n, Object.assign({}, e.settings.characterReferences, {
		attribute: !0,
		subset: cn.unquoted[s][c]
	}))), u !== n && (e.settings.quoteSmart && d(n, l) > d(n, e.alternative) && (l = e.alternative), u = l + f(n, Object.assign({}, e.settings.characterReferences, {
		subset: (l === "'" ? cn.single : cn.double)[s][c],
		attribute: !0
	})) + l), p + (u && "=" + u));
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/hast-util-to-html/9.0.5/d974b148a338f62647e86468060113373485d7b344318d975a7e678af9482f05/node_modules/hast-util-to-html/lib/handle/text.js
var fn = ["<", "&"];
function pn(e, t, n, r) {
	return n && n.type === "element" && (n.tagName === "script" || n.tagName === "style") ? e.value : f(e.value, Object.assign({}, r.settings.characterReferences, { subset: fn }));
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/hast-util-to-html/9.0.5/d974b148a338f62647e86468060113373485d7b344318d975a7e678af9482f05/node_modules/hast-util-to-html/lib/handle/raw.js
function mn(e, t, n, r) {
	return r.settings.allowDangerousHtml ? e.value : pn(e, t, n, r);
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/hast-util-to-html/9.0.5/d974b148a338f62647e86468060113373485d7b344318d975a7e678af9482f05/node_modules/hast-util-to-html/lib/handle/root.js
function hn(e, t, n, r) {
	return r.all(e);
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/hast-util-to-html/9.0.5/d974b148a338f62647e86468060113373485d7b344318d975a7e678af9482f05/node_modules/hast-util-to-html/lib/handle/index.js
var gn = u("type", {
	invalid: _n,
	unknown: vn,
	handlers: {
		comment: Mt,
		doctype: Nt,
		element: ln,
		raw: mn,
		root: hn,
		text: pn
	}
});
function _n(e) {
	throw Error("Expected node, not `" + e + "`");
}
function vn(e) {
	throw Error("Cannot compile unknown node `" + e.type + "`");
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/hast-util-to-html/9.0.5/d974b148a338f62647e86468060113373485d7b344318d975a7e678af9482f05/node_modules/hast-util-to-html/lib/index.js
var yn = {}, bn = {}, xn = [];
function Sn(e, t) {
	let n = t || yn, r = n.quote || "\"", i = r === "\"" ? "'" : "\"";
	if (r !== "\"" && r !== "'") throw Error("Invalid quote `" + r + "`, expected `'` or `\"`");
	return {
		one: Cn,
		all: wn,
		settings: {
			omitOptionalTags: n.omitOptionalTags || !1,
			allowParseErrors: n.allowParseErrors || !1,
			allowDangerousCharacters: n.allowDangerousCharacters || !1,
			quoteSmart: n.quoteSmart || !1,
			preferUnquoted: n.preferUnquoted || !1,
			tightAttributes: n.tightAttributes || !1,
			upperDoctype: n.upperDoctype || !1,
			tightDoctype: n.tightDoctype || !1,
			bogusComments: n.bogusComments || !1,
			tightCommaSeparatedLists: n.tightCommaSeparatedLists || !1,
			tightSelfClosing: n.tightSelfClosing || !1,
			collapseEmptyAttributes: n.collapseEmptyAttributes || !1,
			allowDangerousHtml: n.allowDangerousHtml || !1,
			voids: n.voids || l,
			characterReferences: n.characterReferences || bn,
			closeSelfClosing: n.closeSelfClosing || !1,
			closeEmptyElements: n.closeEmptyElements || !1
		},
		schema: n.space === "svg" ? s : a,
		quote: r,
		alternative: i
	}.one(Array.isArray(e) ? {
		type: "root",
		children: e
	} : e, void 0, void 0);
}
function Cn(e, t, n) {
	return gn(e, t, n, this);
}
function wn(e) {
	let t = [], n = e && e.children || xn, r = -1;
	for (; ++r < n.length;) t[r] = this.one(n[r], r, e);
	return t.join("");
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/emoji-regex-xs/1.0.0/f0b704780b707f304abcee547c33a21a98117d56d8ce06e82152208a0836df4e/node_modules/emoji-regex-xs/index.mjs
var Tn = String.raw, En = Tn`(?:\p{Emoji}\uFE0F\u20E3?|\p{Emoji_Modifier_Base}\p{Emoji_Modifier}?|\p{Emoji_Presentation})`, Dn = Tn`\u{E0061}-\u{E007A}`, On = () => new RegExp(Tn`[\u{1F1E6}-\u{1F1FF}]{2}|\u{1F3F4}[${Dn}]{2}[\u{E0030}-\u{E0039}${Dn}]{1,3}\u{E007F}|${En}(?:\u200D${En})*`, "gu"), T = Object.freeze({
	DEFAULT: "DEFAULT",
	CHAR_CLASS: "CHAR_CLASS"
});
function kn(e, t, n, r) {
	let i = new RegExp(String.raw`${t}|(?<$skip>\[\^?|\\?.)`, "gsu"), a = [!1], o = 0, s = "";
	for (let t of e.matchAll(i)) {
		let { 0: e, groups: { $skip: i } } = t;
		!i && (!r || r === T.DEFAULT == !o) ? n instanceof Function ? s += n(t, {
			context: o ? T.CHAR_CLASS : T.DEFAULT,
			negated: a[a.length - 1]
		}) : s += n : (e[0] === "[" ? (o++, a.push(e[1] === "^")) : e === "]" && o && (o--, a.pop()), s += e);
	}
	return s;
}
function An(e, t, n, r) {
	kn(e, t, n, r);
}
function jn(e, t, n = 0, r) {
	if (!new RegExp(t, "su").test(e)) return null;
	let i = RegExp(`${t}|(?<$skip>\\\\?.)`, "gsu");
	i.lastIndex = n;
	let a = 0, o;
	for (; o = i.exec(e);) {
		let { 0: e, groups: { $skip: t } } = o;
		if (!t && (!r || r === T.DEFAULT == !a)) return o;
		e === "[" ? a++ : e === "]" && a && a--, i.lastIndex == o.index && i.lastIndex++;
	}
	return null;
}
function Mn(e, t, n) {
	return !!jn(e, t, 0, n);
}
function Nn(e, t) {
	let n = /\\?./gsu;
	n.lastIndex = t;
	let r = e.length, i = 0, a = 1, o;
	for (; o = n.exec(e);) {
		let [e] = o;
		if (e === "[") i++;
		else if (i) e === "]" && i--;
		else if (e === "(") a++;
		else if (e === ")" && (a--, !a)) {
			r = o.index;
			break;
		}
	}
	return e.slice(t, r);
}
var Pn = class e extends RegExp {
	_captureMap;
	_namesByIndex;
	constructor(t, n, r) {
		if (t instanceof RegExp && r) throw Error("Cannot provide options when copying a regexp");
		let i = !!r?.useEmulationGroups, a = i ? Fn(t) : null;
		super(a?.expression || t, n);
		let o = i ? a : t instanceof e ? t : null;
		o && (this._captureMap = o._captureMap, this._namesByIndex = o._namesByIndex);
	}
	exec(e) {
		let t = RegExp.prototype.exec.call(this, e);
		if (!t || !this._captureMap) return t;
		let n = [...t];
		t.length = 1;
		let r;
		this.hasIndices && (r = [...t.indices], t.indices.length = 1);
		for (let e = 1; e < n.length; e++) if (this._captureMap[e].exclude) {
			let i = this._captureMap[e].transfer;
			if (i && t.length > i) {
				t[i] = n[e];
				let a = this._namesByIndex[i];
				a && (t.groups[a] = n[e], this.hasIndices && (t.indices.groups[a] = r[e])), this.hasIndices && (t.indices[i] = r[e]);
			}
		} else t.push(n[e]), this.hasIndices && t.indices.push(r[e]);
		return t;
	}
};
function Fn(e) {
	let t = "$E$".replace(/\$/g, "\\$"), n = [{ exclude: !1 }], r = { 0: "" }, i = 0;
	return e = kn(e, String.raw`\((?:(?!\?)|\?<(?![=!])(?<name>[^>]+)>)(?<mark>(?:\$(?<transfer>[1-9]\d*))?${t})?`, ({ 0: e, groups: { name: t, mark: a, transfer: o } }) => a ? (n.push({
		exclude: !0,
		transfer: o && +o
	}), e.slice(0, -a.length)) : (i++, t && (r[i] = t), n.push({ exclude: !1 }), e), T.DEFAULT), {
		_captureMap: n,
		_namesByIndex: r,
		expression: e
	};
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/regex/5.1.1/e2c83eb20fd1e3ac29759eb945405ec8ba419afc7adf9a8209ca49c1a18b8a03/node_modules/regex/src/utils-internals.js
var In = String.raw`\(\?(?:[:=!>A-Za-z\-]|<[=!]|\(DEFINE\))`;
function Ln(e, t, n, r) {
	return e.slice(0, t) + r + e.slice(t + n.length);
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/regex/5.1.1/e2c83eb20fd1e3ac29759eb945405ec8ba419afc7adf9a8209ca49c1a18b8a03/node_modules/regex/src/atomic.js
var Rn = new RegExp(String.raw`(?<noncapturingStart>${In})|(?<capturingStart>\((?:\?<[^>]+>)?)|\\?.`, "gsu");
function zn(e, t) {
	if (!/\(\?>/.test(e)) return e;
	let n = `(?:(?=(${t?.useEmulationGroups ? "$E$" : ""}`, r = [0], i = 0, a = 0, o = NaN, s;
	do {
		s = !1;
		let t = 0, c = 0, l = !1, u;
		for (Rn.lastIndex = Number.isNaN(o) ? 0 : o + n.length; u = Rn.exec(e);) {
			let { 0: d, index: f, groups: { capturingStart: p, noncapturingStart: m } } = u;
			if (d === "[") t++;
			else if (t) d === "]" && t--;
			else if (d === "(?>" && !l) o = f, l = !0;
			else if (l && m) c++;
			else if (p) l ? c++ : (i++, r.push(i + a));
			else if (d === ")" && l) {
				if (!c) {
					a++, e = `${e.slice(0, o)}${n}${e.slice(o + 3, f)}))<$$${a + i}>)${e.slice(f + 1)}`, s = !0;
					break;
				}
				c--;
			}
		}
	} while (s);
	return e = kn(e, String.raw`\\(?<backrefNum>[1-9]\d*)|<\$\$(?<wrappedBackrefNum>\d+)>`, ({ 0: e, groups: { backrefNum: t, wrappedBackrefNum: n } }) => {
		if (t) {
			let n = +t;
			if (n > r.length - 1) throw Error(`Backref "${e}" greater than number of captures`);
			return `\\${r[n]}`;
		}
		return `\\${n}`;
	}, T.DEFAULT), e;
}
var Bn = String.raw`(?:[?*+]|\{\d+(?:,\d*)?\})`, Vn = new RegExp(String.raw`
\\(?: \d+
  | c[A-Za-z]
  | [gk]<[^>]+>
  | [pPu]\{[^\}]+\}
  | u[A-Fa-f\d]{4}
  | x[A-Fa-f\d]{2}
  )
| \((?: \? (?: [:=!>]
  | <(?:[=!]|[^>]+>)
  | [A-Za-z\-]+:
  | \(DEFINE\)
  ))?
| (?<qBase>${Bn})(?<qMod>[?+]?)(?<invalidQ>[?*+\{]?)
| \\?.
`.replace(/\s+/g, ""), "gsu");
function Hn(e) {
	if (!RegExp(`${Bn}\\+`).test(e)) return e;
	let t = [], n = null, r = null, i = "", a = 0, o;
	for (Vn.lastIndex = 0; o = Vn.exec(e);) {
		let { 0: s, index: c, groups: { qBase: l, qMod: u, invalidQ: d } } = o;
		if (s === "[") a || (r = c), a++;
		else if (s === "]") a ? a-- : r = null;
		else if (!a) {
			if (u === "+" && i && !i.startsWith("(")) {
				if (d) throw Error(`Invalid quantifier "${s}"`);
				let t = -1;
				if (/^\{\d+\}$/.test(l)) e = Ln(e, c + l.length, u, "");
				else {
					if (i === ")" || i === "]") {
						let t = i === ")" ? n : r;
						if (t === null) throw Error(`Invalid unmatched "${i}"`);
						e = `${e.slice(0, t)}(?>${e.slice(t, c)}${l})${e.slice(c + s.length)}`;
					} else e = `${e.slice(0, c - i.length)}(?>${i}${l})${e.slice(c + s.length)}`;
					t += 4;
				}
				Vn.lastIndex += t;
			} else s[0] === "(" ? t.push(c) : s === ")" && (n = t.length ? t.pop() : null);
		}
		i = s;
	}
	return e;
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/regex-recursion/5.1.1/8c7c371f701a702af1f9a4116ed8267d30505b305545d80d8163ad2ec3fba86b/node_modules/regex-recursion/src/index.js
var E = String.raw, Un = E`\(\?R=(?<rDepth>[^\)]+)\)|${E`\\g<(?<gRNameOrNum>[^>&]+)&R=(?<gRDepth>[^>]+)>`}`, Wn = E`\(\?<(?![=!])(?<captureName>[^>]+)>`, D = new RegExp(E`${Wn}|${Un}|\(\?|\\?.`, "gsu"), Gn = "Cannot use multiple overlapping recursions", Kn = new RegExp(E`(?:\$[1-9]\d*)?${"$E$".replace(/\$/g, E`\$`)}`, "y");
function qn(e, t) {
	if (!new RegExp(Un, "su").test(e)) return e;
	if (Mn(e, E`\(\?\(DEFINE\)`, T.DEFAULT)) throw Error("DEFINE groups cannot be used with recursion");
	let n = !!t?.useEmulationGroups, r = Mn(e, E`\\[1-9]`, T.DEFAULT), i = /* @__PURE__ */ new Map(), a = [], o = !1, s = 0, c = 0, l;
	for (D.lastIndex = 0; l = D.exec(e);) {
		let { 0: t, groups: { captureName: u, rDepth: d, gRNameOrNum: f, gRDepth: p } } = l;
		if (t === "[") s++;
		else if (s) t === "]" && s--;
		else {
			if (d) {
				if (Jn(d), o) throw Error(Gn);
				if (r) throw Error("Numbered backrefs cannot be used with global recursion");
				let t = e.slice(0, l.index), i = e.slice(D.lastIndex);
				if (Mn(i, Un, T.DEFAULT)) throw Error(Gn);
				return Yn(t, i, +d, !1, n);
			}
			if (f) {
				Jn(p);
				let s = !1;
				for (let e of a) if (e.name === f || e.num === +f) {
					if (s = !0, e.hasRecursedWithin) throw Error(Gn);
					break;
				}
				if (!s) throw Error(E`Recursive \g cannot be used outside the referenced group "\g<${f}&R=${p}>"`);
				let c = i.get(f), u = Nn(e, c);
				if (r && Mn(u, E`${Wn}|\((?!\?)`, T.DEFAULT)) throw Error("Numbered backrefs cannot be used with recursion of capturing groups");
				let d = e.slice(c, l.index), m = u.slice(d.length + t.length), h = Yn(d, m, +p, !0, n);
				e = `${e.slice(0, c)}${h}${e.slice(c + u.length)}`, D.lastIndex += h.length - t.length - d.length - m.length, a.forEach((e) => e.hasRecursedWithin = !0), o = !0;
			} else if (u) c++, i.set(String(c), D.lastIndex), i.set(u, D.lastIndex), a.push({
				num: c,
				name: u
			});
			else if (t.startsWith("(")) {
				let r = t === "(";
				r && (c++, i.set(String(c), D.lastIndex + (n ? Zn(e, D.lastIndex) : 0))), a.push(r ? { num: c } : {});
			} else t === ")" && a.pop();
		}
	}
	return e;
}
function Jn(e) {
	let t = `Max depth must be integer between 2 and 100; used ${e}`;
	if (!/^[1-9]\d*$/.test(e) || (e = +e, e < 2 || e > 100)) throw Error(t);
}
function Yn(e, t, n, r, i) {
	let a = /* @__PURE__ */ new Set();
	r && An(e + t, Wn, ({ groups: { captureName: e } }) => {
		a.add(e);
	}, T.DEFAULT);
	let o = n - 1;
	return `${e}${Xn(`(?:${e}`, o, r ? a : null, "forward", i)}(?:)${Xn(`${t})`, o, r ? a : null, "backward", i)}${t}`;
}
function Xn(e, t, n, r, i) {
	let a = (e) => r === "backward" ? t - e + 2 - 1 : e + 2, o = "";
	for (let r = 0; r < t; r++) {
		let t = a(r);
		o += kn(e, E`${Wn}|\\k<(?<backref>[^>]+)>${i ? E`|(?<unnamed>\()(?!\?)(?:${Kn.source})?` : ""}`, ({ 0: e, index: r, groups: { captureName: a, backref: o, unnamed: s } }) => {
			if (o && n && !n.has(o)) return e;
			if (s) return "($E$";
			let c = `_$${t}`;
			return a ? `(?<${a}${c}>${i ? "$E$" : ""}` : E`\k<${o}${c}>`;
		}, T.DEFAULT);
	}
	return o;
}
function Zn(e, t) {
	Kn.lastIndex = t;
	let n = Kn.exec(e);
	return n ? n[0].length : 0;
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/oniguruma-to-es/2.3.0/4823180bb5974976a3d462055f3cfa576691e21d31f4a5b86d0528fd1db4f921/node_modules/oniguruma-to-es/dist/esm/index.js
var O = String.fromCodePoint, k = String.raw, Qn = (() => {
	try {
		RegExp("(?i:)");
	} catch {
		return !1;
	}
	return !0;
})(), $n = (() => {
	try {
		RegExp("", "v");
	} catch {
		return !1;
	}
	return !0;
})();
function er(e, { enable: t, disable: n }) {
	return {
		dotAll: !n?.dotAll && !!(t?.dotAll || e.dotAll),
		ignoreCase: !n?.ignoreCase && !!(t?.ignoreCase || e.ignoreCase)
	};
}
function tr(e, t, n) {
	return e.has(t) || e.set(t, n), e.get(t);
}
function nr(e, t) {
	return rr[e] >= rr[t];
}
function A(e, t) {
	if (!e) throw Error(t ?? "Value expected");
	return e;
}
var rr = {
	ES2025: 2025,
	ES2024: 2024,
	ES2018: 2018
}, ir = {
	auto: "auto",
	ES2025: "ES2025",
	ES2024: "ES2024",
	ES2018: "ES2018"
};
function ar(e) {
	if (e?.target !== void 0 && !ir[e.target]) throw Error(`Unexpected target "${e.target}"`);
	let t = {
		accuracy: "default",
		avoidSubclass: !1,
		flags: "",
		global: !1,
		hasIndices: !1,
		target: "auto",
		verbose: !1,
		...e,
		rules: {
			allowOrphanBackrefs: !1,
			asciiWordBoundaries: !1,
			captureGroup: !1,
			recursionLimit: 20,
			singleline: !1,
			...e?.rules
		}
	};
	return t.target === "auto" && (t.target = Qn ? "ES2025" : $n ? "ES2024" : "ES2018"), t;
}
var or = /* @__PURE__ */ new Set([O(304), O(305)]);
function sr(e) {
	if (or.has(e)) return [e];
	let t = /* @__PURE__ */ new Set(), n = e.toLowerCase(), r = n.toUpperCase(), i = mr.get(n), a = fr.get(n), o = pr.get(n);
	return [...r].length === 1 && t.add(r), o && t.add(o), i && t.add(i), t.add(n), a && t.add(a), [...t];
}
var cr = new Set("C Other\nCc Control cntrl\nCf Format\nCn Unassigned\nCo Private_Use\nCs Surrogate\nL Letter\nLC Cased_Letter\nLl Lowercase_Letter\nLm Modifier_Letter\nLo Other_Letter\nLt Titlecase_Letter\nLu Uppercase_Letter\nM Mark Combining_Mark\nMc Spacing_Mark\nMe Enclosing_Mark\nMn Nonspacing_Mark\nN Number\nNd Decimal_Number digit\nNl Letter_Number\nNo Other_Number\nP Punctuation punct\nPc Connector_Punctuation\nPd Dash_Punctuation\nPe Close_Punctuation\nPf Final_Punctuation\nPi Initial_Punctuation\nPo Other_Punctuation\nPs Open_Punctuation\nS Symbol\nSc Currency_Symbol\nSk Modifier_Symbol\nSm Math_Symbol\nSo Other_Symbol\nZ Separator\nZl Line_Separator\nZp Paragraph_Separator\nZs Space_Separator\nASCII\nASCII_Hex_Digit AHex\nAlphabetic Alpha\nAny\nAssigned\nBidi_Control Bidi_C\nBidi_Mirrored Bidi_M\nCase_Ignorable CI\nCased\nChanges_When_Casefolded CWCF\nChanges_When_Casemapped CWCM\nChanges_When_Lowercased CWL\nChanges_When_NFKC_Casefolded CWKCF\nChanges_When_Titlecased CWT\nChanges_When_Uppercased CWU\nDash\nDefault_Ignorable_Code_Point DI\nDeprecated Dep\nDiacritic Dia\nEmoji\nEmoji_Component EComp\nEmoji_Modifier EMod\nEmoji_Modifier_Base EBase\nEmoji_Presentation EPres\nExtended_Pictographic ExtPict\nExtender Ext\nGrapheme_Base Gr_Base\nGrapheme_Extend Gr_Ext\nHex_Digit Hex\nIDS_Binary_Operator IDSB\nIDS_Trinary_Operator IDST\nID_Continue IDC\nID_Start IDS\nIdeographic Ideo\nJoin_Control Join_C\nLogical_Order_Exception LOE\nLowercase Lower\nMath\nNoncharacter_Code_Point NChar\nPattern_Syntax Pat_Syn\nPattern_White_Space Pat_WS\nQuotation_Mark QMark\nRadical\nRegional_Indicator RI\nSentence_Terminal STerm\nSoft_Dotted SD\nTerminal_Punctuation Term\nUnified_Ideograph UIdeo\nUppercase Upper\nVariation_Selector VS\nWhite_Space space\nXID_Continue XIDC\nXID_Start XIDS".split(/\s/)), lr = /* @__PURE__ */ new Map();
for (let e of cr) lr.set(vr(e), e);
var ur = /* @__PURE__ */ new Set([
	"Basic_Emoji",
	"Emoji_Keycap_Sequence",
	"RGI_Emoji",
	"RGI_Emoji_Flag_Sequence",
	"RGI_Emoji_Modifier_Sequence",
	"RGI_Emoji_Tag_Sequence",
	"RGI_Emoji_ZWJ_Sequence"
]), dr = /* @__PURE__ */ new Map();
for (let e of ur) dr.set(vr(e), e);
var fr = /* @__PURE__ */ new Map([["s", O(383)], [O(383), "s"]]), pr = /* @__PURE__ */ new Map([
	[O(223), O(7838)],
	[O(107), O(8490)],
	[O(229), O(8491)],
	[O(969), O(8486)]
]), mr = new Map([
	j(453),
	j(456),
	j(459),
	j(498),
	...yr(8072, 8079),
	...yr(8088, 8095),
	...yr(8104, 8111),
	j(8124),
	j(8140),
	j(8188)
]), hr = /* @__PURE__ */ new Map([
	["alnum", k`[\p{Alpha}\p{Nd}]`],
	["alpha", k`\p{Alpha}`],
	["ascii", k`\p{ASCII}`],
	["blank", k`[\p{Zs}\t]`],
	["cntrl", k`\p{cntrl}`],
	["digit", k`\p{Nd}`],
	["graph", k`[\P{space}&&\P{cntrl}&&\P{Cn}&&\P{Cs}]`],
	["lower", k`\p{Lower}`],
	["print", k`[[\P{space}&&\P{cntrl}&&\P{Cn}&&\P{Cs}]\p{Zs}]`],
	["punct", k`[\p{P}\p{S}]`],
	["space", k`\p{space}`],
	["upper", k`\p{Upper}`],
	["word", k`[\p{Alpha}\p{M}\p{Nd}\p{Pc}]`],
	["xdigit", k`\p{AHex}`]
]), gr = /* @__PURE__ */ new Set([
	"alnum",
	"blank",
	"graph",
	"print",
	"word",
	"xdigit"
]);
function _r(e, t) {
	let n = [];
	for (let r = e; r <= t; r++) n.push(r);
	return n;
}
function vr(e) {
	return e.replace(/[- _]+/g, "").toLowerCase();
}
function j(e) {
	let t = O(e);
	return [t.toLowerCase(), t];
}
function yr(e, t) {
	return _r(e, t).map((e) => j(e));
}
var br = /* @__PURE__ */ new Set([
	"Lower",
	"Lowercase",
	"Upper",
	"Uppercase",
	"Ll",
	"Lowercase_Letter",
	"Lt",
	"Titlecase_Letter",
	"Lu",
	"Uppercase_Letter"
]), M = {
	Alternator: "Alternator",
	Assertion: "Assertion",
	Backreference: "Backreference",
	Character: "Character",
	CharacterClassClose: "CharacterClassClose",
	CharacterClassHyphen: "CharacterClassHyphen",
	CharacterClassIntersector: "CharacterClassIntersector",
	CharacterClassOpen: "CharacterClassOpen",
	CharacterSet: "CharacterSet",
	Directive: "Directive",
	GroupClose: "GroupClose",
	GroupOpen: "GroupOpen",
	Subroutine: "Subroutine",
	Quantifier: "Quantifier",
	VariableLengthCharacterSet: "VariableLengthCharacterSet",
	EscapedNumber: "EscapedNumber"
}, N = {
	any: "any",
	digit: "digit",
	dot: "dot",
	hex: "hex",
	non_newline: "non_newline",
	posix: "posix",
	property: "property",
	space: "space",
	word: "word"
}, xr = {
	flags: "flags",
	keep: "keep"
}, P = {
	absent_repeater: "absent_repeater",
	atomic: "atomic",
	capturing: "capturing",
	group: "group",
	lookahead: "lookahead",
	lookbehind: "lookbehind"
}, Sr = /* @__PURE__ */ new Map([
	["a", 7],
	["b", 8],
	["e", 27],
	["f", 12],
	["n", 10],
	["r", 13],
	["t", 9],
	["v", 11]
]), Cr = k`\[\^?`, wr = `c.? | C(?:-.?)?|${k`[pP]\{(?:\^?[-\x20_]*[A-Za-z][-\x20\w]*\})?`}|${k`x[89A-Fa-f]\p{AHex}(?:\\x[89A-Fa-f]\p{AHex})*`}|${k`u(?:\p{AHex}{4})? | x\{[^\}]*\}? | x\p{AHex}{0,2}`}|${k`o\{[^\}]*\}?`}|${k`\d{1,3}`}`, Tr = /[?*+][?+]?|\{(?:\d+(?:,\d*)?|,\d+)\}\??/, Er = new RegExp(k`
  \\ (?:
    ${wr}
    | [gk]<[^>]*>?
    | [gk]'[^']*'?
    | .
  )
  | \( (?: \? (?:
    [:=!>(]
    | <[=!]
    | <[^>]*>
    | '[^']*'
    | ~\|?
    | # (?:[^)\\] | \\.?)*
    | [imx\-]+[:)]
  )?)?
  | ${Tr.source}
  | ${Cr}
  | .
`.replace(/\s+/g, ""), "gsu"), Dr = new RegExp(k`
  \\ (?:
    ${wr}
    | .
  )
  | \[:[^:]*:\]
  | ${Cr}
  | &&
  | .
`.replace(/\s+/g, ""), "gsu");
function Or(e, t = "", n) {
	if (n = {
		captureGroup: !1,
		singleline: !1,
		...n
	}, typeof e != "string") throw Error("String expected as pattern");
	if (!/^[imxDSW]*$/.test(t)) throw Error(`Flags "${t}" includes unsupported value`);
	let r = t.includes("x"), i = [r], a = {
		captureGroup: n.captureGroup,
		getCurrentModX: () => i.at(-1),
		numOpenGroups: 0,
		popModX() {
			i.pop();
		},
		pushModX(e) {
			i.push(e);
		},
		replaceCurrentModX(e) {
			i[i.length - 1] = e;
		},
		singleline: n.singleline
	}, o = [], s;
	for (Er.lastIndex = 0; s = Er.exec(e);) {
		let t = kr(a, e, s[0], Er.lastIndex);
		t.tokens ? o.push(...t.tokens) : t.token && o.push(t.token), t.lastIndex !== void 0 && (Er.lastIndex = t.lastIndex);
	}
	let c = [], l = 0;
	o.forEach((e) => {
		e.type === M.GroupOpen && (e.kind === P.capturing ? e.number = ++l : e.raw === "(" && c.push(e));
	}), l || c.forEach((e, t) => {
		e.kind = P.capturing, e.number = t + 1;
	});
	let u = l || c.length;
	return o = o.map((e) => e.type === M.EscapedNumber ? Br(e, u) : e).flat(), {
		tokens: o,
		flags: {
			ignoreCase: t.includes("i"),
			dotAll: t.includes("m"),
			extended: r,
			digitIsAscii: t.includes("D"),
			spaceIsAscii: t.includes("S"),
			wordIsAscii: t.includes("W")
		},
		rules: n
	};
}
function kr(e, t, n, r) {
	let [i, a, o] = n;
	if (i === "[") {
		let e = Ar(t, n, r);
		return {
			tokens: e.tokens,
			lastIndex: e.lastIndex
		};
	}
	if (i === "\\") {
		if ("AbBGzZ".includes(a)) return { token: F(M.Assertion, n, { kind: n }) };
		if (/^\\g[<']/.test(n)) {
			if (!/^\\g(?:<[^>]+>|'[^']+')$/.test(n)) throw Error(`Invalid group name "${n}"`);
			return { token: F(M.Subroutine, n) };
		}
		if (/^\\k[<']/.test(n)) {
			if (!/^\\k(?:<[^>]+>|'[^']+')$/.test(n)) throw Error(`Invalid group name "${n}"`);
			return { token: F(M.Backreference, n) };
		}
		if (a === "K") return { token: F(M.Directive, n, { kind: xr.keep }) };
		if (a === "N") return { token: F(M.CharacterSet, n, { kind: N.non_newline }) };
		if (a === "O") return { token: F(M.CharacterSet, n, { kind: N.any }) };
		if ("RX".includes(a)) return { token: F(M.VariableLengthCharacterSet, n, { kind: n }) };
		if ("yY".includes(a)) throw Error(`Unsupported grapheme boundary "${n}"`);
		let e = Mr(n, { inCharClass: !1 });
		return Array.isArray(e) ? { tokens: e } : { token: e };
	}
	if (i === "(") {
		if (o === "#") {
			if (t[r] !== ")") throw Error("Unclosed comment group \"(?#\"");
			return { lastIndex: r + 1 };
		}
		if ("-imx".includes(o)) return { token: Pr(n, e) };
		if (e.pushModX(e.getCurrentModX()), e.numOpenGroups++, n === "(" && !e.captureGroup || n === "(?:") return { token: F(M.GroupOpen, n, { kind: P.group }) };
		if (n === "(?>") return { token: F(M.GroupOpen, n, { kind: P.atomic }) };
		if (n === "(?=" || n === "(?!" || n === "(?<=" || n === "(?<!") return { token: F(M.GroupOpen, n, {
			kind: o === "<" ? P.lookbehind : P.lookahead,
			negate: n.endsWith("!")
		}) };
		if (o === "<" || o === "'" || n === "(" && e.captureGroup) {
			let e = F(M.GroupOpen, n, { kind: P.capturing });
			return n !== "(" && (e.name = n.slice(3, -1)), { token: e };
		}
		if (o === "~") {
			if (n === "(?~|") throw Error(`Unsupported absent function kind "${n}"`);
			return { token: F(M.GroupOpen, n, { kind: P.absent_repeater }) };
		}
		throw Error(o === "(" ? `Unsupported conditional "${n}"` : n === "(?" ? "Invalid group" : `Unexpected group "${n}"`);
	}
	if (n === ")") {
		if (e.popModX(), e.numOpenGroups--, e.numOpenGroups < 0) throw Error("Unmatched \")\"");
		return { token: F(M.GroupClose, n) };
	}
	if (n === "#" && e.getCurrentModX()) {
		let e = t.indexOf("\n", r);
		return { lastIndex: e === -1 ? t.length : e };
	}
	if (/^\s$/.test(n) && e.getCurrentModX()) {
		let e = /\s+/y;
		return e.lastIndex = r, { lastIndex: e.exec(t) ? e.lastIndex : r };
	}
	if (n === ".") return { token: F(M.CharacterSet, n, { kind: N.dot }) };
	if (n === "^" || n === "$") {
		let t = e.singleline ? {
			"^": k`\A`,
			$: k`\Z`
		}[n] : n;
		return { token: F(M.Assertion, n, { kind: t }) };
	}
	return n === "|" ? { token: F(M.Alternator, n) } : Tr.test(n) ? { token: Fr(n) } : (Vr(n), { token: F(M.Character, n, { value: n.codePointAt(0) }) });
}
function Ar(e, t, n) {
	let r = [F(M.CharacterClassOpen, t, { negate: t[1] === "^" })], i = 1, a;
	for (Dr.lastIndex = n; a = Dr.exec(e);) {
		let e = a[0];
		if (e[0] === "[" && e[1] !== ":") i++, r.push(F(M.CharacterClassOpen, e, { negate: e[1] === "^" }));
		else if (e === "]") {
			if (r.at(-1).type === M.CharacterClassOpen) r.push(F(M.Character, e, { value: 93 }));
			else if (i--, r.push(F(M.CharacterClassClose, e)), !i) break;
		} else {
			let t = jr(e);
			Array.isArray(t) ? r.push(...t) : r.push(t);
		}
	}
	return {
		tokens: r,
		lastIndex: Dr.lastIndex || e.length
	};
}
function jr(e) {
	if (e[0] === "\\") return Mr(e, { inCharClass: !0 });
	if (e[0] === "[") {
		let t = /\[:(?<negate>\^?)(?<name>[a-z]+):\]/.exec(e);
		if (!t || !hr.get(t.groups.name)) throw Error(`Invalid POSIX class "${e}"`);
		return F(M.CharacterSet, e, {
			kind: N.posix,
			negate: !!t.groups.negate,
			value: t.groups.name
		});
	}
	return e === "-" ? F(M.CharacterClassHyphen, e) : e === "&&" ? F(M.CharacterClassIntersector, e) : (Vr(e), F(M.Character, e, { value: e.codePointAt(0) }));
}
function Mr(e, { inCharClass: t }) {
	let n = e[1];
	if (n === "c" || n === "C") return Nr(e);
	if ("dDhHsSwW".includes(n)) return Ir(e);
	if (e.startsWith(k`\o{`)) throw Error(`Incomplete, invalid, or unsupported octal code point "${e}"`);
	if (/^\\[pP]\{/.test(e)) {
		if (e.length === 3) throw Error(`Incomplete or invalid Unicode property "${e}"`);
		return Lr(e);
	}
	if (/^\\x[89A-Fa-f]\p{AHex}/u.test(e)) try {
		let t = e.split(/\\x/).slice(1).map((e) => parseInt(e, 16)), n = new TextDecoder("utf-8", {
			ignoreBOM: !0,
			fatal: !0
		}).decode(new Uint8Array(t)), r = new TextEncoder();
		return [...n].map((e) => {
			let t = [...r.encode(e)].map((e) => `\\x${e.toString(16)}`).join("");
			return F(M.Character, t, { value: e.codePointAt(0) });
		});
	} catch {
		throw Error(`Multibyte code "${e}" incomplete or invalid in Oniguruma`);
	}
	if (n === "u" || n === "x") return F(M.Character, e, { value: zr(e) });
	if (Sr.has(n)) return F(M.Character, e, { value: Sr.get(n) });
	if (/\d/.test(n)) return F(M.EscapedNumber, e, { inCharClass: t });
	if (e === "\\") throw Error(k`Incomplete escape "\"`);
	if (n === "M") throw Error(`Unsupported meta "${e}"`);
	if ([...e].length === 2) return F(M.Character, e, { value: e.codePointAt(1) });
	throw Error(`Unexpected escape "${e}"`);
}
function F(e, t, n) {
	return {
		type: e,
		raw: t,
		...n
	};
}
function Nr(e) {
	let t = e[1] === "c" ? e[2] : e[3];
	if (!t || !/[A-Za-z]/.test(t)) throw Error(`Unsupported control character "${e}"`);
	return F(M.Character, e, { value: t.toUpperCase().codePointAt(0) - 64 });
}
function Pr(e, t) {
	let { on: n, off: r } = /^\(\?(?<on>[imx]*)(?:-(?<off>[imx\-]*))?/.exec(e).groups;
	r ??= "";
	let i = (t.getCurrentModX() || n.includes("x")) && !r.includes("x"), a = Rr(n), o = Rr(r), s = {};
	if (a && (s.enable = a), o && (s.disable = o), e.endsWith(")")) return t.replaceCurrentModX(i), F(M.Directive, e, {
		kind: xr.flags,
		flags: s
	});
	if (e.endsWith(":")) {
		t.pushModX(i), t.numOpenGroups++;
		let n = F(M.GroupOpen, e, { kind: P.group });
		return (a || o) && (n.flags = s), n;
	}
	throw Error(`Unexpected flag modifier "${e}"`);
}
function Fr(e) {
	let t = {};
	if (e[0] === "{") {
		let { min: n, max: r } = /^\{(?<min>\d*)(?:,(?<max>\d*))?/.exec(e).groups, i = 1e5;
		if (+n > i || +r > i) throw Error("Quantifier value unsupported in Oniguruma");
		t.min = +n, t.max = r === void 0 ? +n : r === "" ? Infinity : +r, t.greedy = !e.endsWith("?"), t.possessive = !1;
	} else t.min = +(e[0] === "+"), t.max = e[0] === "?" ? 1 : Infinity, t.greedy = e[1] !== "?", t.possessive = e[1] === "+";
	return F(M.Quantifier, e, t);
}
function Ir(e) {
	let t = e[1].toLowerCase();
	return F(M.CharacterSet, e, {
		kind: {
			d: N.digit,
			h: N.hex,
			s: N.space,
			w: N.word
		}[t],
		negate: e[1] !== t
	});
}
function Lr(e) {
	let { p: t, neg: n, value: r } = /^\\(?<p>[pP])\{(?<neg>\^?)(?<value>[^}]+)/.exec(e).groups, i = t === "P" && !n || t === "p" && !!n;
	return F(M.CharacterSet, e, {
		kind: N.property,
		negate: i,
		value: r
	});
}
function Rr(e) {
	let t = {};
	return e.includes("i") && (t.ignoreCase = !0), e.includes("m") && (t.dotAll = !0), e.includes("x") && (t.extended = !0), Object.keys(t).length ? t : null;
}
function zr(e) {
	if (/^(?:\\u(?!\p{AHex}{4})|\\x(?!\p{AHex}{1,2}|\{\p{AHex}{1,8}\}))/u.test(e)) throw Error(`Incomplete or invalid escape "${e}"`);
	let t = e[2] === "{" ? /^\\x\{\s*(?<hex>\p{AHex}+)/u.exec(e).groups.hex : e.slice(2);
	return parseInt(t, 16);
}
function Br(e, t) {
	let { raw: n, inCharClass: r } = e, i = n.slice(1);
	if (!r && (i !== "0" && i.length === 1 || i[0] !== "0" && +i <= t)) return [F(M.Backreference, n)];
	let a = [], o = i.match(/^[0-7]+|\d/g);
	for (let e = 0; e < o.length; e++) {
		let t = o[e], r;
		if (e === 0 && t !== "8" && t !== "9") {
			if (r = parseInt(t, 8), r > 127) throw Error(k`Octal encoded byte above 177 unsupported "${n}"`);
		} else r = t.codePointAt(0);
		a.push(F(M.Character, (e === 0 ? "\\" : "") + t, { value: r }));
	}
	return a;
}
function Vr(e) {
	if ([...e].length !== 1) throw Error(`Expected "${e}" to be a single code point`);
}
function Hr({ alternatives: e }, t) {
	return e.length === 1 && e[0].elements.length === 1 && (!t || t(e[0].elements[0]));
}
function Ur({ type: e }) {
	return e === R.Assertion || e === R.Directive;
}
function Wr(e) {
	let t = [
		R.Character,
		R.CharacterClass,
		R.CharacterSet
	];
	return t.includes(e.type) || e.type === R.Quantifier && e.min && t.includes(e.element.type);
}
function Gr({ type: e }) {
	return e === R.CapturingGroup || e === R.Group;
}
function I({ type: e, kind: t }) {
	return e === R.Assertion && (t === z.lookahead || t === z.lookbehind);
}
function L(e, t, n) {
	let r = e.node;
	for (; r.parent;) r = r.parent;
	function i(e, t) {
		for (let n = 0; n < e.length; n++) {
			let r = a(e[n], t, n, e);
			n = Math.max(-1, n + r);
		}
	}
	function a(e, o = null, s = null, c = null) {
		let l = 0, u = !1, d = {
			node: e,
			parent: o,
			key: s,
			container: c,
			ast: r,
			remove() {
				A(c, "Container expected").splice(Math.max(0, s + l), 1), --l;
			},
			removeAllNextSiblings() {
				return A(c, "Container expected").splice(s + 1);
			},
			removeAllPrevSiblings() {
				let e = s + l;
				return l -= e, A(c, "Container expected").splice(0, Math.max(0, e));
			},
			replaceWith(e) {
				Jr(e, o), c ? c[Math.max(0, s + l)] = e : o[s] = e;
			},
			skip() {
				u = !0;
			}
		}, f = qr(e).find((e) => !!n[e]), p = f && n[f], m = typeof p == "function" ? p : p?.enter, h = p?.exit;
		if (m?.(d, t), !u) switch (e.type) {
			case R.Regex:
				a(e.pattern, e, "pattern"), a(e.flags, e, "flags");
				break;
			case R.Alternative:
			case R.CharacterClass:
				i(e.elements, e);
				break;
			case R.Assertion:
				I(e) && i(e.alternatives, e);
				break;
			case R.Backreference:
			case R.Character:
			case R.CharacterSet:
			case R.Directive:
			case R.Flags:
			case R.Recursion:
			case R.Subroutine:
			case R.VariableLengthCharacterSet: break;
			case R.AbsentFunction:
			case R.CapturingGroup:
			case R.Group:
			case R.Pattern:
				i(e.alternatives, e);
				break;
			case R.CharacterClassIntersection:
				i(e.classes, e);
				break;
			case R.CharacterClassRange:
				a(e.min, e, "min"), a(e.max, e, "max");
				break;
			case R.Quantifier:
				a(e.element, e, "element");
				break;
			default: throw Error(`Unexpected node type "${e.type}"`);
		}
		return h?.(d, t), l;
	}
	a(e.node, e.parent, e.key, e.container);
}
var Kr = {
	AnyGroup: "AnyGroup",
	AnyNode: "AnyNode"
};
function qr(e) {
	let t = [Kr.AnyNode];
	return (Gr(e) || I(e)) && t.push(Kr.AnyGroup), t.push(e.type), t;
}
function Jr(e, t) {
	"parent" in t && (e.parent = t);
}
var R = {
	AbsentFunction: "AbsentFunction",
	Alternative: "Alternative",
	Assertion: "Assertion",
	Backreference: "Backreference",
	CapturingGroup: "CapturingGroup",
	Character: "Character",
	CharacterClass: "CharacterClass",
	CharacterClassIntersection: "CharacterClassIntersection",
	CharacterClassRange: "CharacterClassRange",
	CharacterSet: "CharacterSet",
	Directive: "Directive",
	Flags: "Flags",
	Group: "Group",
	Pattern: "Pattern",
	Quantifier: "Quantifier",
	Regex: "Regex",
	Subroutine: "Subroutine",
	VariableLengthCharacterSet: "VariableLengthCharacterSet",
	Recursion: "Recursion"
}, Yr = { repeater: "repeater" }, z = {
	line_end: "line_end",
	line_start: "line_start",
	lookahead: "lookahead",
	lookbehind: "lookbehind",
	search_start: "search_start",
	string_end: "string_end",
	string_end_newline: "string_end_newline",
	string_start: "string_start",
	word_boundary: "word_boundary"
}, B = N, Xr = xr, Zr = {
	grapheme: "grapheme",
	newline: "newline"
};
function Qr({ tokens: e, flags: t, rules: n }, r) {
	let i = {
		skipBackrefValidation: !1,
		skipLookbehindValidation: !1,
		skipPropertyNameValidation: !1,
		verbose: !1,
		...r
	}, a = {
		capturingGroups: [],
		current: 0,
		hasNumberedRef: !1,
		namedGroupsByName: /* @__PURE__ */ new Map(),
		parent: null,
		skipBackrefValidation: i.skipBackrefValidation,
		skipLookbehindValidation: i.skipLookbehindValidation,
		skipPropertyNameValidation: i.skipPropertyNameValidation,
		subroutines: [],
		token: null,
		tokens: e,
		verbose: i.verbose,
		walk: o
	};
	function o(t, n) {
		let r = e[a.current];
		switch (a.parent = t, a.token = r, a.current++, r.type) {
			case M.Alternator: return V();
			case M.Assertion: return ci(r);
			case M.Backreference: return $r(a);
			case M.Character: return fi(r.value, { useLastValid: !!n.isCheckingRangeEnd });
			case M.CharacterClassHyphen: return ei(a, n);
			case M.CharacterClassOpen: return ti(a, n);
			case M.CharacterSet: return ni(a);
			case M.Directive: return _i(r);
			case M.GroupOpen: return ri(a, n);
			case M.Quantifier: return ii(a);
			case M.Subroutine: return ai(a);
			case M.VariableLengthCharacterSet: return Ci(r.kind);
			default: throw Error(`Unexpected token type "${r.type}"`);
		}
	}
	let s = xi(yi(), vi(t)), c = s.pattern.alternatives[0];
	for (; a.current < e.length;) {
		let e = o(c, {});
		e.type === R.Alternative ? (s.pattern.alternatives.push(e), c = e) : c.elements.push(e);
	}
	let { capturingGroups: l, hasNumberedRef: u, namedGroupsByName: d, subroutines: f } = a;
	if (u && d.size && !n.captureGroup) throw Error("Numbered backref/subroutine not allowed when using named capture");
	for (let { ref: e } of f) if (typeof e == "number") {
		if (e > l.length) throw Error("Subroutine uses a group number that's not defined");
	} else if (!d.has(e)) throw Error(k`Subroutine uses a group name that's not defined "\g<${e}>"`);
	else if (d.get(e).length > 1) throw Error(k`Subroutine uses a duplicate group name "\g<${e}>"`);
	return L({ node: s }, null, { AnyNode({ node: e, parent: t }) {
		e.parent = t;
	} }), s;
}
function $r(e) {
	let { raw: t } = e.token, n = /^\\k[<']/.test(t), r = n ? t.slice(3, -1) : t.slice(1), i = (n, r = !1) => {
		let i = e.capturingGroups.length, a = !1;
		if (n > i) {
			if (e.skipBackrefValidation) a = !0;
			else throw Error(`Not enough capturing groups defined to the left "${t}"`);
		}
		return e.hasNumberedRef = !0, li(r ? i + 1 - n : n, { orphan: a });
	};
	if (n) {
		let n = /^(?<sign>-?)0*(?<num>[1-9]\d*)$/.exec(r);
		if (n) return i(+n.groups.num, !!n.groups.sign);
		if (/[-+]/.test(r)) throw Error(`Invalid backref name "${t}"`);
		if (!e.namedGroupsByName.has(r)) throw Error(`Group name not defined to the left "${t}"`);
		return li(r);
	}
	return i(+r);
}
function ei(e, t) {
	let { parent: n, tokens: r, walk: i } = e, a = n.elements.at(-1), o = r[e.current];
	if (!t.isCheckingRangeEnd && a && a.type !== R.CharacterClass && a.type !== R.CharacterClassRange && o && o.type !== M.CharacterClassOpen && o.type !== M.CharacterClassClose && o.type !== M.CharacterClassIntersector) {
		let e = i(n, {
			...t,
			isCheckingRangeEnd: !0
		});
		if (a.type === R.Character && e.type === R.Character) return n.elements.pop(), hi(a, e);
		throw Error("Invalid character class range");
	}
	return fi(45);
}
function ti(e, t) {
	let { token: n, tokens: r, verbose: i, walk: a } = e, o = r[e.current], s = pi({ negate: n.negate }), c = s.elements[0], l = Oi(o);
	for (; l.type !== M.CharacterClassClose;) {
		if (l.type === M.CharacterClassIntersector) c.classes.push(pi({
			negate: !1,
			baseOnly: !0
		})), e.current++;
		else {
			let e = c.classes.at(-1);
			e.elements.push(a(e, t));
		}
		l = Oi(r[e.current], o);
	}
	if (i || Di(c), c.classes.length === 1) {
		let e = c.classes[0];
		e.negate = s.negate !== e.negate, s = e;
	}
	return e.current++, s;
}
function ni({ token: e, skipPropertyNameValidation: t }) {
	let { kind: n, negate: r, value: i } = e;
	if (n === N.property) {
		let e = vr(i);
		if (gr.has(e)) n = N.posix, i = e;
		else return W(i, {
			negate: r,
			skipPropertyNameValidation: t
		});
	}
	return n === N.posix ? {
		type: R.CharacterSet,
		kind: B.posix,
		negate: r,
		value: i
	} : gi(n, { negate: r });
}
function ri(e, t) {
	let { token: n, tokens: r, capturingGroups: i, namedGroupsByName: a, skipLookbehindValidation: o, verbose: s, walk: c } = e, l = ui(n), u = l.type === R.AbsentFunction, d = l.kind === z.lookbehind, f = d && l.negate;
	if (l.type === R.CapturingGroup && (i.push(l), l.name && tr(a, l.name, []).push(l)), u && t.isInAbsentFunction) throw Error("Nested absent function not supported by Oniguruma");
	let p = ki(r[e.current]);
	for (; p.type !== M.GroupClose;) {
		if (p.type === M.Alternator) l.alternatives.push(V()), e.current++;
		else {
			let e = l.alternatives.at(-1), n = c(e, {
				...t,
				isInAbsentFunction: t.isInAbsentFunction || u,
				isInLookbehind: t.isInLookbehind || d,
				isInNegLookbehind: t.isInNegLookbehind || f
			});
			if (e.elements.push(n), (d || t.isInLookbehind) && !o) {
				let e = "Lookbehind includes a pattern not allowed by Oniguruma";
				if (f || t.isInNegLookbehind) {
					if (n.kind === z.lookahead || n.type === R.CapturingGroup) throw Error(e);
				} else if (n.kind === z.lookahead || n.kind === z.lookbehind && n.negate) throw Error(e);
			}
		}
		p = ki(r[e.current]);
	}
	return s || (l = Ti(l)), e.current++, l;
}
function ii({ token: e, parent: t }) {
	let { min: n, max: r, greedy: i, possessive: a } = e, o = t.elements.at(-1);
	if (!o || o.type === R.Assertion || o.type === R.Directive) throw Error("Quantifier requires a repeatable token");
	let s = bi(o, n, r, i, a);
	return t.elements.pop(), s;
}
function ai(e) {
	let { token: t, capturingGroups: n, subroutines: r } = e, i = t.raw.slice(3, -1), a = /^(?<sign>[-+]?)0*(?<num>[1-9]\d*)$/.exec(i);
	if (a) {
		let t = +a.groups.num, r = n.length;
		if (e.hasNumberedRef = !0, i = {
			"": t,
			"+": r + t,
			"-": r + 1 - t
		}[a.groups.sign], i < 1) throw Error("Invalid subroutine number");
	} else i === "0" && (i = 0);
	let o = Si(i);
	return r.push(o), o;
}
function oi(e) {
	if (e !== Yr.repeater) throw Error(`Unexpected absent function kind "${e}"`);
	return {
		type: R.AbsentFunction,
		kind: e,
		alternatives: [V()]
	};
}
function V() {
	return {
		type: R.Alternative,
		elements: []
	};
}
function si(e, t) {
	let n = !!t?.negate;
	return {
		type: R.Assertion,
		kind: e,
		...e === z.word_boundary && { negate: n }
	};
}
function ci({ kind: e }) {
	return si(A({
		"^": z.line_start,
		$: z.line_end,
		"\\A": z.string_start,
		"\\b": z.word_boundary,
		"\\B": z.word_boundary,
		"\\G": z.search_start,
		"\\z": z.string_end,
		"\\Z": z.string_end_newline
	}[e], `Unexpected assertion kind "${e}"`), { negate: e === k`\B` });
}
function li(e, t) {
	let n = !!t?.orphan;
	return {
		type: R.Backreference,
		...n && { orphan: n },
		ref: e
	};
}
function ui({ flags: e, kind: t, name: n, negate: r, number: i }) {
	switch (t) {
		case P.absent_repeater: return oi(Yr.repeater);
		case P.atomic: return H({ atomic: !0 });
		case P.capturing: return di(i, n);
		case P.group: return H({ flags: e });
		case P.lookahead:
		case P.lookbehind: return U({
			behind: t === P.lookbehind,
			negate: r
		});
		default: throw Error(`Unexpected group kind "${t}"`);
	}
}
function di(e, t) {
	let n = t !== void 0;
	if (n && !Ei(t)) throw Error(`Group name "${t}" invalid in Oniguruma`);
	return {
		type: R.CapturingGroup,
		number: e,
		...n && { name: t },
		alternatives: [V()]
	};
}
function fi(e, t) {
	let n = {
		useLastValid: !1,
		...t
	};
	if (e > 1114111) {
		let t = e.toString(16);
		if (n.useLastValid) e = 1114111;
		else if (e > 1310719) throw Error(`Invalid code point out of range "\\x{${t}}"`);
		else throw Error(`Invalid code point out of range in JS "\\x{${t}}"`);
	}
	return {
		type: R.Character,
		value: e
	};
}
function pi(e) {
	let t = {
		baseOnly: !1,
		negate: !1,
		...e
	};
	return {
		type: R.CharacterClass,
		negate: t.negate,
		elements: t.baseOnly ? [] : [mi()]
	};
}
function mi() {
	return {
		type: R.CharacterClassIntersection,
		classes: [pi({
			negate: !1,
			baseOnly: !0
		})]
	};
}
function hi(e, t) {
	if (t.value < e.value) throw Error("Character class range out of order");
	return {
		type: R.CharacterClassRange,
		min: e,
		max: t
	};
}
function gi(e, { negate: t }) {
	let n = {
		type: R.CharacterSet,
		kind: A(B[e], `Unexpected character set kind "${e}"`)
	};
	return (e === N.digit || e === N.hex || e === N.space || e === N.word) && (n.negate = t), n;
}
function _i({ kind: e, flags: t }) {
	let n = {
		type: R.Directive,
		kind: A(Xr[e], `Unexpected directive kind "${e}"`)
	};
	return e === xr.flags && (n.flags = t), n;
}
function vi({ ignoreCase: e, dotAll: t, extended: n, digitIsAscii: r, spaceIsAscii: i, wordIsAscii: a }) {
	return {
		type: R.Flags,
		ignoreCase: e,
		dotAll: t,
		extended: n,
		digitIsAscii: r,
		spaceIsAscii: i,
		wordIsAscii: a
	};
}
function H(e) {
	let t = e?.atomic, n = e?.flags;
	return {
		type: R.Group,
		...t && { atomic: t },
		...n && { flags: n },
		alternatives: [V()]
	};
}
function U(e) {
	let t = {
		behind: !1,
		negate: !1,
		...e
	};
	return {
		type: R.Assertion,
		kind: t.behind ? z.lookbehind : z.lookahead,
		negate: t.negate,
		alternatives: [V()]
	};
}
function yi() {
	return {
		type: R.Pattern,
		alternatives: [V()]
	};
}
function bi(e, t, n, r = !0, i = !1) {
	let a = {
		type: R.Quantifier,
		min: t,
		max: n,
		greedy: r,
		possessive: i,
		element: e
	};
	return n < t ? {
		...a,
		min: n,
		max: t,
		possessive: !0
	} : a;
}
function xi(e, t) {
	return {
		type: R.Regex,
		pattern: e,
		flags: t
	};
}
function Si(e) {
	return {
		type: R.Subroutine,
		ref: e
	};
}
function W(e, t) {
	let n = {
		negate: !1,
		skipPropertyNameValidation: !1,
		...t
	};
	return {
		type: R.CharacterSet,
		kind: B.property,
		value: n.skipPropertyNameValidation ? e : wi(e),
		negate: n.negate
	};
}
function Ci(e) {
	return {
		type: R.VariableLengthCharacterSet,
		kind: A({
			"\\R": Zr.newline,
			"\\X": Zr.grapheme
		}[e], `Unexpected varcharset kind "${e}"`)
	};
}
function wi(e) {
	let t = vr(e);
	if (dr.has(t)) throw Error(k`Unicode property "\p{${e}}" unsupported in Oniguruma`);
	return lr.get(t) || e.trim().replace(/[- _]+/g, "_").replace(/[A-Z][a-z]+(?=[A-Z])/g, "$&_").replace(/[A-Za-z]+/g, (e) => e[0].toUpperCase() + e.slice(1).toLowerCase());
}
function Ti(e) {
	let t = e.alternatives[0].elements[0];
	return e.type === R.Group && Hr(e, (e) => e.type === R.Group) && !(e.atomic && t.flags) && !(e.flags && (t.atomic || t.flags)) ? (e.atomic ? t.atomic = !0 : e.flags && (t.flags = e.flags), t) : e;
}
function Ei(e) {
	return /^[\p{Alpha}\p{Pc}][^)]*$/u.test(e);
}
function Di(e) {
	for (let t = 0; t < e.classes.length; t++) {
		let n = e.classes[t], r = n.elements[0];
		n.elements.length === 1 && r.type === R.CharacterClass && (e.classes[t] = r, r.negate = n.negate !== r.negate);
	}
}
function Oi(e, t) {
	return A(e, `${t?.value === 93 ? "Empty" : "Unclosed"} character class`);
}
function ki(e) {
	return A(e, "Unclosed group");
}
function Ai(e, t) {
	let n = {
		accuracy: "default",
		asciiWordBoundaries: !1,
		avoidSubclass: !1,
		bestEffortTarget: "ES2025",
		...t
	}, r = {
		accuracy: n.accuracy,
		asciiWordBoundaries: n.asciiWordBoundaries,
		avoidSubclass: n.avoidSubclass,
		flagDirectivesByAlt: /* @__PURE__ */ new Map(),
		jsGroupNameMap: /* @__PURE__ */ new Map(),
		minTargetEs2024: nr(n.bestEffortTarget, "ES2024"),
		passedLookbehind: !1,
		strategy: null,
		subroutineRefMap: /* @__PURE__ */ new Map(),
		supportedGNodes: /* @__PURE__ */ new Set(),
		digitIsAscii: e.flags.digitIsAscii,
		spaceIsAscii: e.flags.spaceIsAscii,
		wordIsAscii: e.flags.wordIsAscii
	};
	L({ node: e }, r, ji);
	let i = {
		dotAll: e.flags.dotAll,
		ignoreCase: e.flags.ignoreCase
	}, a = {
		currentFlags: i,
		prevFlags: null,
		globalFlags: i,
		groupOriginByCopy: /* @__PURE__ */ new Map(),
		groupsByName: /* @__PURE__ */ new Map(),
		multiplexCapturesToLeftByRef: /* @__PURE__ */ new Map(),
		openRefs: /* @__PURE__ */ new Map(),
		reffedNodesByReferencer: /* @__PURE__ */ new Map(),
		subroutineRefMap: r.subroutineRefMap
	};
	L({ node: e }, a, Mi);
	let o = {
		groupsByName: a.groupsByName,
		highestOrphanBackref: 0,
		numCapturesToLeft: 0,
		reffedNodesByReferencer: a.reffedNodesByReferencer
	};
	return L({ node: e }, o, Ni), e._strategy = r.strategy, e;
}
var ji = {
	AbsentFunction({ node: e, replaceWith: t }) {
		let n = q(H(), [Fi(U({ negate: !0 }), e.alternatives), W("Any")]), r = bi(n, 0, Infinity);
		n.parent = r, t(q(H(), [r]));
	},
	Alternative: {
		enter({ node: e, parent: t, key: n }, { flagDirectivesByAlt: r }) {
			let i = e.elements.filter((e) => e.kind === Xr.flags);
			for (let e = n + 1; e < t.alternatives.length; e++) {
				let n = t.alternatives[e];
				tr(r, n, []).push(...i);
			}
		},
		exit({ node: e }, { flagDirectivesByAlt: t }) {
			if (t.get(e)?.length) {
				let n = Ui(t.get(e));
				if (n) {
					let t = q(H({ flags: n }), e.elements);
					t.parent = e, e.elements = [t];
				}
			}
		}
	},
	Assertion({ node: e, key: t, container: n, ast: r, remove: i, replaceWith: a }, o) {
		let { kind: s, negate: c } = e, { asciiWordBoundaries: l, avoidSubclass: u, supportedGNodes: d, wordIsAscii: f } = o;
		if (s === z.line_end) a(K(k`(?=\z|\n)`));
		else if (s === z.line_start) a(K(k`(?<=\A|\n(?!\z))`, { skipLookbehindValidation: !0 }));
		else if (s === z.lookbehind) o.passedLookbehind = !0;
		else if (s === z.search_start) {
			if (d.has(e)) r.flags.sticky = !0, i();
			else {
				let e = n[t - 1];
				if (e && Wr(e)) a(q(U({ negate: !0 })));
				else if (u) throw Error(k`Uses "\G" in a way that requires a subclass`);
				else a(si(z.string_start)), o.strategy = "search_start_clip";
			}
		} else if (s === z.string_end_newline) a(K(k`(?=\n?\z)`));
		else if (s === z.word_boundary && !f && !l) {
			let e = `(?:(?<=${G})(?!${G})|(?<!${G})(?=${G}))`, t = `(?:(?<=${G})(?=${G})|(?<!${G})(?!${G}))`;
			a(K(c ? t : e));
		}
	},
	Backreference({ node: e }, { jsGroupNameMap: t }) {
		let { ref: n } = e;
		typeof n == "string" && !Yi(n) && (n = Vi(n, t), e.ref = n);
	},
	CapturingGroup({ node: e }, { jsGroupNameMap: t, subroutineRefMap: n }) {
		let { name: r } = e;
		r && !Yi(r) && (r = Vi(r, t), e.name = r), n.set(e.number, e), r && n.set(r, e);
	},
	CharacterSet({ node: e, replaceWith: t }, { accuracy: n, minTargetEs2024: r, digitIsAscii: i, spaceIsAscii: a, wordIsAscii: o }) {
		let { kind: s, negate: c, value: l } = e;
		if (i && (s === B.digit || l === "digit")) t(gi(B.digit, { negate: c }));
		else if (a && (s === B.space || l === "space")) t(Xi(K(Pi), c));
		else if (o && (s === B.word || l === "word")) t(gi(B.word, { negate: c }));
		else if (s === B.any) t(W("Any"));
		else if (s === B.digit) t(W("Nd", { negate: c }));
		else if (s === B.hex) t(W("AHex", { negate: c }));
		else if (s === B.non_newline) t(K(k`[^\n]`));
		else if (s === B.space) t(W("space", { negate: c }));
		else if (s === B.word) t(Xi(K(G), c));
		else if (s === B.property) cr.has(l) || (e.key = "sc");
		else if (s === B.posix) {
			if (!r && (l === "graph" || l === "print")) {
				if (n === "strict") throw Error(`POSIX class "${l}" requires min target ES2024 or non-strict accuracy`);
				let e = {
					graph: "!-~",
					print: " -~"
				}[l];
				c && (e = `\0-${O(e.codePointAt(0) - 1)}${O(e.codePointAt(2) + 1)}-\u{10FFFF}`), t(K(`[${e}]`));
			} else t(Xi(K(hr.get(l)), c));
		}
	},
	Directive(e, t) {
		let { node: n, parent: r, ast: i, remove: a, replaceWith: o, removeAllPrevSiblings: s, removeAllNextSiblings: c } = e, { kind: l, flags: u } = n;
		if (l === Xr.flags) {
			if (!u.enable && !u.disable) a();
			else {
				let n = q(H({ flags: u }), c());
				o(n), Zi(n, e, t, ji);
			}
		} else if (l === Xr.keep) {
			let e = i.pattern.alternatives[0].elements[0], t = Hr(i.pattern, (e) => e.type === R.Group) && e.alternatives.length === 1 ? e : i.pattern;
			if (r.parent !== t || t.alternatives.length > 1) throw Error(k`Uses "\K" in a way that's unsupported`);
			o(q(U({ behind: !0 }), s()));
		}
	},
	Flags({ node: e, parent: t }) {
		[
			"digitIsAscii",
			"extended",
			"spaceIsAscii",
			"wordIsAscii"
		].forEach((t) => delete e[t]), Object.assign(e, {
			global: !1,
			hasIndices: !1,
			multiline: !1,
			sticky: e.sticky ?? !1
		}), t.options = {
			disable: {
				x: !0,
				n: !0
			},
			force: { v: !0 }
		};
	},
	Group({ node: e }) {
		if (!e.flags) return;
		let { enable: t, disable: n } = e.flags;
		t?.extended && delete t.extended, n?.extended && delete n.extended, t?.dotAll && n?.dotAll && delete t.dotAll, t?.ignoreCase && n?.ignoreCase && delete t.ignoreCase, t && !Object.keys(t).length && delete e.flags.enable, n && !Object.keys(n).length && delete e.flags.disable, !e.flags.enable && !e.flags.disable && delete e.flags;
	},
	Pattern: {
		enter({ node: e }, { supportedGNodes: t }) {
			let n = [], r = !1, i = !1;
			for (let t of e.alternatives) if (t.elements.length === 1 && t.elements[0].kind === z.search_start) t.elements.pop();
			else {
				let e = Ki(t.elements);
				e ? (r = !0, Array.isArray(e) ? n.push(...e) : n.push(e)) : i = !0;
			}
			r && !i && n.forEach((e) => t.add(e));
		},
		exit(e, { accuracy: t, passedLookbehind: n, strategy: r }) {
			if (t === "strict" && n && r) throw Error(k`Uses "\G" in a way that requires non-strict accuracy`);
		}
	},
	Quantifier({ node: e }) {
		if (e.element.type === R.Quantifier) {
			let t = q(H(), [e.element]);
			t.parent = e, e.element = t;
		}
	},
	Subroutine({ node: e }, { jsGroupNameMap: t }) {
		let { ref: n } = e;
		typeof n == "string" && !Yi(n) && (n = Vi(n, t), e.ref = n);
	},
	VariableLengthCharacterSet({ node: e, replaceWith: t }, { accuracy: n, minTargetEs2024: r }) {
		let { kind: i } = e;
		if (i === Zr.newline) t(K("(?>\r\n?|[\n\v\f\u2028\u2029])"));
		else if (i === Zr.grapheme) {
			if (n === "strict") throw Error(k`Use of "\X" requires non-strict accuracy`);
			t(K(k`(?>\r\n|${r ? k`\p{RGI_Emoji}` : On().source.replace(/\\u\{/g, "\\x{")}|\P{M}\p{M}*)`, { skipPropertyNameValidation: !0 }));
		} else throw Error(`Unexpected varcharset kind "${i}"`);
	}
}, Mi = {
	Backreference({ node: e }, { multiplexCapturesToLeftByRef: t, reffedNodesByReferencer: n }) {
		let { orphan: r, ref: i } = e;
		r || n.set(e, [...t.get(i).map(({ node: e }) => e)]);
	},
	CapturingGroup: {
		enter({ node: e, replaceWith: t, skip: n }, { groupOriginByCopy: r, groupsByName: i, multiplexCapturesToLeftByRef: a, openRefs: o, reffedNodesByReferencer: s }) {
			let c = r.get(e);
			if (c && o.has(e.number)) {
				let r = zi(e.number);
				s.set(r, o.get(e.number)), t(r), n();
				return;
			}
			o.set(e.number, e), a.set(e.number, []), e.name && tr(a, e.name, []);
			let l = a.get(e.name ?? e.number);
			for (let t = 0; t < l.length; t++) {
				let n = l[t];
				if (c === n.node || c && c === n.origin || e === n.origin) {
					l.splice(t, 1);
					break;
				}
			}
			if (a.get(e.number).push({
				node: e,
				origin: c
			}), e.name && a.get(e.name).push({
				node: e,
				origin: c
			}), e.name) {
				let t = tr(i, e.name, /* @__PURE__ */ new Map()), n = !1;
				if (c) n = !0;
				else for (let e of t.values()) if (!e.hasDuplicateNameToRemove) {
					n = !0;
					break;
				}
				i.get(e.name).set(e, {
					node: e,
					hasDuplicateNameToRemove: n
				});
			}
			c && (e._originNumber = c.number);
		},
		exit({ node: e }, { openRefs: t }) {
			t.delete(e.number);
		}
	},
	Group: {
		enter({ node: e }, t) {
			t.prevFlags = t.currentFlags, e.flags && (t.currentFlags = er(t.currentFlags, e.flags));
		},
		exit(e, t) {
			t.currentFlags = t.prevFlags;
		}
	},
	Recursion({ node: e, parent: t }, { reffedNodesByReferencer: n }) {
		let { ref: r } = e, i = t;
		for (; (i = i.parent) && (i.type !== R.CapturingGroup || i.name !== r && i.number !== r););
		n.set(e, i);
	},
	Subroutine(e, t) {
		let { node: n, replaceWith: r } = e, { ref: i } = n, a = t.subroutineRefMap.get(i), o = i === 0, s = o ? zi(0) : Ri(a, t.groupOriginByCopy, null), c = s;
		if (!o) {
			let e = Ui(Bi(a, (e) => e.type === R.Group && !!e.flags)), n = e ? er(t.globalFlags, e) : t.globalFlags;
			Ii(n, t.currentFlags) || (c = q(H({ flags: Wi(n) }), [s]));
		}
		r(c), o || Zi(c, e, t, Mi);
	}
}, Ni = {
	Backreference({ node: e, replaceWith: t }, n) {
		if (e.orphan) {
			n.highestOrphanBackref = Math.max(n.highestOrphanBackref, e.ref);
			return;
		}
		let r = n.reffedNodesByReferencer.get(e).filter((t) => Li(t, e));
		if (!r.length) t(q(U({ negate: !0 })));
		else if (r.length > 1) {
			let e = r.map((e) => Fi(V(), [li(e.number)]));
			t(Fi(H(), e));
		} else e.ref = r[0].number;
	},
	CapturingGroup({ node: e }, t) {
		e.number = ++t.numCapturesToLeft, e.name && t.groupsByName.get(e.name).get(e).hasDuplicateNameToRemove && delete e.name;
	},
	Recursion({ node: e }, t) {
		e.ref !== 0 && (e.ref = t.reffedNodesByReferencer.get(e).number);
	},
	Regex: { exit({ node: e }, t) {
		let n = Math.max(t.highestOrphanBackref - t.numCapturesToLeft, 0);
		for (let t = 0; t < n; t++) {
			let t = di();
			e.pattern.alternatives.at(-1).elements.push(t);
		}
	} }
}, Pi = "[	-\r ]", G = k`[\p{L}\p{M}\p{N}\p{Pc}]`;
function Fi(e, t) {
	return t.forEach((t) => t.parent = e), e[Hi(e)] = t, e;
}
function Ii(e, t) {
	return e.dotAll === t.dotAll && e.ignoreCase === t.ignoreCase;
}
function Li(e, t) {
	let n = t;
	do {
		if (n.type === R.Pattern) return !1;
		if (n.type === R.Alternative) continue;
		if (n === e) return !1;
		let t = Gi(n.parent);
		for (let r of t) {
			if (r === n) break;
			if (r === e || qi(r, e)) return !0;
		}
	} while (n = n.parent);
	throw Error("Unexpected path");
}
function Ri(e, t, n, r) {
	let i = Array.isArray(e) ? [] : {};
	for (let [a, o] of Object.entries(e)) a === "parent" ? i.parent = Array.isArray(n) ? r : n : o && typeof o == "object" ? i[a] = Ri(o, t, i, n) : (a === "type" && o === R.CapturingGroup && t.set(i, t.get(e) ?? e), i[a] = o);
	return i;
}
function zi(e) {
	return {
		type: R.Recursion,
		ref: e
	};
}
function Bi(e, t) {
	let n = [];
	for (; e = e.parent;) (!t || t(e)) && n.push(e);
	return n;
}
function Vi(e, t) {
	if (t.has(e)) return t.get(e);
	let n = `$${t.size}_${e.replace(/^[^$_\p{IDS}]|[^$\u200C\u200D\p{IDC}]/gu, "_")}`;
	return t.set(e, n), n;
}
function Hi(e) {
	for (let t of [
		"alternatives",
		"classes",
		"elements"
	]) if (e[t]) return t;
	return null;
}
function Ui(e) {
	let t = ["dotAll", "ignoreCase"], n = {
		enable: {},
		disable: {}
	};
	return e.forEach(({ flags: e }) => {
		t.forEach((t) => {
			e.enable?.[t] && (delete n.disable[t], n.enable[t] = !0), e.disable?.[t] && (n.disable[t] = !0);
		});
	}), Object.keys(n.enable).length || delete n.enable, Object.keys(n.disable).length || delete n.disable, n.enable || n.disable ? n : null;
}
function Wi({ dotAll: e, ignoreCase: t }) {
	let n = {};
	return (e || t) && (n.enable = {}, e && (n.enable.dotAll = !0), t && (n.enable.ignoreCase = !0)), e && t || (n.disable = {}, !e && (n.disable.dotAll = !0), !t && (n.disable.ignoreCase = !0)), n;
}
function Gi(e) {
	if (!e) throw Error("Node expected");
	if (e.type === R.Quantifier) return [e.element];
	let t = Hi(e);
	return t && e[t];
}
function Ki(e) {
	let t = e.find((e) => e.kind === z.search_start || Ji(e, { negate: !1 }) || !Ur(e));
	if (!t) return null;
	if (t.kind === z.search_start) return t;
	if (I(t)) return t.alternatives[0].elements[0];
	if (Gr(t)) {
		let e = [];
		for (let n of t.alternatives) {
			let t = Ki(n.elements);
			if (!t) return null;
			Array.isArray(t) ? e.push(...t) : e.push(t);
		}
		return e;
	}
	return null;
}
function qi(e, t) {
	let n = Gi(e) ?? [];
	for (let e of n) if (e === t || qi(e, t)) return !0;
	return !1;
}
function Ji(e, t) {
	let n = {
		negate: null,
		...t
	};
	return I(e) && (n.negate === null || e.negate === n.negate) && Hr(e, (e) => e.kind === z.search_start);
}
function Yi(e) {
	return /^[$_\p{IDS}][$\u200C\u200D\p{IDC}]*$/u.test(e);
}
function K(e, t) {
	let n = Qr(Or(e), t).pattern.alternatives;
	return n.length > 1 || n[0].elements.length > 1 ? Fi(H(), n) : n[0].elements[0];
}
function q(e, t) {
	let n = Hi(e);
	return e[n][0].parent = e, t && Fi(e[n][0], t), e;
}
function Xi(e, t) {
	return e.negate = t, e;
}
function Zi(e, { parent: t, key: n, container: r }, i, a) {
	L({
		node: e,
		parent: t,
		key: n,
		container: r
	}, i, a);
}
function Qi(e, t) {
	let n = ar(t), r = nr(n.target, "ES2024"), i = nr(n.target, "ES2025"), a = n.rules.recursionLimit;
	if (!Number.isInteger(a) || a < 2 || a > 20) throw Error("Invalid recursionLimit; use 2-20");
	let o = null, s = null;
	if (!i) {
		let t = [e.flags.ignoreCase];
		L({ node: e }, {
			getCurrentModI: () => t.at(-1),
			popModI() {
				t.pop();
			},
			pushModI(e) {
				t.push(e);
			},
			setHasCasedChar() {
				t.at(-1) ? o = !0 : s = !0;
			}
		}, $i);
	}
	let c = {
		dotAll: e.flags.dotAll,
		ignoreCase: !(!e.flags.ignoreCase && !o || s)
	}, l = null, u = {
		accuracy: n.accuracy,
		appliedGlobalFlags: c,
		avoidSubclass: n.avoidSubclass,
		captureMap: /* @__PURE__ */ new Map(),
		currentFlags: {
			dotAll: e.flags.dotAll,
			ignoreCase: e.flags.ignoreCase
		},
		inCharClass: !1,
		lastNode: l,
		recursionLimit: a,
		useAppliedIgnoreCase: !!(!i && o && s),
		useFlagMods: i,
		useFlagV: r,
		verbose: n.verbose
	};
	function d(e) {
		switch (u.lastNode = l, l = e, e.type) {
			case R.Regex: return {
				pattern: d(e.pattern),
				flags: d(e.flags),
				options: { ...e.options }
			};
			case R.Alternative: return e.elements.map(d).join("");
			case R.Assertion: return oa(e, u, d);
			case R.Backreference: return sa(e, u);
			case R.CapturingGroup: return ca(e, u, d);
			case R.Character: return la(e, u);
			case R.CharacterClass: return ua(e, u, d);
			case R.CharacterClassIntersection:
				if (!u.useFlagV) throw Error("Use of class intersection requires min target ES2024");
				return e.classes.map(d).join("&&");
			case R.CharacterClassRange: return da(e, u);
			case R.CharacterSet: return fa(e, u);
			case R.Flags: return pa(e, u);
			case R.Group: return ma(e, u, d);
			case R.Pattern: return e.alternatives.map(d).join("|");
			case R.Quantifier: return d(e.element) + ya(e);
			case R.Recursion: return ha(e, u);
			default: throw Error(`Unexpected node type "${e.type}"`);
		}
	}
	let f = d(e);
	return r || (delete f.options.force.v, f.options.disable.v = !0, f.options.unicodeSetsPlugin = null), f;
}
var $i = {
	AnyGroup: {
		enter({ node: e }, t) {
			let n = t.getCurrentModI();
			t.pushModI(e.flags ? er({ ignoreCase: n }, e.flags).ignoreCase : n);
		},
		exit(e, t) {
			t.popModI();
		}
	},
	Backreference(e, t) {
		t.setHasCasedChar();
	},
	Character({ node: e }, t) {
		aa(O(e.value)) && t.setHasCasedChar();
	},
	CharacterClassRange({ node: e, skip: t }, n) {
		t(), ga(e, { firstOnly: !0 }).length && n.setHasCasedChar();
	},
	CharacterSet({ node: e }, t) {
		e.kind === B.property && br.has(e.value) && t.setHasCasedChar();
	}
}, ea = /* @__PURE__ */ new Set([
	"$",
	"(",
	")",
	"*",
	"+",
	".",
	"?",
	"[",
	"\\",
	"]",
	"^",
	"{",
	"|",
	"}"
]), ta = /* @__PURE__ */ new Set([
	"-",
	"\\",
	"]",
	"^",
	"["
]), na = /* @__PURE__ */ new Set(/* @__PURE__ */ "()-/[\\]^{|}!#$%&*+,.:;<=>?@`~".split("")), ra = /* @__PURE__ */ new Map([
	[9, k`\t`],
	[10, k`\n`],
	[11, k`\v`],
	[12, k`\f`],
	[13, k`\r`],
	[8232, k`\u2028`],
	[8233, k`\u2029`],
	[65279, k`\uFEFF`]
]), ia = /^\p{Cased}$/u;
function aa(e) {
	return ia.test(e);
}
function oa(e, t, n) {
	let { kind: r, negate: i, alternatives: a } = e;
	if (I(e)) return `(?${`${r === z.lookahead ? "" : "<"}${i ? "!" : "="}`}${a.map(n).join("|")})`;
	if (r === z.string_end) return "$";
	if (r === z.string_start) return "^";
	if (r === z.word_boundary) return i ? k`\B` : k`\b`;
	throw Error(`Unexpected assertion kind "${r}"`);
}
function sa({ ref: e }, t) {
	if (typeof e != "number") throw Error("Unexpected named backref in transformed AST");
	if (!t.useFlagMods && t.accuracy === "strict" && t.currentFlags.ignoreCase && !t.captureMap.get(e).ignoreCase) throw Error("Use of case-insensitive backref to case-sensitive group requires target ES2025 or non-strict accuracy");
	return "\\" + e;
}
function ca({ name: e, number: t, alternatives: n, _originNumber: r }, i, a) {
	return i.captureMap.set(t, { ignoreCase: i.currentFlags.ignoreCase }), `(${e ? `?<${e}>` : ""}${!i.avoidSubclass && r ? `${r < t ? `$${r}` : ""}\$E\$` : ""}${n.map(a).join("|")})`;
}
function la({ value: e }, t) {
	let n = O(e), r = J(e, {
		isAfterBackref: t.lastNode.type === R.Backreference,
		inCharClass: t.inCharClass,
		useFlagV: t.useFlagV
	});
	if (r !== n) return r;
	if (t.useAppliedIgnoreCase && t.currentFlags.ignoreCase && aa(n)) {
		let e = sr(n);
		return t.inCharClass ? e.join("") : e.length > 1 ? `[${e.join("")}]` : e[0];
	}
	return n;
}
function ua({ negate: e, parent: t, elements: n }, r, i) {
	let a = () => `[${e ? "^" : ""}${n.map(i).join("")}]`;
	if (!r.inCharClass) {
		r.inCharClass = !0;
		let e = a();
		return r.inCharClass = !1, e;
	}
	let o = n[0]?.type;
	if (!e && o && ((!r.useFlagV || !r.verbose) && t.type === R.CharacterClass && o !== R.CharacterClassIntersection || !r.verbose && t.type === R.CharacterClassIntersection && n.length === 1 && o !== R.CharacterClass && o !== R.CharacterClassRange)) return n.map(i).join("");
	if (!r.useFlagV && t.type === R.CharacterClass) throw Error("Use of nested character class requires min target ES2024");
	return a();
}
function da(e, t) {
	let n = e.min.value, r = e.max.value, i = {
		isAfterBackref: !1,
		inCharClass: !0,
		useFlagV: t.useFlagV
	}, a = J(n, i), o = J(r, i), s = /* @__PURE__ */ new Set();
	return t.useAppliedIgnoreCase && t.currentFlags.ignoreCase && _a(ga(e)).forEach((e) => {
		s.add(Array.isArray(e) ? `${J(e[0], i)}-${J(e[1], i)}` : J(e, i));
	}), `${a}-${o}${[...s].join("")}`;
}
function fa({ kind: e, negate: t, value: n, key: r }, i) {
	if (e === B.dot) return i.currentFlags.dotAll ? i.appliedGlobalFlags.dotAll || i.useFlagMods ? "." : "[^]" : k`[^\n]`;
	if (e === B.digit) return t ? k`\D` : k`\d`;
	if (e === B.property) {
		if (i.useAppliedIgnoreCase && i.currentFlags.ignoreCase && br.has(n)) throw Error(`Unicode property "${n}" can't be case-insensitive when other chars have specific case`);
		return `${t ? k`\P` : k`\p`}{${r ? `${r}=` : ""}${n}}`;
	}
	if (e === B.word) return t ? k`\W` : k`\w`;
	throw Error(`Unexpected character set kind "${e}"`);
}
function pa(e, t) {
	return (t.appliedGlobalFlags.ignoreCase ? "i" : "") + (e.dotAll ? "s" : "") + (e.sticky ? "y" : "");
}
function ma({ atomic: e, flags: t, parent: n, alternatives: r }, i, a) {
	let o = i.currentFlags;
	t && (i.currentFlags = er(o, t));
	let s = r.map(a).join("|"), c = !i.verbose && r.length === 1 && n.type !== R.Quantifier && !e && (!i.useFlagMods || !t) ? s : `(?${va(e, t, i.useFlagMods)}${s})`;
	return i.currentFlags = o, c;
}
function ha({ ref: e }, t) {
	let n = t.recursionLimit;
	return e === 0 ? `(?R=${n})` : k`\g<${e}&R=${n}>`;
}
function ga(e, t) {
	let n = !!t?.firstOnly, r = e.min.value, i = e.max.value, a = [];
	if (r < 65 && (i === 65535 || i >= 131071) || r === 65536 && i >= 131071) return a;
	for (let e = r; e <= i; e++) {
		let t = O(e);
		if (!aa(t)) continue;
		let o = sr(t).filter((e) => {
			let t = e.codePointAt(0);
			return t < r || t > i;
		});
		if (o.length && (a.push(...o), n)) break;
	}
	return a;
}
function J(e, { isAfterBackref: t, inCharClass: n, useFlagV: r }) {
	if (ra.has(e)) return ra.get(e);
	if (e < 32 || e > 126 && e < 160 || e > 262143 || t && ba(e)) return e > 255 ? `\\u{${e.toString(16).toUpperCase()}}` : `\\x${e.toString(16).toUpperCase().padStart(2, "0")}`;
	let i = n ? r ? na : ta : ea, a = O(e);
	return (i.has(a) ? "\\" : "") + a;
}
function _a(e) {
	let t = e.map((e) => e.codePointAt(0)).sort((e, t) => e - t), n = [], r = null;
	for (let e = 0; e < t.length; e++) t[e + 1] === t[e] + 1 ? r ??= t[e] : r === null ? n.push(t[e]) : (n.push([r, t[e]]), r = null);
	return n;
}
function va(e, t, n) {
	if (e) return ">";
	let r = "";
	if (t && n) {
		let { enable: e, disable: n } = t;
		r = (e?.ignoreCase ? "i" : "") + (e?.dotAll ? "s" : "") + (n ? "-" : "") + (n?.ignoreCase ? "i" : "") + (n?.dotAll ? "s" : "");
	}
	return `${r}:`;
}
function ya({ min: e, max: t, greedy: n, possessive: r }) {
	let i;
	return i = !e && t === 1 ? "?" : !e && t === Infinity ? "*" : e === 1 && t === Infinity ? "+" : e === t ? `{${e}}` : `{${e},${t === Infinity ? "" : t}}`, i + (r ? "+" : n ? "" : "?");
}
function ba(e) {
	return e > 47 && e < 58;
}
var xa = class e extends Pn {
	#e;
	rawArgs;
	constructor(t, n, r) {
		if (t instanceof RegExp) {
			if (r) throw Error("Cannot provide options when copying a regexp");
			super(t, n), t instanceof e ? (this.#e = t.#e, this.rawArgs = t.rawArgs) : (this.#e = null, this.rawArgs = {
				pattern: t.source,
				flags: t.flags,
				options: {}
			}), n !== void 0 && (this.rawArgs.flags = n);
		} else {
			let e = {
				strategy: null,
				useEmulationGroups: !1,
				...r
			};
			super(t, n, { useEmulationGroups: e.useEmulationGroups }), this.#e = e.strategy, this.rawArgs = {
				pattern: t,
				flags: n ?? "",
				options: {
					...e.strategy && { strategy: e.strategy },
					...e.useEmulationGroups && { useEmulationGroups: !0 }
				}
			};
		}
	}
	exec(e) {
		let t = super.exec, n = this.global || this.sticky, r = this.lastIndex;
		if (this.#e === "search_start_clip" && n && r) {
			this.lastIndex = 0;
			let n = t.call(this, e.slice(r));
			return n && (Sa(n, this, e, r), this.lastIndex += r), n;
		}
		return t.call(this, e);
	}
};
function Sa(e, t, n, r) {
	if (e.input = n, e.index += r, t.hasIndices) {
		let t = e.indices;
		for (let e = 0; e < t.length; e++) {
			let n = t[e];
			n && (t[e] = [n[0] + r, n[1] + r]);
		}
		let n = t.groups;
		n && Object.keys(n).forEach((e) => {
			let t = n[e];
			t && (n[e] = [t[0] + r, t[1] + r]);
		});
	}
}
function Ca(e, t) {
	let n = ar(t), r = n.avoidSubclass, i = Ai(Qr(Or(e, n.flags, {
		captureGroup: n.rules.captureGroup,
		singleline: n.rules.singleline
	}), {
		skipBackrefValidation: n.rules.allowOrphanBackrefs,
		verbose: n.verbose
	}), {
		accuracy: n.accuracy,
		asciiWordBoundaries: n.rules.asciiWordBoundaries,
		avoidSubclass: r,
		bestEffortTarget: n.target
	}), a = Qi(i, n), o = { useEmulationGroups: !r }, s = {
		pattern: zn(Hn(qn(a.pattern, o)), o),
		flags: `${n.hasIndices ? "d" : ""}${n.global ? "g" : ""}${a.flags}${a.options.disable.v ? "u" : "v"}`
	}, c = !r && s.pattern.includes("$E$"), l = i._strategy;
	return (c || l) && (s.options = {
		...l && { strategy: l },
		...c && { useEmulationGroups: c }
	}), s;
}
function wa(e, t) {
	let n = Ca(e, t);
	return n.options ? new xa(n.pattern, n.flags, n.options) : new RegExp(n.pattern, n.flags);
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@shikijs/engine-javascript/1.29.2/97934766166e2c5c5f4d1a1d77d99382f2c7438e6cb55c207f9a2ae918a59437/node_modules/@shikijs/engine-javascript/dist/shared/engine-javascript.hzpS1_41.mjs
var Ta = 4294967295, Ea = class {
	constructor(e, t = {}) {
		this.patterns = e, this.options = t;
		let { forgiving: n = !1, cache: r, regexConstructor: i } = t;
		if (!i) throw Error("Option `regexConstructor` is not provided");
		this.regexps = e.map((e) => {
			if (typeof e != "string") return e;
			let t = r?.get(e);
			if (t) {
				if (t instanceof RegExp) return t;
				if (n) return null;
				throw t;
			}
			try {
				let t = i(e);
				return r?.set(e, t), t;
			} catch (t) {
				if (r?.set(e, t), n) return null;
				throw t;
			}
		});
	}
	regexps;
	findNextMatchSync(e, t, n) {
		let r = typeof e == "string" ? e : e.content, i = [];
		function a(e, t, n = 0) {
			return {
				index: e,
				captureIndices: t.indices.map((e) => e == null ? {
					start: Ta,
					end: Ta,
					length: 0
				} : {
					start: e[0] + n,
					end: e[1] + n,
					length: e[1] - e[0]
				})
			};
		}
		for (let e = 0; e < this.regexps.length; e++) {
			let n = this.regexps[e];
			if (n) try {
				n.lastIndex = t;
				let o = n.exec(r);
				if (!o) continue;
				if (o.index === t) return a(e, o, 0);
				i.push([
					e,
					o,
					0
				]);
			} catch (e) {
				if (this.options.forgiving) continue;
				throw e;
			}
		}
		if (i.length) {
			let e = Math.min(...i.map((e) => e[1].index));
			for (let [t, n, r] of i) if (n.index === e) return a(t, n, r);
		}
		return null;
	}
};
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@shikijs/engine-javascript/1.29.2/97934766166e2c5c5f4d1a1d77d99382f2c7438e6cb55c207f9a2ae918a59437/node_modules/@shikijs/engine-javascript/dist/engine-compile.mjs
function Da(e, t) {
	return wa(e, {
		global: !0,
		hasIndices: !0,
		rules: {
			allowOrphanBackrefs: !0,
			asciiWordBoundaries: !0,
			captureGroup: !0,
			recursionLimit: 5,
			singleline: !0
		},
		...t
	});
}
function Oa(e = {}) {
	let t = Object.assign({
		target: "auto",
		cache: /* @__PURE__ */ new Map()
	}, e);
	return t.regexConstructor ||= (e) => Da(e, { target: t.target }), {
		createScanner(e) {
			return new Ea(e, t);
		},
		createString(e) {
			return { content: e };
		}
	};
}
// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@shikijs/core/1.29.2/4f53f01a066794cc51277f29e62c6a722662d3d199ebf17033264b74446fbec1/node_modules/@shikijs/core/dist/index.mjs
function ka(e) {
	return _("import `createOnigurumaEngine` from `@shikijs/engine-oniguruma` or `shiki/engine/oniguruma` instead"), n(e);
}
function Aa(e) {
	return _("import `createOnigurumaEngine` from `@shikijs/engine-oniguruma` or `shiki/engine/oniguruma` instead"), n(e);
}
function ja(t) {
	return _("import `loadWasm` from `@shikijs/engine-oniguruma` or `shiki/engine/oniguruma` instead"), e(t);
}
function Ma(e) {
	return Array.isArray(e) ? e : [e];
}
function Na(e, t = !1) {
	let n = e.split(/(\r?\n)/g), r = 0, i = [];
	for (let e = 0; e < n.length; e += 2) {
		let a = t ? n[e] + (n[e + 1] || "") : n[e];
		i.push([a, r]), r += n[e].length, r += n[e + 1]?.length || 0;
	}
	return i;
}
function Pa(e) {
	return !e || [
		"plaintext",
		"txt",
		"text",
		"plain"
	].includes(e);
}
function Fa(e) {
	return e === "ansi" || Pa(e);
}
function Ia(e) {
	return e === "none";
}
function La(e) {
	return Ia(e);
}
function Ra(e, t) {
	if (!t) return e;
	e.properties ||= {}, e.properties.class ||= [], typeof e.properties.class == "string" && (e.properties.class = e.properties.class.split(/\s+/g)), Array.isArray(e.properties.class) || (e.properties.class = []);
	let n = Array.isArray(t) ? t : t.split(/\s+/g);
	for (let t of n) t && !e.properties.class.includes(t) && e.properties.class.push(t);
	return e;
}
function za(e, t) {
	let n = 0, r = [];
	for (let i of t) i > n && r.push({
		...e,
		content: e.content.slice(n, i),
		offset: e.offset + n
	}), n = i;
	return n < e.content.length && r.push({
		...e,
		content: e.content.slice(n),
		offset: e.offset + n
	}), r;
}
function Ba(e, t) {
	let n = Array.from(t instanceof Set ? t : new Set(t)).sort((e, t) => e - t);
	return n.length ? e.map((e) => e.flatMap((e) => {
		let t = n.filter((t) => e.offset < t && t < e.offset + e.content.length).map((t) => t - e.offset).sort((e, t) => e - t);
		return t.length ? za(e, t) : e;
	})) : e;
}
async function Va(e) {
	return Promise.resolve(typeof e == "function" ? e() : e).then((e) => e.default || e);
}
function Ha(e, t) {
	let n = typeof e == "string" ? {} : { ...e.colorReplacements }, r = typeof e == "string" ? e : e.name;
	for (let [e, i] of Object.entries(t?.colorReplacements || {})) typeof i == "string" ? n[e] = i : e === r && Object.assign(n, i);
	return n;
}
function Y(e, t) {
	return e && (t?.[e?.toLowerCase()] || e);
}
function Ua(e) {
	let t = {};
	return e.color && (t.color = e.color), e.bgColor && (t["background-color"] = e.bgColor), e.fontStyle && (e.fontStyle & v.Italic && (t["font-style"] = "italic"), e.fontStyle & v.Bold && (t["font-weight"] = "bold"), e.fontStyle & v.Underline && (t["text-decoration"] = "underline")), t;
}
function Wa(e) {
	return typeof e == "string" ? e : Object.entries(e).map(([e, t]) => `${e}:${t}`).join(";");
}
function Ga(e) {
	let t = Na(e, !0).map(([e]) => e);
	function n(n) {
		if (n === e.length) return {
			line: t.length - 1,
			character: t[t.length - 1].length
		};
		let r = n, i = 0;
		for (let e of t) {
			if (r < e.length) break;
			r -= e.length, i++;
		}
		return {
			line: i,
			character: r
		};
	}
	function r(e, n) {
		let r = 0;
		for (let n = 0; n < e; n++) r += t[n].length;
		return r += n, r;
	}
	return {
		lines: t,
		indexToPos: n,
		posToIndex: r
	};
}
var X = class extends Error {
	constructor(e) {
		super(e), this.name = "ShikiError";
	}
}, Ka = /* @__PURE__ */ new WeakMap();
function qa(e, t) {
	Ka.set(e, t);
}
function Ja(e) {
	return Ka.get(e);
}
var Ya = class e {
	_stacks = {};
	lang;
	get themes() {
		return Object.keys(this._stacks);
	}
	get theme() {
		return this.themes[0];
	}
	get _stack() {
		return this._stacks[this.theme];
	}
	static initial(t, n) {
		return new e(Object.fromEntries(Ma(n).map((e) => [e, Ot])), t);
	}
	constructor(...e) {
		if (e.length === 2) {
			let [t, n] = e;
			this.lang = n, this._stacks = t;
		} else {
			let [t, n, r] = e;
			this.lang = n, this._stacks = { [r]: t };
		}
	}
	getInternalStack(e = this.theme) {
		return this._stacks[e];
	}
	get scopes() {
		return _("GrammarState.scopes is deprecated, use GrammarState.getScopes() instead"), Xa(this._stacks[this.theme]);
	}
	getScopes(e = this.theme) {
		return Xa(this._stacks[e]);
	}
	toJSON() {
		return {
			lang: this.lang,
			theme: this.theme,
			themes: this.themes,
			scopes: this.scopes
		};
	}
};
function Xa(e) {
	let t = [], n = /* @__PURE__ */ new Set();
	function r(e) {
		if (n.has(e)) return;
		n.add(e);
		let i = e?.nameScopesList?.scopeName;
		i && t.push(i), e.parent && r(e.parent);
	}
	return r(e), t;
}
function Za(e, t) {
	if (!(e instanceof Ya)) throw new X("Invalid grammar state");
	return e.getInternalStack(t);
}
function Qa() {
	let e = /* @__PURE__ */ new WeakMap();
	function t(t) {
		if (!e.has(t.meta)) {
			let n = function(e) {
				if (typeof e == "number") {
					if (e < 0 || e > t.source.length) throw new X(`Invalid decoration offset: ${e}. Code length: ${t.source.length}`);
					return {
						...r.indexToPos(e),
						offset: e
					};
				}
				{
					let t = r.lines[e.line];
					if (t === void 0) throw new X(`Invalid decoration position ${JSON.stringify(e)}. Lines length: ${r.lines.length}`);
					if (e.character < 0 || e.character > t.length) throw new X(`Invalid decoration position ${JSON.stringify(e)}. Line ${e.line} length: ${t.length}`);
					return {
						...e,
						offset: r.posToIndex(e.line, e.character)
					};
				}
			}, r = Ga(t.source), i = (t.options.decorations || []).map((e) => ({
				...e,
				start: n(e.start),
				end: n(e.end)
			}));
			$a(i), e.set(t.meta, {
				decorations: i,
				converter: r,
				source: t.source
			});
		}
		return e.get(t.meta);
	}
	return {
		name: "shiki:decorations",
		tokens(e) {
			if (this.options.decorations?.length) return Ba(e, t(this).decorations.flatMap((e) => [e.start.offset, e.end.offset]));
		},
		code(e) {
			if (!this.options.decorations?.length) return;
			let n = t(this), r = Array.from(e.children).filter((e) => e.type === "element" && e.tagName === "span");
			if (r.length !== n.converter.lines.length) throw new X(`Number of lines in code element (${r.length}) does not match the number of lines in the source (${n.converter.lines.length}). Failed to apply decorations.`);
			function i(e, t, n, i) {
				let a = r[e], s = "", c = -1, l = -1;
				if (t === 0 && (c = 0), n === 0 && (l = 0), n === Infinity && (l = a.children.length), c === -1 || l === -1) for (let e = 0; e < a.children.length; e++) s += eo(a.children[e]), c === -1 && s.length === t && (c = e + 1), l === -1 && s.length === n && (l = e + 1);
				if (c === -1) throw new X(`Failed to find start index for decoration ${JSON.stringify(i.start)}`);
				if (l === -1) throw new X(`Failed to find end index for decoration ${JSON.stringify(i.end)}`);
				let u = a.children.slice(c, l);
				if (!i.alwaysWrap && u.length === a.children.length) o(a, i, "line");
				else if (!i.alwaysWrap && u.length === 1 && u[0].type === "element") o(u[0], i, "token");
				else {
					let e = {
						type: "element",
						tagName: "span",
						properties: {},
						children: u
					};
					o(e, i, "wrapper"), a.children.splice(c, u.length, e);
				}
			}
			function a(e, t) {
				r[e] = o(r[e], t, "line");
			}
			function o(e, t, n) {
				let r = t.properties || {}, i = t.transform || ((e) => e);
				return e.tagName = t.tagName || "span", e.properties = {
					...e.properties,
					...r,
					class: e.properties.class
				}, t.properties?.class && Ra(e, t.properties.class), e = i(e, n) || e, e;
			}
			let s = [], c = n.decorations.sort((e, t) => t.start.offset - e.start.offset);
			for (let e of c) {
				let { start: t, end: n } = e;
				if (t.line === n.line) i(t.line, t.character, n.character, e);
				else if (t.line < n.line) {
					i(t.line, t.character, Infinity, e);
					for (let r = t.line + 1; r < n.line; r++) s.unshift(() => a(r, e));
					i(n.line, 0, n.character, e);
				}
			}
			s.forEach((e) => e());
		}
	};
}
function $a(e) {
	for (let t = 0; t < e.length; t++) {
		let n = e[t];
		if (n.start.offset > n.end.offset) throw new X(`Invalid decoration range: ${JSON.stringify(n.start)} - ${JSON.stringify(n.end)}`);
		for (let r = t + 1; r < e.length; r++) {
			let t = e[r], i = n.start.offset < t.start.offset && t.start.offset < n.end.offset, a = n.start.offset < t.end.offset && t.end.offset < n.end.offset, o = t.start.offset < n.start.offset && n.start.offset < t.end.offset, s = t.start.offset < n.end.offset && n.end.offset < t.end.offset;
			if (i || a || o || s) {
				if (a && a || o && s) continue;
				throw new X(`Decorations ${JSON.stringify(n.start)} and ${JSON.stringify(t.start)} intersect.`);
			}
		}
	}
}
function eo(e) {
	return e.type === "text" ? e.value : e.type === "element" ? e.children.map(eo).join("") : "";
}
var to = [/* @__PURE__ */ Qa()];
function no(e) {
	return [...e.transformers || [], ...to];
}
var Z = [
	"black",
	"red",
	"green",
	"yellow",
	"blue",
	"magenta",
	"cyan",
	"white",
	"brightBlack",
	"brightRed",
	"brightGreen",
	"brightYellow",
	"brightBlue",
	"brightMagenta",
	"brightCyan",
	"brightWhite"
], ro = {
	1: "bold",
	2: "dim",
	3: "italic",
	4: "underline",
	7: "reverse",
	9: "strikethrough"
};
function io(e, t) {
	let n = e.indexOf("\x1B[", t);
	if (n !== -1) {
		let t = e.indexOf("m", n);
		return {
			sequence: e.substring(n + 2, t).split(";"),
			startPosition: n,
			position: t + 1
		};
	}
	return { position: e.length };
}
function ao(e, t) {
	let n = 1, r = e[t + n++], i;
	if (r === "2") {
		let r = [
			e[t + n++],
			e[t + n++],
			e[t + n]
		].map((e) => Number.parseInt(e));
		r.length === 3 && !r.some((e) => Number.isNaN(e)) && (i = {
			type: "rgb",
			rgb: r
		});
	} else if (r === "5") {
		let r = Number.parseInt(e[t + n]);
		Number.isNaN(r) || (i = {
			type: "table",
			index: Number(r)
		});
	}
	return [n, i];
}
function oo(e) {
	let t = [];
	for (let n = 0; n < e.length; n++) {
		let r = e[n], i = Number.parseInt(r);
		if (!Number.isNaN(i)) {
			if (i === 0) t.push({ type: "resetAll" });
			else if (i <= 9) ro[i] && t.push({
				type: "setDecoration",
				value: ro[i]
			});
			else if (i <= 29) {
				let e = ro[i - 20];
				e && t.push({
					type: "resetDecoration",
					value: e
				});
			} else if (i <= 37) t.push({
				type: "setForegroundColor",
				value: {
					type: "named",
					name: Z[i - 30]
				}
			});
			else if (i === 38) {
				let [r, i] = ao(e, n);
				i && t.push({
					type: "setForegroundColor",
					value: i
				}), n += r;
			} else if (i === 39) t.push({ type: "resetForegroundColor" });
			else if (i <= 47) t.push({
				type: "setBackgroundColor",
				value: {
					type: "named",
					name: Z[i - 40]
				}
			});
			else if (i === 48) {
				let [r, i] = ao(e, n);
				i && t.push({
					type: "setBackgroundColor",
					value: i
				}), n += r;
			} else i === 49 ? t.push({ type: "resetBackgroundColor" }) : i >= 90 && i <= 97 ? t.push({
				type: "setForegroundColor",
				value: {
					type: "named",
					name: Z[i - 90 + 8]
				}
			}) : i >= 100 && i <= 107 && t.push({
				type: "setBackgroundColor",
				value: {
					type: "named",
					name: Z[i - 100 + 8]
				}
			});
		}
	}
	return t;
}
function so() {
	let e = null, t = null, n = /* @__PURE__ */ new Set();
	return { parse(r) {
		let i = [], a = 0;
		do {
			let o = io(r, a), s = o.sequence ? r.substring(a, o.startPosition) : r.substring(a);
			if (s.length > 0 && i.push({
				value: s,
				foreground: e,
				background: t,
				decorations: new Set(n)
			}), o.sequence) {
				let r = oo(o.sequence);
				for (let i of r) i.type === "resetAll" ? (e = null, t = null, n.clear()) : i.type === "resetForegroundColor" ? e = null : i.type === "resetBackgroundColor" ? t = null : i.type === "resetDecoration" && n.delete(i.value);
				for (let i of r) i.type === "setForegroundColor" ? e = i.value : i.type === "setBackgroundColor" ? t = i.value : i.type === "setDecoration" && n.add(i.value);
			}
			a = o.position;
		} while (a < r.length);
		return i;
	} };
}
var co = {
	black: "#000000",
	red: "#bb0000",
	green: "#00bb00",
	yellow: "#bbbb00",
	blue: "#0000bb",
	magenta: "#ff00ff",
	cyan: "#00bbbb",
	white: "#eeeeee",
	brightBlack: "#555555",
	brightRed: "#ff5555",
	brightGreen: "#00ff00",
	brightYellow: "#ffff55",
	brightBlue: "#5555ff",
	brightMagenta: "#ff55ff",
	brightCyan: "#55ffff",
	brightWhite: "#ffffff"
};
function lo(e = co) {
	function t(t) {
		return e[t];
	}
	function n(e) {
		return `#${e.map((e) => Math.max(0, Math.min(e, 255)).toString(16).padStart(2, "0")).join("")}`;
	}
	let r;
	function i() {
		if (r) return r;
		r = [];
		for (let e = 0; e < Z.length; e++) r.push(t(Z[e]));
		let e = [
			0,
			95,
			135,
			175,
			215,
			255
		];
		for (let t = 0; t < 6; t++) for (let i = 0; i < 6; i++) for (let a = 0; a < 6; a++) r.push(n([
			e[t],
			e[i],
			e[a]
		]));
		let i = 8;
		for (let e = 0; e < 24; e++, i += 10) r.push(n([
			i,
			i,
			i
		]));
		return r;
	}
	function a(e) {
		return i()[e];
	}
	function o(e) {
		switch (e.type) {
			case "named": return t(e.name);
			case "rgb": return n(e.rgb);
			case "table": return a(e.index);
		}
	}
	return { value: o };
}
function uo(e, t, n) {
	let r = Ha(e, n), i = Na(t), a = lo(Object.fromEntries(Z.map((t) => [t, e.colors?.[`terminal.ansi${t[0].toUpperCase()}${t.substring(1)}`]]))), o = so();
	return i.map((t) => o.parse(t[0]).map((n) => {
		let i, o;
		n.decorations.has("reverse") ? (i = n.background ? a.value(n.background) : e.bg, o = n.foreground ? a.value(n.foreground) : e.fg) : (i = n.foreground ? a.value(n.foreground) : e.fg, o = n.background ? a.value(n.background) : void 0), i = Y(i, r), o = Y(o, r), n.decorations.has("dim") && (i = fo(i));
		let s = v.None;
		return n.decorations.has("bold") && (s |= v.Bold), n.decorations.has("italic") && (s |= v.Italic), n.decorations.has("underline") && (s |= v.Underline), {
			content: n.value,
			offset: t[1],
			color: i,
			bgColor: o,
			fontStyle: s
		};
	}));
}
function fo(e) {
	let t = e.match(/#([0-9a-f]{3})([0-9a-f]{3})?([0-9a-f]{2})?/);
	if (t) {
		if (t[3]) {
			let e = Math.round(Number.parseInt(t[3], 16) / 2).toString(16).padStart(2, "0");
			return `#${t[1]}${t[2]}${e}`;
		}
		return t[2] ? `#${t[1]}${t[2]}80` : `#${Array.from(t[1]).map((e) => `${e}${e}`).join("")}80`;
	}
	let n = e.match(/var\((--[\w-]+-ansi-[\w-]+)\)/);
	return n ? `var(${n[1]}-dim)` : e;
}
function po(e, t, n = {}) {
	let { lang: r = "text", theme: i = e.getLoadedThemes()[0] } = n;
	if (Pa(r) || Ia(i)) return Na(t).map((e) => [{
		content: e[0],
		offset: e[1]
	}]);
	let { theme: a, colorMap: o } = e.setTheme(i);
	if (r === "ansi") return uo(a, t, n);
	let s = e.getLanguage(r);
	if (n.grammarState) {
		if (n.grammarState.lang !== s.name) throw new p(`Grammar state language "${n.grammarState.lang}" does not match highlight language "${s.name}"`);
		if (!n.grammarState.themes.includes(a.name)) throw new p(`Grammar state themes "${n.grammarState.themes}" do not contain highlight theme "${a.name}"`);
	}
	return ho(t, s, a, o, n);
}
function mo(...e) {
	if (e.length === 2) return Ja(e[1]);
	let [t, n, r = {}] = e, { lang: i = "text", theme: a = t.getLoadedThemes()[0] } = r;
	if (Pa(i) || Ia(a)) throw new p("Plain language does not have grammar state");
	if (i === "ansi") throw new p("ANSI language does not have grammar state");
	let { theme: o, colorMap: s } = t.setTheme(a), c = t.getLanguage(i);
	return new Ya(go(n, c, o, s, r).stateStack, c.name, o.name);
}
function ho(e, t, n, r, i) {
	let a = go(e, t, n, r, i), o = new Ya(go(e, t, n, r, i).stateStack, t.name, n.name);
	return qa(a.tokens, o), a.tokens;
}
function go(e, t, n, r, i) {
	let a = Ha(n, i), { tokenizeMaxLineLength: o = 0, tokenizeTimeLimit: s = 500 } = i, c = Na(e), l = i.grammarState ? Za(i.grammarState, n.name) ?? Ot : i.grammarContextCode == null ? Ot : go(i.grammarContextCode, t, n, r, {
		...i,
		grammarState: void 0,
		grammarContextCode: void 0
	}).stateStack, u = [], d = [];
	for (let e = 0, f = c.length; e < f; e++) {
		let [f, p] = c[e];
		if (f === "") {
			u = [], d.push([]);
			continue;
		}
		if (o > 0 && f.length >= o) {
			u = [], d.push([{
				content: f,
				offset: p,
				color: "",
				fontStyle: 0
			}]);
			continue;
		}
		let m, h, g;
		i.includeExplanation && (m = t.tokenizeLine(f, l), h = m.tokens, g = 0);
		let _ = t.tokenizeLine2(f, l, s), ee = _.tokens.length / 2;
		for (let e = 0; e < ee; e++) {
			let t = _.tokens[2 * e], o = e + 1 < ee ? _.tokens[2 * e + 2] : f.length;
			if (t === o) continue;
			let s = _.tokens[2 * e + 1], c = Y(r[y.getForeground(s)], a), l = y.getFontStyle(s), d = {
				content: f.substring(t, o),
				offset: p + t,
				color: c,
				fontStyle: l
			};
			if (i.includeExplanation) {
				let e = [];
				if (i.includeExplanation !== "scopeName") for (let t of n.settings) {
					let n;
					switch (typeof t.scope) {
						case "string":
							n = t.scope.split(/,/).map((e) => e.trim());
							break;
						case "object":
							n = t.scope;
							break;
						default: continue;
					}
					e.push({
						settings: t,
						selectors: n.map((e) => e.split(/ /))
					});
				}
				d.explanation = [];
				let r = 0;
				for (; t + r < o;) {
					let t = h[g], n = f.substring(t.startIndex, t.endIndex);
					r += n.length, d.explanation.push({
						content: n,
						scopes: i.includeExplanation === "scopeName" ? _o(t.scopes) : vo(e, t.scopes)
					}), g += 1;
				}
			}
			u.push(d);
		}
		d.push(u), u = [], l = _.ruleStack;
	}
	return {
		tokens: d,
		stateStack: l
	};
}
function _o(e) {
	return e.map((e) => ({ scopeName: e }));
}
function vo(e, t) {
	let n = [];
	for (let r = 0, i = t.length; r < i; r++) {
		let i = t[r];
		n[r] = {
			scopeName: i,
			themeMatches: xo(e, i, t.slice(0, r))
		};
	}
	return n;
}
function yo(e, t) {
	return e === t || t.substring(0, e.length) === e && t[e.length] === ".";
}
function bo(e, t, n) {
	if (!yo(e[e.length - 1], t)) return !1;
	let r = e.length - 2, i = n.length - 1;
	for (; r >= 0 && i >= 0;) yo(e[r], n[i]) && --r, --i;
	return r === -1;
}
function xo(e, t, n) {
	let r = [];
	for (let { selectors: i, settings: a } of e) for (let e of i) if (bo(e, t, n)) {
		r.push(a);
		break;
	}
	return r;
}
function So(e, t, n) {
	let r = Object.entries(n.themes).filter((e) => e[1]).map((e) => ({
		color: e[0],
		theme: e[1]
	})), i = r.map((r) => {
		let i = po(e, t, {
			...n,
			theme: r.theme
		});
		return {
			tokens: i,
			state: Ja(i),
			theme: typeof r.theme == "string" ? r.theme : r.theme.name
		};
	}), a = Co(...i.map((e) => e.tokens)), o = a[0].map((e, t) => e.map((e, i) => {
		let o = {
			content: e.content,
			variants: {},
			offset: e.offset
		};
		return "includeExplanation" in n && n.includeExplanation && (o.explanation = e.explanation), a.forEach((e, n) => {
			let { content: a, explanation: s, offset: c, ...l } = e[t][i];
			o.variants[r[n].color] = l;
		}), o;
	})), s = i[0].state ? new Ya(Object.fromEntries(i.map((e) => [e.theme, e.state?.getInternalStack(e.theme)])), i[0].state.lang) : void 0;
	return s && qa(o, s), o;
}
function Co(...e) {
	let t = e.map(() => []), n = e.length;
	for (let r = 0; r < e[0].length; r++) {
		let i = e.map((e) => e[r]), a = t.map(() => []);
		t.forEach((e, t) => e.push(a[t]));
		let o = i.map(() => 0), s = i.map((e) => e[0]);
		for (; s.every((e) => e);) {
			let e = Math.min(...s.map((e) => e.content.length));
			for (let t = 0; t < n; t++) {
				let n = s[t];
				n.content.length === e ? (a[t].push(n), o[t] += 1, s[t] = i[t][o[t]]) : (a[t].push({
					...n,
					content: n.content.slice(0, e)
				}), s[t] = {
					...n,
					content: n.content.slice(e),
					offset: n.offset + e
				});
			}
		}
	}
	return t;
}
function Q(e, t, n) {
	let r, i, a, o, s, c;
	if ("themes" in n) {
		let { defaultColor: l = "light", cssVariablePrefix: u = "--shiki-" } = n, d = Object.entries(n.themes).filter((e) => e[1]).map((e) => ({
			color: e[0],
			theme: e[1]
		})).sort((e, t) => e.color === l ? -1 : +(t.color === l));
		if (d.length === 0) throw new p("`themes` option must not be empty");
		let f = So(e, t, n);
		if (c = Ja(f), l && !d.find((e) => e.color === l)) throw new p(`\`themes\` option must contain the defaultColor key \`${l}\``);
		let m = d.map((t) => e.getTheme(t.theme)), h = d.map((e) => e.color);
		a = f.map((e) => e.map((e) => wo(e, h, u, l))), c && qa(a, c);
		let g = d.map((e) => Ha(e.theme, n));
		i = d.map((e, t) => (t === 0 && l ? "" : `${u + e.color}:`) + (Y(m[t].fg, g[t]) || "inherit")).join(";"), r = d.map((e, t) => (t === 0 && l ? "" : `${u + e.color}-bg:`) + (Y(m[t].bg, g[t]) || "inherit")).join(";"), o = `shiki-themes ${m.map((e) => e.name).join(" ")}`, s = l ? void 0 : [i, r].join(";");
	} else if ("theme" in n) {
		let s = Ha(n.theme, n);
		a = po(e, t, n);
		let l = e.getTheme(n.theme);
		r = Y(l.bg, s), i = Y(l.fg, s), o = l.name, c = Ja(a);
	} else throw new p("Invalid options, either `theme` or `themes` must be provided");
	return {
		tokens: a,
		fg: i,
		bg: r,
		themeName: o,
		rootStyle: s,
		grammarState: c
	};
}
function wo(e, t, n, r) {
	let i = {
		content: e.content,
		explanation: e.explanation,
		offset: e.offset
	}, a = t.map((t) => Ua(e.variants[t])), o = new Set(a.flatMap((e) => Object.keys(e))), s = {};
	return a.forEach((e, i) => {
		for (let a of o) {
			let o = e[a] || "inherit";
			if (i === 0 && r) s[a] = o;
			else {
				let e = a === "color" ? "" : a === "background-color" ? "-bg" : `-${a}`, r = n + t[i] + (a === "color" ? "" : e);
				s[r] = o;
			}
		}
	}), i.htmlStyle = s, i;
}
function $(e, t, n, r = {
	meta: {},
	options: n,
	codeToHast: (t, n) => $(e, t, n),
	codeToTokens: (t, n) => Q(e, t, n)
}) {
	let i = t;
	for (let e of no(n)) i = e.preprocess?.call(r, i, n) || i;
	let { tokens: a, fg: o, bg: s, themeName: c, rootStyle: l, grammarState: u } = Q(e, i, n), { mergeWhitespaces: d = !0 } = n;
	d === !0 ? a = Eo(a) : d === "never" && (a = Do(a));
	let f = {
		...r,
		get source() {
			return i;
		}
	};
	for (let e of no(n)) a = e.tokens?.call(f, a) || a;
	return To(a, {
		...n,
		fg: o,
		bg: s,
		themeName: c,
		rootStyle: l
	}, f, u);
}
function To(e, t, n, r = Ja(e)) {
	let i = no(t), a = [], o = {
		type: "root",
		children: []
	}, { structure: s = "classic", tabindex: c = "0" } = t, l = {
		type: "element",
		tagName: "pre",
		properties: {
			class: `shiki ${t.themeName || ""}`,
			style: t.rootStyle || `background-color:${t.bg};color:${t.fg}`,
			...c !== !1 && c != null ? { tabindex: c.toString() } : {},
			...Object.fromEntries(Array.from(Object.entries(t.meta || {})).filter(([e]) => !e.startsWith("_")))
		},
		children: []
	}, u = {
		type: "element",
		tagName: "code",
		properties: {},
		children: a
	}, d = [], f = {
		...n,
		structure: s,
		addClassToHast: Ra,
		get source() {
			return n.source;
		},
		get tokens() {
			return e;
		},
		get options() {
			return t;
		},
		get root() {
			return o;
		},
		get pre() {
			return l;
		},
		get code() {
			return u;
		},
		get lines() {
			return d;
		}
	};
	if (e.forEach((e, t) => {
		t && (s === "inline" ? o.children.push({
			type: "element",
			tagName: "br",
			properties: {},
			children: []
		}) : s === "classic" && a.push({
			type: "text",
			value: "\n"
		}));
		let n = {
			type: "element",
			tagName: "span",
			properties: { class: "line" },
			children: []
		}, r = 0;
		for (let a of e) {
			let e = {
				type: "element",
				tagName: "span",
				properties: { ...a.htmlAttrs },
				children: [{
					type: "text",
					value: a.content
				}]
			};
			typeof a.htmlStyle == "string" && _("`htmlStyle` as a string is deprecated. Use an object instead.");
			let c = Wa(a.htmlStyle || Ua(a));
			c && (e.properties.style = c);
			for (let o of i) e = o?.span?.call(f, e, t + 1, r, n, a) || e;
			s === "inline" ? o.children.push(e) : s === "classic" && n.children.push(e), r += a.content.length;
		}
		if (s === "classic") {
			for (let e of i) n = e?.line?.call(f, n, t + 1) || n;
			d.push(n), a.push(n);
		}
	}), s === "classic") {
		for (let e of i) u = e?.code?.call(f, u) || u;
		l.children.push(u);
		for (let e of i) l = e?.pre?.call(f, l) || l;
		o.children.push(l);
	}
	let p = o;
	for (let e of i) p = e?.root?.call(f, p) || p;
	return r && qa(p, r), p;
}
function Eo(e) {
	return e.map((e) => {
		let t = [], n = "", r = 0;
		return e.forEach((i, a) => {
			let o = !(i.fontStyle && i.fontStyle & v.Underline);
			o && i.content.match(/^\s+$/) && e[a + 1] ? (r ||= i.offset, n += i.content) : n ? (o ? t.push({
				...i,
				offset: r,
				content: n + i.content
			}) : t.push({
				content: n,
				offset: r
			}, i), r = 0, n = "") : t.push(i);
		}), t;
	});
}
function Do(e) {
	return e.map((e) => e.flatMap((e) => {
		if (e.content.match(/^\s+$/)) return e;
		let t = e.content.match(/^(\s*)(.*?)(\s*)$/);
		if (!t) return e;
		let [, n, r, i] = t;
		if (!n && !i) return e;
		let a = [{
			...e,
			offset: e.offset + n.length,
			content: r
		}];
		return n && a.unshift({
			content: n,
			offset: e.offset
		}), i && a.push({
			content: i,
			offset: e.offset + n.length + r.length
		}), a;
	}));
}
function Oo(e, t, n) {
	let r = {
		meta: {},
		options: n,
		codeToHast: (t, n) => $(e, t, n),
		codeToTokens: (t, n) => Q(e, t, n)
	}, i = Sn($(e, t, n, r));
	for (let e of no(n)) i = e.postprocess?.call(r, i, n) || i;
	return i;
}
var ko = {
	light: "#333333",
	dark: "#bbbbbb"
}, Ao = {
	light: "#fffffe",
	dark: "#1e1e1e"
}, jo = "__shiki_resolved";
function Mo(e) {
	if (e?.[jo]) return e;
	let t = { ...e };
	t.tokenColors && !t.settings && (t.settings = t.tokenColors, delete t.tokenColors), t.type ||= "dark", t.colorReplacements = { ...t.colorReplacements }, t.settings ||= [];
	let { bg: n, fg: r } = t;
	if (!n || !r) {
		let e = t.settings ? t.settings.find((e) => !e.name && !e.scope) : void 0;
		e?.settings?.foreground && (r = e.settings.foreground), e?.settings?.background && (n = e.settings.background), !r && t?.colors?.["editor.foreground"] && (r = t.colors["editor.foreground"]), !n && t?.colors?.["editor.background"] && (n = t.colors["editor.background"]), r ||= t.type === "light" ? ko.light : ko.dark, n ||= t.type === "light" ? Ao.light : Ao.dark, t.fg = r, t.bg = n;
	}
	t.settings[0] && t.settings[0].settings && !t.settings[0].scope || t.settings.unshift({ settings: {
		foreground: t.fg,
		background: t.bg
	} });
	let i = 0, a = /* @__PURE__ */ new Map();
	function o(e) {
		if (a.has(e)) return a.get(e);
		i += 1;
		let n = `#${i.toString(16).padStart(8, "0").toLowerCase()}`;
		return t.colorReplacements?.[`#${n}`] ? o(e) : (a.set(e, n), n);
	}
	t.settings = t.settings.map((e) => {
		let n = e.settings?.foreground && !e.settings.foreground.startsWith("#"), r = e.settings?.background && !e.settings.background.startsWith("#");
		if (!n && !r) return e;
		let i = {
			...e,
			settings: { ...e.settings }
		};
		if (n) {
			let n = o(e.settings.foreground);
			t.colorReplacements[n] = e.settings.foreground, i.settings.foreground = n;
		}
		if (r) {
			let n = o(e.settings.background);
			t.colorReplacements[n] = e.settings.background, i.settings.background = n;
		}
		return i;
	});
	for (let e of Object.keys(t.colors || {})) if ((e === "editor.foreground" || e === "editor.background" || e.startsWith("terminal.ansi")) && !t.colors[e]?.startsWith("#")) {
		let n = o(t.colors[e]);
		t.colorReplacements[n] = t.colors[e], t.colors[e] = n;
	}
	return Object.defineProperty(t, jo, {
		enumerable: !1,
		writable: !1,
		value: !0
	}), t;
}
async function No(e) {
	return Array.from(new Set((await Promise.all(e.filter((e) => !Fa(e)).map(async (e) => await Va(e).then((e) => Array.isArray(e) ? e : [e])))).flat()));
}
async function Po(e) {
	return (await Promise.all(e.map(async (e) => La(e) ? null : Mo(await Va(e))))).filter((e) => !!e);
}
var Fo = class extends Dt {
	constructor(e, t, n, r = {}) {
		super(e), this._resolver = e, this._themes = t, this._langs = n, this._alias = r, this._themes.map((e) => this.loadTheme(e)), this.loadLanguages(this._langs);
	}
	_resolvedThemes = /* @__PURE__ */ new Map();
	_resolvedGrammars = /* @__PURE__ */ new Map();
	_langMap = /* @__PURE__ */ new Map();
	_langGraph = /* @__PURE__ */ new Map();
	_textmateThemeCache = /* @__PURE__ */ new WeakMap();
	_loadedThemesCache = null;
	_loadedLanguagesCache = null;
	getTheme(e) {
		return typeof e == "string" ? this._resolvedThemes.get(e) : this.loadTheme(e);
	}
	loadTheme(e) {
		let t = Mo(e);
		return t.name && (this._resolvedThemes.set(t.name, t), this._loadedThemesCache = null), t;
	}
	getLoadedThemes() {
		return this._loadedThemesCache ||= [...this._resolvedThemes.keys()], this._loadedThemesCache;
	}
	setTheme(e) {
		let t = this._textmateThemeCache.get(e);
		t || (t = pe.createFromRawTheme(e), this._textmateThemeCache.set(e, t)), this._syncRegistry.setTheme(t);
	}
	getGrammar(e) {
		if (this._alias[e]) {
			let t = /* @__PURE__ */ new Set([e]);
			for (; this._alias[e];) {
				if (e = this._alias[e], t.has(e)) throw new X(`Circular alias \`${Array.from(t).join(" -> ")} -> ${e}\``);
				t.add(e);
			}
		}
		return this._resolvedGrammars.get(e);
	}
	loadLanguage(e) {
		if (this.getGrammar(e.name)) return;
		let t = new Set([...this._langMap.values()].filter((t) => t.embeddedLangsLazy?.includes(e.name)));
		this._resolver.addLanguage(e);
		let n = {
			balancedBracketSelectors: e.balancedBracketSelectors || ["*"],
			unbalancedBracketSelectors: e.unbalancedBracketSelectors || []
		};
		this._syncRegistry._rawGrammars.set(e.scopeName, e);
		let r = this.loadGrammarWithConfiguration(e.scopeName, 1, n);
		if (r.name = e.name, this._resolvedGrammars.set(e.name, r), e.aliases && e.aliases.forEach((t) => {
			this._alias[t] = e.name;
		}), this._loadedLanguagesCache = null, t.size) for (let e of t) this._resolvedGrammars.delete(e.name), this._loadedLanguagesCache = null, this._syncRegistry?._injectionGrammars?.delete(e.scopeName), this._syncRegistry?._grammars?.delete(e.scopeName), this.loadLanguage(this._langMap.get(e.name));
	}
	dispose() {
		super.dispose(), this._resolvedThemes.clear(), this._resolvedGrammars.clear(), this._langMap.clear(), this._langGraph.clear(), this._loadedThemesCache = null;
	}
	loadLanguages(e) {
		for (let t of e) this.resolveEmbeddedLanguages(t);
		let t = Array.from(this._langGraph.entries()), n = t.filter(([e, t]) => !t);
		if (n.length) {
			let e = t.filter(([e, t]) => t && t.embeddedLangs?.some((e) => n.map(([e]) => e).includes(e))).filter((e) => !n.includes(e));
			throw new X(`Missing languages ${n.map(([e]) => `\`${e}\``).join(", ")}, required by ${e.map(([e]) => `\`${e}\``).join(", ")}`);
		}
		for (let [e, n] of t) this._resolver.addLanguage(n);
		for (let [e, n] of t) this.loadLanguage(n);
	}
	getLoadedLanguages() {
		return this._loadedLanguagesCache ||= [.../* @__PURE__ */ new Set([...this._resolvedGrammars.keys(), ...Object.keys(this._alias)])], this._loadedLanguagesCache;
	}
	resolveEmbeddedLanguages(e) {
		if (this._langMap.set(e.name, e), this._langGraph.set(e.name, e), e.embeddedLangs) for (let t of e.embeddedLangs) this._langGraph.set(t, this._langMap.get(t));
	}
}, Io = class {
	_langs = /* @__PURE__ */ new Map();
	_scopeToLang = /* @__PURE__ */ new Map();
	_injections = /* @__PURE__ */ new Map();
	_onigLib;
	constructor(e, t) {
		this._onigLib = {
			createOnigScanner: (t) => e.createScanner(t),
			createOnigString: (t) => e.createString(t)
		}, t.forEach((e) => this.addLanguage(e));
	}
	get onigLib() {
		return this._onigLib;
	}
	getLangRegistration(e) {
		return this._langs.get(e);
	}
	loadGrammar(e) {
		return this._scopeToLang.get(e);
	}
	addLanguage(e) {
		this._langs.set(e.name, e), e.aliases && e.aliases.forEach((t) => {
			this._langs.set(t, e);
		}), this._scopeToLang.set(e.scopeName, e), e.injectTo && e.injectTo.forEach((t) => {
			this._injections.get(t) || this._injections.set(t, []), this._injections.get(t).push(e.scopeName);
		});
	}
	getInjections(e) {
		let t = e.split("."), n = [];
		for (let e = 1; e <= t.length; e++) {
			let r = t.slice(0, e).join(".");
			n = [...n, ...this._injections.get(r) || []];
		}
		return n;
	}
}, Lo = 0;
function Ro(e) {
	Lo += 1, e.warnings !== !1 && Lo >= 10 && Lo % 10 == 0 && console.warn(`[Shiki] ${Lo} instances have been created. Shiki is supposed to be used as a singleton, consider refactoring your code to cache your highlighter instance; Or call \`highlighter.dispose()\` to release unused instances.`);
	let t = !1;
	if (!e.engine) throw new X("`engine` option is required for synchronous mode");
	let n = (e.langs || []).flat(1), r = (e.themes || []).flat(1).map(Mo), i = new Fo(new Io(e.engine, n), r, n, e.langAlias), a;
	function o(e) {
		h();
		let t = i.getGrammar(typeof e == "string" ? e : e.name);
		if (!t) throw new X(`Language \`${e}\` not found, you may need to load it first`);
		return t;
	}
	function s(e) {
		if (e === "none") return {
			bg: "",
			fg: "",
			name: "none",
			settings: [],
			type: "dark"
		};
		h();
		let t = i.getTheme(e);
		if (!t) throw new X(`Theme \`${e}\` not found, you may need to load it first`);
		return t;
	}
	function c(e) {
		h();
		let t = s(e);
		return a !== e && (i.setTheme(t), a = e), {
			theme: t,
			colorMap: i.getColorMap()
		};
	}
	function l() {
		return h(), i.getLoadedThemes();
	}
	function u() {
		return h(), i.getLoadedLanguages();
	}
	function d(...e) {
		h(), i.loadLanguages(e.flat(1));
	}
	async function f(...e) {
		return d(await No(e));
	}
	function p(...e) {
		h();
		for (let t of e.flat(1)) i.loadTheme(t);
	}
	async function m(...e) {
		return h(), p(await Po(e));
	}
	function h() {
		if (t) throw new X("Shiki instance has been disposed");
	}
	function g() {
		t || (t = !0, i.dispose(), --Lo);
	}
	return {
		setTheme: c,
		getTheme: s,
		getLanguage: o,
		getLoadedThemes: l,
		getLoadedLanguages: u,
		loadLanguage: f,
		loadLanguageSync: d,
		loadTheme: m,
		loadThemeSync: p,
		dispose: g,
		[Symbol.dispose]: g
	};
}
async function zo(e = {}) {
	e.loadWasm && _("`loadWasm` option is deprecated. Use `engine: createOnigurumaEngine(loadWasm)` instead.");
	let [r, i, a] = await Promise.all([
		Po(e.themes || []),
		No(e.langs || []),
		e.engine || n(e.loadWasm || t())
	]);
	return Ro({
		...e,
		loadWasm: void 0,
		themes: r,
		langs: i,
		engine: a
	});
}
function Bo(e = {}) {
	return _("`getShikiInternal` is deprecated. Use `createShikiInternal` instead."), zo(e);
}
async function Vo(e = {}) {
	let t = await zo(e);
	return {
		getLastGrammarState: (...e) => mo(t, ...e),
		codeToTokensBase: (e, n) => po(t, e, n),
		codeToTokensWithThemes: (e, n) => So(t, e, n),
		codeToTokens: (e, n) => Q(t, e, n),
		codeToHast: (e, n) => $(t, e, n),
		codeToHtml: (e, n) => Oo(t, e, n),
		...t,
		getInternalContext: () => t
	};
}
function Ho(e = {}) {
	let t = Ro(e);
	return {
		getLastGrammarState: (...e) => mo(t, ...e),
		codeToTokensBase: (e, n) => po(t, e, n),
		codeToTokensWithThemes: (e, n) => So(t, e, n),
		codeToTokens: (e, n) => Q(t, e, n),
		codeToHast: (e, n) => $(t, e, n),
		codeToHtml: (e, n) => Oo(t, e, n),
		...t,
		getInternalContext: () => t
	};
}
function Uo(e) {
	let t;
	async function n(n = {}) {
		if (t) {
			let e = await t;
			return await Promise.all([e.loadTheme(...n.themes || []), e.loadLanguage(...n.langs || [])]), e;
		}
		return t = e({
			...n,
			themes: n.themes || [],
			langs: n.langs || []
		}), t;
	}
	return n;
}
var Wo = /* @__PURE__ */ Uo(Vo);
function Go(e = {}) {
	return _("`getHighlighterCore` is deprecated. Use `createHighlighterCore` or `getSingletonHighlighterCore` instead."), Vo(e);
}
function Ko(e, t, n) {
	let r, i, a;
	if (t) _("`createdBundledHighlighter` signature with `bundledLanguages` and `bundledThemes` is deprecated. Use the options object signature instead."), r = e, i = t, a = () => ka(n);
	else {
		let t = e;
		r = t.langs, i = t.themes, a = t.engine;
	}
	async function o(e) {
		function t(e) {
			if (typeof e == "string") {
				if (Fa(e)) return [];
				let t = r[e];
				if (!t) throw new p(`Language \`${e}\` is not included in this bundle. You may want to load it from external source.`);
				return t;
			}
			return e;
		}
		function n(e) {
			if (La(e)) return "none";
			if (typeof e == "string") {
				let t = i[e];
				if (!t) throw new p(`Theme \`${e}\` is not included in this bundle. You may want to load it from external source.`);
				return t;
			}
			return e;
		}
		let o = (e.themes ?? []).map((e) => n(e)), s = (e.langs ?? []).map((e) => t(e)), c = await Vo({
			engine: e.engine ?? a(),
			...e,
			themes: o,
			langs: s
		});
		return {
			...c,
			loadLanguage(...e) {
				return c.loadLanguage(...e.map(t));
			},
			loadTheme(...e) {
				return c.loadTheme(...e.map(n));
			}
		};
	}
	return o;
}
function qo(e) {
	let t;
	async function n(n = {}) {
		if (t) {
			let e = await t;
			return await Promise.all([e.loadTheme(...n.themes || []), e.loadLanguage(...n.langs || [])]), e;
		}
		return t = e({
			...n,
			themes: n.themes || [],
			langs: n.langs || []
		}), t;
	}
	return n;
}
function Jo(e) {
	let t = qo(e);
	return {
		getSingletonHighlighter(e) {
			return t(e);
		},
		async codeToHtml(e, n) {
			return (await t({
				langs: [n.lang],
				themes: "theme" in n ? [n.theme] : Object.values(n.themes)
			})).codeToHtml(e, n);
		},
		async codeToHast(e, n) {
			return (await t({
				langs: [n.lang],
				themes: "theme" in n ? [n.theme] : Object.values(n.themes)
			})).codeToHast(e, n);
		},
		async codeToTokens(e, n) {
			return (await t({
				langs: [n.lang],
				themes: "theme" in n ? [n.theme] : Object.values(n.themes)
			})).codeToTokens(e, n);
		},
		async codeToTokensBase(e, n) {
			return (await t({
				langs: [n.lang],
				themes: [n.theme]
			})).codeToTokensBase(e, n);
		},
		async codeToTokensWithThemes(e, n) {
			return (await t({
				langs: [n.lang],
				themes: Object.values(n.themes).filter(Boolean)
			})).codeToTokensWithThemes(e, n);
		},
		async getLastGrammarState(e, n) {
			return (await t({
				langs: [n.lang],
				themes: [n.theme]
			})).getLastGrammarState(e, n);
		}
	};
}
function Yo(e) {
	return _("import `createJavaScriptRegexEngine` from `@shikijs/engine-javascript` or `shiki/engine/javascript` instead"), Oa(e);
}
function Xo(e) {
	return _("import `defaultJavaScriptRegexConstructor` from `@shikijs/engine-javascript` or `shiki/engine/javascript` instead"), Da(e);
}
function Zo(e = {}) {
	let { name: t = "css-variables", variablePrefix: n = "--shiki-", fontStyle: r = !0 } = e, i = (t) => e.variableDefaults?.[t] ? `var(${n}${t}, ${e.variableDefaults[t]})` : `var(${n}${t})`, a = {
		name: t,
		type: "dark",
		colors: {
			"editor.foreground": i("foreground"),
			"editor.background": i("background"),
			"terminal.ansiBlack": i("ansi-black"),
			"terminal.ansiRed": i("ansi-red"),
			"terminal.ansiGreen": i("ansi-green"),
			"terminal.ansiYellow": i("ansi-yellow"),
			"terminal.ansiBlue": i("ansi-blue"),
			"terminal.ansiMagenta": i("ansi-magenta"),
			"terminal.ansiCyan": i("ansi-cyan"),
			"terminal.ansiWhite": i("ansi-white"),
			"terminal.ansiBrightBlack": i("ansi-bright-black"),
			"terminal.ansiBrightRed": i("ansi-bright-red"),
			"terminal.ansiBrightGreen": i("ansi-bright-green"),
			"terminal.ansiBrightYellow": i("ansi-bright-yellow"),
			"terminal.ansiBrightBlue": i("ansi-bright-blue"),
			"terminal.ansiBrightMagenta": i("ansi-bright-magenta"),
			"terminal.ansiBrightCyan": i("ansi-bright-cyan"),
			"terminal.ansiBrightWhite": i("ansi-bright-white")
		},
		tokenColors: [
			{
				scope: [
					"keyword.operator.accessor",
					"meta.group.braces.round.function.arguments",
					"meta.template.expression",
					"markup.fenced_code meta.embedded.block"
				],
				settings: { foreground: i("foreground") }
			},
			{
				scope: "emphasis",
				settings: { fontStyle: "italic" }
			},
			{
				scope: [
					"strong",
					"markup.heading.markdown",
					"markup.bold.markdown"
				],
				settings: { fontStyle: "bold" }
			},
			{
				scope: ["markup.italic.markdown"],
				settings: { fontStyle: "italic" }
			},
			{
				scope: "meta.link.inline.markdown",
				settings: {
					fontStyle: "underline",
					foreground: i("token-link")
				}
			},
			{
				scope: [
					"string",
					"markup.fenced_code",
					"markup.inline"
				],
				settings: { foreground: i("token-string") }
			},
			{
				scope: ["comment", "string.quoted.docstring.multi"],
				settings: { foreground: i("token-comment") }
			},
			{
				scope: [
					"constant.numeric",
					"constant.language",
					"constant.other.placeholder",
					"constant.character.format.placeholder",
					"variable.language.this",
					"variable.other.object",
					"variable.other.class",
					"variable.other.constant",
					"meta.property-name",
					"meta.property-value",
					"support"
				],
				settings: { foreground: i("token-constant") }
			},
			{
				scope: [
					"keyword",
					"storage.modifier",
					"storage.type",
					"storage.control.clojure",
					"entity.name.function.clojure",
					"entity.name.tag.yaml",
					"support.function.node",
					"support.type.property-name.json",
					"punctuation.separator.key-value",
					"punctuation.definition.template-expression"
				],
				settings: { foreground: i("token-keyword") }
			},
			{
				scope: "variable.parameter.function",
				settings: { foreground: i("token-parameter") }
			},
			{
				scope: [
					"support.function",
					"entity.name.type",
					"entity.other.inherited-class",
					"meta.function-call",
					"meta.instance.constructor",
					"entity.other.attribute-name",
					"entity.name.function",
					"constant.keyword.clojure"
				],
				settings: { foreground: i("token-function") }
			},
			{
				scope: [
					"entity.name.tag",
					"string.quoted",
					"string.regexp",
					"string.interpolated",
					"string.template",
					"string.unquoted.plain.out.yaml",
					"keyword.other.template"
				],
				settings: { foreground: i("token-string-expression") }
			},
			{
				scope: [
					"punctuation.definition.arguments",
					"punctuation.definition.dict",
					"punctuation.separator",
					"meta.function-call.arguments"
				],
				settings: { foreground: i("token-punctuation") }
			},
			{
				scope: ["markup.underline.link", "punctuation.definition.metadata.markdown"],
				settings: { foreground: i("token-link") }
			},
			{
				scope: ["beginning.punctuation.definition.list.markdown"],
				settings: { foreground: i("token-string") }
			},
			{
				scope: [
					"punctuation.definition.string.begin.markdown",
					"punctuation.definition.string.end.markdown",
					"string.other.link.title.markdown",
					"string.other.link.description.markdown"
				],
				settings: { foreground: i("token-keyword") }
			},
			{
				scope: [
					"markup.inserted",
					"meta.diff.header.to-file",
					"punctuation.definition.inserted"
				],
				settings: { foreground: i("token-inserted") }
			},
			{
				scope: [
					"markup.deleted",
					"meta.diff.header.from-file",
					"punctuation.definition.deleted"
				],
				settings: { foreground: i("token-deleted") }
			},
			{
				scope: ["markup.changed", "punctuation.definition.changed"],
				settings: { foreground: i("token-changed") }
			}
		]
	};
	return r || (a.tokenColors = a.tokenColors?.map((e) => (e.settings?.fontStyle && delete e.settings.fontStyle, e))), a;
}
export { v as FontStyle, p as ShikiError, y as StackElementMetadata, Ra as addClassToHast, Y as applyColorReplacements, $ as codeToHast, Oo as codeToHtml, Q as codeToTokens, po as codeToTokensBase, So as codeToTokensWithThemes, Zo as createCssVariablesTheme, Vo as createHighlighterCore, Ho as createHighlighterCoreSync, Yo as createJavaScriptRegexEngine, ka as createOnigurumaEngine, Ga as createPositionConverter, zo as createShikiInternal, Ro as createShikiInternalSync, Jo as createSingletonShorthands, Aa as createWasmOnigEngine, Ko as createdBundledHighlighter, Xo as defaultJavaScriptRegexConstructor, g as enableDeprecationWarnings, Go as getHighlighterCore, Bo as getShikiInternal, Wo as getSingletonHighlighterCore, Ua as getTokenStyleObject, Sn as hastToHtml, Ia as isNoneTheme, Pa as isPlainLang, Fa as isSpecialLang, La as isSpecialTheme, ja as loadWasm, qo as makeSingletonHighlighter, Uo as makeSingletonHighlighterCore, Va as normalizeGetter, Mo as normalizeTheme, Ha as resolveColorReplacements, Na as splitLines, za as splitToken, Ba as splitTokens, Wa as stringifyTokenStyle, Ma as toArray, uo as tokenizeAnsiWithTheme, ho as tokenizeWithTheme, To as tokensToHast, Qa as transformerDecorations, _ as warnDeprecated };
