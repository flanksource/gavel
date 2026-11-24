import { a as e } from "./rolldown-runtime-DUcC5eAW.js";
import { n as t, t as n } from "./jsx-runtime-B4NxTvNR.js";
import { t as r } from "./button-BLoR2pgR.js";
import { n as i, t as a } from "./MonacoEditor-0E72ufhs.js";
// ../../../clicky-ui/packages/ui/src/monaco/schema-normalizer.ts
var o = /* @__PURE__ */ e(t(), 1), s = /* @__PURE__ */ new Set([
	"unevaluatedProperties",
	"unevaluatedItems",
	"dependentSchemas",
	"dependentRequired",
	"prefixItems",
	"minContains",
	"maxContains",
	"$dynamicRef",
	"$dynamicAnchor"
]);
function c(e) {
	let t = /* @__PURE__ */ new Set();
	function n(e) {
		if (Array.isArray(e)) return e.map(n);
		if (!e || typeof e != "object") return e;
		let r = {};
		for (let [i, a] of Object.entries(e)) {
			s.has(i) && t.add(i);
			let e = i === "$defs" ? "definitions" : i;
			r[e] = e === "$schema" ? "http://json-schema.org/draft-07/schema#" : e === "$ref" && typeof a == "string" ? a.replace(/^#\/\$defs\//, "#/definitions/") : n(a);
		}
		return r;
	}
	let r = n(e), i = [...t].sort();
	return i.length > 0 ? { unsupportedKeywords: i } : {
		schema: r,
		unsupportedKeywords: i
	};
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/monaco-marker-data-provider/1.2.5/70c15ff026c309d3f01fa130ba691f4623caef604eef1223d0aba3ced9a4b59e/node_modules/monaco-marker-data-provider/dist/monaco-marker-data-provider.js
function l(e, t, n) {
	let r = /* @__PURE__ */ new Map(), i = (e) => t === "*" ? !0 : Array.isArray(t) ? t.includes(e) : t === e, a = async (t) => {
		let r = t.getVersionId(), a = await n.provideMarkerData(t);
		!t.isDisposed() && r === t.getVersionId() && i(t.getLanguageId()) && e.editor.setModelMarkers(t, n.owner, a ?? []);
	}, o = (e) => {
		if (!i(e.getLanguageId())) return;
		let t, n = e.onDidChangeContent(() => {
			clearTimeout(t), t = setTimeout(() => {
				a(e);
			}, 500);
		});
		r.set(e, { dispose() {
			clearTimeout(t), n.dispose();
		} }), a(e);
	}, s = (t) => {
		e.editor.setModelMarkers(t, n.owner, []);
		let i = r.get(t);
		i && (i.dispose(), r.delete(t));
	};
	function c(e, t = e.getLanguageId()) {
		n.doReset && i(t) && n.doReset(e);
	}
	let l = e.editor.onDidCreateModel(o), u = e.editor.onWillDisposeModel((e) => {
		s(e), c(e);
	}), d = e.editor.onDidChangeModelLanguage(({ model: e, oldLanguage: t }) => {
		s(e), o(e), c(e, t);
	});
	for (let t of e.editor.getModels()) o(t);
	return {
		dispose() {
			for (let e of r.keys()) s(e);
			l.dispose(), u.dispose(), d.dispose();
		},
		async revalidate() {
			await Promise.all(e.editor.getModels().map(a));
		}
	};
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/monaco-worker-manager/2.0.1/c2733b573fe2fe6bc656c23c14b04c2fcb62c4b4cbaf5bcdd4588c8933fc91d1/node_modules/monaco-worker-manager/index.js
function u(e, t) {
	let { createData: n, interval: r = 3e4, label: i, moduleId: a, stopWhenIdleFor: o = 12e4 } = t, s, c = 0, l = !1, u = () => {
		s &&= (s.dispose(), void 0);
	}, d = setInterval(() => {
		s && Date.now() - c > o && u();
	}, r);
	return {
		dispose() {
			l = !0, clearInterval(d), u();
		},
		getWorker(...t) {
			if (l) throw Error("Worker manager has been disposed");
			return c = Date.now(), s ||= e.editor.createWebWorker({
				createData: n,
				label: i,
				moduleId: a
			}), s.withSyncedResources(t);
		},
		updateCreateData(e) {
			n = e, u();
		}
	};
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/vscode-uri/3.2.0/e764f4c050fef71d11e10348600b127bec9fd2a4279fdbdadae515e23b9356d4/node_modules/vscode-uri/lib/esm/index.mjs
var d = { 975(e) {
	function t(e) {
		if (typeof e != "string") throw TypeError("Path must be a string. Received " + JSON.stringify(e));
	}
	function n(e, t) {
		for (var n, r = "", i = 0, a = -1, o = 0, s = 0; s <= e.length; ++s) {
			if (s < e.length) n = e.charCodeAt(s);
			else {
				if (n === 47) break;
				n = 47;
			}
			if (n === 47) {
				if (a !== s - 1 && o !== 1) {
					if (a !== s - 1 && o === 2) {
						if (r.length < 2 || i !== 2 || r.charCodeAt(r.length - 1) !== 46 || r.charCodeAt(r.length - 2) !== 46) {
							if (r.length > 2) {
								var c = r.lastIndexOf("/");
								if (c !== r.length - 1) {
									c === -1 ? (r = "", i = 0) : i = (r = r.slice(0, c)).length - 1 - r.lastIndexOf("/"), a = s, o = 0;
									continue;
								}
							} else if (r.length === 2 || r.length === 1) {
								r = "", i = 0, a = s, o = 0;
								continue;
							}
						}
						t && (r.length > 0 ? r += "/.." : r = "..", i = 2);
					} else r.length > 0 ? r += "/" + e.slice(a + 1, s) : r = e.slice(a + 1, s), i = s - a - 1;
				}
				a = s, o = 0;
			} else n === 46 && o !== -1 ? ++o : o = -1;
		}
		return r;
	}
	var r = {
		resolve: function() {
			for (var e, r = "", i = !1, a = arguments.length - 1; a >= -1 && !i; a--) {
				var o;
				a >= 0 ? o = arguments[a] : (e === void 0 && (e = process.cwd()), o = e), t(o), o.length !== 0 && (r = o + "/" + r, i = o.charCodeAt(0) === 47);
			}
			return r = n(r, !i), i ? r.length > 0 ? "/" + r : "/" : r.length > 0 ? r : ".";
		},
		normalize: function(e) {
			if (t(e), e.length === 0) return ".";
			var r = e.charCodeAt(0) === 47, i = e.charCodeAt(e.length - 1) === 47;
			return (e = n(e, !r)).length !== 0 || r || (e = "."), e.length > 0 && i && (e += "/"), r ? "/" + e : e;
		},
		isAbsolute: function(e) {
			return t(e), e.length > 0 && e.charCodeAt(0) === 47;
		},
		join: function() {
			if (arguments.length === 0) return ".";
			for (var e, n = 0; n < arguments.length; ++n) {
				var i = arguments[n];
				t(i), i.length > 0 && (e === void 0 ? e = i : e += "/" + i);
			}
			return e === void 0 ? "." : r.normalize(e);
		},
		relative: function(e, n) {
			if (t(e), t(n), e === n || (e = r.resolve(e)) === (n = r.resolve(n))) return "";
			for (var i = 1; i < e.length && e.charCodeAt(i) === 47; ++i);
			for (var a = e.length, o = a - i, s = 1; s < n.length && n.charCodeAt(s) === 47; ++s);
			for (var c = n.length - s, l = o < c ? o : c, u = -1, d = 0; d <= l; ++d) {
				if (d === l) {
					if (c > l) {
						if (n.charCodeAt(s + d) === 47) return n.slice(s + d + 1);
						if (d === 0) return n.slice(s + d);
					} else o > l && (e.charCodeAt(i + d) === 47 ? u = d : d === 0 && (u = 0));
					break;
				}
				var f = e.charCodeAt(i + d);
				if (f !== n.charCodeAt(s + d)) break;
				f === 47 && (u = d);
			}
			var p = "";
			for (d = i + u + 1; d <= a; ++d) d !== a && e.charCodeAt(d) !== 47 || (p.length === 0 ? p += ".." : p += "/..");
			return p.length > 0 ? p + n.slice(s + u) : (s += u, n.charCodeAt(s) === 47 && ++s, n.slice(s));
		},
		_makeLong: function(e) {
			return e;
		},
		dirname: function(e) {
			if (t(e), e.length === 0) return ".";
			for (var n = e.charCodeAt(0), r = n === 47, i = -1, a = !0, o = e.length - 1; o >= 1; --o) if ((n = e.charCodeAt(o)) === 47) {
				if (!a) {
					i = o;
					break;
				}
			} else a = !1;
			return i === -1 ? r ? "/" : "." : r && i === 1 ? "//" : e.slice(0, i);
		},
		basename: function(e, n) {
			if (n !== void 0 && typeof n != "string") throw TypeError("\"ext\" argument must be a string");
			t(e);
			var r, i = 0, a = -1, o = !0;
			if (n !== void 0 && n.length > 0 && n.length <= e.length) {
				if (n.length === e.length && n === e) return "";
				var s = n.length - 1, c = -1;
				for (r = e.length - 1; r >= 0; --r) {
					var l = e.charCodeAt(r);
					if (l === 47) {
						if (!o) {
							i = r + 1;
							break;
						}
					} else c === -1 && (o = !1, c = r + 1), s >= 0 && (l === n.charCodeAt(s) ? --s === -1 && (a = r) : (s = -1, a = c));
				}
				return i === a ? a = c : a === -1 && (a = e.length), e.slice(i, a);
			}
			for (r = e.length - 1; r >= 0; --r) if (e.charCodeAt(r) === 47) {
				if (!o) {
					i = r + 1;
					break;
				}
			} else a === -1 && (o = !1, a = r + 1);
			return a === -1 ? "" : e.slice(i, a);
		},
		extname: function(e) {
			t(e);
			for (var n = -1, r = 0, i = -1, a = !0, o = 0, s = e.length - 1; s >= 0; --s) {
				var c = e.charCodeAt(s);
				if (c !== 47) i === -1 && (a = !1, i = s + 1), c === 46 ? n === -1 ? n = s : o !== 1 && (o = 1) : n !== -1 && (o = -1);
				else if (!a) {
					r = s + 1;
					break;
				}
			}
			return n === -1 || i === -1 || o === 0 || o === 1 && n === i - 1 && n === r + 1 ? "" : e.slice(n, i);
		},
		format: function(e) {
			if (typeof e != "object" || !e) throw TypeError("The \"pathObject\" argument must be of type Object. Received type " + typeof e);
			return function(e, t) {
				var n = t.dir || t.root, r = t.base || (t.name || "") + (t.ext || "");
				return n ? n === t.root ? n + r : n + "/" + r : r;
			}(0, e);
		},
		parse: function(e) {
			t(e);
			var n = {
				root: "",
				dir: "",
				base: "",
				ext: "",
				name: ""
			};
			if (e.length === 0) return n;
			var r, i = e.charCodeAt(0), a = i === 47;
			a ? (n.root = "/", r = 1) : r = 0;
			for (var o = -1, s = 0, c = -1, l = !0, u = e.length - 1, d = 0; u >= r; --u) if ((i = e.charCodeAt(u)) !== 47) c === -1 && (l = !1, c = u + 1), i === 46 ? o === -1 ? o = u : d !== 1 && (d = 1) : o !== -1 && (d = -1);
			else if (!l) {
				s = u + 1;
				break;
			}
			return o === -1 || c === -1 || d === 0 || d === 1 && o === c - 1 && o === s + 1 ? c !== -1 && (n.base = n.name = s === 0 && a ? e.slice(1, c) : e.slice(s, c)) : (s === 0 && a ? (n.name = e.slice(1, o), n.base = e.slice(1, c)) : (n.name = e.slice(s, o), n.base = e.slice(s, c)), n.ext = e.slice(o, c)), s > 0 ? n.dir = e.slice(0, s - 1) : a && (n.dir = "/"), n;
		},
		sep: "/",
		delimiter: ":",
		win32: null,
		posix: null
	};
	r.posix = r, e.exports = r;
} }, f = {};
function p(e) {
	var t = f[e];
	if (t !== void 0) return t.exports;
	var n = f[e] = { exports: {} };
	return d[e](n, n.exports, p), n.exports;
}
var m;
p.d = (e, t) => {
	for (var n in t) p.o(t, n) && !p.o(e, n) && Object.defineProperty(e, n, {
		enumerable: !0,
		get: t[n]
	});
}, p.o = (e, t) => Object.prototype.hasOwnProperty.call(e, t), typeof process == "object" ? m = process.platform === "win32" : typeof navigator == "object" && (m = navigator.userAgent.indexOf("Windows") >= 0);
var h = /^\w[\w\d+.-]*$/, g = /^\//, _ = /^\/\//;
function v(e, t) {
	if (!e.scheme && t) throw Error(`[UriError]: Scheme is missing: {scheme: "", authority: "${e.authority}", path: "${e.path}", query: "${e.query}", fragment: "${e.fragment}"}`);
	if (e.scheme && !h.test(e.scheme)) throw Error("[UriError]: Scheme contains illegal characters.");
	if (e.path) {
		if (e.authority) {
			if (!g.test(e.path)) throw Error("[UriError]: If a URI contains an authority component, then the path component must either be empty or begin with a slash (\"/\") character");
		} else if (_.test(e.path)) throw Error("[UriError]: If a URI does not contain an authority component, then the path cannot begin with two slash characters (\"//\")");
	}
}
var y = "", b = "/", x = /^(([^:/?#]+?):)?(\/\/([^/?#]*))?([^?#]*)(\?([^#]*))?(#(.*))?/, S = class e {
	static isUri(t) {
		return t instanceof e || !!t && typeof t.authority == "string" && typeof t.fragment == "string" && typeof t.path == "string" && typeof t.query == "string" && typeof t.scheme == "string" && typeof t.fsPath == "string" && typeof t.with == "function" && typeof t.toString == "function";
	}
	scheme;
	authority;
	path;
	query;
	fragment;
	constructor(e, t, n, r, i, a = !1) {
		typeof e == "object" ? (this.scheme = e.scheme || y, this.authority = e.authority || y, this.path = e.path || y, this.query = e.query || y, this.fragment = e.fragment || y) : (this.scheme = function(e, t) {
			return e || t ? e : "file";
		}(e, a), this.authority = t || y, this.path = function(e, t) {
			switch (e) {
				case "https":
				case "http":
				case "file": t ? t[0] !== b && (t = b + t) : t = b;
			}
			return t;
		}(this.scheme, n || y), this.query = r || y, this.fragment = i || y, v(this, a));
	}
	get fsPath() {
		return D(this, !1);
	}
	with(e) {
		if (!e) return this;
		let { scheme: t, authority: n, path: r, query: i, fragment: a } = e;
		return t === void 0 ? t = this.scheme : t === null && (t = y), n === void 0 ? n = this.authority : n === null && (n = y), r === void 0 ? r = this.path : r === null && (r = y), i === void 0 ? i = this.query : i === null && (i = y), a === void 0 ? a = this.fragment : a === null && (a = y), t === this.scheme && n === this.authority && r === this.path && i === this.query && a === this.fragment ? this : new w(t, n, r, i, a);
	}
	static parse(e, t = !1) {
		let n = x.exec(e);
		return n ? new w(n[2] || y, j(n[4] || y), j(n[5] || y), j(n[7] || y), j(n[9] || y), t) : new w(y, y, y, y, y);
	}
	static file(e) {
		let t = y;
		if (m && (e = e.replace(/\\/g, b)), e[0] === b && e[1] === b) {
			let n = e.indexOf(b, 2);
			n === -1 ? (t = e.substring(2), e = b) : (t = e.substring(2, n), e = e.substring(n) || b);
		}
		return new w("file", t, e, y, y);
	}
	static from(e) {
		let t = new w(e.scheme, e.authority, e.path, e.query, e.fragment);
		return v(t, !0), t;
	}
	toString(e = !1) {
		return O(this, e);
	}
	toJSON() {
		return this;
	}
	static revive(t) {
		if (t) {
			if (t instanceof e) return t;
			{
				let e = new w(t);
				return e._formatted = t.external, e._fsPath = t._sep === C ? t.fsPath : null, e;
			}
		}
		return t;
	}
}, C = m ? 1 : void 0, w = class extends S {
	_formatted = null;
	_fsPath = null;
	get fsPath() {
		return this._fsPath ||= D(this, !1), this._fsPath;
	}
	toString(e = !1) {
		return e ? O(this, !0) : (this._formatted ||= O(this, !1), this._formatted);
	}
	toJSON() {
		let e = { $mid: 1 };
		return this._fsPath && (e.fsPath = this._fsPath, e._sep = C), this._formatted && (e.external = this._formatted), this.path && (e.path = this.path), this.scheme && (e.scheme = this.scheme), this.authority && (e.authority = this.authority), this.query && (e.query = this.query), this.fragment && (e.fragment = this.fragment), e;
	}
}, T = {
	58: "%3A",
	47: "%2F",
	63: "%3F",
	35: "%23",
	91: "%5B",
	93: "%5D",
	64: "%40",
	33: "%21",
	36: "%24",
	38: "%26",
	39: "%27",
	40: "%28",
	41: "%29",
	42: "%2A",
	43: "%2B",
	44: "%2C",
	59: "%3B",
	61: "%3D",
	32: "%20"
};
function E(e, t, n) {
	let r, i = -1;
	for (let a = 0; a < e.length; a++) {
		let o = e.charCodeAt(a);
		if (o >= 97 && o <= 122 || o >= 65 && o <= 90 || o >= 48 && o <= 57 || o === 45 || o === 46 || o === 95 || o === 126 || t && o === 47 || n && o === 91 || n && o === 93 || n && o === 58) i !== -1 && (r += encodeURIComponent(e.substring(i, a)), i = -1), r !== void 0 && (r += e.charAt(a));
		else {
			r === void 0 && (r = e.substr(0, a));
			let t = T[o];
			t === void 0 ? i === -1 && (i = a) : (i !== -1 && (r += encodeURIComponent(e.substring(i, a)), i = -1), r += t);
		}
	}
	return i !== -1 && (r += encodeURIComponent(e.substring(i))), r === void 0 ? e : r;
}
function ee(e) {
	let t;
	for (let n = 0; n < e.length; n++) {
		let r = e.charCodeAt(n);
		r === 35 || r === 63 ? (t === void 0 && (t = e.substr(0, n)), t += T[r]) : t !== void 0 && (t += e[n]);
	}
	return t === void 0 ? e : t;
}
function D(e, t) {
	let n;
	return n = e.authority && e.path.length > 1 && e.scheme === "file" ? `//${e.authority}${e.path}` : e.path.charCodeAt(0) === 47 && (e.path.charCodeAt(1) >= 65 && e.path.charCodeAt(1) <= 90 || e.path.charCodeAt(1) >= 97 && e.path.charCodeAt(1) <= 122) && e.path.charCodeAt(2) === 58 ? t ? e.path.substr(1) : e.path[1].toLowerCase() + e.path.substr(2) : e.path, m && (n = n.replace(/\//g, "\\")), n;
}
function O(e, t) {
	let n = t ? ee : E, r = "", { scheme: i, authority: a, path: o, query: s, fragment: c } = e;
	if (i && (r += i, r += ":"), (a || i === "file") && (r += b, r += b), a) {
		let e = a.indexOf("@");
		if (e !== -1) {
			let t = a.substr(0, e);
			a = a.substr(e + 1), e = t.lastIndexOf(":"), e === -1 ? r += n(t, !1, !1) : (r += n(t.substr(0, e), !1, !1), r += ":", r += n(t.substr(e + 1), !1, !0)), r += "@";
		}
		a = a.toLowerCase(), e = a.lastIndexOf(":"), e === -1 ? r += n(a, !1, !0) : (r += n(a.substr(0, e), !1, !0), r += a.substr(e));
	}
	if (o) {
		if (o.length >= 3 && o.charCodeAt(0) === 47 && o.charCodeAt(2) === 58) {
			let e = o.charCodeAt(1);
			e >= 65 && e <= 90 && (o = `/${String.fromCharCode(e + 32)}:${o.substr(3)}`);
		} else if (o.length >= 2 && o.charCodeAt(1) === 58) {
			let e = o.charCodeAt(0);
			e >= 65 && e <= 90 && (o = `${String.fromCharCode(e + 32)}:${o.substr(2)}`);
		}
		r += n(o, !0, !1);
	}
	return s && (r += "?", r += n(s, !1, !1)), c && (r += "#", r += t ? c : E(c, !1, !1)), r;
}
function k(e) {
	try {
		return decodeURIComponent(e);
	} catch {
		return e.length > 3 ? e.substr(0, 3) + k(e.substr(3)) : e;
	}
}
var A = /(%[0-9A-Za-z][0-9A-Za-z])+/g;
function j(e) {
	return e.match(A) ? e.replace(A, (e) => k(e)) : e;
}
var M = p(975), N = M.posix || M, P = "/", te;
(function(e) {
	e.joinPath = function(e, ...t) {
		return e.with({ path: N.join(e.path, ...t) });
	}, e.resolvePath = function(e, ...t) {
		let n = e.path, r = !1;
		n[0] !== P && (n = P + n, r = !0);
		let i = N.resolve(n, ...t);
		return r && i[0] === P && !e.authority && (i = i.substring(1)), e.with({ path: i });
	}, e.dirname = function(e) {
		if (e.path.length === 0 || e.path === P) return e;
		let t = N.dirname(e.path);
		return t.length === 1 && t.charCodeAt(0) === 46 && (t = ""), e.with({ path: t });
	}, e.basename = function(e) {
		return N.basename(e.path);
	}, e.extname = function(e) {
		return N.extname(e.path);
	};
})(te ||= {});
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/monaco-languageserver-types/0.3.4/b06e18c35d5080f1a6bf8b374e28919a96160905148709f59c2649de9ad22eb5/node_modules/monaco-languageserver-types/dist/markerSeverity.js
function ne(e) {
	return e === 8 ? 1 : e === 4 ? 2 : e === 2 ? 3 : 4;
}
function re(e) {
	return e === 1 ? 8 : e === 2 ? 4 : e === 3 ? 2 : 1;
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/monaco-languageserver-types/0.3.4/b06e18c35d5080f1a6bf8b374e28919a96160905148709f59c2649de9ad22eb5/node_modules/monaco-languageserver-types/dist/markerTag.js
function ie(e) {
	return e;
}
function ae(e) {
	return e;
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/monaco-languageserver-types/0.3.4/b06e18c35d5080f1a6bf8b374e28919a96160905148709f59c2649de9ad22eb5/node_modules/monaco-languageserver-types/dist/range.js
function F(e) {
	return {
		start: {
			line: e.startLineNumber - 1,
			character: e.startColumn - 1
		},
		end: {
			line: e.endLineNumber - 1,
			character: e.endColumn - 1
		}
	};
}
function I(e) {
	return {
		startLineNumber: e.start.line + 1,
		startColumn: e.start.character + 1,
		endLineNumber: e.end.line + 1,
		endColumn: e.end.character + 1
	};
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/monaco-languageserver-types/0.3.4/b06e18c35d5080f1a6bf8b374e28919a96160905148709f59c2649de9ad22eb5/node_modules/monaco-languageserver-types/dist/relatedInformation.js
function oe(e) {
	return {
		location: {
			range: F(e),
			uri: String(e.resource)
		},
		message: e.message
	};
}
function se(e) {
	return {
		...I(e.location.range),
		message: e.message,
		resource: S.parse(e.location.uri)
	};
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/monaco-languageserver-types/0.3.4/b06e18c35d5080f1a6bf8b374e28919a96160905148709f59c2649de9ad22eb5/node_modules/monaco-languageserver-types/dist/markerData.js
function ce(e) {
	let t = {
		message: e.message,
		range: F(e),
		severity: ne(e.severity)
	};
	return typeof e.code == "string" ? t.code = e.code : e.code != null && (t.code = e.code.value, t.codeDescription = { href: String(e.code.target) }), e.relatedInformation && (t.relatedInformation = e.relatedInformation.map(oe)), e.tags && (t.tags = e.tags.map(ie)), e.source != null && (t.source = e.source), t;
}
function L(e, t) {
	let n = {
		...I(e.range),
		message: e.message,
		severity: e.severity ? re(e.severity) : t?.defaultSeverity ?? 8
	};
	return e.code != null && (n.code = e.codeDescription == null ? String(e.code) : {
		value: String(e.code),
		target: S.parse(e.codeDescription.href)
	}), e.relatedInformation && (n.relatedInformation = e.relatedInformation.map(se)), e.tags && (n.tags = e.tags.map(ae)), e.source != null && (n.source = e.source), n;
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/monaco-languageserver-types/0.3.4/b06e18c35d5080f1a6bf8b374e28919a96160905148709f59c2649de9ad22eb5/node_modules/monaco-languageserver-types/dist/textEdit.js
function R(e) {
	return {
		range: I(e.range),
		text: e.newText
	};
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/monaco-languageserver-types/0.3.4/b06e18c35d5080f1a6bf8b374e28919a96160905148709f59c2649de9ad22eb5/node_modules/monaco-languageserver-types/dist/workspaceFileEditOptions.js
function le(e) {
	let t = {};
	return e.ignoreIfExists != null && (t.ignoreIfExists = e.ignoreIfExists), e.ignoreIfNotExists != null && (t.ignoreIfNotExists = e.ignoreIfNotExists), e.overwrite != null && (t.overwrite = e.overwrite), e.recursive != null && (t.recursive = e.recursive), t;
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/monaco-languageserver-types/0.3.4/b06e18c35d5080f1a6bf8b374e28919a96160905148709f59c2649de9ad22eb5/node_modules/monaco-languageserver-types/dist/workspaceFileEdit.js
function ue(e) {
	let t = e.kind === "create" ? { newResource: S.parse(e.uri) } : e.kind === "delete" ? { oldResource: S.parse(e.uri) } : {
		oldResource: S.parse(e.oldUri),
		newResource: S.parse(e.newUri)
	};
	return e.options && (t.options = le(e.options)), t;
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/monaco-languageserver-types/0.3.4/b06e18c35d5080f1a6bf8b374e28919a96160905148709f59c2649de9ad22eb5/node_modules/monaco-languageserver-types/dist/workspaceEdit.js
function z(e, t, n) {
	return {
		resource: S.parse(t),
		versionId: n,
		textEdit: R(e)
	};
}
function de(e) {
	let t = [];
	if (e.changes) for (let [n, r] of Object.entries(e.changes)) for (let e of r) t.push(z(e, n));
	if (e.documentChanges) for (let n of e.documentChanges) if ("textDocument" in n) for (let e of n.edits) t.push(z(e, n.textDocument.uri, n.textDocument.version ?? void 0));
	else t.push(ue(n));
	return { edits: t };
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/monaco-languageserver-types/0.3.4/b06e18c35d5080f1a6bf8b374e28919a96160905148709f59c2649de9ad22eb5/node_modules/monaco-languageserver-types/dist/codeAction.js
function fe(e, t) {
	let n = {
		title: e.title,
		isPreferred: e.isPreferred
	};
	return e.diagnostics && (n.diagnostics = e.diagnostics.map((e) => L(e, t))), e.disabled && (n.disabled = e.disabled.reason), e.edit && (n.edit = de(e.edit)), e.isPreferred != null && (n.isPreferred = e.isPreferred), e.kind && (n.kind = e.kind), n;
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/monaco-languageserver-types/0.3.4/b06e18c35d5080f1a6bf8b374e28919a96160905148709f59c2649de9ad22eb5/node_modules/monaco-languageserver-types/dist/codeActionTriggerType.js
function pe(e) {
	return e;
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/monaco-languageserver-types/0.3.4/b06e18c35d5080f1a6bf8b374e28919a96160905148709f59c2649de9ad22eb5/node_modules/monaco-languageserver-types/dist/codeActionContext.js
function me(e) {
	let t = {
		diagnostics: e.markers.map(ce),
		triggerKind: pe(e.trigger)
	};
	return e.only != null && (t.only = [e.only]), t;
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/monaco-languageserver-types/0.3.4/b06e18c35d5080f1a6bf8b374e28919a96160905148709f59c2649de9ad22eb5/node_modules/monaco-languageserver-types/dist/command.js
function he(e) {
	let t = {
		title: e.title,
		id: e.command
	};
	return e.arguments && (t.arguments = e.arguments), t;
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/monaco-languageserver-types/0.3.4/b06e18c35d5080f1a6bf8b374e28919a96160905148709f59c2649de9ad22eb5/node_modules/monaco-languageserver-types/dist/completionItemKind.js
function ge(e) {
	return e === 1 ? 18 : e === 2 ? 0 : e === 3 ? 1 : e === 4 ? 2 : e === 5 ? 3 : e === 6 ? 4 : e === 7 ? 5 : e === 8 ? 7 : e === 9 ? 8 : e === 10 ? 9 : e === 11 ? 12 : e === 12 ? 13 : e === 13 ? 15 : e === 14 ? 17 : e === 15 ? 27 : e === 16 ? 19 : e === 17 ? 20 : e === 18 ? 21 : e === 19 ? 23 : e === 20 ? 16 : e === 21 ? 14 : e === 22 ? 6 : e === 23 ? 10 : e === 24 ? 11 : 24;
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/monaco-languageserver-types/0.3.4/b06e18c35d5080f1a6bf8b374e28919a96160905148709f59c2649de9ad22eb5/node_modules/monaco-languageserver-types/dist/completionItemTag.js
function _e(e) {
	return e;
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/monaco-languageserver-types/0.3.4/b06e18c35d5080f1a6bf8b374e28919a96160905148709f59c2649de9ad22eb5/node_modules/monaco-languageserver-types/dist/markdownString.js
function B(e) {
	return { value: e.value };
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/monaco-languageserver-types/0.3.4/b06e18c35d5080f1a6bf8b374e28919a96160905148709f59c2649de9ad22eb5/node_modules/monaco-languageserver-types/dist/singleEditOperation.js
function ve(e) {
	return {
		range: I(e.range),
		text: e.newText
	};
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/monaco-languageserver-types/0.3.4/b06e18c35d5080f1a6bf8b374e28919a96160905148709f59c2649de9ad22eb5/node_modules/monaco-languageserver-types/dist/completionItem.js
function ye(e) {
	return "range" in e ? I(e.range) : "insert" in e && "replace" in e ? {
		insert: I(e.insert),
		replace: I(e.replace)
	} : I(e);
}
function be(e, t) {
	let n = t.itemDefaults ?? {}, r = e.textEdit ?? n.editRange, i = e.commitCharacters ?? n.commitCharacters, a = e.insertTextFormat ?? n.insertTextFormat, o = e.insertTextMode ?? n.insertTextMode, s = e.insertText, c;
	r ? (c = ye(r), "newText" in r && (s = r.newText)) : c = { ...t.range };
	let l = {
		insertText: s ?? e.label,
		kind: e.kind == null ? 18 : ge(e.kind),
		label: e.label,
		range: c
	};
	return e.additionalTextEdits && (l.additionalTextEdits = e.additionalTextEdits.map(ve)), e.command && (l.command = he(e.command)), i && (l.commitCharacters = i), e.detail != null && (l.detail = e.detail), typeof e.documentation == "string" ? l.documentation = e.documentation : e.documentation && (l.documentation = B(e.documentation)), e.filterText != null && (l.filterText = e.filterText), a === 2 ? l.insertTextRules = 4 : o === 2 && (l.insertTextRules = 1), e.preselect != null && (l.preselect = e.preselect), e.sortText != null && (l.sortText = e.sortText), e.tags && (l.tags = e.tags.map(_e)), l;
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/monaco-languageserver-types/0.3.4/b06e18c35d5080f1a6bf8b374e28919a96160905148709f59c2649de9ad22eb5/node_modules/monaco-languageserver-types/dist/completionList.js
function xe(e, t) {
	return {
		incomplete: !!e.isIncomplete,
		suggestions: e.items.map((n) => be(n, {
			range: t.range,
			itemDefaults: e.itemDefaults
		}))
	};
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/monaco-languageserver-types/0.3.4/b06e18c35d5080f1a6bf8b374e28919a96160905148709f59c2649de9ad22eb5/node_modules/monaco-languageserver-types/dist/symbolKind.js
function Se(e) {
	return e === 1 ? 0 : e === 2 ? 1 : e === 3 ? 2 : e === 4 ? 3 : e === 5 ? 4 : e === 6 ? 5 : e === 7 ? 6 : e === 8 ? 7 : e === 9 ? 8 : e === 10 ? 9 : e === 11 ? 10 : e === 12 ? 11 : e === 13 ? 12 : e === 14 ? 13 : e === 15 ? 14 : e === 16 ? 15 : e === 17 ? 16 : e === 18 ? 17 : e === 19 ? 18 : e === 20 ? 19 : e === 21 ? 20 : e === 22 ? 21 : e === 23 ? 22 : e === 24 ? 23 : e === 25 ? 24 : 25;
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/monaco-languageserver-types/0.3.4/b06e18c35d5080f1a6bf8b374e28919a96160905148709f59c2649de9ad22eb5/node_modules/monaco-languageserver-types/dist/symbolTag.js
function Ce(e) {
	return e;
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/monaco-languageserver-types/0.3.4/b06e18c35d5080f1a6bf8b374e28919a96160905148709f59c2649de9ad22eb5/node_modules/monaco-languageserver-types/dist/documentSymbol.js
function V(e) {
	let t = {
		detail: e.detail ?? "",
		kind: Se(e.kind),
		name: e.name,
		range: I(e.range),
		selectionRange: I(e.selectionRange),
		tags: e.tags?.map(Ce) ?? []
	};
	return e.children && (t.children = e.children.map(V)), t;
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/monaco-languageserver-types/0.3.4/b06e18c35d5080f1a6bf8b374e28919a96160905148709f59c2649de9ad22eb5/node_modules/monaco-languageserver-types/dist/foldingRange.js
function we(e) {
	let t = {
		start: e.startLine + 1,
		end: e.endLine + 1
	};
	return e.kind != null && (t.kind = { value: e.kind }), t;
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/monaco-languageserver-types/0.3.4/b06e18c35d5080f1a6bf8b374e28919a96160905148709f59c2649de9ad22eb5/node_modules/monaco-languageserver-types/dist/hover.js
function H(e) {
	return typeof e == "string" ? { value: e } : { value: `\`\`\`${e.language}\n${e.value}\n\`\`\`` };
}
function Te(e) {
	return typeof e == "string" || "language" in e ? [H(e)] : Array.isArray(e) ? e.map(H) : [B(e)];
}
function Ee(e) {
	let t = { contents: Te(e.contents) };
	return e.range && (t.range = I(e.range)), t;
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/monaco-languageserver-types/0.3.4/b06e18c35d5080f1a6bf8b374e28919a96160905148709f59c2649de9ad22eb5/node_modules/monaco-languageserver-types/dist/position.js
function U(e) {
	return {
		character: e.column - 1,
		line: e.lineNumber - 1
	};
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/monaco-languageserver-types/0.3.4/b06e18c35d5080f1a6bf8b374e28919a96160905148709f59c2649de9ad22eb5/node_modules/monaco-languageserver-types/dist/link.js
function De(e) {
	let t = { range: I(e.range) };
	return e.tooltip != null && (t.tooltip = e.tooltip), e.target != null && (t.url = S.parse(e.target)), t;
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/monaco-languageserver-types/0.3.4/b06e18c35d5080f1a6bf8b374e28919a96160905148709f59c2649de9ad22eb5/node_modules/monaco-languageserver-types/dist/locationLink.js
function Oe(e) {
	let t = {
		range: I(e.targetRange),
		targetSelectionRange: I(e.targetSelectionRange),
		uri: S.parse(e.targetUri)
	};
	return e.originSelectionRange && (t.originSelectionRange = I(e.originSelectionRange)), t;
}
// ../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@/monaco-yaml/5.1.1/5bbccd846a8839d39056d52e17970f321f830b1cd4f05d0145fe830a307eb1ce/node_modules/monaco-yaml/index.js
function W(e) {
	return {
		owner: "yaml",
		async provideMarkerData(t) {
			return (await (await e(t.uri)).doValidation(String(t.uri)))?.map((e) => L(e));
		},
		async doReset(t) {
			await (await e(t.uri)).resetSchema(String(t.uri));
		}
	};
}
function ke(e) {
	return {
		triggerCharacters: [" ", ":"],
		async provideCompletionItems(t, n) {
			let r = t.uri, i = t.getWordUntilPosition(n), a = await (await e(r)).doComplete(String(r), U(n));
			if (a) return xe(a, { range: {
				startLineNumber: n.lineNumber,
				startColumn: i.startColumn,
				endLineNumber: n.lineNumber,
				endColumn: i.endColumn
			} });
		}
	};
}
function Ae(e) {
	return { async provideDefinition(t, n) {
		let r = t.uri;
		return (await (await e(r)).doDefinition(String(r), U(n)))?.map(Oe);
	} };
}
function je(e) {
	return { async provideHover(t, n) {
		let r = t.uri, i = await (await e(r)).doHover(String(r), U(n));
		if (i) return Ee(i);
	} };
}
function Me(e) {
	return {
		displayName: "yaml",
		async provideDocumentSymbols(t) {
			let n = t.uri;
			return (await (await e(n)).findDocumentSymbols(String(n)))?.map(V);
		}
	};
}
function Ne(e) {
	return {
		displayName: "yaml",
		async provideDocumentFormattingEdits(t) {
			let n = t.uri;
			return (await (await e(n)).format(String(n)))?.map(R);
		}
	};
}
function Pe(e) {
	return { async provideLinks(t) {
		let n = t.uri, r = await (await e(n)).findLinks(String(n));
		if (r) return { links: r.map(De) };
	} };
}
function Fe(e) {
	return { async provideCodeActions(t, n, r) {
		let i = t.uri, a = await (await e(i)).getCodeAction(String(i), F(n), me(r));
		if (a) return {
			actions: a.map((e) => fe(e)),
			dispose() {}
		};
	} };
}
function Ie(e) {
	return { async provideFoldingRanges(t) {
		let n = t.uri;
		return (await (await e(n)).getFoldingRanges(String(n)))?.map(we);
	} };
}
function Le(e, t) {
	let n = {
		completion: !0,
		customTags: [],
		enableSchemaRequest: !1,
		format: !0,
		isKubernetes: !1,
		hover: !0,
		schemas: [],
		validate: !0,
		yamlVersion: "1.2",
		...t
	};
	e.languages.register({
		id: "yaml",
		extensions: [".yaml", ".yml"],
		aliases: [
			"YAML",
			"yaml",
			"YML",
			"yml"
		],
		mimetypes: ["application/x-yaml"]
	});
	let r = u(e, {
		label: "yaml",
		moduleId: "monaco-yaml/yaml.worker",
		createData: n
	}), i = l(e, "yaml", W(r.getWorker)), a = [
		r,
		e.languages.registerCompletionItemProvider("yaml", ke(r.getWorker)),
		e.languages.registerHoverProvider("yaml", je(r.getWorker)),
		e.languages.registerDefinitionProvider("yaml", Ae(r.getWorker)),
		e.languages.registerDocumentSymbolProvider("yaml", Me(r.getWorker)),
		e.languages.registerDocumentFormattingEditProvider("yaml", Ne(r.getWorker)),
		e.languages.registerLinkProvider("yaml", Pe(r.getWorker)),
		e.languages.registerCodeActionProvider("yaml", Fe(r.getWorker)),
		e.languages.registerFoldingRangeProvider("yaml", Ie(r.getWorker)),
		e.languages.setLanguageConfiguration("yaml", {
			comments: { lineComment: "#" },
			brackets: [
				["{", "}"],
				["[", "]"],
				["(", ")"]
			],
			autoClosingPairs: [
				{
					open: "{",
					close: "}"
				},
				{
					open: "[",
					close: "]"
				},
				{
					open: "(",
					close: ")"
				},
				{
					open: "\"",
					close: "\""
				},
				{
					open: "'",
					close: "'"
				}
			],
			surroundingPairs: [
				{
					open: "{",
					close: "}"
				},
				{
					open: "[",
					close: "]"
				},
				{
					open: "(",
					close: ")"
				},
				{
					open: "\"",
					close: "\""
				},
				{
					open: "'",
					close: "'"
				}
			],
			onEnterRules: [{
				beforeText: /:\s*$/,
				action: { indentAction: e.languages.IndentAction.Indent }
			}]
		})
	];
	return {
		dispose() {
			for (let e of a) e.dispose();
			i.dispose();
		},
		update(t) {
			r.updateCreateData(Object.assign(n, t)), i.dispose(), i = l(e, "yaml", W(r.getWorker));
		}
	};
}
// ../../../clicky-ui/packages/ui/src/monaco/schema-registry.ts
var G = /* @__PURE__ */ new Map(), K = /* @__PURE__ */ new Map(), q, J = Promise.resolve();
function Re(e, t, n, r, i) {
	if (t === "json") {
		let t = Symbol(n);
		return K.set(n, {
			token: t,
			uri: r,
			schema: i
		}), Y(e), () => {
			K.get(n)?.token === t && (K.delete(n), Y(e));
		};
	}
	let a = Symbol(n);
	return G.set(n, {
		token: a,
		uri: r,
		schema: i
	}), q ? X() : q = Le(e, {
		completion: !0,
		hover: !0,
		validate: !0,
		format: !0,
		yamlVersion: "1.2",
		schemas: Z()
	}), () => {
		G.get(n)?.token === a && (G.delete(n), X());
	};
}
function Y(e) {
	e.languages.json.jsonDefaults.setDiagnosticsOptions({
		validate: !0,
		schemas: [...K.entries()].map(([e, t]) => ({
			uri: t.uri,
			fileMatch: Q(e),
			schema: t.schema
		}))
	});
}
function X() {
	let e = Z();
	J = J.then(() => q?.update({ schemas: e })).then(() => void 0);
}
function Z() {
	return [...G.entries()].map(([e, t]) => ({
		uri: t.uri,
		fileMatch: Q(e),
		schema: t.schema
	}));
}
function Q(e) {
	let t = e.split("/").pop();
	return t ? [e, `**/${t}`] : [e];
}
// ../../../clicky-ui/packages/ui/src/monaco/MonacoSchemaEditor.tsx
var $ = n();
function ze(e) {
	let { value: t, language: n, path: r, schema: s, schemaUri: l, onValidationChange: u } = e, [d, f] = (0, o.useState)(0), p = i(), m = (0, o.useRef)(void 0), h = (0, o.useRef)(void 0), g = (0, o.useRef)(void 0), _ = (0, o.useRef)(void 0), v = (0, o.useRef)(null), y = (0, o.useMemo)(() => c(s), [s]);
	if ((0, o.useEffect)(() => (g.current && clearTimeout(g.current), () => {
		h.current?.(), m.current?.(), _.current && clearTimeout(_.current);
		let e = v.current;
		g.current = setTimeout(() => e?.dispose(), 0);
	}), []), (0, o.useEffect)(() => {
		u(y.schema ? p ? {
			status: "loading",
			errors: []
		} : {
			status: "unavailable",
			errors: ["Monaco workers are not configured."]
		} : {
			status: "unavailable",
			errors: [`Unsupported schema keywords: ${y.unsupportedKeywords.join(", ")}`]
		});
	}, [
		t,
		d,
		p,
		u,
		y.schema
	]), !p) return /* @__PURE__ */ (0, $.jsx)(Be, {
		errors: ["Monaco workers are not configured."],
		onRetry: () => f((e) => e + 1)
	});
	if (!y.schema) {
		let e = [`Unsupported schema keywords: ${y.unsupportedKeywords.join(", ")}`];
		return /* @__PURE__ */ (0, $.jsx)(Be, {
			errors: e,
			onRetry: () => f((e) => e + 1)
		});
	}
	let b = (e) => {
		let t = v.current;
		if (!t) return;
		let n = t.getVersionId();
		queueMicrotask(() => {
			if (v.current?.getVersionId() !== n) return;
			let t = e.filter((e) => e.severity >= 4).map((e) => e.message);
			u({
				status: t.length > 0 ? "invalid" : "valid",
				errors: t
			});
		});
	}, x = (e) => {
		m.current?.(), m.current = Re(e, n, r, l, y.schema);
	}, S = (e, t) => {
		g.current && clearTimeout(g.current), v.current = e.getModel(), h.current?.();
		let n = t.editor.onDidChangeMarkers((e) => {
			let n = v.current;
			n && e.some((e) => e.toString() === n.uri.toString()) && b(t.editor.getModelMarkers({ resource: n.uri }));
		});
		h.current = () => n.dispose(), _.current = setTimeout(() => {
			let e = v.current;
			e && b(t.editor.getModelMarkers({ resource: e.uri }));
		}, 1500);
	};
	return /* @__PURE__ */ (0, $.jsx)(a, {
		...e,
		beforeMount: (e) => {
			try {
				x(e);
			} catch (e) {
				u({
					status: "unavailable",
					errors: [e instanceof Error ? e.message : String(e)]
				});
			}
		},
		onMount: S,
		onValidate: b
	}, d);
}
function Be({ errors: e, onRetry: t }) {
	return /* @__PURE__ */ (0, $.jsxs)("div", {
		role: "alert",
		className: "space-y-2 rounded-md border border-destructive/40 p-3 text-sm text-destructive",
		children: [/* @__PURE__ */ (0, $.jsx)("div", { children: e.join("; ") }), /* @__PURE__ */ (0, $.jsx)(r, {
			type: "button",
			variant: "outline",
			size: "sm",
			onClick: t,
			children: "Retry editor"
		})]
	});
}
export { ze as MonacoSchemaEditor };
