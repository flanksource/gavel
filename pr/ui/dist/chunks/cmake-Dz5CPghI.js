// ../../../../../../../../../../Volumes/ugreen/cache/pnpm/store/v11/links/@codemirror/legacy-modes/6.5.3/6ddc4bb142089e0be26db8a3e2cc436edc8eddd035daf11477b72a920064aef3/node_modules/@codemirror/legacy-modes/mode/cmake.js
var e = /({)?[a-zA-Z0-9_]+(})?/;
function t(e, t) {
	for (var n, r, i = !1; !e.eol() && (n = e.next()) != t.pending;) {
		if (n === "$" && r != "\\" && t.pending == "\"") {
			i = !0;
			break;
		}
		r = n;
	}
	return i && e.backUp(1), t.continueString = n != t.pending, "string";
}
function n(n, r) {
	var i = n.next();
	return i === "$" ? n.match(e) ? "variableName.special" : "variable" : r.continueString ? (n.backUp(1), t(n, r)) : n.match(/(\s+)?\w+\(/) || n.match(/(\s+)?\w+\ \(/) ? (n.backUp(1), "def") : i == "#" ? (n.skipToEnd(), "comment") : i == "'" || i == "\"" ? (r.pending = i, t(n, r)) : i == "(" || i == ")" ? "bracket" : i.match(/[0-9]/) ? "number" : (n.eatWhile(/[\w-]/), null);
}
var r = {
	name: "cmake",
	startState: function() {
		var e = {};
		return e.inDefinition = !1, e.inInclude = !1, e.continueString = !1, e.pending = !1, e;
	},
	token: function(e, t) {
		return e.eatSpace() ? null : n(e, t);
	}
};
export { r as cmake };
